package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/niktin06sash/VoidLink/internal/config"
	"github.com/spf13/cobra"
)

var peerCmd = &cobra.Command{
	Use:   "peer [role]",
	Short: "Manage Peer ID in the whitelist in specific role",
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
		return nil
	},
}
var addPeerCmd = &cobra.Command{
	Use:   "add [role] [peerID]",
	Short: "Add a new PeerID to the whitelist in specific role",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(2)(cmd, args); err != nil {
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
		reloadRunningVlink()
		return nil
	}}
var listPeersCmd = &cobra.Command{
	Use:   "list [role]",
	Short: "List all PeerIDs in the whitelist in specific role",
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
		for peerID, peerInfo := range cfg.Whitelist {
			fmt.Printf("%s: %s\n", peerID, peerInfo.Name)
		}
		return nil
	}}
var removePeerCmd = &cobra.Command{
	Use:   "remove [role] [peerID]",
	Short: "Remove a PeerID from the whitelist in specific role",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(2)(cmd, args); err != nil {
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
		peerID := args[1]
		finalPath := cfgFile
		if finalPath == "" {
			finalPath = config.GetConfigPath(role)
		}
		cfg, err := config.LoadConfig(finalPath)
		if err != nil {
			return err
		}
		delete(cfg.Whitelist, peerID)
		if err := config.SaveConfig(finalPath, *cfg); err != nil {
			return err
		}
		fmt.Printf("Successfully removed peer %s from %s\n", peerID, finalPath)
		reloadRunningVlink()
		return nil
	}}

func reloadRunningVlink() {
	out, err := exec.Command("pgrep", "-x", "vlink").Output()
	if err != nil {
		log.Printf("peer: whitelist updated (no reload); pgrep failed err=%v", err)
		return
	}
	self := os.Getpid()
	var signaled int
	for line := range strings.SplitSeq(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		pid, err := strconv.Atoi(strings.TrimSpace(line))
		if err != nil || pid <= 0 || pid == self {
			continue
		}
		if err := syscall.Kill(pid, syscall.SIGHUP); err != nil {
			log.Printf("peer: failed to send SIGHUP pid=%d err=%v", pid, err)
			continue
		}
		signaled++
	}
	if signaled > 0 {
		fmt.Printf("Sent SIGHUP to %d running vlink process(es) to reload whitelist.\n", signaled)
	}
}
