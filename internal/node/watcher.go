package node

import (
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
				newCfg, err := config.LoadConfig(n.path)
				if err != nil {
					continue
				}
				n.wm.Update(newCfg.Whitelist)
			}
		}
	}()
}
