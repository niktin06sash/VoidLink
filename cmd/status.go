package cmd

import (
	"fmt"
	"slices"

	"github.com/niktin06sash/VoidLink/internal/config"
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

		fmt.Printf("Config: %s\n", finalPath)
		fmt.Printf("Role: %s\n", cfg.Role)
		fmt.Printf("Peer ID: %s\n", peerID)
		fmt.Printf("Local IP: %s\n", cfg.LocalIP)
		fmt.Printf("Interface Name: %s\n", cfg.InterfaceName)
		fmt.Printf("Key Path: %s\n", cfg.KeyPath)
		fmt.Printf("Rendezvous: %s\n", cfg.Rendezvous)
		fmt.Printf("Whitelist entries: %d\n", len(cfg.Whitelist))
		return nil
	},
}
