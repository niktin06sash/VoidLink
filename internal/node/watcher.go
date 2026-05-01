package node

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/niktin06sash/VoidLink/internal/config"
)

func (n *Node) WatchSignal() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP)
	go func() {
		for {
			select {
			case <-n.ctx.Done():
				return
			case <-sigs:
				log.Printf("config: received SIGHUP, reloading whitelist path=%s", n.path)
				newCfg, err := config.LoadConfig(n.path)
				if err != nil {
					log.Printf("config: reload failed err=%v", err)
					continue
				}
				n.wm.Update(newCfg.Whitelist)
				log.Printf("config: whitelist reloaded entries=%d", len(newCfg.Whitelist))
			}
		}
	}()
}
