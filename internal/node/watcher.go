package node

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/niktin06sash/VoidLink/internal/config"
)

func (n *Node) watchSignal() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP)
	go func() {
		for {
			select {
			case <-n.ctx.Done():
				return
			case <-sigs:
				log.Printf("watcher: received SIGHUP, reloading whitelist path=%s", n.sets.path)
				newCfg, err := config.LoadConfig(n.sets.path)
				if err != nil {
					log.Printf("watcher: reload failed err=%v", err)
					continue
				}
				n.wm.Update(newCfg.Whitelist)
				log.Printf("watcher: whitelist reloaded entries=%d", len(newCfg.Whitelist))
				if n.sets.routeSplited {
					if err := n.Tun.ReloadSplitedRoutes(); err != nil {
						log.Printf("watcher: routes reload failed err=%v", err)
					}
				}
			}
		}
	}()
}
