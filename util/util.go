// Package util provides shared utility functions and data for the application.
package util

import (
	"os"

	"github.com/mitchellh/go-homedir"
)

// DefaultEmacsCommandLine is the default emacs command line when not provided.
const DefaultEmacsCommandLine = "emacs"

// DefaultEmacsConfigDir defines the default emacs configuration directory
// when not provided.
var DefaultEmacsConfigDir, _ = homedir.Expand("~/.emacs.d")

// EnsureDir ensures directory exists.
func EnsureDir(path string) error {
	path, err := homedir.Expand(path)
	if err != nil {
		return err
	}

	_, err = os.Stat(path)
	if !os.IsNotExist(err) {
		return nil
	}

	return os.MkdirAll(path, 0755)
}
