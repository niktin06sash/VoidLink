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

func (m *WhitelistManager) Update(newData map[string]config.PeerInfo) {
	m.Lock()
	defer m.Unlock()
	m.data = newData
}
