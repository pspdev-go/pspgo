package target

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMaterialize(t *testing.T) {
	buildDir := t.TempDir()
	path, err := Materialize(buildDir)
	if err != nil {
		t.Fatal(err)
	}
	if path != filepath.Join(buildDir, "target", "psp.json") {
		t.Fatalf("path = %q", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("embedded target is empty")
	}
}
