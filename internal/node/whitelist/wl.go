package whitelist

import (
	"sync"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/niktin06sash/VoidLink/internal/config"
)

func NewWhitelistManager(data map[string]config.PeerInfo) *WhitelistManager {
	return &WhitelistManager{data: data}
}

type WhitelistManager struct {
	sync.RWMutex
	data map[string]config.PeerInfo
}

func (m *WhitelistManager) IsAllowed(p peer.ID) bool {
	m.RLock()
	defer m.RUnlock()
	_, ok := m.data[p.String()]
	return ok
}

func (m *WhitelistManager) Update(newData map[string]config.PeerInfo) []peer.ID {
	m.Lock()
	defer m.Unlock()

	removed := make([]peer.ID, 0)
	for rawID := range m.data {
		if _, ok := newData[rawID]; ok {
			continue
		}
		peerID, err := peer.Decode(rawID)
		if err == nil {
			removed = append(removed, peerID)
		}
	}
	m.data = newData
	return removed
}
