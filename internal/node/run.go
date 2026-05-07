package node

import (
	"fmt"
	"log"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/niktin06sash/VoidLink/internal/config"
)

const ProtocolID = "/voidlink/1.0"

func (n *Node) Run() error {
	var kdht *dht.IpfsDHT
	var err error
	directMode := n.sets.serverAddress != ""
	if n.role == config.Server || !directMode {
		if n.role == config.Client {
			kdht, err = dht.New(n.ctx, n.Host, dht.Mode(dht.ModeAutoServer))
			log.Printf("node: dht created mode=autoserver")
		} else {
			kdht, err = dht.New(n.ctx, n.Host, dht.Mode(dht.ModeServer))
			log.Printf("node: dht created mode=server")
		}
		if err != nil {
			return fmt.Errorf("failed to create DHT: %w", err)
		}
		if kdht != nil {
			n.DHT = kdht
		}
		if err = n.bootstrap(); err != nil {
			return err
		}
	} else {
		log.Printf("node: direct mode, skipping DHT and bootstrap")
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
	} else {
		log.Printf("node: role=client starting discovery loop...")
		n.startClient()
	}
	return nil
}
