package cmd

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/niktin06sash/VoidLink/internal/config"
	"github.com/spf13/cobra"
)

var routeCmd = &cobra.Command{
	Use:   "route",
	Short: "Manage routes",
}
var addRouteCmd = &cobra.Command{
	Use:   "add [cidr]",
	Short: "Add a route to default.lst",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cidr := args[0]
		finalDir := cfgFile
		if finalDir == "" {
			finalDir = config.GetConfigDir()
		}
		routesPath := config.GetRoutesFilePath(finalDir)
		routes, err := config.ReadRoutes(routesPath)
		if err != nil {
			return err
		}
		if slices.Contains(routes, cidr) {
			return fmt.Errorf("route %s already exists", cidr)
		}
		f, err := os.OpenFile(routesPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open routes file: %w", err)
		}
		defer f.Close()
		_, err = fmt.Fprintln(f, cidr)
		if err != nil {
			return fmt.Errorf("failed to write route: %w", err)
		}
		fmt.Printf("Successfully added route %s\n", cidr)
		return nil
	},
}

var removeRouteCmd = &cobra.Command{
	Use:   "remove [cidr]",
	Short: "Remove a route from default.lst",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cidr := args[0]
		finalDir := cfgFile
		if finalDir == "" {
			finalDir = config.GetConfigDir()
		}
		routesPath := config.GetRoutesFilePath(finalDir)
		routes, err := config.ReadRoutes(routesPath)
		if err != nil {
			return err
		}
		var newRoutes []string
		found := false
		for _, r := range routes {
			if r == cidr {
				found = true
				continue
			}
			newRoutes = append(newRoutes, r)
		}
		if !found {
			return fmt.Errorf("route %s not found", cidr)
		}
		content := strings.Join(newRoutes, "\n") + "\n"
		if err := os.WriteFile(routesPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write routes file: %w", err)
		}
		fmt.Printf("Successfully removed route %s\n", cidr)
		return nil
	},
}

var listRoutesCmd = &cobra.Command{
	Use:   "list",
	Short: "List all routes in default.lst",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		finalDir := cfgFile
		if finalDir == "" {
			finalDir = config.GetConfigDir()
		}
		routesPath := config.GetRoutesFilePath(finalDir)
		routes, err := config.ReadRoutes(routesPath)
		if err != nil {
			return err
		}
		if len(routes) == 0 {
			fmt.Println("No routes found")
			return nil
		}
		for _, r := range routes {
			fmt.Println(r)
		}
		return nil
	},
}
