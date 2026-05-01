package mdns

import (
	"context"
	"log"

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
		log.Printf("mdns: peer found but not whitelisted peer=%s", pi.ID)
		return
	}
	log.Printf("mdns: attempting connect peer=%s addrs=%d", pi.ID, len(pi.Addrs))
	err := n.h.Connect(n.ctx, pi)
	if err != nil {
		log.Printf("mdns: connect failed peer=%s err=%v", pi.ID, err)
		return
	}
	log.Printf("mdns: connected peer=%s", pi.ID)
}
