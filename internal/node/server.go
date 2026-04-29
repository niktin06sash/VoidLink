package node

import (
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/libp2p/go-libp2p/p2p/discovery/util"
)

func (n *Node) startServer() {
	routingDiscovery := routing.NewRoutingDiscovery(n.DHT)
	util.Advertise(n.ctx, routingDiscovery, n.rendezvous)
}
