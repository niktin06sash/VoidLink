package cmd

import (
	"fmt"

	"github.com/niktin06sash/VoidLink/internal/config"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:           "init [role]",
	Short:         "Initialize VoidLink configuration",
	Args:          cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	ValidArgs:     []string{"server", "client"},
	SilenceErrors: false,
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		userRole := args[0]
		path, cfg, priv, err := config.InitConfig(userRole)
		if err != nil {
			return err
		}
		id, err := config.GetPeerID(priv)
		if err != nil {
			return err
		}
		fmt.Printf("Success! Config created at %s\n", path)
		fmt.Printf("Your Peer ID: %s\n", id)
		fmt.Printf("Local IP set to: %s\n", cfg.LocalIP)
		return nil
	},
}
