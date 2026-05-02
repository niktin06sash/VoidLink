package gater

import (
	"log"

	"github.com/libp2p/go-libp2p/core/control"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/niktin06sash/VoidLink/internal/node/whitelist"
)

func NewSecurityGater(wl *whitelist.WhitelistManager) *SecurityGater {
	return &SecurityGater{wm: wl}
}

type SecurityGater struct {
	wm *whitelist.WhitelistManager
}

func (g *SecurityGater) InterceptPeerDial(p peer.ID) bool {
	return g.wm.IsAllowed(p) || isBootstrapPeer(p)
}

func (g *SecurityGater) InterceptAddrDial(p peer.ID, a ma.Multiaddr) bool {
	return g.wm.IsAllowed(p) || isBootstrapPeer(p)
}

func (g *SecurityGater) InterceptAccept(c network.ConnMultiaddrs) bool {
	return true
}

func (g *SecurityGater) InterceptSecured(dir network.Direction, p peer.ID, c network.ConnMultiaddrs) bool {
	if isBootstrapPeer(p) {
		return true
	}
	allowed := g.wm.IsAllowed(p)
	if !allowed {
		log.Printf("gater: %s secured rejected peer=%s", dir, p)
	}
	return allowed
}
func (g *SecurityGater) InterceptUpgraded(c network.Conn) (bool, control.DisconnectReason) {
	return true, 0
}
