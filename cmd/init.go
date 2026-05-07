package cmd

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"slices"

	"github.com/niktin06sash/VoidLink/internal/config"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [role]",
	Short: "Initialize VoidLink configuration",
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
		userRole := args[0]
		if secretPhrase == "" {
			s, err := generateRandomString(12)
			if err != nil {
				return fmt.Errorf("failed to generate rendezvous secret: %w", err)
			}
			secretPhrase = "vlink-" + s
		}
		finalPath := cfgFile
		if finalPath == "" {
			finalPath = config.GetConfigPath(userRole)
		}
		log.Printf("init: role=%s config=%s", userRole, finalPath)
		cfg, priv, err := config.InitConfig(userRole, secretPhrase, finalPath, port)
		if err != nil {
			return err
		}
		id, err := config.GetPeerID(priv)
		if err != nil {
			return err
		}
		fmt.Printf("Success! Config created at %s\n", finalPath)
		fmt.Printf("Your Peer ID: %s\n", id)
		fmt.Printf("Local IP set to: %s\n", cfg.LocalIP)
		fmt.Printf("Rendezvous: %s\n", cfg.Rendezvous)
		return nil
	},
}

func generateRandomString(n int) (string, error) {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	ret := make([]byte, n)
	for i := range ret {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		ret[i] = letters[num.Int64()]
	}
	return string(ret), nil
}
