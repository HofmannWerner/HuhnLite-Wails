//go:build !windows

package utils

// SetupServerConsole is a no-op on non-Windows platforms.
func SetupServerConsole(title string, minimize bool, disableCloseButton bool) {
	// No-op
}
