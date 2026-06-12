package node

import (
	"context"
	"log"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"github.com/niktin06sash/VoidLink/internal/engine"
)

func (n *Node) startTunnel(s network.Stream) {
	remotePeer := s.Conn().RemotePeer()
	if !n.acquireTunnel(remotePeer) {
		log.Printf("tunnel: rejected peer=%s reason=tunnel_already_active", remotePeer)
		if err := s.Reset(); err != nil {
			log.Printf("tunnel: reset rejected stream peer=%s err=%v", remotePeer, err)
		}
		return
	}

	log.Printf("tunnel: start peer=%s active=%d", remotePeer, n.tunnelCount())
	defer func() {
		n.releaseTunnel(remotePeer)
		s.Close()
		log.Printf("tunnel: stop peer=%s active=%d", remotePeer, n.tunnelCount())
	}()
	streamctx, cancel := context.WithCancel(n.ctx)
	defer cancel()
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-streamctx.Done():
				return
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(n.ctx, 5*time.Second)
				result := <-ping.Ping(ctx, n.Host, remotePeer)
				cancel()
				if result.Error == nil {
					if ps, ok := n.Host.Peerstore().(peerstore.Metrics); ok {
						ps.RecordLatency(remotePeer, result.RTT)
					}
				} else {
					log.Printf("tunnel: ping error peer=%s err=%v", remotePeer, result.Error)
				}
			}
		}
	}()
	donechan := make(chan struct{}, 2)
	go func() {
		engine.StreamToTun(streamctx, n.Tun.Iface, s, &n.sets.rxBytes)
		donechan <- struct{}{}
	}()
	go func() {
		engine.TunToStream(streamctx, n.Tun.Iface, s, &n.sets.txBytes)
		donechan <- struct{}{}
	}()
	select {
	case <-n.ctx.Done():
		log.Printf("tunnel: stop (node context done) peer=%s", remotePeer)
	case <-donechan:
		log.Printf("tunnel: stop (direction ended) peer=%s", remotePeer)
		cancel()
	}
}

func (n *Node) acquireTunnel(remotePeer peer.ID) bool {
	n.sets.tunnelsMu.Lock()
	defer n.sets.tunnelsMu.Unlock()

	if len(n.sets.activeTunnels) != 0 {
		return false
	}
	if n.sets.activeTunnels == nil {
		n.sets.activeTunnels = make(map[peer.ID]struct{}, 1)
	}
	n.sets.activeTunnels[remotePeer] = struct{}{}
	return true
}

func (n *Node) releaseTunnel(remotePeer peer.ID) {
	n.sets.tunnelsMu.Lock()
	defer n.sets.tunnelsMu.Unlock()
	delete(n.sets.activeTunnels, remotePeer)
}

func (n *Node) tunnelCount() int {
	n.sets.tunnelsMu.RLock()
	defer n.sets.tunnelsMu.RUnlock()
	return len(n.sets.activeTunnels)
}

func (n *Node) isTunnelActive() bool {
	return n.tunnelCount() > 0
}
