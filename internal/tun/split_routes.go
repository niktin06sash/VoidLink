package tun

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/niktin06sash/VoidLink/internal/config"
)

const splitListURL = "https://antifilter.download/list/subnet.lst"

func (t *Tun) AddSplitRoutes(ctx context.Context, routePath string) error {
	log.Printf("tun: downloading split IP list...")
	req, err := http.NewRequestWithContext(ctx, "GET", splitListURL, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("download split list: %w", err)
	}
	defer resp.Body.Close()
	var antifilter []string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		antifilter = append(antifilter, line)
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan list error: %w", err)
	}
	defRoutes, err := config.ReadRoutes(routePath)
	if err != nil {
		log.Printf("tun: failed to read default routes: %v", err)
	}
	log.Printf("tun: adding %d antifilter + %d default routes...", len(antifilter), len(defRoutes))
	var addedAntifilter int
	for _, route := range antifilter {
		cmd := exec.Command("ip", "route", "add", route, "via", serverLocalIP, "dev", t.Iface.Name())
		if cmd.Run() == nil {
			t.antifilterRoutes = append(t.antifilterRoutes, route)
			addedAntifilter++
		}
	}
	var addedDefault int
	for _, route := range defRoutes {
		cmd := exec.Command("ip", "route", "add", route, "via", serverLocalIP, "dev", t.Iface.Name())
		if cmd.Run() == nil {
			t.defaultRoutes = append(t.defaultRoutes, route)
			addedDefault++
		}
	}
	t.routePath = routePath
	if err := t.setDNS("8.8.8.8"); err != nil {
		log.Printf("tun: failed to set DNS: %v", err)
	}
	log.Printf("tun: split routes added antifilter=%d/%d default=%d/%d",
		addedAntifilter, len(antifilter), addedDefault, len(defRoutes))
	return nil
}

func (t *Tun) RemoveSplitedRoutes() {
	total := len(t.antifilterRoutes) + len(t.defaultRoutes)
	if total == 0 {
		return
	}
	log.Printf("tun: removing %d split routes...", total)
	for _, route := range t.antifilterRoutes {
		if err := exec.Command("ip", "route", "del", route, "via", serverLocalIP, "dev", t.Iface.Name()).Run(); err != nil {
			log.Printf("tun: failed to remove antifilter route %s: %v", route, err)
		}
	}
	for _, route := range t.defaultRoutes {
		if err := exec.Command("ip", "route", "del", route, "via", serverLocalIP, "dev", t.Iface.Name()).Run(); err != nil {
			log.Printf("tun: failed to remove default route %s: %v", route, err)
		}
	}
	t.antifilterRoutes = nil
	t.defaultRoutes = nil
	log.Println("tun: split routes removed")
	t.restoreDNS()
}

func (t *Tun) ReloadSplitedRoutes() error {
	newRoutes, err := config.ReadRoutes(t.routePath)
	if err != nil {
		return fmt.Errorf("reload routes: %w", err)
	}
	oldSet := make(map[string]struct{}, len(t.defaultRoutes))
	for _, r := range t.defaultRoutes {
		oldSet[r] = struct{}{}
	}
	newSet := make(map[string]struct{}, len(newRoutes))
	for _, r := range newRoutes {
		newSet[r] = struct{}{}
	}
	for _, r := range t.defaultRoutes {
		if _, ok := newSet[r]; !ok {
			exec.Command("ip", "route", "del", r, "via", serverLocalIP, "dev", t.Iface.Name()).Run()
		}
	}
	var result []string
	for _, r := range newRoutes {
		if _, ok := oldSet[r]; !ok {
			cmd := exec.Command("ip", "route", "add", r, "via", serverLocalIP, "dev", t.Iface.Name())
			if cmd.Run() == nil {
				result = append(result, r)
			}
		} else {
			result = append(result, r)
		}
	}

	t.defaultRoutes = result
	log.Printf("tun: default routes reloaded total=%d", len(result))
	return nil
}
