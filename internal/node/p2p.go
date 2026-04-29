package node

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p/core/crypto"

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
	"github.com/niktin06sash/VoidLink/internal/tun"
)

type Node struct {
	Host host.Host
	DHT  *dht.IpfsDHT
	Tun  *tun.Tun
	Cfg  *config.Config
	Mdns mdns.Service
	ctx  context.Context
}

func NewNode(ctx context.Context, cfg *config.Config, tun *tun.Tun, privkey crypto.PrivKey) (*Node, error) {
	host, err := libp2p.New(
		libp2p.ConnectionGater(gater.NewSecurityGater(cfg.Whitelist)),
		libp2p.Identity(privkey),
		libp2p.ListenAddrStrings(
			"/ip4/0.0.0.0/tcp/0",
			"/ip4/0.0.0.0/udp/0/quic-v1",
		),
		libp2p.Security(noise.ID, noise.New),
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.Transport(quic.NewTransport))
	if err != nil {
		return nil, fmt.Errorf("error while create node: %w", err)
	}
	locmdns := implmDNS.NewDiscoveryNotifee(ctx, host, cfg.Whitelist)
	ser := mdns.NewMdnsService(host, implmDNS.MDNSName, locmdns)
	err = ser.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start mDNS: %w", err)
	}
	kdht, err := dht.New(ctx, host, dht.Mode(dht.ModeAuto))
	if err != nil {
		return nil, fmt.Errorf("failed to create DHT: %w", err)
	}
	return &Node{Host: host, DHT: kdht, Tun: tun, Cfg: cfg}, nil
}

func (n *Node) Close() error {
	n.Mdns.Close()
	n.DHT.Close()
	return n.Host.Close()
}
