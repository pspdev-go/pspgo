package build

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pspdev-go/pspgo/internal/command"
	"github.com/pspdev-go/pspgo/internal/config"
	"github.com/pspdev-go/pspgo/internal/resolver"
	"github.com/pspdev-go/pspgo/internal/target"
	"github.com/pspdev-go/pspgo/internal/toolchain"
)

type Builder struct {
	Config config.Config
	Runner command.Runner
	PSPDEV string
}

func (b Builder) Build(ctx context.Context) (string, error) {
	cfg := b.Config
	if cfg.Root == "" {
		cfg.Root = cfg.ProjectDir
	}
	if cfg.PSPNM == "" {
		cfg.PSPNM = cfg.NM
	}
	if cfg.OutputName == "" {
		cfg.OutputName = "app"
	}
	if cfg.Title == "" {
		cfg.Title = "PSP Go Application"
	}
	if cfg.OutputName == "" {
		cfg.OutputName = "app"
	}
	cfg.Target = target.Resolve(cfg.Target)
	if cfg.Package == "" {
		cfg.Package = "."
	}
	if b.PSPDEV == "" {
		b.PSPDEV = os.Getenv("PSPDEV")
	}
	b.Config = cfg
	objectDir := filepath.Join(cfg.BuildDir, "obj")
	cmakeDir := filepath.Join(cfg.BuildDir, "cmake")
	if err := os.MkdirAll(objectDir, 0o755); err != nil {
		return "", err
	}
	if err := os.MkdirAll(cmakeDir, 0o755); err != nil {
		return "", err
	}
	b.Config = cfg
	object := filepath.Join(objectDir, "go.o")
	env, err := toolchain.GoEnvironment(ctx, b.Runner, cfg)
	if err != nil {
		return "", fmt.Errorf("compile stage: configure Go environment: %w", err)
	}
	env = append(env, "GOMIPS=softfloat")
	tinyArgs := []string{"build", "-no-debug", "-scheduler=none", "-gc=psp", "-target", cfg.Target, "-o", object, cfg.Package}
	if err := b.Runner.Run(ctx, command.Spec{Path: cfg.TinyGo, Args: tinyArgs, Dir: cfg.Root, Env: env}); err != nil {
		return "", fmt.Errorf("compile stage: %w", err)
	}
	resolved, err := resolver.Resolve(ctx, b.Runner, cfg.PSPNM, b.PSPDEV, []string{object})
	if err != nil {
		return "", fmt.Errorf("dependency stage: %w", err)
	}
	cmakePath := filepath.Join(cmakeDir, "CMakeLists.txt")
	content, err := b.cmake(object, resolved)
	if err != nil {
		return "", err
	}
	if err := writeIfChanged(cmakePath, []byte(content)); err != nil {
		return "", err
	}
	args := []string{"-S", cmakeDir, "-B", cmakeDir, "-DPSP_KERNEL_MODE=" + onOff(cfg.KernelMode)}
	if err := b.Runner.Run(ctx, command.Spec{Path: cfg.PSPCMake, Args: args, Dir: cfg.Root}); err != nil {
		return "", fmt.Errorf("configure stage: %w", err)
	}
	if err := b.Runner.Run(ctx, command.Spec{Path: cfg.CMake, Args: []string{"--build", cmakeDir}, Dir: cfg.Root}); err != nil {
		return "", fmt.Errorf("link/package stage: %w", err)
	}
	pbp := filepath.Join(cmakeDir, "EBOOT.PBP")
	if _, err := os.Stat(pbp); err != nil {
		return "", fmt.Errorf("package stage did not produce %s", pbp)
	}
	return pbp, nil
}

func (b Builder) cmake(object string, result resolver.Result) (string, error) {
	cfg := b.Config
	mainSource := filepath.Join(cfg.SDKRoot, "bridge", "main.c")
	markers := filepath.Join(cfg.SDKRoot, "bridge", "library_markers.c")
	sources := []string{mainSource}
	if _, err := os.Stat(markers); err == nil {
		sources = append(sources, markers)
	}
	for _, relative := range result.Bridges {
		path := filepath.Join(cfg.SDKRoot, filepath.FromSlash(relative))
		sources = append(sources, path)
	}
	sort.Strings(sources)
	var sourceLines, libraryLines strings.Builder
	for _, source := range sources {
		fmt.Fprintf(&sourceLines, "  %s\n", cmakeQuote(source))
	}
	for _, library := range result.Libraries {
		fmt.Fprintf(&libraryLines, "  %s\n", library)
	}
	return fmt.Sprintf(`cmake_minimum_required(VERSION 3.16)
include("$ENV{PSPDEV}/psp/share/pspdev.cmake")
project(pspgo_generated C)
option(PSP_KERNEL_MODE "Build with kernel module attributes" OFF)
add_executable(%s
%s  %s
)
set_source_files_properties(%s PROPERTIES EXTERNAL_OBJECT TRUE GENERATED TRUE)
target_include_directories(%s PRIVATE %s)
target_link_options(%s PRIVATE
  "LINKER:--defsym,_globals_start=_fdata"
  "LINKER:--defsym,_globals_end=_end"
  "LINKER:--defsym,_stack_top=0x0A000000"
)
if(PSP_KERNEL_MODE)
  target_compile_definitions(%s PRIVATE PSPSDK_GO_KERNEL_MODE=1)
endif()
target_link_libraries(%s
  -Wl,--start-group
  m
  c
%s  -Wl,--end-group
)
create_pbp_file(TARGET %s TITLE %s)
`, cfg.OutputName, sourceLines.String(), cmakeQuote(object), cmakeQuote(object),
		cfg.OutputName, cmakeQuote(cfg.SDKRoot), cfg.OutputName, cfg.OutputName, cfg.OutputName,
		libraryLines.String(), cfg.OutputName, cmakeQuote(cfg.Title)), nil
}

func writeIfChanged(path string, data []byte) error {
	sum := sha256.Sum256(data)
	stamp := path + ".sha256"
	if old, err := os.ReadFile(stamp); err == nil && strings.TrimSpace(string(old)) == hex.EncodeToString(sum[:]) {
		return nil
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return err
	}
	return os.WriteFile(stamp, []byte(hex.EncodeToString(sum[:])+"\n"), 0o644)
}
func cmakeQuote(value string) string { return `"` + strings.ReplaceAll(value, `"`, `\"`) + `"` }
func onOff(value bool) string {
	if value {
		return "ON"
	}
	return "OFF"
}
