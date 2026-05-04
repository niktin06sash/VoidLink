package node

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	quic "github.com/libp2p/go-libp2p/p2p/transport/quic"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"github.com/niktin06sash/VoidLink/internal/config"
	"github.com/niktin06sash/VoidLink/internal/node/gater"
	implmDNS "github.com/niktin06sash/VoidLink/internal/node/mdns"
	"github.com/niktin06sash/VoidLink/internal/node/whitelist"
	"github.com/niktin06sash/VoidLink/internal/tun"
)

type Node struct {
	Host       host.Host
	DHT        *dht.IpfsDHT
	Tun        *tun.Tun
	wm         *whitelist.WhitelistManager
	rendezvous string
	role       config.Role
	mdns       mdns.Service
	ctx        context.Context
	sets       *nodeSettings
}
type nodeSettings struct {
	path         string
	tunnelActive int32
	currentPeer  peer.ID
	peerMu       sync.RWMutex
	startTime    time.Time
	rxBytes      uint64
	txBytes      uint64
}

func NewNode(ctx context.Context, cfg *config.Config, tun *tun.Tun, privkey crypto.PrivKey, path string) (*Node, error) {
	wgh := whitelist.NewWhitelistManager(cfg.Whitelist)
	host, err := libp2p.New(
		libp2p.ConnectionGater(gater.NewSecurityGater(wgh)),
		libp2p.Identity(privkey),
		libp2p.ListenAddrStrings(
			"/ip4/0.0.0.0/tcp/0",
			"/ip4/0.0.0.0/udp/0/quic-v1",
		),
		libp2p.Security(noise.ID, noise.New),
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.Transport(quic.NewTransport),
		libp2p.EnableHolePunching(),
		libp2p.EnableRelay())
	if err != nil {
		return nil, fmt.Errorf("error while create node: %w", err)
	}
	log.Printf("node: created host peer_id=%s", host.ID())
	for _, a := range host.Addrs() {
		log.Printf("node: listen_addr=%s", a)
	}
	locmdns := implmDNS.NewDiscoveryNotifee(ctx, host, wgh)
	ser := mdns.NewMdnsService(host, implmDNS.MDNSName, locmdns)
	err = ser.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start mDNS: %w", err)
	}
	log.Printf("node: mdns started service=%s", implmDNS.MDNSName)
	kdht, err := dht.New(ctx, host, dht.Mode(dht.ModeAuto))
	if err != nil {
		return nil, fmt.Errorf("failed to create DHT: %w", err)
	}
	log.Printf("node: dht created mode=auto")
	sets := &nodeSettings{path: path, startTime: time.Now()}
	return &Node{
		Host:       host,
		DHT:        kdht,
		Tun:        tun,
		rendezvous: cfg.Rendezvous,
		wm:         wgh,
		role:       cfg.Role,
		mdns:       ser,
		ctx:        ctx,
		sets:       sets,
	}, nil
}

func (n *Node) Close() error {
	n.mdns.Close()
	n.DHT.Close()
	return n.Host.Close()
}
