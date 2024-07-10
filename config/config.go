// Package config provides application configuration data and functions used by the app.
package config

import (
	"os"
	"path/filepath"
)

// AppName is the name of the application.
const AppName = "emacsctl"

// AppDescription is the description of the application.
const AppDescription = `This app enables users to manage multiple emacs environments.
It enables you to define different emacs command lines and configuration
directories. These can be combined into environments that can be used to open
files with the desired emacs command and configuration.

This app stores its state in a JSON file in the application directory. The
application directory is located in the user's ~/.config/emacsctl' by default,
but can be overridden with the --app-dir flag. The state file is named state.json
and is located in the application directory.`

// AppDir is the location of the application state file in unexpanded form.
// This variable is set by the app at runtime.
var AppDir string

// DryRun controls whether the application should execute commands or print them.
// This variable is set by the app at runtime.
var DryRun bool

// Verbose controls whether the application should print verbose output.
// This variable is set by the app at runtime.
var Verbose bool

// Context controls the configuration context to use.
// This variable is set by the app at runtime.
var Context string

// DefaultAppDir is the default application directory when not provided.
var DefaultAppDir, _ = HomeDirPath(".config", AppName)

// DefaultEmacsCommandLine is the default emacs command line when not provided.
const DefaultEmacsCommandLine = "emacs"

// DefaultEmacsConfigDir defines the default emacs configuration directory when not provided.
var DefaultEmacsConfigDir, _ = HomeDirPath(".emacs.d")

// AppPath returns the absolute path of the application directory with the provided path parts.
func AppPath(parts ...string) string {
	return filepath.Join(append([]string{AppDir}, parts...)...)
}

// StatePath returns the absolute path of the application state file.
func StatePath() string {
	return AppPath("state.json")
}

// CachePath returns the absolute path of the application cache directory with the provided path parts.
func CachePath(parts ...string) string {
	return AppPath(append([]string{"cache"}, parts...)...)
}

// HomeDirPath returns the absolute path of the home directory with the provided path parts.
func HomeDirPath(parts ...string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	homeDirPath := homeDir
	if len(parts) > 0 {
		homeDirPath = filepath.Join(append([]string{homeDir}, parts...)...)
	}
	return homeDirPath, nil
}
