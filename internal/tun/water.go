package tun

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/niktin06sash/VoidLink/internal/config"
	"github.com/songgao/water"
)

type Tun struct {
	Iface *water.Interface
}

func NewTun(cfg *config.Config) (*Tun, error) {
	waterconf := water.Config{DeviceType: water.TUN}
	waterconf.Name = cfg.InterfaceName
	iface, err := water.New(waterconf)
	if err != nil {
		return nil, fmt.Errorf("error while created TUN-interface: %w", err)
	}
	t := &Tun{
		Iface: iface,
	}
	if err := t.applySettings(cfg); err != nil {
		iface.Close()
		return nil, err
	}
	return t, nil
}
func (t *Tun) applySettings(cfg *config.Config) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("OS %s is not supported yet", runtime.GOOS)
	}
	commands := [][]string{
		{"ip", "addr", "add", cfg.LocalIP + "/24", "dev", cfg.InterfaceName},
		{"ip", "link", "set", "dev", cfg.InterfaceName, "up"},
		{"ip", "link", "set", "dev", cfg.InterfaceName, "mtu", "1400"},
	}

	for _, args := range commands {
		cmd := exec.Command(args[0], args[1:]...)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("command failed %v: %w", args, err)
		}
	}
	return nil
}
func (t *Tun) Close() error {
	if t.Iface != nil {
		return t.Iface.Close()
	}
	return nil
}
