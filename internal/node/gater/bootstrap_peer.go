package gater

import (
	"sync"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/peer"
)

var (
	bootstrapOnce sync.Once
	bootstrapSet  map[peer.ID]struct{}
)

func isBootstrapPeer(p peer.ID) bool {
	bootstrapOnce.Do(func() {
		bootstrapSet = make(map[peer.ID]struct{}, len(dht.DefaultBootstrapPeers))
		for _, addr := range dht.DefaultBootstrapPeers {
			pi, err := peer.AddrInfoFromP2pAddr(addr)
			if err != nil {
				continue
			}
			bootstrapSet[pi.ID] = struct{}{}
		}
	})
	_, ok := bootstrapSet[p]
	return ok
}
