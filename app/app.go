// Package app provides the cli application using v2 of https://cli.urfave.org/.
package app

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/fatih/color"
	"github.com/mitchellh/go-homedir"
	"github.com/rodaine/table"
	"github.com/urfave/cli/v2"

	"github.com/mojochao/emacscfg/cache"
	"github.com/mojochao/emacscfg/state"
)

// appDir controls the location of the application state file in unexpanded form.
var appDir string

// appDirFlag is the flag used to specify an alternate application directory
var appDirFlag = cli.StringFlag{
	Name:        "app-dir",
	Usage:       "Specify application directory",
	Destination: &appDir,
	Value:       "~/.config/emacscfg",
	EnvVars:     []string{"EMACSCFG_DIR"},
}

// dryRun controls whether the application should execute commands or print them.
var dryRun bool

// dryRunFlag is the flag used to specify commands to be printed but not executed.
var dryRunFlag = cli.BoolFlag{
	Name:        "dry-run",
	Usage:       "Display the command that would be executed, but do not execute it",
	Destination: &dryRun,
}

// verbose controls whether the application should print verbose output.
var verbose bool

// verboseFlag is the flag used to specify increased output.
var verboseFlag = cli.BoolFlag{
	Name:        "verbose",
	Aliases:     []string{"v"},
	Usage:       "Display verbose output",
	Destination: &verbose,
}

// context controls the configuration context to use.
var context string

// contextFlag is the flag used to provide name of a configuration context to
// use instead of any active context found in the state.
var contextFlag = cli.StringFlag{
	Name:        "context",
	Aliases:     []string{"c"},
	Usage:       "Use a specific configuration context",
	Destination: &context,
}

// noContextError indicates no configuration context found in the active context or context flag.
var noContextError = fmt.Errorf("no configuration specified or active")

// New creates a new cli application.
func New() *cli.App {
	return &cli.App{
		Name:  "emacscfg",
		Usage: "Manage multiple emacs configuration profiles",
		Flags: []cli.Flag{
			&appDirFlag,
			&dryRunFlag,
			&verboseFlag,
		},
		Commands: []*cli.Command{
			{
				Name:   "state",
				Usage:  "Display application state",
				Action: showState,
			},
			{
				Name:      "context",
				Aliases:   []string{"ctx"},
				Usage:     "Get or set the active configuration context in application state",
				Action:    activeContext,
				Args:      true,
				ArgsUsage: "[NAME]",
			},
			{
				Name:    "list",
				Aliases: []string{"ls"},
				Usage:   "Display table of all configurations in application state",
				Action:  listConfigs,
			},
			{
				Name:      "add",
				Usage:     "Add a new configuration to application state",
				Action:    addConfig,
				Args:      true,
				ArgsUsage: "NAME PATH",
			},
			{
				Name:      "remove",
				Aliases:   []string{"rm"},
				Usage:     "Remove a configuration from application state",
				Action:    removeConfig,
				Args:      true,
				ArgsUsage: "NAME",
			},
			{
				Name:    "path",
				Aliases: []string{"dir"},
				Usage:   "Print the path of the configuration directory",
				Action:  showConfigPath,
				Flags: []cli.Flag{
					&contextFlag,
				},
			},
			{
				Name:      "open",
				Usage:     "Open files in emacs with the desired configuration",
				Action:    openEmacs,
				Args:      true,
				ArgsUsage: "[FILES...]",
				Flags: []cli.Flag{
					&contextFlag,
				},
			},
			{
				Name:   "version",
				Usage:  "Print the version of the application",
				Action: showAppVersion,
			},
		},
	}
}

// showState prints the application state.
func showState(_ *cli.Context) error {
	// Display the path of the application state file if verbose output enabled.

	if verbose {
		fmt.Println(statePath())
	}

	// Load the application state and print it to stdout.
	appState, err := state.Load(statePath())
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(appState, "", "  ")
	if err != nil {
		return err
	}

	_, err = os.Stdout.Write(data)
	return err
}

// listConfigs prints a table of all configurations in the state file.
func listConfigs(_ *cli.Context) error {
	// Load the application state.
	appState, err := state.Load(statePath())
	if err != nil {
		return err
	}

	// If no configurations are found, there's nothing else to do.
	if appState == nil || len(appState.Configs) == 0 {
		return nil
	}

	// Otherwise, print a pretty table of all configurations.
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()
	tbl := table.New("Name", "Path")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)
	for name, path := range appState.Configs {
		tbl.AddRow(name, path)
	}

	tbl.Print()
	return nil
}

// addConfig adds a new configuration to the state file.
func addConfig(c *cli.Context) error {
	// Verify correct usage.
	if c.NArg() != 2 {
		return fmt.Errorf("expected 2 arguments, got %d", c.NArg())
	}
	name := c.Args().Get(0)
	path := c.Args().Get(1)

	// Load the application state.
	appState, err := state.Load(statePath())
	if err != nil {
		return err
	}

	// If is a dry run, there's nothing else to do.
	if dryRun {
		return nil
	}

	// If the path is a git URL, add the repository to the cache.
	if isGitURL(path) {
		// Add the repository to the cache.
		if err := cache.AddRepo(cachePath(), name, path); err != nil {
			return err
		}
	}

	// Otherwise, add the configuration to the application state and save it back to the state file.
	if err := appState.AddConfig(name, path); err != nil {
		return err
	}
	if err := state.Save(appState, statePath()); err != nil {
		return err
	}

	// Success!
	if verbose {
		fmt.Printf("added configuration: %s\n", name)
	}
	return nil

}

