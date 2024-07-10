package errors

import "fmt"

type UnexpectedNumArgsError struct {
	Expected int
	Received int
}

func (e UnexpectedNumArgsError) Error() string {
	return fmt.Sprintf("unexpected number of arguments: expected %d, got %d", e.Expected, e.Received)
}

type MinimumNumArgsError struct {
	Minimum  int
	Received int
}

func (e MinimumNumArgsError) Error() string {
	return fmt.Sprintf("minimum number of arguments not met: minimum %d, got %d", e.Minimum, e.Received)
}

type CommandExistsError struct {
	Name string
}

func (e CommandExistsError) Error() string {
	return "command already exists: " + e.Name
}

type CommandNotFoundError struct {
	Name string
}

func (e CommandNotFoundError) Error() string {
	return "command not found: " + e.Name
}

type ConfigExistsError struct {
	Name string
}

func (e ConfigExistsError) Error() string {
	return "config already exists: " + e.Name
}

type ConfigNotFoundError struct {
	Name string
}

func (e ConfigNotFoundError) Error() string {
	return "config not found: " + e.Name
}

type EnvironmentExistsError struct {
	Name string
}

func (e EnvironmentExistsError) Error() string {
	return "environment already exists: " + e.Name
}

type EnvironmentNotFoundError struct {
	Name string
}

func (e EnvironmentNotFoundError) Error() string {
	return "environment not found: " + e.Name
}

var NoContextError = fmt.Errorf("no environment context specified or active")
