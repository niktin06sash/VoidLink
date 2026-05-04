package node

import (
	"context"
	"log"
	"sync/atomic"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/niktin06sash/VoidLink/internal/engine"
)

func (n *Node) startTunnel(s network.Stream) {
	remotePeer := s.Conn().RemotePeer()
	atomic.StoreInt32(&n.sets.tunnelActive, 1)
	n.sets.peerMu.Lock()
	n.sets.currentPeer = remotePeer
	n.sets.peerMu.Unlock()
	log.Printf("tunnel: start peer=%s", remotePeer)
	defer func() {
		atomic.StoreInt32(&n.sets.tunnelActive, 0)
		n.sets.peerMu.Lock()
		n.sets.currentPeer = ""
		n.sets.peerMu.Unlock()
		s.Close()
	}()
	streamctx, cancel := context.WithCancel(n.ctx)
	defer cancel()
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
		log.Printf("tunnel: stop (node context done) peer=%s", s.Conn().RemotePeer())
		return
	case <-donechan:
		log.Printf("tunnel: stop (direction ended) peer=%s", s.Conn().RemotePeer())
		cancel()
		return
	}
}
