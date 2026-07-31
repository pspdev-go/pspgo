// Package sdk locates the pspsdk-go module used by a project.
package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

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
		Args: []string{"mod", "download", "-json", Module},
		Dir:  cfg.Root,
		Env:  goEnv,
	})
	if err != nil {
		return "", fmt.Errorf("%s is not available from the project module; add it with `go get %s`: %w", Module, Module, err)
	}
	var downloaded struct {
		Dir   string
		Error string
	}
	if err := json.Unmarshal(out, &downloaded); err != nil {
		return "", fmt.Errorf("decode module location: %w", err)
	}
	if downloaded.Error != "" {
		return "", fmt.Errorf("download %s: %s", Module, downloaded.Error)
	}
	root := downloaded.Dir
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
		if len(item) >= len(prefix) && item[:len(prefix)] == prefix {
			return item[len(prefix):]
		}
	}
	return ""
}
