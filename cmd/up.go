package cmd

import (
	"fmt"

	"github.com/niktin06sash/VoidLink/internal/config"
	"github.com/spf13/cobra"
)

var upCmd = &cobra.Command{
	Use:           "up [role]",
	Short:         "Start VoidLink tunnel using a specific profile",
	Args:          cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	ValidArgs:     []string{"server", "client"},
	SilenceErrors: false,
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		roleName := args[0]
		cfg, err := config.LoadConfig(roleName)
		if err != nil {
			return err
		}
		priv, err := config.LoadIdentity(cfg.KeyPath)
		if err != nil {
			return err
		}
		myID, err := config.GetPeerID(priv)
		if err != nil {
			return err
		}
		fmt.Printf("Starting VoidLink as %s\n", cfg.Role)
		fmt.Printf("Node ID: %s\n", myID)
		return nil
	},
}
