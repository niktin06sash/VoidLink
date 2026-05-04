package node

import (
	"log"

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
		log.Printf("stream: inbound opened peer=%s", s.Conn().RemotePeer())
		go n.startTunnel(s)
	})
	n.statusSocket()
	n.watchSignal()
	if n.role == config.Server {
		log.Printf("node: role=server starting advertise...")
		n.startServer()
		<-n.ctx.Done()
	} else {
		log.Printf("node: role=client starting discovery loop...")
		n.startClient()
	}
	return nil
}
