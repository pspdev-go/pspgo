// Package sdk locates the pspsdk-go module used by a project.
package sdk

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pspdev-go/pspgo/internal/command"
	"github.com/pspdev-go/pspgo/internal/config"
	"github.com/pspdev-go/pspgo/internal/toolchain"
)

const Module = "github.com/pspdev-go/pspsdk-go"

func Resolve(ctx context.Context, runner command.Runner, cfg config.Config) (string, error) {
	goEnv, err := toolchain.GoEnvironment(ctx, runner, cfg)
	if err != nil {
		return "", fmt.Errorf("configure Go environment: %w", err)
	}
	goRoot := environmentValue(goEnv, "GOROOT")
	goPath := filepath.Join(goRoot, "bin", "go")
	out, err := runner.Output(ctx, command.Spec{
		Path: goPath,
		Args: []string{"list", "-m", "-f", "{{.Dir}}", Module},
		Dir:  cfg.Root,
		Env:  goEnv,
	})
	if err != nil {
		return "", fmt.Errorf("%s is not available from the project module; add it with `go get %s`: %w", Module, Module, err)
	}
	root := strings.TrimSpace(string(out))
	if root == "" {
		return "", fmt.Errorf("%s did not resolve to a local directory", Module)
	}
	if _, err := os.Stat(filepath.Join(root, "bridge", "main.c")); err != nil {
		return "", fmt.Errorf("resolved %s at %s, but bridge/main.c is missing", Module, root)
	}
	return root, nil
}

func environmentValue(env []string, key string) string {
	prefix := key + "="
	for _, item := range env {
		if strings.HasPrefix(item, prefix) {
			return strings.TrimPrefix(item, prefix)
		}
	}
	return ""
}
