package node

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	mh "github.com/multiformats/go-multihash"

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
	key        cid.Cid
}
type nodeSettings struct {
	path           string
	activeTunnels  map[peer.ID]struct{}
	tunnelsMu      sync.RWMutex
	startTime      time.Time
	rxBytes        uint64
	txBytes        uint64
	serverAddress  string
	lastRxBytes    uint64
	lastTxBytes    uint64
	lastCheckTime  time.Time
	currentRxSpeed float64
	currentTxSpeed float64
	speedMu        sync.Mutex
	routeSplited   bool
}

func NewNode(ctx context.Context, cfg *config.Config, tun *tun.Tun, privkey crypto.PrivKey, path string, serveradr string, routesplit bool) (*Node, error) {
	wgh := whitelist.NewWhitelistManager(cfg.Whitelist)
	host, err := libp2p.New(
		libp2p.ConnectionGater(gater.NewSecurityGater(wgh)),
		libp2p.Identity(privkey),
		libp2p.ListenAddrStrings(buildListenAddrs(cfg)...),
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
	sets := &nodeSettings{path: path, startTime: time.Now(), serverAddress: serveradr, routeSplited: routesplit}
	prefix := cid.Prefix{
		Version:  1,
		Codec:    cid.Raw,
		MhType:   mh.SHA2_256,
		MhLength: -1,
	}
	key, err := prefix.Sum([]byte(cfg.Rendezvous))
	if err != nil {
		log.Printf("node: cid generation error: %v", err)
		return nil, err
	}
	return &Node{
		Host:       host,
		Tun:        tun,
		rendezvous: cfg.Rendezvous,
		wm:         wgh,
		role:       cfg.Role,
		mdns:       ser,
		ctx:        ctx,
		sets:       sets,
		key:        key,
	}, nil
}

func (n *Node) Close() error {
	if n.mdns != nil {
		n.mdns.Close()
	}
	if n.DHT != nil {
		n.DHT.Close()
	}
	return n.Host.Close()
}
func buildListenAddrs(cfg *config.Config) []string {
	port := 0
	if cfg.ListenPort != 0 {
		port = cfg.ListenPort
	}
	return []string{
		fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port),
		fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic-v1", port),
	}
}
