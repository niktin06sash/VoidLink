package node

import (
	"context"
	"log"
	"sync/atomic"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
)

func (n *Node) startClient() {
	if n.sets.serverAddress != "" {
		n.direct()
	} else {
		n.discover()
	}
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-n.ctx.Done():
			return
		case <-ticker.C:
			if n.sets.serverAddress != "" {
				n.direct()
			} else {
				n.discover()
			}
		}
	}
}
func (n *Node) direct() {
	if atomic.LoadInt32(&n.sets.tunnelActive) == 1 {
		return
	}
	maddr, err := multiaddr.NewMultiaddr(n.sets.serverAddress)
	if err != nil {
		log.Printf("client: invalid server address: %v", err)
		return
	}
	pi, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		log.Printf("client: parse server address failed: %v", err)
		return
	}
	if !n.wm.IsAllowed(pi.ID) {
		log.Printf("client: server peer not in whitelist: %s", pi.ID)
		return
	}
	if err := n.Host.Connect(n.ctx, *pi); err != nil {
		log.Printf("client: direct connect failed: %v", err)
		return
	}
	s, err := n.newTunnelStream(pi.ID)
	if err != nil {
		log.Printf("client: stream failed: %v", err)
		return
	}
	log.Printf("client: direct tunnel established peer=%s", pi.ID)
	n.startTunnel(s)
}
func (n *Node) discover() {
	if atomic.LoadInt32(&n.sets.tunnelActive) == 1 {
		return
	}
	conns := n.Host.Network().Conns()
	log.Printf("client: total active connections: %d, DHT RT size: %d", len(conns), n.DHT.RoutingTable().Size())
	log.Printf("client: searching for key: %s (rendezvous='%s')", n.key.String(), n.rendezvous)
	ctx, cancel := context.WithTimeout(n.ctx, time.Second*10)
	defer cancel()
	provChan := n.DHT.FindProvidersAsync(ctx, n.key, 1)
	for p := range provChan {
		log.Printf("client: found provider: id=%s addrs=%d allowed=%v", p.ID, len(p.Addrs), n.wm.IsAllowed(p.ID))
		if p.ID == n.Host.ID() || len(p.Addrs) == 0 {
			continue
		}
		log.Println(p.ID)
		if !n.wm.IsAllowed(p.ID) {
			continue
		}
		n.Host.Peerstore().AddAddrs(p.ID, p.Addrs, time.Hour)
		log.Printf("client: attempting stream peer=%s addrs=%d", p.ID, len(p.Addrs))
		s, err := n.newTunnelStream(p.ID)
		if err != nil {
			log.Printf("client: new stream failed peer=%s err=%v", p.ID, err)
			continue
		}
		log.Printf("client: stream established peer=%s", p.ID)
		n.startTunnel(s)
		log.Printf("client: tunnel ended peer=%s; will continue discovery", p.ID)
		break
	}
	if atomic.LoadInt32(&n.sets.tunnelActive) == 0 {
		for _, pid := range n.Host.Network().Peers() {
			if pid == n.Host.ID() {
				continue
			}
			if !n.wm.IsAllowed(pid) {
				continue
			}
			log.Printf("client: attempting stream to connected peer=%s", pid)
			s, err := n.newTunnelStream(pid)
			if err != nil {
				log.Printf("client: new stream (connected peer) failed peer=%s err=%v", pid, err)
				continue
			}
			log.Printf("client: stream established (connected peer) peer=%s", pid)
			n.startTunnel(s)
			log.Printf("client: tunnel ended peer=%s; will continue discovery", pid)
			break
		}
	}
	log.Println("client: discovery tick complete")
}
func (n *Node) newTunnelStream(pid peer.ID) (network.Stream, error) {
	ctx, cancel := context.WithTimeout(n.ctx, 10*time.Second)
	defer cancel()
	return n.Host.NewStream(ctx, pid, protocol.ID(ProtocolID))
}
