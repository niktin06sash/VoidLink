package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "vlink",
	SilenceErrors: false,
	SilenceUsage:  true,
}

func Execute() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}

var secretPhrase string
var cfgFile string
var peerName string

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(upCmd)
	rootCmd.AddCommand(addPeerCmd)
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "path to config file (default is $HOME/.voidlink/role.yaml)")
	addPeerCmd.Flags().StringVarP(&peerName, "name", "n", "new-device", "Friendly name for the peer")
	initCmd.Flags().StringVarP(&secretPhrase, "secret", "s", "", "secret phrase for rendezvous (leave empty for random)")
}
