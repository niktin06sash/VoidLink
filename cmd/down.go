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

	"github.com/niktin06sash/VoidLink/internal/config"
	"github.com/niktin06sash/VoidLink/internal/tun"
	"github.com/spf13/cobra"
)

var downCmd = &cobra.Command{
	Use:   "down [role]",
	Short: "Stop running VoidLink tunnel using a specific role",
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
		if forceDown {
			return forceCleanup(role, serverAddress)
		}
		out, err := exec.Command("pgrep", "-f", "vlink up "+role).Output()
		if err != nil {
			return fmt.Errorf("down: no running vlink %s process found", role)
		}
		var stopped int
		self := os.Getpid()
		for line := range strings.SplitSeq(strings.TrimSpace(string(out)), "\n") {
			if line == "" {
				continue
			}
			pid, err := strconv.Atoi(strings.TrimSpace(line))
			if err != nil || pid <= 0 || pid == self {
				continue
			}
			if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
				log.Printf("down: failed to stop pid=%d err=%v", pid, err)
				continue
			}
			stopped++
		}
		if stopped == 0 {
			return fmt.Errorf("down: no running vlink %s process found", role)
		}
		fmt.Printf("Stopped %d vlink %s process(es)\n", stopped, role)
		return nil
	},
}

func forceCleanup(role string, serveraddress string) error {
	iface := config.ClientInterface
	if config.Role(role) == config.Server {
		iface = config.ServerInterface
	}
	log.Printf("down: force cleanup interface=%s", iface)
	err := tun.ForceCleanup(iface, serveraddress)
	if err != nil {
		return err
	}
	return nil
}
