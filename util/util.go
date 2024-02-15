// Package util provides utility functions for the application.
package util

import (
	"os"

	"github.com/mitchellh/go-homedir"
)

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
