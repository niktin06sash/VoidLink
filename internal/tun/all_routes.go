package tun

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

const etcResolvConf = "/etc/resolv.conf"

func (t *Tun) AddAllRoutes(serveraddr string) error {
	serverIP, err := extractIP(serveraddr)
	if err != nil {
		return err
	}
	oldData, err := os.ReadFile(etcResolvConf)
	if err == nil {
		t.originalDNS = make([]byte, len(oldData))
		copy(t.originalDNS, oldData)
	}
	gateway, gwIface, err := getDefaultGateway()
	if err != nil {
		return fmt.Errorf("get gateway: %w", err)
	}
	t.gateway = gateway
	t.gwIface = gwIface
	t.serverIP = serverIP
	commands := [][]string{
		{"ip", "route", "add", serverIP + "/32", "via", gateway, "dev", gwIface},
		{"ip", "route", "add", "default", "via", serverLocalIP, "dev", t.Iface.Name(), "metric", metric},
	}
	for _, args := range commands {
		cmd := exec.Command(args[0], args[1:]...)
		log.Printf("tun: exec %s", strings.Join(args, " "))
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("command failed %v: %w (output=%s)", args, err, strings.TrimSpace(string(out)))
		}
	}
	err = os.WriteFile(etcResolvConf, []byte("nameserver 8.8.8.8\n"), 0644)
	if err != nil {
		return fmt.Errorf("failed to write DNS configuration: %w", err)
	}
	log.Printf("tun: added routes to server %s via gateway %s dev %s", serverIP, gateway, gwIface)
	return nil
}
func (t *Tun) RemoveAllRoutes() {
	if t.serverIP == "" {
		return
	}
	commands := [][]string{
		{"ip", "route", "del", "default", "via", serverLocalIP, "dev", t.Iface.Name()},
		{"ip", "route", "del", t.serverIP + "/32"},
	}
	for _, args := range commands {
		_, err := exec.Command(args[0], args[1:]...).CombinedOutput()
		if err != nil {
			log.Printf("tun: failed to remove route: %v", err)
		}
	}
	if len(t.originalDNS) > 0 {
		err := os.WriteFile(etcResolvConf, t.originalDNS, 0644)
		if err != nil {
			log.Printf("tun: failed to restore backup DNS: %v", err)
		} else {
			log.Println("tun: DNS restored from backup")
		}
	} else {
		os.WriteFile(etcResolvConf, []byte("nameserver "+t.gateway+"\n"), 0644)
		log.Println("tun: DNS set to gateway")
	}
}
func extractIP(addr string) (string, error) {
	parts := strings.Split(addr, "/")
	for i, p := range parts {
		if p == "ip4" && i+1 < len(parts) {
			return parts[i+1], nil
		}
	}
	return "", fmt.Errorf("no ip4 in address: %s", addr)
}
func getDefaultGateway() (gateway, iface string, err error) {
	out, err := exec.Command("ip", "route", "show", "default").Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to get default gateway: %w", err)
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
		return "", "", fmt.Errorf("could not parse default gateway")
	}
	return gateway, iface, nil
}
