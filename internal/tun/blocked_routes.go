package tun

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

var extraRoutes = []string{
	// YouTube / Google
	"173.194.0.0/16", "74.125.0.0/16", "142.250.0.0/15",
	"172.217.0.0/16", "216.58.192.0/19", "64.233.160.0/19",
	// Telegram
	"91.108.4.0/22", "91.108.8.0/22", "91.108.56.0/22",
	"91.108.12.0/22", "91.108.16.0/22", "91.108.20.0/22",
	"149.154.160.0/20", "149.154.128.0/17",
	// Instagram / Facebook
	"157.240.0.0/16", "31.13.24.0/21", "31.13.64.0/18",
	// Twitter/X
	"104.244.40.0/21", "192.133.76.0/22",
	// --- CLOUDFLARE
	"173.245.48.0/20",
	"103.21.244.0/22",
	"103.22.200.0/22",
	"103.31.4.0/22",
	"141.101.64.0/18",
	"108.162.192.0/18",
	"190.93.240.0/20",
	"188.114.96.0/20",
	"197.234.240.0/22",
	"198.41.128.0/17",
	"162.158.0.0/15",
	"104.16.0.0/13",
	"172.64.0.0/13",
	"131.0.72.0/22",
	//OPENSEA
	"77.94.160.0/19",
	//CLAUDE
	"160.79.104.0/21",
	//PANEL
	"8.47.69.0/24",
}

const blockedListURL = "https://antifilter.download/list/subnet.lst"

func (t *Tun) AddBlockedRoutes(ctx context.Context) error {
	log.Printf("tun: downloading blocked IP list...")
	req, err := http.NewRequestWithContext(ctx, "GET", blockedListURL, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("download blocked list: %w", err)
	}
	defer resp.Body.Close()
	var routes []string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		routes = append(routes, line)
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan list error: %w", err)
	}
	routes = append(routes, extraRoutes...)
	log.Printf("tun: adding %d blocked routes...", len(routes))
	var added int
	for _, route := range routes {
		cmd := exec.Command("ip", "route", "add", route, "via", "10.1.1.1", "dev", t.Iface.Name())
		if err := cmd.Run(); err == nil {
			t.blockedRoutes = append(t.blockedRoutes, route)
			added++
		} else {
			log.Printf("failed to add route %s: %v", route, err)
		}
	}
	log.Printf("tun: blocked routes added=%d/%d", added, len(routes))
	return nil
}
func (t *Tun) RemoveBlockedRoutes() {
	if len(t.blockedRoutes) == 0 {
		return
	}
	log.Printf("tun: removing %d blocked routes...", len(t.blockedRoutes))
	for _, route := range t.blockedRoutes {
		err := exec.Command("ip", "route", "del", route, "via", "10.1.1.1", "dev", t.Iface.Name()).Run()
		if err != nil {
			log.Printf("failed to remove route %s: %v", route, err)
		}
	}
	t.blockedRoutes = nil
	log.Printf("tun: blocked routes removed")
}
