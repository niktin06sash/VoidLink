package node

import (
	"log"
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
			peerChan, err := routingDiscovery.FindPeers(n.ctx, n.rendezvous)
			if err != nil {
				log.Printf("client: find peers failed rendezvous=%s err=%v", n.rendezvous, err)
				continue
			}
			var found, allowed int
			for p := range peerChan {
				found++
				if p.ID == n.Host.ID() || len(p.Addrs) == 0 {
					continue
				}
				if !n.wm.IsAllowed(p.ID) {
					continue
				}
				allowed++
				log.Printf("client: attempting stream peer=%s addrs=%d", p.ID, len(p.Addrs))
				s, err := n.Host.NewStream(n.ctx, p.ID, protocol.ID(ProtocolID))
				if err != nil {
					log.Printf("client: new stream failed peer=%s err=%v", p.ID, err)
					continue
				}
				log.Printf("client: stream established peer=%s", p.ID)
				n.startTunnel(s)
				log.Printf("client: tunnel ended peer=%s; will continue discovery", p.ID)
				break
			}
			log.Printf("client: discovery tick complete found=%d allowed=%d", found, allowed)
		}
	}
}
