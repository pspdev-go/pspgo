package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	Root       string
	ProjectDir string
	Package    string
	BuildDir   string
	OutputName string
	Title      string
	SDKRoot    string
	Target     string
	KernelMode bool
	Go         string
	TinyGo     string
	PSPCMake   string
	PSPNM      string
	NM         string
	CMake      string
	Make       string
	PPSSPP     string
}

func Load(start string) (Config, error) {
	root, err := filepath.Abs(start)
	if err != nil {
		return Config{}, err
	}
	if info, statErr := os.Stat(root); statErr == nil && !info.IsDir() {
		root = filepath.Dir(root)
	}
	cfg := Config{
		Root: root, Package: ".", BuildDir: "build/pspgo",
		OutputName: "app", Title: "PSP Go Application",
		Go: env("PSPGO_GO", "go"), TinyGo: env("PSPGO_TINYGO", "tinygo"),
		PSPCMake: env("PSPGO_PSP_CMAKE", "psp-cmake"),
		PSPNM:    env("PSPGO_PSP_NM", env("PSPGO_NM", "psp-nm")),
		CMake:    env("PSPGO_CMAKE", "cmake"), Make: env("PSPGO_MAKE", "make"),
		PPSSPP: env("PSPGO_PPSSPP", "PPSSPPSDL"),
	}
	cfg.ProjectDir = cfg.Root
	cfg.NM = cfg.PSPNM
	path := filepath.Join(root, "pspgo.toml")
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		cfg.autoDetect()
		return cfg, nil
	}
	if err != nil {
		return Config{}, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(strings.SplitN(scanner.Text(), "#", 2)[0])
		if line == "" || strings.HasPrefix(line, "[") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return Config{}, fmt.Errorf("%s: invalid setting %q", path, line)
		}
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		switch key {
		case "package":
			cfg.Package = value
		case "build_dir":
			cfg.BuildDir = value
		case "output":
			cfg.OutputName = value
		case "title":
			cfg.Title = value
		case "target":
			cfg.Target = value
		case "kernel_mode":
			cfg.KernelMode, err = strconv.ParseBool(value)
			if err != nil {
				return Config{}, fmt.Errorf("%s: kernel_mode: %w", path, err)
			}
		case "go":
			cfg.Go = value
		case "tinygo":
			cfg.TinyGo = value
		case "psp_cmake":
			cfg.PSPCMake = value
		case "psp_nm":
			cfg.PSPNM = value
		case "cmake":
			cfg.CMake = value
		case "make":
			cfg.Make = value
		case "ppsspp":
			cfg.PPSSPP = value
		default:
			return Config{}, fmt.Errorf("%s: unknown setting %q", path, key)
		}
	}
	if err := scanner.Err(); err != nil {
		return Config{}, err
	}
	cfg.resolvePaths()
	cfg.autoDetect()
	return cfg, nil
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func (c *Config) resolvePaths() {
	for _, field := range []*string{&c.BuildDir, &c.Target} {
		if *field != "" && !filepath.IsAbs(*field) {
			*field = filepath.Join(c.Root, *field)
		}
	}
}

func (c *Config) autoDetect() {
	if c.Root == "" {
		c.Root = c.ProjectDir
	}
	if c.ProjectDir == "" {
		c.ProjectDir = c.Root
	}
	if c.PSPNM == "" {
		c.PSPNM = c.NM
	}
	if c.NM == "" {
		c.NM = c.PSPNM
	}
	c.resolvePaths()
}
