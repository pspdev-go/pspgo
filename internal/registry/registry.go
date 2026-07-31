package registry

type Bridge struct {
	Symbol   string
	Source   string
	Requires []string
}

// Bridges is deliberately data, not resolver control flow. SDKs can later
// supply the same schema from a manifest without changing the resolver.
var Bridges = []Bridge{
	{Symbol: "psp_debugscreen_kputs", Source: "bridge/printf.c", Requires: []string{"pspDebugScreenKprintf"}},
	{Symbol: "psp_kdebug_puts", Source: "bridge/printf.c", Requires: []string{"Kprintf"}},
	{Symbol: "pspsdk_go_fdputs", Source: "bridge/printf.c", Requires: []string{"fdprintf"}},
	{Symbol: "pspsdk_go_vlf_add_text", Source: "bridge/vlf_printf.c"},
	{Symbol: "pspsdk_go_vlf_set_text", Source: "bridge/vlf_printf.c"},
	{Symbol: "pspsdk_go_gum_draw_array", Source: "bridge/gum_abi.c", Requires: []string{"sceGumDrawArray"}},
	{Symbol: "pspsdk_go_gum_draw_array_3d", Source: "bridge/gum_abi.c", Requires: []string{"sceGumDrawArray"}},
}

func Select(undefined []string) (sources, requirements []string) {
	seenSource, seenReq := map[string]bool{}, map[string]bool{}
	for _, symbol := range undefined {
		for _, bridge := range Bridges {
			if bridge.Symbol != symbol {
				continue
			}
			if !seenSource[bridge.Source] {
				seenSource[bridge.Source] = true
				sources = append(sources, bridge.Source)
			}
			for _, req := range bridge.Requires {
				if !seenReq[req] {
					seenReq[req] = true
					requirements = append(requirements, req)
				}
			}
		}
	}
	return sources, requirements
}
