package config

import (
	"fmt"
	"os"

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
	ListenPort    int                 `yaml:"listen_port,omitempty"`
}

const serverLocalIP = "10.1.1.1"
const clientLocalIP = "10.1.1.2"
const ClientInterface = "void1"
const ServerInterface = "void0"

func InitConfig(role string, secret string, dir string, port int) (*Config, crypto.PrivKey, error) {
	path := GetConfigFilePath(dir, role)
	keyPath := GetKeyPath(dir, role)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, nil, fmt.Errorf("config: failed to create config dir: %w", err)
	}
	var cfg *Config
	if _, err := os.Stat(path); os.IsNotExist(err) {
		localIP := clientLocalIP
		if Role(role) == Server {
			localIP = serverLocalIP
		}
		ifaceName := ClientInterface
		if Role(role) == Server {
			ifaceName = ServerInterface
		}
		cfg = &Config{
			ListenPort:    port,
			Role:          Role(role),
			LocalIP:       localIP,
			InterfaceName: ifaceName,
			Whitelist:     map[string]PeerInfo{},
			KeyPath:       keyPath,
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
		priv, err = LoadIdentity(keyPath)
		if err != nil {
			return nil, nil, err
		}
	}
	routesPath := GetRoutesFilePath(dir)
	if _, err := os.Stat(routesPath); os.IsNotExist(err) {
		if err := writeDefaultRoutes(routesPath); err != nil {
			return nil, nil, err
		}
	}
	return cfg, priv, nil
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
		return "", fmt.Errorf("config: failed to get peer id from private key: %w", err)
	}
	return id.String(), nil
}
func LoadIdentity(keypath string) (crypto.PrivKey, error) {
	data, err := os.ReadFile(keypath)
	if err != nil {
		return nil, fmt.Errorf("config: failed to read key: %w", err)
	}
	key, err := crypto.UnmarshalPrivateKey(data)
	if err != nil {
		return nil, fmt.Errorf("config: failed to unmarshal private key: %w", err)
	}
	return key, nil
}
func SaveConfig(configPath string, cfg Config) error {
	yamlData, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("config: failed to marshal config: %w", err)
	}
	err = os.WriteFile(configPath, yamlData, 0644)
	if err != nil {
		return fmt.Errorf("config: failed to write file: %w", err)
	}
	return nil
}
