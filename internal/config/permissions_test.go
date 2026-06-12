package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitConfigCreatesPrivateDirectoryAndKey(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "voidlink")
	if _, _, err := InitConfig(string(Client), "test-secret", dir, 0); err != nil {
		t.Fatal(err)
	}

	assertPermissions(t, dir, privateDirMode)
	assertPermissions(t, GetKeyPath(dir, string(Client)), privateKeyMode)
}

func TestLoadConfigAndIdentityTightenExistingPermissions(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "voidlink")
	if _, _, err := InitConfig(string(Client), "test-secret", dir, 0); err != nil {
		t.Fatal(err)
	}

	keyPath := GetKeyPath(dir, string(Client))
	if err := os.Chmod(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(keyPath, 0644); err != nil {
		t.Fatal(err)
	}

	if _, err := LoadConfig(GetConfigFilePath(dir, string(Client))); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadIdentity(keyPath); err != nil {
		t.Fatal(err)
	}

	assertPermissions(t, dir, privateDirMode)
	assertPermissions(t, keyPath, privateKeyMode)
}

func TestLoadIdentityRejectsSymlink(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "voidlink")
	if _, _, err := InitConfig(string(Client), "test-secret", dir, 0); err != nil {
		t.Fatal(err)
	}

	linkPath := filepath.Join(dir, "linked.key")
	if err := os.Symlink(GetKeyPath(dir, string(Client)), linkPath); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadIdentity(linkPath); err == nil {
		t.Fatal("Expected symlinked private key to be rejected")
	}
}

func assertPermissions(t *testing.T, path string, expected os.FileMode) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != expected {
		t.Fatalf("Expected %s permissions %04o, got %04o", path, expected, got)
	}
}
