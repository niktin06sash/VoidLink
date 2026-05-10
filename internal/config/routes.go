package config

import (
	"fmt"
	"os"
	"strings"
)

var defaultRoutes = []string{
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

func ReadRoutes(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read default routes file: %w", err)
	}
	var routes []string
	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		routes = append(routes, line)
	}
	return routes, nil
}
func writeDefaultRoutes(path string) error {
	content := strings.Join(defaultRoutes, "\n") + "\n"
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write default routes: %w", err)
	}
	return nil
}
