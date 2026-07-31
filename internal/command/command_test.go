package command

import "testing"

func TestSpecStringQuotesArguments(t *testing.T) {
	got := (Spec{Path: "tinygo", Args: []string{"build", "-o", "/tmp/a b.o"}}).String()
	if got != `tinygo build -o "/tmp/a b.o"` {
		t.Fatalf("got %q", got)
	}
}
