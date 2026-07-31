package sdk

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/pspdev-go/pspgo/internal/command"
	"github.com/pspdev-go/pspgo/internal/config"
)

type fakeRunner struct {
	sdk string
}

func (f fakeRunner) Run(context.Context, command.Spec) error { return nil }

func (f fakeRunner) Output(_ context.Context, spec command.Spec) ([]byte, error) {
	if len(spec.Args) >= 2 && spec.Args[0] == "env" && spec.Args[1] == "GOROOT" {
		return []byte("/toolchains/go1.25.9\n"), nil
	}
	return json.Marshal(struct {
		Dir string
	}{Dir: f.sdk})
}

func TestResolve(t *testing.T) {
	root := t.TempDir()
	sdkRoot := filepath.Join(root, "module-cache", "pspsdk-go")
	if err := os.MkdirAll(filepath.Join(sdkRoot, "bridge"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sdkRoot, "bridge", "main.c"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	project := filepath.Join(root, "project")
	if err := os.MkdirAll(project, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, "go.mod"), []byte("module example.com/app\ngo 1.25.9\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := Resolve(context.Background(), fakeRunner{sdk: sdkRoot}, config.Config{Root: project, Go: "go"})
	if err != nil {
		t.Fatal(err)
	}
	if got != sdkRoot {
		t.Fatalf("Resolve() = %q, want %q", got, sdkRoot)
	}
}
