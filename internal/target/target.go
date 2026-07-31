// Package target resolves the TinyGo target used for PSP builds.
package target

const Default = "psp"

// Resolve returns an explicitly configured target or TinyGo's built-in PSP
// target name when no override is configured.
func Resolve(override string) string {
	if override != "" {
		return override
	}
	return Default
}
