package node

import (
	"context"
	"fmt"
	"log"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/peer"
)

func (n *Node) bootstrap() error {
	log.Printf("bootstrap: starting dht bootstrap...")
	if err := n.DHT.Bootstrap(n.ctx); err != nil {
		return fmt.Errorf("error while create bootstrap: %w", err)
	}
	log.Printf("bootstrap: dht bootstrap complete, connecting to default bootstrap peers=%d", len(dht.DefaultBootstrapPeers))
	var ok int
	for _, addr := range dht.DefaultBootstrapPeers {
		pi, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			log.Printf("bootstrap: parse bootstrap peer failed addr=%s err=%v", addr, err)
			continue
		}
		ctx, cancel := context.WithTimeout(n.ctx, 30*time.Second)
		err = n.Host.Connect(ctx, *pi)
		cancel()
		if err != nil {
			log.Printf("bootstrap: connect failed peer=%s err=%v", pi.ID, err)
			continue
		}
		ok++
	}
	if ok == 0 {
		log.Printf("bootstrap: WARNING no default bootstrap peers reachable; continuing without public DHT connectivity")
		return nil
	}
	log.Printf("bootstrap: connected bootstrap_peers=%d/%d", ok, len(dht.DefaultBootstrapPeers))
	return nil
}
