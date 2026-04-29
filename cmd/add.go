package cmd

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/niktin06sash/VoidLink/internal/config"
	"github.com/spf13/cobra"
)

var addPeerCmd = &cobra.Command{
	Use:   "add-peer [role] [peerID]",
	Short: "Add a new PeerID to the whitelist",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		role := args[0]
		newID := args[1]
		finalPath := cfgFile
		if finalPath == "" {
			finalPath = config.GetConfigPath(role)
		}
		cfg, err := config.LoadConfig(finalPath)
		if err != nil {
			return err
		}
		if cfg.Whitelist == nil {
			cfg.Whitelist = make(map[string]config.PeerInfo)
		}
		cfg.Whitelist[newID] = config.PeerInfo{
			Name:    peerName,
			AddedAt: time.Now().Format("2006-01-02 15:04:05"),
		}
		if err := config.SaveConfig(finalPath, *cfg); err != nil {
			return err
		}
		fmt.Printf("Successfully added peer %s to %s\n", newID, finalPath)
		exec.Command("sudo", "pkill", "-HUP", "voidlink").Run()
		fmt.Println("Sent SIGHUP to voidlink process to reload whitelist.")
		return nil
	}}
