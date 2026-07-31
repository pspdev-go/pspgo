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
	text := `title = "Cube"
output = "cube"
kernel_mode = true
icon = "assets/ICON0.PNG"
animation = "assets/ICON1.PMF"
preview = "assets/PIC0.PNG"
background = "assets/PIC1.PNG"
music = "assets/SND0.AT3"
`
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
	for name, got := range map[string]string{
		"icon": cfg.Icon, "animation": cfg.Animation, "preview": cfg.Preview,
		"background": cfg.Background, "music": cfg.Music,
	} {
		want := filepath.Join(dir, "assets", map[string]string{
			"icon": "ICON0.PNG", "animation": "ICON1.PMF", "preview": "PIC0.PNG",
			"background": "PIC1.PNG", "music": "SND0.AT3",
		}[name])
		if got != want {
			t.Errorf("%s = %q, want %q", name, got, want)
		}
	}
}
