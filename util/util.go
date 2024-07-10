// Package util provides shared utility functions for the application.
package util

import (
	"os"
	"runtime/debug"
	"strings"
)

// EnsureDir ensures directory exists.
func EnsureDir(path string) error {
	_, err := os.Stat(path)
	if !os.IsNotExist(err) {
		return nil
	}
	return os.MkdirAll(path, 0755)
}

// GetBuildInfo returns the build information for the application.
func GetBuildInfo() map[string]string {
	var results map[string]string
	if info, ok := debug.ReadBuildInfo(); ok {
		results = make(map[string]string)
		for _, setting := range info.Settings {
			if strings.HasPrefix(setting.Key, "vcs.") {
				results[setting.Key] = setting.Value
			}
		}
	}
	return results
}

// IsGitURL checks if the input is a valid git URL.
func IsGitURL(input string) bool {
	return strings.HasPrefix(input, "git@") || strings.HasPrefix(input, "https://")
}
