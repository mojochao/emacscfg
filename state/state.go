// Package state provides application state management.
package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mojochao/emacscfg/util"
)

// State represents the state of the application.
type State struct {
	Configs map[string]string `json:"configs"` // map of config name to internal directory of repositories cloned by the application
	Context string            `json:"context"` // name of current, active configuration context
}

// New returns a new, empty application state.
func New() *State {
	return &State{
		Configs: make(map[string]string),
	}
}

// AddConfig adds a configuration to the state.
func (s *State) AddConfig(name, path string) error {
	// Ensure the configuration does not already exist.
	if _, exists := s.Configs[name]; exists {
		return fmt.Errorf("configuration %s exists", name)
	}

	// Add the configuration to the state.
	s.Configs[name] = path

	// Set the context to the new configuration.
	s.Context = name
	return nil
}

// RemoveConfig removes a configuration from the state.
func (s *State) RemoveConfig(name string) error {
	// Ensure the configuration exists.
	if _, exists := s.Configs[name]; !exists {
		return fmt.Errorf("configuration %s not found", name)
	}

	// Remove the configuration from the state.
	delete(s.Configs, name)

	// Reset the context to nothing.
	s.Context = ""

	return nil
}

// GetConfigPath returns the path to a managed configuration in the state.
func (s *State) GetConfigPath(name string) (string, error) {
	path, exists := s.Configs[name]
	if !exists {
		return "", fmt.Errorf("configuration %s not found", name)
	}
	return path, nil
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
