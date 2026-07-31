package registry

import (
	"reflect"
	"testing"
)

func TestSelectGUMBridge(t *testing.T) {
	sources, requirements := Select([]string{"sceGuStart", "pspsdk_go_gum_draw_array_3d"})
	if !reflect.DeepEqual(sources, []string{"bridge/gum_abi.c"}) {
		t.Fatalf("sources = %#v", sources)
	}
	if !reflect.DeepEqual(requirements, []string{"sceGumDrawArray"}) {
		t.Fatalf("requirements = %#v", requirements)
	}
}
