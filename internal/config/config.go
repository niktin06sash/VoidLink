package config

import (
	"fmt"
	"os"
	"path/filepath"

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

func InitConfig(role string, secret string) (string, *Config, crypto.PrivKey, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to find home directory: %w", err)
	}
	configDir := filepath.Join(home, ".voidlink")
	keyPath := filepath.Join(configDir, role+".key")
	configPath := filepath.Join(configDir, role+".yaml")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", nil, nil, fmt.Errorf("failed to create config dir: %w", err)
	}
	var priv crypto.PrivKey
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		priv, err = createIdentity(keyPath)
		if err != nil {
			return "", nil, nil, err
		}
	} else {
		priv, err = LoadIdentity(keyPath)
		if err != nil {
			return "", nil, nil, err
		}
	}
	var cfg *Config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		localIP := "10.1.1.2"
		if Role(role) == Server {
			localIP = "10.1.1.1"
		}

		ifaceName := "void1"
		if Role(role) == Server {
			ifaceName = "void0"
		}
		cfg = &Config{
			Role:          Role(role),
			LocalIP:       localIP,
			InterfaceName: ifaceName,
			Whitelist:     map[string]PeerInfo{},
			KeyPath:       keyPath,
			Rendezvous:    secret,
		}
		if err := saveConfig(configPath, *cfg); err != nil {
			return "", nil, nil, err
		}
	} else {
		cfg, err = LoadConfig(role)
		if err != nil {
			return "", nil, nil, err
		}
	}
	return configPath, cfg, priv, nil
}

func LoadConfig(profileName string) (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home dir: %w", err)
	}
	path := filepath.Join(home, ".voidlink", profileName+".yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read config file: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("could not parse yaml: %w", err)
	}
	return &cfg, nil
}

func GetPeerID(priv crypto.PrivKey) (string, error) {
	id, err := peer.IDFromPrivateKey(priv)
	if err != nil {
		return "", fmt.Errorf("failet to get id from private key: %w", err)
	}
	return id.String(), nil
}
func LoadIdentity(targetPath string) (crypto.PrivKey, error) {
	data, err := os.ReadFile(targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read key: %w", err)
	}
	key, err := crypto.UnmarshalPrivateKey(data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal private key: %w", err)
	}
	return key, nil
}

func saveConfig(configPath string, cfg Config) error {
	yamlData, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	err = os.WriteFile(configPath, yamlData, 0600)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
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
	err = os.WriteFile(targetPath, data, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}
	return priv, nil
}
