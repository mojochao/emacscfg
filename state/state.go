// Package state provides application state management.
package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mojochao/emacscfg/util"
)

// CommandLine represents an emacs command line.
type CommandLine = string

// ConfigDir represents a path to an emacs configuration directory.
type ConfigDir = string

// Environment represents an emacs command and config directory environment.
type Environment struct {
	CommandName string `json:"command_name"` // name of the command to use
	ConfigName  string `json:"config_name"`  // name of the config to use
	Description string `json:"description"`  // description of the config
}

// State represents the state of the application.
type State struct {
	CommandLines map[string]CommandLine `json:"command_lines"`
	ConfigDirs   map[string]ConfigDir   `json:"config_dirs"`
	Environments map[string]Environment `json:"environments"`
	Context      string                 `json:"context"`
}

// New returns a new, empty application state.
func New() *State {
	return &State{
		CommandLines: map[string]CommandLine{
			"default": util.DefaultEmacsCommandLine,
		},
		ConfigDirs: map[string]ConfigDir{
			"default": util.DefaultEmacsConfigDir,
		},
		Environments: map[string]Environment{
			"default": {
				CommandName: "default",
				ConfigName:  "default",
				Description: "default emacs environment",
			},
		},
		Context: "default",
	}
}

// CommandExists checks if a command line exists in the state.
func (s *State) CommandExists(name string) bool {
	_, exists := s.CommandLines[name]
	return exists
}

// AddCommandLine adds a command line to the state.
func (s *State) AddCommandLine(name string, commandLine []string) error {
	// Ensure the named command line does not already exist.
	if _, exists := s.CommandLines[name]; exists {
		return fmt.Errorf("command %s exists", name)
	}

	// Add the command line to the state.
	s.CommandLines[name] = strings.Join(commandLine, " ")
	return nil
}

// RemoveCommandLine removes a command line from the state.
func (s *State) RemoveCommandLine(name string) error {
	// Ensure the command line exists.
	if _, exists := s.ConfigDirs[name]; !exists {
		return fmt.Errorf("command %s not found", name)
	}

	// Remove the configuration from the state.
	delete(s.ConfigDirs, name)

	// Reset the context to nothing.
	s.Context = ""

	return nil
}

// ConfigDirExists checks if a configuration directory exists in the state.
func (s *State) ConfigDirExists(name string) bool {
	_, exists := s.ConfigDirs[name]
	return exists
}

// AddConfigDir adds a configuration directory to the state.
func (s *State) AddConfigDir(name, path string) error {
	// Ensure the configuration does not already exist.
	if _, exists := s.ConfigDirs[name]; exists {
		return fmt.Errorf("configuration %s exists", name)
	}

	// Add the configuration to the state.
	s.ConfigDirs[name] = path

	//// Set the context to the new configuration.
	//s.Context = name
	return nil
}

// RemoveConfigDir removes a configuration directory rom the state.
func (s *State) RemoveConfigDir(name string) error {
	// Ensure the configuration exists.
	if _, exists := s.ConfigDirs[name]; !exists {
		return fmt.Errorf("configuration %s not found", name)
	}

	// Remove the configuration from the state.
	delete(s.ConfigDirs, name)

	// Reset the context to nothing.
	s.Context = ""

	return nil
}

// EnvironmentExists checks if an emacs environment exists in the state.
func (s *State) EnvironmentExists(name string) bool {
	_, exists := s.Environments[name]
	return exists
}

// AddEnvironment adds an emacs environment to the state.
func (s *State) AddEnvironment(name, command, config, description string) error {
	// Ensure the environment does not already exist.
	if _, exists := s.Environments[name]; exists {
		return fmt.Errorf("environment %s exists", name)
	}

	// Ensure the command exists.
	if _, exists := s.CommandLines[name]; !exists {
		return fmt.Errorf("command %s not found", name)
	}

	// Ensure the configuration exists.
	if _, exists := s.ConfigDirs[name]; !exists {
		return fmt.Errorf("configuration %s not found", name)
	}

	// Add the configuration to the state.
	s.Environments[name] = Environment{
		CommandName: command,
		ConfigName:  config,
		Description: description,
	}

	// Set the context to the new configuration.
	s.Context = name
	return nil
}

// RemoveEnvironment removes an emacs environment from the state.
func (s *State) RemoveEnvironment(name string) error {
	// Ensure the environment exists.
	if _, exists := s.Environments[name]; !exists {
		return fmt.Errorf("environment %s not found", name)
	}

	// Remove the environment from the state.
	delete(s.Environments, name)

	// Reset the context to nothing.
	s.Context = ""

	return nil
}

// Load loads the state from the state file.
func Load(path string) (*State, error) {
	// If the state file does not exist, return a new state.
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return New(), nil
	}

	// Otherwise, open the state file and decode the state.
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config State
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, err
	}
	return &config, nil
}

// Save saves the state to the state file.
func Save(state *State, path string) error {
	// Ensure the containing directory exists and create the state file.
	dir := filepath.Dir(path)
	if err := util.EnsureDir(dir); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	return err
}
