package cmd

import (
	"fmt"
	"log"
	"slices"

	"github.com/niktin06sash/VoidLink/internal/config"
	"github.com/niktin06sash/VoidLink/internal/node"
	"github.com/niktin06sash/VoidLink/internal/tun"
	"github.com/spf13/cobra"
)

var upCmd = &cobra.Command{
	Use:   "up [role]",
	Short: "Start VoidLink tunnel using a specific role",
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
	SilenceErrors: false,
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		role := args[0]
		finalDir := cfgFile
		if finalDir == "" {
			finalDir = config.GetConfigDir()
		}
		finalPath := config.GetConfigFilePath(finalDir, role)
		log.Printf("up: role=%s config=%s", role, finalDir)
		cfg, err := config.LoadConfig(finalPath)
		if err != nil {
			return err
		}
		priv, err := config.LoadIdentity(cfg.KeyPath)
		if err != nil {
			return err
		}
		peerID, err := config.GetPeerID(priv)
		if err != nil {
			return err
		}
		log.Printf("up: peer_id=%s iface=%s ip=%s rendezvous=%s whitelist=%d", peerID, cfg.InterfaceName, cfg.LocalIP, cfg.Rendezvous, len(cfg.Whitelist))
		tunnel, err := tun.NewTun(cfg)
		if err != nil {
			return err
		}
		defer tunnel.Close()
		if routeAll && config.Role(role) == config.Client && serverAddress != "" {
			if err := tunnel.AddAllRoutes(serverAddress); err != nil {
				return err
			}
			defer tunnel.RemoveAllRoutes()
		}
		if routeSplit && config.Role(role) == config.Client && serverAddress != "" {
			routepath := config.GetRoutesFilePath(finalDir)
			if err := tunnel.AddSplitRoutes(cmd.Context(), routepath); err != nil {
				return err
			}
			defer tunnel.RemoveSplitedRoutes()
		}
		noda, err := node.NewNode(cmd.Context(), cfg, tunnel, priv, finalPath, serverAddress, routeSplit)
		if err != nil {
			return err
		}
		defer noda.Close()
		err = noda.Run()
		if err != nil {
			return err
		}
		return nil
	},
}
