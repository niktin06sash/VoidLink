package gater

import (
	"github.com/libp2p/go-libp2p/core/control"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/niktin06sash/VoidLink/internal/config"
)

func NewSecurityGater(wl map[string]config.PeerInfo) *SecurityGater {
	return &SecurityGater{whitelist: wl}
}

type SecurityGater struct {
	whitelist map[string]config.PeerInfo
}

func (g *SecurityGater) InterceptPeerDial(p peer.ID) bool {
	return g.isAllowed(p)
}

func (g *SecurityGater) InterceptAddrDial(p peer.ID, a ma.Multiaddr) bool {
	return g.isAllowed(p)
}

func (g *SecurityGater) InterceptAccept(c network.ConnMultiaddrs) bool {
	return true
}

func (g *SecurityGater) InterceptSecured(dir network.Direction, p peer.ID, c network.ConnMultiaddrs) bool {
	if dir == network.DirInbound {
		return g.isAllowed(p)
	}
	return true
}

func (g *SecurityGater) InterceptUpgraded(c network.Conn) (bool, control.DisconnectReason) {
	return true, 0
}
func (g *SecurityGater) isAllowed(p peer.ID) bool {
	_, ok := g.whitelist[p.String()]
	return ok
}
