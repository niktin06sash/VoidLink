package node

import (
	"context"
	"errors"
	"io"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/niktin06sash/VoidLink/internal/config"
	"github.com/niktin06sash/VoidLink/internal/node/whitelist"
)

func TestUpdateWhitelistDisconnectsRemovedPeer(t *testing.T) {
	mn := mocknet.New()
	t.Cleanup(func() { _ = mn.Close() })

	server, err := mn.GenPeer()
	if err != nil {
		t.Fatal(err)
	}
	client, err := mn.GenPeer()
	if err != nil {
		t.Fatal(err)
	}
	if err := mn.LinkAll(); err != nil {
		t.Fatal(err)
	}
	if err := mn.ConnectAllButSelf(); err != nil {
		t.Fatal(err)
	}

	n := &Node{
		Host: server,
		wm: whitelist.NewWhitelistManager(map[string]config.PeerInfo{
			client.ID().String(): {},
		}),
	}
	n.updateWhitelist(map[string]config.PeerInfo{})

	if connectedness := server.Network().Connectedness(client.ID()); connectedness != network.NotConnected {
		t.Fatalf("Expected removed peer to be disconnected, got %s", connectedness)
	}
}

func TestHandleTunnelStreamRejectsPeerOutsideWhitelist(t *testing.T) {
	mn := mocknet.New()
	t.Cleanup(func() { _ = mn.Close() })

	server, err := mn.GenPeer()
	if err != nil {
		t.Fatal(err)
	}
	client, err := mn.GenPeer()
	if err != nil {
		t.Fatal(err)
	}
	if err := mn.LinkAll(); err != nil {
		t.Fatal(err)
	}
	if err := mn.ConnectAllButSelf(); err != nil {
		t.Fatal(err)
	}

	n := &Node{
		Host: server,
		wm:   whitelist.NewWhitelistManager(nil),
	}
	server.SetStreamHandler(protocol.ID(ProtocolID), n.handleTunnelStream)

	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	stream, err := client.NewStream(ctx, server.ID(), protocol.ID(ProtocolID))
	if err != nil {
		if errors.Is(err, network.ErrReset) {
			return
		}
		t.Fatalf("Expected rejected stream to be reset, got %v", err)
	}
	defer stream.Close()

	if err := stream.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	_, err = stream.Read(make([]byte, 1))
	if !errors.Is(err, network.ErrReset) && !errors.Is(err, io.EOF) {
		t.Fatalf("Expected rejected stream to be reset, got %v", err)
	}
}

func TestAcquireTunnelAllowsOnlyOneConcurrentTunnel(t *testing.T) {
	n := &Node{sets: &nodeSettings{}}

	const attempts = 100
	var acquired atomic.Int32
	var wg sync.WaitGroup
	wg.Add(attempts)
	for range attempts {
		go func() {
			defer wg.Done()
			if n.acquireTunnel(peer.ID("peer-a")) {
				acquired.Add(1)
			}
		}()
	}
	wg.Wait()

	if got := acquired.Load(); got != 1 {
		t.Fatalf("Expected exactly one acquired tunnel, got %d", got)
	}
	if got := n.tunnelCount(); got != 1 {
		t.Fatalf("Expected one active tunnel, got %d", got)
	}

	n.releaseTunnel(peer.ID("peer-a"))
	if !n.acquireTunnel(peer.ID("peer-b")) {
		t.Fatal("Expected tunnel slot to be reusable after release")
	}
}

func TestHandleTunnelStreamRejectsSecondActiveTunnel(t *testing.T) {
	mn := mocknet.New()
	t.Cleanup(func() { _ = mn.Close() })

	server, err := mn.GenPeer()
	if err != nil {
		t.Fatal(err)
	}
	client, err := mn.GenPeer()
	if err != nil {
		t.Fatal(err)
	}
	if err := mn.LinkAll(); err != nil {
		t.Fatal(err)
	}
	if err := mn.ConnectAllButSelf(); err != nil {
		t.Fatal(err)
	}

	n := &Node{
		Host: server,
		wm: whitelist.NewWhitelistManager(map[string]config.PeerInfo{
			client.ID().String(): {},
		}),
		sets: &nodeSettings{},
	}
	if !n.acquireTunnel(peer.ID("existing-peer")) {
		t.Fatal("Failed to reserve initial tunnel slot")
	}
	server.SetStreamHandler(protocol.ID(ProtocolID), n.handleTunnelStream)

	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	stream, err := client.NewStream(ctx, server.ID(), protocol.ID(ProtocolID))
	if err != nil {
		if errors.Is(err, network.ErrReset) {
			return
		}
		t.Fatalf("Expected second tunnel stream to be reset, got %v", err)
	}
	defer stream.Close()

	if err := stream.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	_, err = stream.Read(make([]byte, 1))
	if !errors.Is(err, network.ErrReset) && !errors.Is(err, io.EOF) {
		t.Fatalf("Expected second tunnel stream to be reset, got %v", err)
	}
}
