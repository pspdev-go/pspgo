package resolver

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/pspdev-go/pspgo/internal/command"
	"github.com/pspdev-go/pspgo/internal/registry"
)

var undefinedLine = regexp.MustCompile(`\bU\s+(\S+)$`)
var archiveLine = regexp.MustCompile(`^(.+?\.a):([^:]+):(?:[0-9A-Fa-f]+)?\s*([A-Za-zU])\s+(\S+)$`)

type index struct {
	definitions map[string][]member
	undefined   map[member][]string
}
type member struct{ archive, object string }

type Result struct {
	Libraries []string
	Bridges   []string
	Undefined []string
}

func Resolve(ctx context.Context, runner command.Runner, nm, pspdev string, objects []string) (Result, error) {
	args := append([]string{"-u"}, objects...)
	out, err := runner.Output(ctx, command.Spec{Path: nm, Args: args})
	if err != nil {
		return Result{}, err
	}
	var symbols []string
	for line := range strings.SplitSeq(string(out), "\n") {
		if match := undefinedLine.FindStringSubmatch(line); match != nil {
			symbols = append(symbols, match[1])
		}
	}
	sort.Strings(symbols)
	bridges, requirements := registry.Select(symbols)
	libDir := filepath.Join(pspdev, "psp", "sdk", "lib")
	archives, err := filepath.Glob(filepath.Join(libDir, "*.a"))
	if err != nil || len(archives) == 0 {
		return Result{}, fmt.Errorf("no PSPSDK archives found in %s", libDir)
	}
	scanArgs := append([]string{"-A", "-g"}, archives...)
	scan, err := runner.Output(ctx, command.Spec{Path: nm, Args: scanArgs})
	if err != nil {
		return Result{}, err
	}
	idx := parseIndex(string(scan))
	pending := append(append([]string{}, symbols...), requirements...)
	selected := map[string]bool{}
	visitedSymbols, visitedMembers := map[string]bool{}, map[member]bool{}
	var unresolved []string
	for len(pending) > 0 {
		symbol := pending[0]
		pending = pending[1:]
		if visitedSymbols[symbol] {
			continue
		}
		visitedSymbols[symbol] = true
		if after, ok := strings.CutPrefix(symbol, "pspsdk_go_require_"); ok {
			name := after
			path := filepath.Join(libDir, "lib"+name+".a")
			if !contains(archives, path) {
				return Result{}, fmt.Errorf("required PSPSDK library does not exist: %s", path)
			}
			selected[path] = true
			continue
		}
		if isIgnored(symbol) {
			continue
		}
		candidates := idx.definitions[symbol]
		if len(candidates) == 0 {
			if strings.HasPrefix(symbol, "sce") || strings.HasPrefix(symbol, "psp") || symbol == "Kprintf" {
				unresolved = append(unresolved, symbol)
			}
			continue
		}
		chosen := choose(symbol, candidates, selected)
		if filepath.Base(chosen.archive) == "libpspuser.a" {
			continue
		}
		selected[chosen.archive] = true
		if !visitedMembers[chosen] {
			visitedMembers[chosen] = true
			pending = append(pending, idx.undefined[chosen]...)
		}
	}
	if len(unresolved) > 0 {
		sort.Strings(unresolved)
		return Result{}, fmt.Errorf("PSPSDK libraries not found for: %s", strings.Join(unresolved, ", "))
	}
	var libraries []string
	for path := range selected {
		name := strings.TrimSuffix(strings.TrimPrefix(filepath.Base(path), "lib"), ".a")
		libraries = append(libraries, name)
	}
	sort.Strings(libraries)
	return Result{Libraries: libraries, Bridges: bridges, Undefined: symbols}, nil
}

func parseIndex(text string) index {
	idx := index{definitions: map[string][]member{}, undefined: map[member][]string{}}
	for line := range strings.SplitSeq(text, "\n") {
		m := archiveLine.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		key := member{m[1], m[2]}
		if m[3] == "U" {
			idx.undefined[key] = append(idx.undefined[key], m[4])
		} else {
			idx.definitions[m[4]] = append(idx.definitions[m[4]], key)
		}
	}
	return idx
}

func choose(symbol string, candidates []member, selected map[string]bool) member {
	sort.SliceStable(candidates, func(i, j int) bool {
		return score(symbol, candidates[i], selected) < score(symbol, candidates[j], selected)
	})
	return candidates[0]
}

func score(symbol string, item member, selected map[string]bool) int {
	name := strings.TrimSuffix(strings.TrimPrefix(filepath.Base(item.archive), "lib"), ".a")
	value := len(name)
	if strings.Contains(name, "driver") {
		value += 1000
	}
	if strings.Contains(name, "_660") || strings.HasSuffix(name, "kernel") {
		value += 500
	}
	if name == "pspuser" {
		value += 100
	}
	if selected[item.archive] {
		value -= 10000
	}
	normalizedName := strings.ReplaceAll(strings.ToLower(name), "_", "")
	normalizedSymbol := strings.ReplaceAll(strings.ToLower(symbol), "_", "")
	if strings.HasPrefix(normalizedName, "psp") && strings.Contains(normalizedSymbol, strings.TrimPrefix(normalizedName, "psp")) {
		value -= 200
	}
	return value
}

func isIgnored(symbol string) bool {
	if strings.HasPrefix(symbol, "__") || strings.HasPrefix(symbol, "pspsdk_go_") {
		return true
	}
	_, ok := ignored[symbol]
	return ok
}

var ignored = func() map[string]bool {
	names := strings.Fields("abort calloc free malloc memalign memchr memcmp memcpy memmove memset printf putchar puts realloc snprintf sprintf strcat strchr strcmp strcpy strlen strncmp strncpy strrchr strstr vsnprintf acosf asinf atan2f atanf ceilf cosf expf fabsf floorf fmodf logf powf sinf sqrtf tanf")
	out := map[string]bool{}
	for _, name := range names {
		out[name] = true
	}
	return out
}()

func contains(items []string, wanted string) bool {
	return slices.Contains(items, wanted)
}
