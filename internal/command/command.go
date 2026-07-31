package command

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type Spec struct {
	Path string
	Args []string
	Dir  string
	Env  []string
}

func (s Spec) String() string {
	parts := append([]string{s.Path}, s.Args...)
	for i, part := range parts {
		if strings.ContainsAny(part, " \t\"'") {
			parts[i] = fmt.Sprintf("%q", part)
		}
	}
	return strings.Join(parts, " ")
}

type Runner interface {
	Run(context.Context, Spec) error
	Output(context.Context, Spec) ([]byte, error)
}

type ExecRunner struct {
	Stdout  io.Writer
	Stderr  io.Writer
	Verbose bool
}

// OSRunner is kept as the public name used by integration fixtures.
type OSRunner = ExecRunner

func (r ExecRunner) command(ctx context.Context, s Spec) *exec.Cmd {
	cmd := exec.CommandContext(ctx, s.Path, s.Args...)
	cmd.Dir = s.Dir
	cmd.Env = append(os.Environ(), s.Env...)
	return cmd
}

func (r ExecRunner) Run(ctx context.Context, s Spec) error {
	if r.Verbose {
		fmt.Fprintln(r.Stderr, "+", s.String())
	}
	cmd := r.command(ctx, s)
	cmd.Stdout, cmd.Stderr = r.Stdout, r.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %w", s.Path, err)
	}
	return nil
}

func (r ExecRunner) Output(ctx context.Context, s Spec) ([]byte, error) {
	if r.Verbose {
		fmt.Fprintln(r.Stderr, "+", s.String())
	}
	cmd := r.command(ctx, s)
	cmd.Stderr = r.Stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", s.Path, err)
	}
	return out, nil
}
