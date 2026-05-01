package node

import (
	"context"
	"log"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/niktin06sash/VoidLink/internal/engine"
)

func (n *Node) startTunnel(s network.Stream) {
	defer s.Close()
	log.Printf("tunnel: start peer=%s", s.Conn().RemotePeer())
	streamctx, cancel := context.WithCancel(n.ctx)
	defer cancel()
	donechan := make(chan struct{}, 2)
	go func() {
		engine.StreamToTun(streamctx, n.Tun.Iface, s)
		donechan <- struct{}{}
	}()
	go func() {
		engine.TunToStream(streamctx, n.Tun.Iface, s)
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
