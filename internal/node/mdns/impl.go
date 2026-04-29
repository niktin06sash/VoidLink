package mdns

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/niktin06sash/VoidLink/internal/node/whitelist"
)

const MDNSName = "voidlink-local"

func NewDiscoveryNotifee(ctx context.Context, h host.Host, wm *whitelist.WhitelistManager) *DiscoveryNotifee {
	return &DiscoveryNotifee{ctx: ctx, h: h, wm: wm}
}

type DiscoveryNotifee struct {
	h   host.Host
	ctx context.Context
	wm  *whitelist.WhitelistManager
}

func (n *DiscoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	if pi.ID == n.h.ID() {
		return
	}
	if !n.wm.IsAllowed(pi.ID) {
		return
	}
	err := n.h.Connect(n.ctx, pi)
	if err != nil {
		fmt.Printf("error while connect to mDNS %s: %v\n", pi.ID, err)
	}
}
