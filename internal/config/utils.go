package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/libp2p/go-libp2p/core/crypto"
	"golang.org/x/sys/unix"
)

const (
	voidlink       = ".voidlink"
	defaultLST     = "default.lst"
	sudoEnv        = "SUDO_USER"
	home           = "/home/"
	yamlExt        = ".yaml"
	keyExt         = ".key"
	sockExt        = ".sock"
	privateDirMode = 0700
	privateKeyMode = 0600
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

func secureConfigDir(dir string, create bool) error {
	if create {
		if err := os.MkdirAll(dir, privateDirMode); err != nil {
			return fmt.Errorf("config: failed to create config dir: %w", err)
		}
	}
	info, err := os.Lstat(dir)
	if err != nil {
		return fmt.Errorf("config: failed to inspect config dir: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return fmt.Errorf("config: config path is not a directory: %s", dir)
	}
	if !isAllowedOwner(info) {
		return fmt.Errorf("config: config dir has unexpected owner: %s", dir)
	}
	if err := os.Chmod(dir, privateDirMode); err != nil {
		return fmt.Errorf("config: failed to secure config dir: %w", err)
	}
	return nil
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
	file, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL|unix.O_NOFOLLOW, privateKeyMode)
	if err != nil {
		return nil, fmt.Errorf("config: failed to create key file: %w", err)
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("config: failed to write key file: %w", err)
	}
	if err := file.Close(); err != nil {
		return nil, fmt.Errorf("config: failed to close key file: %w", err)
	}
	return priv, nil
}

func readPrivateKeyFile(path string) ([]byte, error) {
	file, err := os.OpenFile(path, os.O_RDONLY|unix.O_NOFOLLOW, 0)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("key path is not a regular file: %s", path)
	}
	if !isAllowedOwner(info) {
		return nil, fmt.Errorf("key file has unexpected owner: %s", path)
	}
	if err := file.Chmod(privateKeyMode); err != nil {
		return nil, fmt.Errorf("failed to secure key permissions: %w", err)
	}
	return io.ReadAll(file)
}

func isAllowedOwner(info os.FileInfo) bool {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return false
	}
	if stat.Uid == uint32(os.Geteuid()) {
		return true
	}
	if os.Geteuid() != 0 {
		return false
	}
	sudoUID, err := strconv.ParseUint(os.Getenv("SUDO_UID"), 10, 32)
	return err == nil && stat.Uid == uint32(sudoUID)
}
