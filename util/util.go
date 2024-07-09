// Package util provides shared utility functions and data for the application.
package util

import (
	"os"
	"path"
)

// DefaultEmacsCommandLine is the default emacs command line when not provided.
const DefaultEmacsCommandLine = "emacs"

// DefaultEmacsConfigDir defines the default emacs configuration directory
// when not provided.
var DefaultEmacsConfigDir, _ = HomeDirPath(".emacs.d")

// EnsureDir ensures directory exists.
func EnsureDir(path string) error {
	homeDirPath, err := HomeDirPath(path)
	if err != nil {
		return err
	}

	_, err = os.Stat(homeDirPath)
	if !os.IsNotExist(err) {
		return nil
	}

	return os.MkdirAll(homeDirPath, 0755)
}

func HomeDirPath(parts ...string) (string, error) {
	// Get the home directory.
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	// If no parts are provided, return the home directory.
	if len(parts) == 0 {
		return homeDir, nil
	}

	//// Otherwise, ensure the tilde is replaced with the home directory
	////in all parts and return the joined path.
	//for _, part := range parts {
	//	strings.Replace(part, "~", homeDir, 1)
	//}
	return path.Join(append([]string{homeDir}, parts...)...), nil
}
