package target

import (
	"bytes"
	_ "embed"
	"os"
	"path/filepath"
)

//go:embed psp.json
var pspJSON []byte

// Materialize writes the embedded TinyGo target into the build directory.
func Materialize(buildDir string) (string, error) {
	path := filepath.Join(buildDir, "target", "psp.json")
	if current, err := os.ReadFile(path); err == nil && bytes.Equal(current, pspJSON) {
		return path, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, pspJSON, 0o644); err != nil {
		return "", err
	}
	return path, nil
}
