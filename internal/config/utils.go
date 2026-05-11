package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/libp2p/go-libp2p/core/crypto"
)

const (
	voidlink   = ".voidlink"
	defaultLST = "default.lst"
	sudoEnv    = "SUDO_USER"
	home       = "/home/"
	yamlExt    = ".yaml"
	keyExt     = ".key"
	sockExt    = ".sock"
)

func GetConfigDir() string {
	return filepath.Join(getHomeDir(), voidlink)
}
func GetConfigFilePath(dir string, role string) string {
	return filepath.Join(dir, role+yamlExt)
}
func GetRoutesFilePath(dir string) string {
	return filepath.Join(dir, defaultLST)
}
func GetKeyPath(dir string, role string) string {
	return filepath.Join(dir, role+keyExt)
}
func GetSocketPath(configPath string) string {
	ext := filepath.Ext(configPath)
	pathWithoutExt := strings.TrimSuffix(configPath, ext)
	return pathWithoutExt + sockExt
}
func getHomeDir() string {
	if os.Geteuid() == 0 {
		if sudoUser := os.Getenv(sudoEnv); sudoUser != "" {
			return home + sudoUser
		}
	}
	home, _ := os.UserHomeDir()
	return home
}
func createIdentity(targetPath string) (crypto.PrivKey, error) {
	priv, _, err := crypto.GenerateKeyPair(crypto.Ed25519, -1)
	if err != nil {
		return nil, fmt.Errorf("config: key generation error: %w", err)
	}
	data, err := crypto.MarshalPrivateKey(priv)
	if err != nil {
		return nil, fmt.Errorf("config: failed to marshal private key: %w", err)
	}
	err = os.WriteFile(targetPath, data, 0644)
	if err != nil {
		return nil, fmt.Errorf("config: failed to write file: %w", err)
	}
	return priv, nil
}
