package whitelist

import (
	"sync"
	"testing"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/niktin06sash/VoidLink/internal/config"
)

func TestWhitelistManager(t *testing.T) {
	p1, _ := generateTestPeerID(t)
	p2, _ := generateTestPeerID(t)
	data := map[string]config.PeerInfo{
		p1.String(): {},
	}
	wm := NewWhitelistManager(data)

	t.Run("IsAllowed_Valid", func(t *testing.T) {
		if !wm.IsAllowed(p1) {
			t.Errorf("Peer %s should be allowed", p1.String())
		}
	})

	t.Run("IsAllowed_Blocked", func(t *testing.T) {
		if wm.IsAllowed(p2) {
			t.Errorf("Peer %s should be blocked", p2.String())
		}
	})

	t.Run("Update_Logic", func(t *testing.T) {
		newData := map[string]config.PeerInfo{
			p2.String(): {},
		}
		removed := wm.Update(newData)

		if wm.IsAllowed(p1) {
			t.Error("Old peer p1 should now be blocked")
		}
		if !wm.IsAllowed(p2) {
			t.Error("New peer p2 should now be allowed")
		}
		if len(removed) != 1 || removed[0] != p1 {
			t.Fatalf("Expected removed peer %s, got %v", p1, removed)
		}
	})
}

func TestWhitelistManager_Race(t *testing.T) {
	p, _ := generateTestPeerID(t)
	wm := NewWhitelistManager(map[string]config.PeerInfo{p.String(): {}})

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for range 1000 {
			wm.IsAllowed(p)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			tmp, _ := generateTestPeerID(t)
			wm.Update(map[string]config.PeerInfo{
				tmp.String(): {},
			})
		}
	}()

	wg.Wait()
}
func generateTestPeerID(t *testing.T) (peer.ID, error) {
	t.Helper()
	_, pub, err := crypto.GenerateKeyPair(crypto.Ed25519, -1)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	id, err := peer.IDFromPublicKey(pub)
	if err != nil {
		t.Fatalf("failed to get peer id: %v", err)
	}
	return id, nil
}
