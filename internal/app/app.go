package app

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pspdev-go/pspgo/internal/build"
	"github.com/pspdev-go/pspgo/internal/command"
	"github.com/pspdev-go/pspgo/internal/config"
	psptarget "github.com/pspdev-go/pspgo/internal/target"
	"github.com/pspdev-go/pspgo/internal/toolchain"
)

func Run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return usage(stderr)
	}
	name, rest := args[0], args[1:]
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	verbose := fs.Bool("v", false, "print external commands")
	root := fs.String("C", ".", "project directory")
	if err := fs.Parse(rest); err != nil {
		return err
	}
	cfg, err := config.Load(*root)
	if err != nil {
		return err
	}
	runner := command.ExecRunner{Stdout: stdout, Stderr: stderr, Verbose: *verbose}
	report := toolchain.Inspect(ctx, runner, cfg)
	switch name {
	case "env":
		printReport(stdout, cfg, report)
		return nil
	case "doctor":
		printReport(stdout, cfg, report)
		if err := report.Validate(cfg); err != nil {
			return err
		}
		fmt.Fprintln(stdout, "toolchain: OK")
		return nil
	case "build", "run", "test":
		if err := report.Validate(cfg); err != nil {
			return err
		}
		if name == "test" {
			cfg.Package = packageArg(fs.Args(), "./...")
		} else {
			cfg.Package = packageArg(fs.Args(), cfg.Package)
		}
		pbp, err := (build.Builder{Config: cfg, Runner: runner, PSPDEV: report.PSPDEV}).Build(ctx)
		if err != nil {
			return err
		}
		fmt.Fprintln(stdout, "built", pbp)
		if name == "run" {
			if err := runner.Run(ctx, command.Spec{Path: cfg.PPSSPP, Args: []string{pbp}, Dir: cfg.Root}); err != nil {
				return fmt.Errorf("run stage: %w", err)
			}
		}
		return nil
	case "clean":
		return clean(cfg)
	default:
		return usage(stderr)
	}
}

func usage(w io.Writer) error {
	fmt.Fprintln(w, "usage: pspgo <build|run|test|env|doctor|clean> [-v] [-C dir] [package]")
	return fmt.Errorf("invalid or missing command")
}
func packageArg(args []string, fallback string) string {
	if len(args) > 0 {
		return args[0]
	}
	return fallback
}
func printReport(w io.Writer, cfg config.Config, report toolchain.Report) {
	target := cfg.Target
	if target == "" {
		target = psptarget.Default + " (TinyGo built-in)"
	}
	fmt.Fprintf(w, "root: %s\nSDK root: %s\ntarget: %s\nbuild dir: %s\nPSPDEV: %s\n", cfg.Root, cfg.SDKRoot, target, cfg.BuildDir, report.PSPDEV)
	for _, item := range report.Items {
		if item.Err != nil {
			fmt.Fprintf(w, "%s: ERROR: %v\n", item.Name, item.Err)
		} else {
			fmt.Fprintf(w, "%s: %s (%s)\n", item.Name, item.Version, item.Path)
		}
	}
}
func clean(cfg config.Config) error {
	path, err := filepath.Abs(cfg.BuildDir)
	if err != nil {
		return err
	}
	root, err := filepath.Abs(cfg.Root)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(root, path)
	if err != nil || rel == "." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return fmt.Errorf("refusing to clean build directory outside project: %s", path)
	}
	if err := os.RemoveAll(path); err != nil {
		return err
	}
	fmt.Println("removed", path)
	return nil
}
