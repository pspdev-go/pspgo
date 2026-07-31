package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadWithoutConfig(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Root != dir || cfg.Package != "." {
		t.Fatalf("config = %#v", cfg)
	}
}

func TestLoadConfig(t *testing.T) {
	dir := t.TempDir()
	text := "title = \"Cube\"\noutput = \"cube\"\nkernel_mode = true\nsdk_root = \"sdk\"\n"
	if err := os.WriteFile(filepath.Join(dir, "pspgo.toml"), []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Title != "Cube" || cfg.OutputName != "cube" || !cfg.KernelMode {
		t.Fatalf("config = %#v", cfg)
	}
	if cfg.SDKRoot != filepath.Join(dir, "sdk") {
		t.Fatalf("sdk root = %s", cfg.SDKRoot)
	}
}
