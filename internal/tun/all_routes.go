package tun

import (
	"fmt"
	"log"
	"strings"
)

func (t *Tun) AddAllRoutes(serveraddr string) error {
	serverIP, err := extractIP(serveraddr)
	if err != nil {
		return err
	}
	gateway, gwIface, err := t.getDefaultGateway()
	if err != nil {
		return err
	}

	serverRoute := []string{"route", "add", serverIP + "/32", "via", gateway, "dev", gwIface}
	if err := t.runLogged("ip", serverRoute...); err != nil {
		return err
	}

	defaultRoute := []string{"route", "add", "default", "via", serverLocalIP, "dev", t.name(), "metric", metric}
	if err := t.runLogged("ip", defaultRoute...); err != nil {
		t.deleteRoute(serverIP+"/32", "via", gateway, "dev", gwIface)
		return err
	}

	if err := t.setDNS("8.8.8.8"); err != nil {
		t.restoreDNS()
		t.deleteRoute("default", "via", serverLocalIP, "dev", t.name())
		t.deleteRoute(serverIP+"/32", "via", gateway, "dev", gwIface)
		return err
	}

	t.gateway = gateway
	t.gwIface = gwIface
	t.serverIP = serverIP
	log.Printf("tun: added routes to server %s via gateway %s dev %s", serverIP, gateway, gwIface)
	return nil
}

func (t *Tun) RemoveAllRoutes() {
	if t.serverIP == "" {
		t.restoreDNS()
		return
	}
	t.deleteRoute("default", "via", serverLocalIP, "dev", t.name())
	t.deleteRoute(t.serverIP+"/32", "via", t.gateway, "dev", t.gwIface)
	t.restoreDNS()
	t.gateway = ""
	t.gwIface = ""
	t.serverIP = ""
}
func extractIP(addr string) (string, error) {
	parts := strings.Split(addr, "/")
	for i, p := range parts {
		if p == "ip4" && i+1 < len(parts) {
			return parts[i+1], nil
		}
	}
	return "", fmt.Errorf("tun: no ip4 in address: %s", addr)
}
func (t *Tun) getDefaultGateway() (gateway, iface string, err error) {
	out, err := t.run("ip", "route", "show", "default")
	if err != nil {
		return "", "", fmt.Errorf("tun: failed to get default gateway: %w", err)
	}
	fields := strings.Fields(string(out))
	for i, f := range fields {
		if f == "via" && i+1 < len(fields) {
			gateway = fields[i+1]
		}
		if f == "dev" && i+1 < len(fields) {
			iface = fields[i+1]
		}
	}
	if gateway == "" || iface == "" {
		return "", "", fmt.Errorf("tun: could not parse default gateway")
	}
	return gateway, iface, nil
}

func (t *Tun) runLogged(name string, args ...string) error {
	command := append([]string{name}, args...)
	log.Printf("tun: exec %s", strings.Join(command, " "))
	out, err := t.run(name, args...)
	if err != nil {
		return fmt.Errorf("tun: command failed %v: %w (output=%s)", command, err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (t *Tun) deleteRoute(destination string, args ...string) {
	commandArgs := append([]string{"route", "del", destination}, args...)
	if _, err := t.run("ip", commandArgs...); err != nil {
		log.Printf("tun: failed to remove route %s: %v", destination, err)
	}
}
