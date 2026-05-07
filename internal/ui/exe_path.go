package ui

import "os"

// executablePath returns the absolute path of the running binary, or "" on error.
// Wrapped for testability and to keep settings/wizard logic clean.
func executablePath() string {
	p, err := os.Executable()
	if err != nil {
		return ""
	}
	return p
}
