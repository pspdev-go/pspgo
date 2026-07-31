package resolver

import "testing"

func TestParseIndex(t *testing.T) {
	idx := parseIndex("/sdk/lib/libpspgu.a:gu.o:00000000 T sceGuStart\n/sdk/lib/libpspgu.a:gu.o:         U sceKernelFoo\n")
	defs := idx.definitions["sceGuStart"]
	if len(defs) != 1 || defs[0].object != "gu.o" {
		t.Fatalf("definitions = %#v", defs)
	}
	if got := idx.undefined[member{"/sdk/lib/libpspgu.a", "gu.o"}]; len(got) != 1 || got[0] != "sceKernelFoo" {
		t.Fatalf("undefined = %#v", got)
	}
}

func TestChoosePrefersSelectedArchive(t *testing.T) {
	items := []member{{"/sdk/lib/liblongname.a", "a.o"}, {"/sdk/lib/libx.a", "x.o"}}
	got := choose("symbol", items, map[string]bool{items[0].archive: true})
	if got.archive != items[0].archive {
		t.Fatalf("got %s", got.archive)
	}
}
