package target

import "testing"

func TestResolve(t *testing.T) {
	for _, test := range []struct {
		name, override, want string
	}{
		{name: "TinyGo default", want: "psp"},
		{name: "custom target", override: "/tmp/custom.json", want: "/tmp/custom.json"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := Resolve(test.override); got != test.want {
				t.Fatalf("Resolve(%q) = %q, want %q", test.override, got, test.want)
			}
		})
	}
}
