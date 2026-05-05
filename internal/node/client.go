package node

import (
	"context"
	"log"
	"sync/atomic"
	"time"

	cid "github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	mh "github.com/multiformats/go-multihash"
)

func (n *Node) startClient() {
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-n.ctx.Done():
			return
		case <-ticker.C:
			n.tick()
		}
	}
}
func (n *Node) tick() {
	if atomic.LoadInt32(&n.sets.tunnelActive) == 1 {
		return
	}
	conns := n.Host.Network().Conns()
	log.Printf("client: total active connections: %d", len(conns))
	for _, c := range conns {
		log.Printf("   -> connected to: %s (Dir: %s)", c.RemotePeer(), c.Stat().Direction)
	}
	rtSize := n.DHT.RoutingTable().Size()
	log.Printf("client: DHT Routing Table size: %d", rtSize)
	prefix := cid.Prefix{
		Version:  1,
		Codec:    cid.Raw,
		MhType:   mh.SHA2_256,
		MhLength: -1,
	}
	key, err := prefix.Sum([]byte(n.rendezvous))
	if err != nil {
		log.Printf("client: cid generation error: %v", err)
		return
	}
	log.Printf("client: searching for key: %s (rendezvous='%s')", key.String(), n.rendezvous)
	ctx, cancel := context.WithTimeout(n.ctx, time.Second*10)
	defer cancel()
	provChan := n.DHT.FindProvidersAsync(ctx, key, 1)
	for p := range provChan {
		if p.ID == n.Host.ID() || len(p.Addrs) == 0 {
			continue
		}
		log.Println(p.ID)
		if !n.wm.IsAllowed(p.ID) {
			continue
		}
		n.Host.Peerstore().AddAddrs(p.ID, p.Addrs, time.Hour)
		log.Printf("client: attempting stream peer=%s addrs=%d", p.ID, len(p.Addrs))
		s, err := n.newTunnelStream(p.ID)
		if err != nil {
			log.Printf("client: new stream failed peer=%s err=%v", p.ID, err)
			continue
		}
		log.Printf("client: stream established peer=%s", p.ID)
		n.startTunnel(s)
		log.Printf("client: tunnel ended peer=%s; will continue discovery", p.ID)
		break
	}
	if atomic.LoadInt32(&n.sets.tunnelActive) == 0 {
		for _, pid := range n.Host.Network().Peers() {
			if pid == n.Host.ID() {
				continue
			}
			if !n.wm.IsAllowed(pid) {
				continue
			}
			log.Printf("client: attempting stream to connected peer=%s", pid)
			s, err := n.newTunnelStream(pid)
			if err != nil {
				log.Printf("client: new stream (connected peer) failed peer=%s err=%v", pid, err)
				continue
			}
			log.Printf("client: stream established (connected peer) peer=%s", pid)
			n.startTunnel(s)
			log.Printf("client: tunnel ended peer=%s; will continue discovery", pid)
			break
		}
	}
	log.Println("client: discovery tick complete")
}
func (n *Node) newTunnelStream(pid peer.ID) (network.Stream, error) {
	ctx, cancel := context.WithTimeout(n.ctx, 10*time.Second)
	defer cancel()
	return n.Host.NewStream(ctx, pid, protocol.ID(ProtocolID))
}
