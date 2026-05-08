package node

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	"github.com/niktin06sash/VoidLink/internal/config"
)

type StatusResponse struct {
	Uptime       string       `json:"uptime"`
	MemoryUsage  uint64       `json:"memory_usage"`
	RxBytes      uint64       `json:"rx_bytes"`
	TxBytes      uint64       `json:"tx_bytes"`
	RxSpeed      float64      `json:"rx_speed"`
	TxSpeed      float64      `json:"tx_speed"`
	PublicAddrs  []string     `json:"public_addrs"`
	ActivePeers  []PeerDetail `json:"active_peers"`
	TunnelActive bool         `json:"tunnel_active"`
	CurrentPeer  Peer         `json:"current_peer"`
}

type Transport string
type Address string
type Latency string
type Peer string

const (
	NoneActivePeer   Peer      = "none"
	LatencyUnknown   Latency   = "n/a"
	AdressUnknown    Address   = "unknown"
	TransportUnknown Transport = "unknown"
	TransportQUIC    Transport = "QUIC"
	TransportTCP     Transport = "TCP"
	TransportUDP     Transport = "UDP"
)

type PeerDetail struct {
	ID        string    `json:"id"`
	Addr      Address   `json:"addr"`
	Latency   Latency   `json:"latency"`
	Transport Transport `json:"transport"`
}

func (n *Node) statusSocket() error {
	socketpath := config.GetSocketPath(n.sets.path)
	if err := os.Remove(socketpath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("socket: failed to remove old socket: %w", err)
	}
	l, err := net.Listen("unix", socketpath)
	if err != nil {
		return fmt.Errorf("socket: error while starts socket: %v", err)
	}
	n.startStatsInformer()
	go func() {
		for {
			select {
			case <-n.ctx.Done():
				l.Close()
				os.Remove(socketpath)
				return
			default:
			}
			conn, err := l.Accept()
			if err != nil {
				log.Printf("socket: error while accepted connection: %v", err)
				continue
			}
			go func(c net.Conn) {
				defer c.Close()
				resp := n.newStatusResponse()
				if err := json.NewEncoder(c).Encode(resp); err != nil {
					log.Printf("socket: encode error: %v", err)
				}
			}(conn)
		}
	}()
	return nil
}

func (n *Node) newStatusResponse() StatusResponse {
	n.sets.peerMu.RLock()
	cp := Peer(n.sets.currentPeer.String())
	n.sets.peerMu.RUnlock()
	if cp == "" {
		cp = NoneActivePeer
	}
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	alloc := m.Alloc
	var peerDetails []PeerDetail
	for _, p := range n.Host.Network().Peers() {
		latStr := LatencyUnknown
		if ps, ok := n.Host.Peerstore().(peerstore.Metrics); ok {
			latency := ps.LatencyEWMA(p)
			if latency > 0 {
				latStr = Latency(latency.Truncate(time.Microsecond).String())
			}
		}
		transport := TransportUnknown
		remoteAddr := AdressUnknown
		if conns := n.Host.Network().ConnsToPeer(p); len(conns) > 0 {
			conn := conns[0]
			remoteAddr = Address(conn.RemoteMultiaddr().String())
			ma := conn.RemoteMultiaddr()
			if _, err := ma.ValueForProtocol(multiaddr.P_QUIC_V1); err == nil {
				transport = TransportQUIC
			} else if _, err := ma.ValueForProtocol(multiaddr.P_TCP); err == nil {
				transport = TransportTCP
			} else if _, err := ma.ValueForProtocol(multiaddr.P_UDP); err == nil {
				transport = TransportUDP
			}
		}
		peerDetails = append(peerDetails, PeerDetail{
			ID:        p.String(),
			Addr:      remoteAddr,
			Latency:   latStr,
			Transport: transport,
		})
	}
	rx := atomic.LoadUint64(&n.sets.rxBytes)
	tx := atomic.LoadUint64(&n.sets.txBytes)
	uptime := time.Since(n.sets.startTime)
	n.sets.speedMu.Lock()
	rxSpeed := n.sets.currentRxSpeed
	txSpeed := n.sets.currentTxSpeed
	n.sets.speedMu.Unlock()
	var publicAddrs []string
	for _, addr := range n.Host.Addrs() {
		if !manet.IsThinWaist(addr) || manet.IsPublicAddr(addr) {
			publicAddrs = append(publicAddrs, addr.String())
		}
	}
	return StatusResponse{
		Uptime:       uptime.Truncate(time.Second).String(),
		MemoryUsage:  alloc,
		RxBytes:      rx,
		TxBytes:      tx,
		ActivePeers:  peerDetails,
		RxSpeed:      rxSpeed,
		TxSpeed:      txSpeed,
		PublicAddrs:  publicAddrs,
		TunnelActive: atomic.LoadInt32(&n.sets.tunnelActive) == 1,
		CurrentPeer:  cp,
	}
}
func (n *Node) startStatsInformer() {
	n.sets.speedMu.Lock()
	n.sets.lastRxBytes = atomic.LoadUint64(&n.sets.rxBytes)
	n.sets.lastTxBytes = atomic.LoadUint64(&n.sets.txBytes)
	n.sets.lastCheckTime = time.Now()
	n.sets.speedMu.Unlock()
	ticker := time.NewTicker(time.Second * 1)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-n.ctx.Done():
				return
			case <-ticker.C:
				now := time.Now()
				currRx := atomic.LoadUint64(&n.sets.rxBytes)
				currTx := atomic.LoadUint64(&n.sets.txBytes)

				n.sets.speedMu.Lock()
				duration := now.Sub(n.sets.lastCheckTime).Seconds()
				if duration > 0 {
					n.sets.currentRxSpeed = float64(currRx-n.sets.lastRxBytes) / duration
					n.sets.currentTxSpeed = float64(currTx-n.sets.lastTxBytes) / duration
				}
				n.sets.lastRxBytes = currRx
				n.sets.lastTxBytes = currTx
				n.sets.lastCheckTime = now
				n.sets.speedMu.Unlock()
			}
		}
	}()
}
