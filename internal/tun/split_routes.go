package tun

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/niktin06sash/VoidLink/internal/config"
)

const splitListURL = "https://antifilter.download/list/subnet.lst"

func (t *Tun) AddSplitRoutes(ctx context.Context, routePath string) error {
	log.Printf("tun: downloading split IP list...")
	req, err := http.NewRequestWithContext(ctx, "GET", splitListURL, nil)
	if err != nil {
		return fmt.Errorf("tun: create request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("tun: download split list: %w", err)
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
		return fmt.Errorf("tun: scan list error: %w", err)
	}
	defRoutes, err := config.ReadRoutes(routePath)
	if err != nil {
		log.Println(err)
	}
	return t.addSplitRoutes(antifilter, defRoutes, routePath)
}

func (t *Tun) addSplitRoutes(antifilter, defRoutes []string, routePath string) error {
	log.Printf("tun: adding %d antifilter + %d default routes...", len(antifilter), len(defRoutes))
	addedAntifilter := make([]string, 0, len(antifilter))
	for _, route := range antifilter {
		_, err := t.run("ip", "route", "add", route, "via", serverLocalIP, "dev", t.name())
		if err == nil {
			addedAntifilter = append(addedAntifilter, route)
		} else {
			log.Printf("tun: failed to add antifilter route %s: %v", route, err)
		}
	}
	addedDefault := make([]string, 0, len(defRoutes))
	for _, route := range defRoutes {
		_, err := t.run("ip", "route", "add", route, "via", serverLocalIP, "dev", t.name())
		if err == nil {
			addedDefault = append(addedDefault, route)
		} else {
			log.Printf("tun: failed to add default route %s: %v", route, err)
		}
	}

	if err := t.setDNS("8.8.8.8"); err != nil {
		t.restoreDNS()
		t.rollbackSplitRoutes(addedAntifilter, addedDefault)
		return err
	}

	t.antifilterRoutes = addedAntifilter
	t.defaultRoutes = addedDefault
	t.routePath = routePath
	log.Printf("tun: split routes added antifilter=%d/%d default=%d/%d",
		len(addedAntifilter), len(antifilter), len(addedDefault), len(defRoutes))
	return nil
}

func (t *Tun) RemoveSplitedRoutes() {
	total := len(t.antifilterRoutes) + len(t.defaultRoutes)
	if total == 0 && !t.dnsConfigured {
		return
	}
	log.Printf("tun: removing %d split routes...", total)
	for _, route := range t.antifilterRoutes {
		t.deleteRoute(route, "via", serverLocalIP, "dev", t.name())
	}
	for _, route := range t.defaultRoutes {
		t.deleteRoute(route, "via", serverLocalIP, "dev", t.name())
	}
	t.antifilterRoutes = nil
	t.defaultRoutes = nil
	log.Println("tun: split routes removed")
	t.restoreDNS()
}

func (t *Tun) ReloadSplitedRoutes() error {
	newRoutes, err := config.ReadRoutes(t.routePath)
	if err != nil {
		return err
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
			_, err := t.run("ip", "route", "del", r, "via", serverLocalIP, "dev", t.name())
			if err != nil {
				log.Printf("tun: failed to remove old route %s: %v", r, err)
			}
		}
	}
	var result []string
	for _, r := range newRoutes {
		if _, ok := oldSet[r]; !ok {
			_, err := t.run("ip", "route", "add", r, "via", serverLocalIP, "dev", t.name())
			if err == nil {
				result = append(result, r)
			} else {
				log.Printf("tun: failed to add new route %s: %v", r, err)
			}
		} else {
			result = append(result, r)
		}
	}

	t.defaultRoutes = result
	log.Printf("tun: default routes reloaded total=%d", len(result))
	return nil
}

func (t *Tun) rollbackSplitRoutes(antifilter, defaults []string) {
	for i := len(defaults) - 1; i >= 0; i-- {
		t.deleteRoute(defaults[i], "via", serverLocalIP, "dev", t.name())
	}
	for i := len(antifilter) - 1; i >= 0; i-- {
		t.deleteRoute(antifilter[i], "via", serverLocalIP, "dev", t.name())
	}
}
