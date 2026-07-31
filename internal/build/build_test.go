package build

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/pspdev-go/pspgo/internal/command"
	"github.com/pspdev-go/pspgo/internal/config"
)

type fakeRunner struct {
	specs []command.Spec
	lib   string
}

func (f *fakeRunner) Run(_ context.Context, spec command.Spec) error {
	f.specs = append(f.specs, spec)
	if spec.Path == "tinygo" {
		for i, arg := range spec.Args {
			if arg == "-o" {
				return os.WriteFile(spec.Args[i+1], nil, 0o644)
			}
		}
	}
	if spec.Path == "cmake" {
		if err := os.MkdirAll(spec.Args[1], 0o755); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(spec.Args[1], "EBOOT.PBP"), []byte("pbp"), 0o644)
	}
	return nil
}

func (f *fakeRunner) Output(_ context.Context, spec command.Spec) ([]byte, error) {
	f.specs = append(f.specs, spec)
	switch {
	case spec.Path == "go" && spec.Args[0] == "version":
		return []byte("go version go1.25.9 darwin/arm64\n"), nil
	case spec.Path == "go":
		return []byte("/fake/goroot\n"), nil
	case spec.Path == "tinygo":
		return []byte("tinygo version 0.41.0 (using go version go1.25.9)\n"), nil
	case len(spec.Args) > 0 && spec.Args[0] == "-u":
		return []byte("         U pspsdk_go_gum_draw_array_3d\n"), nil
	default:
		return []byte(f.lib + ":gum.o:00000000 T sceGumDrawArray\n"), nil
	}
}

func TestBuildPipelineCommandsAndBridgeSelection(t *testing.T) {
	root := t.TempDir()
	sdk, pspdev := filepath.Join(root, "sdk"), filepath.Join(root, "pspdev")
	libDir := filepath.Join(pspdev, "psp", "sdk", "lib")
	for _, dir := range []string{filepath.Join(sdk, "bridge"), libDir, filepath.Join(pspdev, "bin")} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for name, body := range map[string]string{"go.mod": "module github.com/pspdev-go/pspsdk-go\ngo 1.25.9\n", "psp.json": "{}"} {
		if err := os.WriteFile(filepath.Join(sdk, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	nm := filepath.Join(pspdev, "bin", "psp-nm")
	if err := os.WriteFile(nm, nil, 0o755); err != nil {
		t.Fatal(err)
	}
	lib := filepath.Join(libDir, "libpspgum.a")
	if err := os.WriteFile(lib, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PSPDEV", pspdev)
	fake := &fakeRunner{lib: lib}
	cfg := config.Config{
		ProjectDir: root, Package: "./example", SDKRoot: sdk, BuildDir: filepath.Join(root, "build"),
		Title: "Cube", Go: "go", TinyGo: "tinygo", PSPCMake: "psp-cmake", CMake: "cmake", NM: nm,
	}
	pbp, err := (Builder{Config: cfg, Runner: fake}).Build(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(pbp); err != nil {
		t.Fatal(err)
	}
	cmake, err := os.ReadFile(filepath.Join(cfg.BuildDir, "cmake", "CMakeLists.txt"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(cmake)
	for _, want := range []string{"gum_abi.c", "pspgum", "create_pbp_file"} {
		if !strings.Contains(text, want) {
			t.Errorf("generated CMake missing %q:\n%s", want, text)
		}
	}
	for _, spec := range fake.specs {
		if spec.Path == "tinygo" && spec.Args[0] == "build" && !containsArg(spec.Args, "-no-debug") {
			t.Error("TinyGo command does not suppress host debug paths")
		}
	}
}

func containsArg(args []string, wanted string) bool {
	return slices.Contains(args, wanted)
}
