package node

import (
	"context"
	"log"
	"time"
)

func (n *Node) startServer() {
	log.Printf("server: addresses:")
	for _, addr := range n.Host.Addrs() {
		log.Printf("server:   -> %s", addr)
	}
	n.doProvide()
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-n.ctx.Done():
			return
		case <-ticker.C:
			n.doProvide()
		}
	}
}

func (n *Node) doProvide() {
	ctx, cancel := context.WithTimeout(n.ctx, 30*time.Second)
	conns := n.Host.Network().Conns()
	log.Printf("server: total active connections: %d, DHT RT size: %d", len(conns), n.DHT.RoutingTable().Size())
	defer cancel()
	if err := n.DHT.Provide(ctx, n.key, true); err != nil {
		log.Printf("server: provide error: %v", err)
	} else {
		log.Printf("server: providing key=%s rendezvous=%s", n.key.String(), n.rendezvous)
	}
}
