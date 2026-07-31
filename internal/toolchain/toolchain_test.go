package toolchain

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/pspdev-go/pspgo/internal/command"
	"github.com/pspdev-go/pspgo/internal/config"
)

type recordingRunner struct {
	specs []command.Spec
}

func (r *recordingRunner) Run(context.Context, command.Spec) error { return nil }

func (r *recordingRunner) Output(_ context.Context, spec command.Spec) ([]byte, error) {
	r.specs = append(r.specs, spec)
	return []byte("/toolchains/go1.25.9\n"), nil
}

func TestGoEnvironmentResolvesProjectToolchain(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/psp\ngo 1.25.9\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runner := &recordingRunner{}
	env, err := GoEnvironment(context.Background(), runner, config.Config{Root: root, Go: "go"})
	if err != nil {
		t.Fatal(err)
	}
	if len(runner.specs) != 1 {
		t.Fatalf("got %d commands, want 1", len(runner.specs))
	}
	resolve := runner.specs[0]
	if !slices.Contains(resolve.Env, "GOTOOLCHAIN=go1.25.9") {
		t.Fatalf("resolver environment = %v", resolve.Env)
	}
	if resolve.Dir != root {
		t.Fatalf("resolver directory = %q, want %q", resolve.Dir, root)
	}
	for _, want := range []string{
		"GOTOOLCHAIN=local",
		"GOROOT=/toolchains/go1.25.9",
		"PATH=/toolchains/go1.25.9/bin" + string(os.PathListSeparator) + os.Getenv("PATH"),
	} {
		if !slices.Contains(env, want) {
			t.Errorf("environment missing %q: %v", want, env)
		}
	}
}
