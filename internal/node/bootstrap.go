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
	if err := n.DHT.Bootstrap(n.ctx); err != nil {
		return fmt.Errorf("error while create bootstrap: %w", err)
	}
	c := n.connectBootstrapPeers()
	if c == 0 {
		log.Printf("bootstrap: WARNING no bootstrap peers reachable")
		return nil
	}
	log.Printf("bootstrap: connected %d/%d peers.", c, len(dht.DefaultBootstrapPeers))
	go func() {
		ticker := time.NewTicker(2 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				log.Printf("bootstrap: re-bootstrapping DHT...")
				err := n.DHT.Bootstrap(n.ctx)
				if err != nil {
					log.Printf("error while create bootstrap: %v", err)
					continue
				}
				if n.DHT.RoutingTable().Size() < 3 {
					n.connectBootstrapPeers()
				}
			case <-n.ctx.Done():
				return
			}
		}
	}()
	return nil
}

func (n *Node) connectBootstrapPeers() int {
	var counter int
	for _, addr := range dht.DefaultBootstrapPeers {
		pi, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			continue
		}
		ctx, cancel := context.WithTimeout(n.ctx, 30*time.Second)
		err = n.Host.Connect(ctx, *pi)
		cancel()
		if err != nil {
			log.Printf("bootstrap: connect failed peer=%s err=%v", pi.ID, err)
			continue
		}
		counter++
		log.Printf("bootstrap: connected peer=%s", pi.ID)
	}
	return counter
}
