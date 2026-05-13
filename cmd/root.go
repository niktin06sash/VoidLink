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
var serverAddress string
var port int
var routeAll bool
var routeSplit bool
var forceDown bool

func init() {
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(peerCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(upCmd)
	rootCmd.AddCommand(downCmd)
	peerCmd.AddCommand(addPeerCmd)
	peerCmd.AddCommand(listPeersCmd)
	peerCmd.AddCommand(removePeerCmd)
	rootCmd.AddCommand(routeCmd)
	routeCmd.AddCommand(addRouteCmd)
	routeCmd.AddCommand(removeRouteCmd)
	routeCmd.AddCommand(listRoutesCmd)
	downCmd.Flags().StringVar(&serverAddress, "server-address", "", "flag to direct connect to server")
	downCmd.Flags().BoolVar(&forceDown, "force", false, "force cleanup routes and interface without stopping process")
	upCmd.Flags().BoolVar(&routeSplit, "route-split", false, "route split resources through VoidLink")
	upCmd.Flags().BoolVar(&routeAll, "route-all", false, "route all traffic through VoidLink")
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "path to config directory (default is $HOME/.voidlink)")
	addPeerCmd.Flags().StringVarP(&peerName, "name", "n", "new-device", "friendly name for the peer")
	upCmd.Flags().StringVar(&serverAddress, "server-address", "", "flag to direct connect to server")
	initCmd.Flags().IntVar(&port, "port", 0, "listening port for p2p connections (default random)")
	initCmd.Flags().StringVarP(&secretPhrase, "secret", "s", "", "secret phrase for rendezvous (leave empty for random)")
}
