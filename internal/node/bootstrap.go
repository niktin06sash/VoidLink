package node

import (
	"fmt"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/peer"
)

func (n *Node) bootstrap() error {
	if err := n.DHT.Bootstrap(n.ctx); err != nil {
		return fmt.Errorf("error while create bootstrap: %w", err)
	}
	for _, addr := range dht.DefaultBootstrapPeers {
		pi, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			return fmt.Errorf("error while get info from p2p address: %w", err)
		}
		err = n.Host.Connect(n.ctx, *pi)
		if err != nil {
			return fmt.Errorf("error while connect to host: %w", err)
		}
	}
	return nil
}
