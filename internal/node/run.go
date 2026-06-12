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
			log.Printf("run: dht created mode=autoserver")
		} else {
			kdht, err = dht.New(n.ctx, n.Host, dht.Mode(dht.ModeServer))
			log.Printf("run: dht created mode=server")
		}
		if err != nil {
			return fmt.Errorf("run: failed to create DHT: %w", err)
		}
		if kdht != nil {
			n.DHT = kdht
		}
		if err = n.bootstrap(); err != nil {
			return err
		}
	} else {
		log.Printf("run: direct mode, skipping DHT and bootstrap")
	}
	n.Host.SetStreamHandler(protocol.ID(ProtocolID), n.handleTunnelStream)
	n.statusSocket()
	n.watchSignal()
	if n.role == config.Server {
		log.Printf("run: role=server starting advertise...")
		n.startServer()
	} else {
		log.Printf("run: role=client starting discovery loop...")
		n.startClient()
	}
	return nil
}

func (n *Node) handleTunnelStream(s network.Stream) {
	remotePeer := s.Conn().RemotePeer()
	if !n.wm.IsAllowed(remotePeer) {
		log.Printf("run: inbound rejected peer=%s reason=not_whitelisted", remotePeer)
		if err := s.Reset(); err != nil {
			log.Printf("run: inbound reset failed peer=%s err=%v", remotePeer, err)
		}
		return
	}
	log.Printf("run: inbound opened peer=%s", remotePeer)
	go n.startTunnel(s)
}
