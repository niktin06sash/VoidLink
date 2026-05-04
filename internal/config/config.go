package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"gopkg.in/yaml.v3"
)

type Role string

const (
	Server Role = "server"
	Client Role = "client"
)

type PeerInfo struct {
	Name    string `yaml:"name"`
	AddedAt string `yaml:"added_at"`
}
type Config struct {
	Role          Role                `yaml:"role"`
	LocalIP       string              `yaml:"local_ip"`
	InterfaceName string              `yaml:"interface_name"`
	Whitelist     map[string]PeerInfo `yaml:"whitelist"`
	KeyPath       string              `yaml:"key_path"`
	Rendezvous    string              `yaml:"rendezvous"`
}

const serverLocalIP = "10.1.1.1"
const clientLocalIP = "10.1.1.2"
const clientInterface = "void1"
const serverInterface = "void0"

func InitConfig(role string, secret string, path string) (*Config, crypto.PrivKey, error) {
	configDir := filepath.Dir(path)
	keyPath := filepath.Join(configDir, role+".key")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, nil, fmt.Errorf("failed to create config dir: %w", err)
	}
	var cfg *Config
	if _, err := os.Stat(path); os.IsNotExist(err) {
		localIP := clientLocalIP
		if Role(role) == Server {
			localIP = serverLocalIP
		}
		ifaceName := clientInterface
		if Role(role) == Server {
			ifaceName = serverInterface
		}
		cfg = &Config{
			Role:          Role(role),
			LocalIP:       localIP,
			InterfaceName: ifaceName,
			Whitelist:     map[string]PeerInfo{},
			KeyPath:       role + ".key",
			Rendezvous:    secret,
		}
		if err := SaveConfig(path, *cfg); err != nil {
			return nil, nil, err
		}
	} else {
		cfg, err = LoadConfig(path)
		if err != nil {
			return nil, nil, err
		}
	}
	var priv crypto.PrivKey
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		priv, err = createIdentity(keyPath)
		if err != nil {
			return nil, nil, err
		}
	} else {
		priv, err = LoadIdentity(cfg.KeyPath, path)
		if err != nil {
			return nil, nil, err
		}
	}
	return cfg, priv, nil
}

func GetConfigPath(role string) string {
	return filepath.Join(getHomeDir(), ".voidlink", role+".yaml")
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func GetPeerID(priv crypto.PrivKey) (string, error) {
	id, err := peer.IDFromPrivateKey(priv)
	if err != nil {
		return "", fmt.Errorf("failet to get peer id from private key: %w", err)
	}
	return id.String(), nil
}
func LoadIdentity(keypath, targetPath string) (crypto.PrivKey, error) {
	configDir := filepath.Dir(targetPath)
	actualKeyPath := filepath.Join(configDir, keypath)
	data, err := os.ReadFile(actualKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read key: %w", err)
	}
	key, err := crypto.UnmarshalPrivateKey(data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal private key: %w", err)
	}
	return key, nil
}
func SaveConfig(configPath string, cfg Config) error {
	yamlData, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	err = os.WriteFile(configPath, yamlData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}
func getHomeDir() string {
	if os.Geteuid() == 0 {
		if sudoUser := os.Getenv("SUDO_USER"); sudoUser != "" {
			return "/home/" + sudoUser
		}
	}
	home, _ := os.UserHomeDir()
	return home
}
func createIdentity(targetPath string) (crypto.PrivKey, error) {
	priv, _, err := crypto.GenerateKeyPair(crypto.Ed25519, -1)
	if err != nil {
		return nil, fmt.Errorf("key generation error: %w", err)
	}
	data, err := crypto.MarshalPrivateKey(priv)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal private key: %w", err)
	}
	err = os.WriteFile(targetPath, data, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}
	return priv, nil
}

func GetSocketPath(configPath string) string {
	ext := filepath.Ext(configPath)
	pathWithoutExt := strings.TrimSuffix(configPath, ext)
	return pathWithoutExt + ".sock"
}