// removeConfig removes a configuration from the state file.
func removeConfig(c *cli.Context) error {
	// Verify correct usage.
	if c.NArg() != 1 {
		return fmt.Errorf("expected 1 argument, got %d", c.NArg())
	}
	name := c.Args().Get(0)

	// Load the application state.
	appState, err := state.Load(statePath())
	if err != nil {
		return err
	}

	// Find the config in the application state.
	_, exists := appState.Configs[name]
	if !exists {
		return fmt.Errorf("configuration %s not found", name)
	}

	// If is a dry run, there's nothing else to do.
	if dryRun {
		return nil
	}

	// Otherwise, remove any cached repository from the filesystem.
	if cache.IsCached(cachePath(), name) {
		if err := cache.RemoveRepo(cachePath(), name); err != nil {
			return err
		}
	}

	// Remove configuration from the application state and save it back to the state file.
	if err := appState.RemoveConfig(name); err != nil {
		return err
	}
	if err := state.Save(appState, statePath()); err != nil {
		return err
	}

	// Success!
	if verbose {
		fmt.Printf("removed configuration: %s\n", name)
	}
	return nil
}

// activeContext gets or sets the active configuration context in the state file.
func activeContext(c *cli.Context) error {
	// Verify correct usage.
	if c.NArg() > 1 {
		return fmt.Errorf("expected 0 or 1 argument, got %d", c.NArg())
	}

	// Load the application state.
	appState, err := state.Load(statePath())
	if err != nil {
		return err
	}

	// If no arguments are provided, print the active context and return.
	if c.NArg() == 0 {
		fmt.Println(appState.Context)
		return nil
	}

	// If is a dry run, there's nothing else to do.
	if dryRun {
		return nil
	}

	// Otherwise, set the active context and save it back to the state file.
	appState.Context = c.Args().Get(0)
	return state.Save(appState, statePath())
}

// openEmacs opens emacs with the desired configuration and all provided arguments.
func openEmacs(c *cli.Context) error {
	// Load the application state.
	appState, err := state.Load(statePath())
	if err != nil {
		return err
	}

	// Ensure a configuration context is available.
	if context == "" {
		context = appState.Context
	}
	if context == "" {
		return noContextError
	}

	// Build the command line to execute.
	cmdline := []string{"emacs"}
	initDir, err := appState.GetConfigPath(context)
	if err != nil {
		return err
	}
	cmdline = append(cmdline, "--init-directory", initDir)
	cmdline = append(cmdline, c.Args().Slice()...)

	// If is a dry run, print the command line and return.
	if dryRun {
		fmt.Println(strings.Join(cmdline, " "))
		return nil
	}

	// Otherwise, execute the command line.
	cmd := exec.Command(cmdline[0], cmdline[1:]...)
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

// showConfigPath prints the path of the configuration directory.
func showConfigPath(_ *cli.Context) error {
	// Load the application state.
	appState, err := state.Load(statePath())
	if err != nil {
		return err
	}

	// Ensure a configuration context is available.
	if context == "" {
		context = appState.Context
	}
	if context == "" {
		return noContextError
	}

	// Get and print the path of the configuration directory.
	configPath, err := appState.GetConfigPath(context)
	if err != nil {
		return err
	}
	fmt.Println(configPath)
	return nil
}

// showAppVersion prints the version of the application set at build time by
// the `go build -ldflags "-X github.com/mojochao/emacscfg/app.version=0.10.0" -o emacscfg .` command.
var version string

// showAppVersion prints the version of the application.
func showAppVersion(_ *cli.Context) error {
	// Print the version of the application.
	fmt.Printf("emacscfg version %s\n", version)
	if !verbose {
		return nil
	}

	// If verbose, print the VCS settings used to build the application.
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if strings.HasPrefix(setting.Key, "vcs.") {
				fmt.Printf("%s: %s\n", setting.Key, setting.Value)
			}
		}
	}
	return nil
}

// isGitURL checks if the input is a valid git URL.
func isGitURL(input string) bool {
	return strings.HasPrefix(input, "git@") || strings.HasPrefix(input, "https://")
}

// appPath returns the absolute path of the components under the application directory after tilde home directory expansion.
func appPath(components ...string) string {
	base, _ := homedir.Expand(appDir)
	return filepath.Join(append([]string{base}, components...)...)
}

// statePath returns the absolute path of the application state file.
func statePath() string {
	return appPath("state.json")
}

// cachePath returns the absolute path of the application cache directory, or repository subdirectory,
// after tilde home directory expansion.
func cachePath(name ...string) string {
	return filepath.Join(appPath("cache"), filepath.Join(name...))
}
