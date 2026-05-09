package tun

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"

	"github.com/niktin06sash/VoidLink/internal/config"
	"github.com/songgao/water"
)

type Tun struct {
	Iface         *water.Interface
	gateway       string
	gwIface       string
	serverIP      string
	originalDNS   []byte
	blockedRoutes []string
}

const MTU = 1400
const metric = "500"

func NewTun(cfg *config.Config) (*Tun, error) {
	waterconf := water.Config{DeviceType: water.TUN}
	waterconf.Name = cfg.InterfaceName
	iface, err := water.New(waterconf)
	if err != nil {
		return nil, fmt.Errorf("error while created TUN-interface: %w", err)
	}
	log.Printf("tun: created interface name=%s", cfg.InterfaceName)
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
		{"ip", "link", "set", "dev", cfg.InterfaceName, "mtu", fmt.Sprintf("%d", MTU)},
	}

	for _, args := range commands {
		cmd := exec.Command(args[0], args[1:]...)
		log.Printf("tun: exec %s", strings.Join(args, " "))
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("command failed %v: %w (output=%s)", args, err, strings.TrimSpace(string(out)))
		}
	}
	log.Printf("tun: configured ip=%s dev=%s mtu=%d", cfg.LocalIP, cfg.InterfaceName, MTU)
	return nil
}

func (t *Tun) Close() error {
	return t.Iface.Close()
}
