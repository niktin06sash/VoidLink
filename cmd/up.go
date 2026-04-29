package cmd

import (
	"github.com/niktin06sash/VoidLink/internal/config"
	"github.com/niktin06sash/VoidLink/internal/node"
	"github.com/niktin06sash/VoidLink/internal/tun"
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
		tunnel, err := tun.NewTun(cfg)
		if err != nil {
			return err
		}
		defer tunnel.Close()
		ctx := cmd.Context()
		noda, err := node.NewNode(ctx, cfg, tunnel, priv)
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
