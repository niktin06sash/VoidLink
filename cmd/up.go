package cmd

import (
	"github.com/niktin06sash/VoidLink/internal/config"
	"github.com/niktin06sash/VoidLink/internal/node"
	"github.com/niktin06sash/VoidLink/internal/tun"
	"github.com/spf13/cobra"
)

var upCmd = &cobra.Command{
	Use:           "up [role]",
	Short:         "Start VoidLink tunnel using a specific role",
	Args:          cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	ValidArgs:     []string{"server", "client"},
	SilenceErrors: false,
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		roleName := args[0]
		finalPath := cfgFile
		if finalPath == "" {
			finalPath = config.GetConfigPath(roleName)
		}
		cfg, err := config.LoadConfig(finalPath)
		if err != nil {
			return err
		}
		priv, err := config.LoadIdentity(cfg.KeyPath, finalPath)
		if err != nil {
			return err
		}
		tunnel, err := tun.NewTun(cfg)
		if err != nil {
			return err
		}
		defer tunnel.Close()
		noda, err := node.NewNode(cmd.Context(), cfg, tunnel, priv, finalPath)
		if err != nil {
			return err
		}
		defer noda.Close()
		noda.WatchSignal()
		err = noda.Run()
		if err != nil {
			return err
		}
		return nil
	},
}
