// Package state provides application state management.
package state

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/mojochao/emacsctl/config"
	"github.com/mojochao/emacsctl/errors"
	"github.com/mojochao/emacsctl/util"
)

// EmacsCommand represents an emacs command.
type EmacsCommand struct {
	BinPath     string   `json:"bin_path"`
	BinArgs     []string `json:"bin_args"`
	Description string   `json:"description"`
}

func (c *EmacsCommand) CommandLine(initDir string) []string {
	args := make([]string, 0, len(c.BinArgs)+2)
	args = append(args, c.BinPath)
	args = append(args, c.BinArgs...)
	args = append(args, "--init-directory", initDir)
	return args
}

// EmacsConfig represents an emacs configuration.
type EmacsConfig struct {
	InitDir     string `json:"init_dir"`
	Description string `json:"description"`
}

// Environment represents an emacs environment consisting of a EmacsCommand and EmacsConfig.
type Environment struct {
	CommandName string `json:"command_name"`
	ConfigName  string `json:"config_name"`
	Description string `json:"description"`
}

// State represents the state of the application.
type State struct {
	Commands     map[string]EmacsCommand `json:"commands"`
	Configs      map[string]EmacsConfig  `json:"configs"`
	Environments map[string]Environment  `json:"environments"`
	Context      string                  `json:"context"`
}

// New returns a new, empty application state.
func New() *State {
	return &State{
		Commands: map[string]EmacsCommand{
			"default": {
				BinPath:     config.DefaultEmacsCommandLine,
				BinArgs:     nil,
				Description: "Default emacs application",
			},
		},
		Configs: map[string]EmacsConfig{
			"default": {
				InitDir:     config.DefaultEmacsConfigDir,
				Description: "Default emacs configuration",
			},
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
	_, exists := s.Commands[name]
	return exists
}

// AddCommand adds a command to the state.
func (s *State) AddCommand(name string, commandLine []string, description string) error {
	if _, exists := s.Commands[name]; exists {
		return errors.CommandExistsError{Name: name}
	}

	s.Commands[name] = EmacsCommand{
		BinPath:     commandLine[0],
		BinArgs:     commandLine[1:],
		Description: description,
	}
	return nil
}

// RemoveCommand removes a command from the state.
func (s *State) RemoveCommand(name string) error {
	if _, exists := s.Configs[name]; !exists {
		return errors.CommandNotFoundError{Name: name}
	}

	delete(s.Configs, name)
	s.Context = ""
	return nil
}

// ConfigExists checks if a configuration exists in the state.
func (s *State) ConfigExists(name string) bool {
	_, exists := s.Configs[name]
	return exists
}

// AddConfig adds a configuration to the state.
func (s *State) AddConfig(name, path, description string) error {
	if _, exists := s.Configs[name]; exists {
		return errors.ConfigExistsError{Name: name}
	}

	s.Configs[name] = EmacsConfig{
		InitDir:     path,
		Description: description,
	}
	s.Context = name
	return nil
}

// RemoveConfig removes a configuration rom the state.
func (s *State) RemoveConfig(name string) error {
	if _, exists := s.Configs[name]; !exists {
		return errors.ConfigNotFoundError{Name: name}
	}

	delete(s.Configs, name)
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
	if _, exists := s.Environments[name]; exists {
		return errors.EnvironmentExistsError{Name: name}
	}
	if _, exists := s.Commands[name]; !exists {
		return errors.CommandNotFoundError{Name: name}
	}
	if _, exists := s.Configs[name]; !exists {
		return errors.ConfigNotFoundError{Name: name}
	}

	s.Environments[name] = Environment{
		CommandName: command,
		ConfigName:  config,
		Description: description,
	}
	s.Context = name
	return nil
}

// RemoveEnvironment removes an emacs environment from the state.
func (s *State) RemoveEnvironment(name string) error {
	if _, exists := s.Environments[name]; !exists {
		return errors.EnvironmentNotFoundError{Name: name}
	}

	delete(s.Environments, name)
	s.Context = ""
	return nil
}

// Load loads the application state from the state file.
func Load(path string) (*State, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return New(), nil
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var state State
	if err := json.NewDecoder(file).Decode(&state); err != nil {
		return nil, err
	}
	return &state, nil
}

// Save saves the application state to the state file.
func Save(state *State, path string) error {
	dir := filepath.Dir(path)
	if err := util.EnsureDir(dir); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	return err
}
