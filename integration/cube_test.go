//go:build integration

package integration

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/pspdev-go/pspgo/internal/build"
	"github.com/pspdev-go/pspgo/internal/command"
	"github.com/pspdev-go/pspgo/internal/config"
)

// TestRotatingCube builds pspsdk-go/example/main.go through the complete PSP
// toolchain. It is opt-in because it requires TinyGo and a PSPSDK installation.
func TestRotatingCube(t *testing.T) {
	sdk := os.Getenv("PSPGO_SDK")
	if sdk == "" {
		sdk = filepath.Clean(filepath.Join("..", "..", "pspsdk-go"))
	}
	sdk, err := filepath.Abs(sdk)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(sdk, "example", "main.go")); err != nil {
		t.Skipf("pspsdk-go rotating cube not available: %v", err)
	}
	cfg := config.Config{
		ProjectDir: sdk, Package: "./example", SDKRoot: sdk, BuildDir: t.TempDir(),
		Title: "pspgo integration cube", Go: env("PSPGO_GO", "go"),
		TinyGo: env("PSPGO_TINYGO", "tinygo"), PSPCMake: env("PSPGO_PSP_CMAKE", "psp-cmake"),
		CMake: env("PSPGO_CMAKE", "cmake"), NM: filepath.Join(os.Getenv("PSPDEV"), "bin", "psp-nm"),
	}
	runner := command.OSRunner{Stdout: io.Discard, Stderr: os.Stderr}
	pbp, err := (build.Builder{Config: cfg, Runner: runner}).Build(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(pbp)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() < 1024 {
		t.Fatalf("EBOOT.PBP is unexpectedly small: %d bytes", info.Size())
	}
}

func env(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
