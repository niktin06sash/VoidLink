package node

import (
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
)

func (n *Node) startClient() {
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-n.ctx.Done():
			return
		case <-ticker.C:
			routingDiscovery := routing.NewRoutingDiscovery(n.DHT)
			peerChan, err := routingDiscovery.FindPeers(n.ctx, n.Cfg.Rendezvous)
			if err != nil {
				continue
			}
			for p := range peerChan {
				if p.ID == n.Host.ID() || len(p.Addrs) == 0 {
					continue
				}
				_, ok := n.Cfg.Whitelist[p.ID.String()]
				if !ok {
					continue
				}
				s, err := n.Host.NewStream(n.ctx, p.ID, protocol.ID(ProtocolID))
				if err != nil {
					fmt.Printf("stream error: %v\n", err)
					continue
				}
				n.startTunnel(s)
				break
			}
		}
	}
}
