package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"slices"

	"github.com/niktin06sash/VoidLink/internal/config"
	"github.com/niktin06sash/VoidLink/internal/node"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status [role]",
	Short: "Show the status of the VoidLink tunnel",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		validRoles := []string{string(config.Server), string(config.Client)}
		if slices.Contains(validRoles, args[0]) {
			return nil
		}
		return fmt.Errorf("invalid role: %s", args[0])
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		role := args[0]
		finalPath := cfgFile
		if finalPath == "" {
			finalPath = config.GetConfigPath(role)
		}
		cfg, err := config.LoadConfig(finalPath)
		if err != nil {
			return err
		}
		priv, err := config.LoadIdentity(cfg.KeyPath, finalPath)
		if err != nil {
			return err
		}
		peerID, err := config.GetPeerID(priv)
		if err != nil {
			return err
		}
		fmt.Println("\n=== VoidLink Node Status ===")
		fmt.Printf("Config: %s\n", finalPath)
		fmt.Printf("Role: %s\n", cfg.Role)
		fmt.Printf("Peer ID: %s\n", peerID)
		fmt.Printf("Local IP: %s\n", cfg.LocalIP)
		fmt.Printf("Interface Name: %s\n", cfg.InterfaceName)
		fmt.Printf("Key Path: %s\n", cfg.KeyPath)
		fmt.Printf("Rendezvous: %s\n", cfg.Rendezvous)
		fmt.Printf("Whitelist entries: %d\n", len(cfg.Whitelist))
		socketPath := config.GetSocketPath(finalPath)
		conn, err := net.Dial("unix", socketPath)
		if err != nil {
			log.Printf("socket: error while connect: %v", err)
			return nil
		}
		defer conn.Close()
		var stats node.StatusResponse
		if err := json.NewDecoder(conn).Decode(&stats); err != nil {
			return fmt.Errorf("status: failed to read node status: %v", err)
		}
		printStatus(stats)
		return nil
	},
}

func printStatus(stats node.StatusResponse) {
	statusStr := "OFFLINE"
	if stats.TunnelActive {
		statusStr = "ONLINE (Connected)"
	} else if stats.Uptime != "" {
		statusStr = "IDLE (Searching)"
	}
	fmt.Printf("Status:         %s\n", statusStr)
	fmt.Printf("Uptime:         %s\n", stats.Uptime)
	if len(stats.PublicAddrs) > 0 {
		fmt.Printf("Public Addr:    %s\n", stats.PublicAddrs[0])
	}
	fmt.Printf("Memory Usage:   %s\n", formatBytes(stats.MemoryUsage))
	fmt.Printf("RX: %-12s | Speed: %s/s\n", formatBytes(stats.RxBytes), formatBytes(uint64(stats.RxSpeed)))
	fmt.Printf("TX: %-12s | Speed: %s/s\n", formatBytes(stats.TxBytes), formatBytes(uint64(stats.TxSpeed)))
	fmt.Println("--- Active Peers ---")
	if len(stats.ActivePeers) == 0 {
		fmt.Println("No active peer connections.")
	} else {
		fmt.Printf("%-15s %-22s %-12s %-8s\n", "PEER ID", "REMOTE ADDRESS", "LATENCY", "PROTO")
		for _, p := range stats.ActivePeers {
			idShort := p.ID
			if len(idShort) > 15 {
				idShort = idShort[:12] + "..."
			}
			fmt.Printf("%-15s %-22s %-12s %-8s\n",
				idShort, p.Addr, p.Latency, p.Transport)
		}
	}
}

func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
