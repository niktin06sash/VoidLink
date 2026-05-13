package tun

import (
	"log"
	"net"
	"os/exec"
	"strings"
)

func ForceCleanup(iface string, serverAddress string) error {
	serverIP, err := extractIP(serverAddress)
	if err != nil {
		return err
	}
	if serverIP != "" {
		err := exec.Command("ip", "route", "del", serverIP+"/32").Run()
		if err != nil {
			log.Printf("force: failed to delete server's route %s: %v", serverIP, err)
		}
	}
	_, err = net.InterfaceByName(iface)
	if err != nil {
		log.Printf("force: interface %s is already gone, nothing to clean up", iface)
		return nil
	}
	out, err := exec.Command("ip", "route", "show", "dev", iface).Output()
	if err == nil {
		for line := range strings.SplitSeq(strings.TrimSpace(string(out)), "\n") {
			if line == "" {
				continue
			}
			fields := strings.Fields(line)
			if len(fields) == 0 {
				continue
			}
			err = exec.Command("ip", "route", "del", fields[0], "dev", iface).Run()
			if err != nil {
				log.Printf("force: failed to delete route %s dev %s: %v", fields[0], iface, err)
			}
		}
	} else {
		log.Printf("force: failed to list routes for interface %s: %v", iface, err)
	}
	err = exec.Command("resolvectl", "revert", iface).Run()
	if err != nil {
		log.Printf("force: failed to revert DNS for interface %s: %v", iface, err)
	}
	err = exec.Command("ip", "link", "delete", iface).Run()
	if err != nil {
		log.Printf("force: failed to delete interface %s: %v", iface, err)
	}
	log.Printf("force: force cleanup done interface=%s\n", iface)
	return nil
}
