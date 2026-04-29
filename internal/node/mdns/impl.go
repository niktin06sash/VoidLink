package mdns

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/niktin06sash/VoidLink/internal/config"
)

const MDNSName = "voidlink-local"

func NewDiscoveryNotifee(ctx context.Context, h host.Host, wl map[string]config.PeerInfo) *DiscoveryNotifee {
	return &DiscoveryNotifee{ctx: ctx, h: h, wl: wl}
}

type DiscoveryNotifee struct {
	h   host.Host
	wl  map[string]config.PeerInfo
	ctx context.Context
}

func (n *DiscoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	if pi.ID == n.h.ID() {
		return
	}
	_, ok := n.wl[pi.ID.String()]
	if !ok {
		return
	}
	err := n.h.Connect(n.ctx, pi)
	if err != nil {
		fmt.Printf("error while connect to mDNS %s: %v\n", pi.ID, err)
	}
}
