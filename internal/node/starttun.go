package node

import (
	"context"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/niktin06sash/VoidLink/internal/engine"
)

func (n *Node) startTunnel(s network.Stream) {
	defer s.Close()
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
		return
	case <-donechan:
		cancel()
		return
	}
}
