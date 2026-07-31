package toolchain

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pspdev-go/pspgo/internal/command"
	"github.com/pspdev-go/pspgo/internal/config"
)

type Report struct {
	PSPDEV string
	Items  []Item
}
type Item struct {
	Name, Path, Version string
	Err                 error
}

func Inspect(ctx context.Context, runner command.Runner, cfg config.Config) Report {
	report := Report{PSPDEV: os.Getenv("PSPDEV")}
	goEnv, goEnvErr := GoEnvironment(ctx, runner, cfg)
	tools := []struct {
		name, path string
		args       []string
	}{
		{"tinygo", cfg.TinyGo, []string{"version"}},
		{"psp-cmake", cfg.PSPCMake, []string{"--version"}}, {"psp-nm", cfg.PSPNM, []string{"--version"}},
		{"cmake", cfg.CMake, []string{"--version"}}, {"make", cfg.Make, []string{"--version"}},
	}
	goItem := Item{Name: "go", Err: goEnvErr}
	if goEnvErr == nil {
		goItem.Path = filepath.Join(environmentValue(goEnv, "GOROOT"), "bin", "go")
		out, err := runner.Output(ctx, command.Spec{Path: goItem.Path, Args: []string{"version"}, Env: goEnv})
		goItem.Err = err
		goItem.Version = firstLine(string(out))
	}
	report.Items = append(report.Items, goItem)
	for _, tool := range tools {
		path, err := exec.LookPath(tool.path)
		item := Item{Name: tool.name, Path: path, Err: err}
		if err == nil {
			var env []string
			if tool.name == "tinygo" {
				env = goEnv
			}
			out, outputErr := runner.Output(ctx, command.Spec{Path: path, Args: tool.args, Env: env})
			item.Err = outputErr
			item.Version = firstLine(string(out))
		}
		report.Items = append(report.Items, item)
	}
	if report.PSPDEV == "" {
		report.Items = append(report.Items, Item{Name: "PSPDEV", Err: fmt.Errorf("PSPDEV is not set")})
	} else if _, err := os.Stat(filepath.Join(report.PSPDEV, "psp", "sdk", "lib")); err != nil {
		report.Items = append(report.Items, Item{Name: "PSPDEV", Path: report.PSPDEV, Err: fmt.Errorf("PSPSDK library directory is missing")})
	}
	return report
}

// GoEnvironment pins TinyGo to the configured Go installation. TinyGo finds
// its host Go through PATH and GOROOT, not through PSPGO_GO directly.
func GoEnvironment(ctx context.Context, runner command.Runner, cfg config.Config) ([]string, error) {
	goPath, err := exec.LookPath(cfg.Go)
	if err != nil {
		return nil, err
	}
	var resolveEnv []string
	if required := moduleGoVersion(cfg.Root); required != "" {
		// Ask Go's built-in toolchain manager for the exact version required by
		// the project. This also works when the "go" found on PATH is newer.
		resolveEnv = []string{"GOTOOLCHAIN=go" + required}
	}
	out, err := runner.Output(ctx, command.Spec{
		Path: goPath, Args: []string{"env", "GOROOT"}, Dir: cfg.Root, Env: resolveEnv,
	})
	if err != nil {
		return nil, err
	}
	goRoot := strings.TrimSpace(string(out))
	return []string{
		"GOTOOLCHAIN=local",
		"GOROOT=" + goRoot,
		"PATH=" + filepath.Join(goRoot, "bin") + string(os.PathListSeparator) + os.Getenv("PATH"),
	}, nil
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

func (r Report) Validate(cfg config.Config) error {
	var failures []string
	for _, item := range r.Items {
		if item.Err != nil {
			failures = append(failures, item.Name+": "+item.Err.Error())
		}
	}
	if cfg.SDKRoot == "" {
		failures = append(failures, "sdk_root: set sdk_root in pspgo.toml (the SDK must provide bridge/main.c)")
	}
	goVersion, tinyGoHost := "", ""
	for _, item := range r.Items {
		if item.Name == "go" {
			goVersion = version(item.Version, `go(\d+\.\d+(?:\.\d+)?)`)
		}
		if item.Name == "tinygo" {
			tinyGoHost = version(item.Version, `using go version go(\d+\.\d+(?:\.\d+)?)`)
		}
	}
	if goVersion != "" && tinyGoHost != "" && majorMinor(goVersion) != majorMinor(tinyGoHost) {
		failures = append(failures, fmt.Sprintf("Go %s does not match TinyGo host Go %s; select a compatible Go with PSPGO_GO/PATH and rebuild or select TinyGo accordingly", goVersion, tinyGoHost))
	}
	if required := moduleGoVersion(cfg.Root); required != "" && goVersion != "" && majorMinor(required) != majorMinor(goVersion) {
		failures = append(failures, fmt.Sprintf("project requires Go %s but selected Go is %s; set PSPGO_GO and PATH to a compatible toolchain", required, goVersion))
	}
	if len(failures) > 0 {
		return fmt.Errorf("%s", strings.Join(failures, "\n"))
	}
	return nil
}

func moduleGoVersion(root string) string {
	for dir := root; dir != ""; dir = filepath.Dir(dir) {
		data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
		if err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				fields := strings.Fields(line)
				if len(fields) == 2 && fields[0] == "go" {
					return fields[1]
				}
			}
			return ""
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
	}
	return ""
}
func firstLine(s string) string {
	if line, _, ok := strings.Cut(strings.TrimSpace(s), "\n"); ok {
		return line
	}
	return strings.TrimSpace(s)
}
func version(s, pattern string) string {
	match := regexp.MustCompile(pattern).FindStringSubmatch(s)
	if len(match) == 2 {
		return match[1]
	}
	return ""
}
func majorMinor(s string) string {
	parts := strings.Split(s, ".")
	if len(parts) >= 2 {
		return strings.Join(parts[:2], ".")
	}
	return s
}
