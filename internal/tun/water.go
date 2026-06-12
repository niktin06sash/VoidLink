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
	Iface            *water.Interface
	ifaceName        string
	runCommand       commandRunner
	gateway          string
	gwIface          string
	serverIP         string
	antifilterRoutes []string
	defaultRoutes    []string
	routePath        string
	dnsConfigured    bool
}

type commandRunner func(name string, args ...string) ([]byte, error)

const MTU = 1400
const metric = "500"
const serverLocalIP = "10.1.1.1"

func NewTun(cfg *config.Config) (*Tun, error) {
	waterconf := water.Config{DeviceType: water.TUN}
	waterconf.Name = cfg.InterfaceName
	iface, err := water.New(waterconf)
	if err != nil {
		return nil, fmt.Errorf("tun: error while created TUN-interface: %w", err)
	}
	log.Printf("tun: created interface name=%s", cfg.InterfaceName)
	t := &Tun{
		Iface:     iface,
		ifaceName: iface.Name(),
	}
	if err := t.applySettings(cfg); err != nil {
		iface.Close()
		return nil, err
	}
	return t, nil
}
func (t *Tun) applySettings(cfg *config.Config) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("tun: OS %s is not supported yet", runtime.GOOS)
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
			return fmt.Errorf("tun: command failed %v: %w (output=%s)", args, err, strings.TrimSpace(string(out)))
		}
	}
	log.Printf("tun: configured ip=%s dev=%s mtu=%d", cfg.LocalIP, cfg.InterfaceName, MTU)
	return nil
}

func (t *Tun) Close() error {
	err := exec.Command("ip", "link", "delete", t.name()).Run()
	if err != nil {
		log.Printf("tun: failed to delete interface %s: %v", t.name(), err)
	}
	return t.Iface.Close()
}

func (t *Tun) setDNS(dns string) error {
	log.Printf("tun: setting DNS to %s for interface %s", dns, t.name())
	if _, err := t.run("resolvectl", "dns", t.name(), dns); err != nil {
		return fmt.Errorf("tun: resolvectl dns: %w", err)
	}
	t.dnsConfigured = true
	if _, err := t.run("resolvectl", "domain", t.name(), "~."); err != nil {
		t.restoreDNS()
		return fmt.Errorf("tun: resolvectl domain: %w", err)
	}
	log.Printf("tun: DNS set to %s", dns)
	return nil
}

func (t *Tun) restoreDNS() {
	if !t.dnsConfigured {
		return
	}
	if err := t.revertDNS(); err == nil {
		t.dnsConfigured = false
	}
}

func (t *Tun) revertDNS() error {
	_, err := t.run("resolvectl", "revert", t.name())
	if err != nil {
		log.Printf("tun: failed to restore DNS: %v", err)
	} else {
		log.Printf("tun: DNS restored")
	}
	return err
}

func (t *Tun) run(name string, args ...string) ([]byte, error) {
	if t.runCommand != nil {
		return t.runCommand(name, args...)
	}
	return exec.Command(name, args...).CombinedOutput()
}

func (t *Tun) name() string {
	if t.ifaceName != "" {
		return t.ifaceName
	}
	return t.Iface.Name()
}
