package node

import (
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/niktin06sash/VoidLink/internal/config"
)

const ProtocolID = "/voidlink/1.0"

func (n *Node) Run() error {
	err := n.bootstrap()
	if err != nil {
		return err
	}
	n.Host.SetStreamHandler(protocol.ID(ProtocolID), func(s network.Stream) {
		go n.startTunnel(s)
	})
	if n.Cfg.Role == config.Server {
		n.startServer()
		<-n.ctx.Done()
	} else {
		n.startClient()
	}
	return nil
}
