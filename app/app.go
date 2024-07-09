// Package app provides the cli application using v2 of https://cli.urfave.org/.
package app

import (
	"encoding/json"
	"fmt"
	"github.com/mojochao/emacscfg/util"
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

// appName is the name of the application.
const appName = "emacscfg"

// appDescription is the description of the application.
const appDescription = `This app enables users to manage multiple emacs environments.
It enables you to define different emacs command lines and configuration
directories. These can be combined into environments that can be used to open
files with the desired emacs command and configuration.

This app stores its state in a JSON file in the application directory. The
application directory is located in the user's ~/.config/emacscfg' by default,
but can be overridden with the --app-dir flag. The state file is named state.json
and is located in the application directory.`

// defaultAppDir is the default application directory when not provided.
var defaultAppDir, _ = homedir.Expand(fmt.Sprintf("~/.config/%s", appName))

// appDir is the location of the application state file in unexpanded form.
var appDir string

// appDirFlag is the flag used to specify an alternate application directory
var appDirFlag = cli.StringFlag{
	Name:        "app-dir",
	Usage:       "Specify application directory",
	Destination: &appDir,
	Value:       defaultAppDir,
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
		Name:        appName,
		Usage:       "Manage multiple emacs environments",
		Description: appDescription,
		Flags: []cli.Flag{
			&appDirFlag,
			&dryRunFlag,
			&verboseFlag,
		},
		Commands: []*cli.Command{
			{
				Name:  "state",
				Usage: "Display application state",
				Subcommands: []*cli.Command{
					{
						Name:    "show",
						Aliases: []string{"cat", "view"},
						Usage:   "Display the content of the application state file",
						Action:  showState,
					},
					{
						Name:    "path",
						Aliases: []string{"file"},
						Usage:   "Display the path of the application state file",
						Action:  showStatePath,
					},
				},
			},
			{
				Name:    "environment",
				Aliases: []string{"env"},
				Usage:   "Manage emacs environments",
				Subcommands: []*cli.Command{
					{
						Name:    "list",
						Aliases: []string{"ls"},
						Usage:   "Display table of all emacs environments in application state",
						Action:  listEnvironments,
					},
					{
						Name:      "add",
						Usage:     "Add a new emacs environment to application state",
						Action:    addEnvironment,
						Args:      true,
						ArgsUsage: "NAME",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "command",
								Aliases: []string{"cmd"},
								Usage:   "Name of existing emacs command line to use for environment",
							},
							&cli.StringFlag{
								Name:    "commandline",
								Aliases: []string{"cmdline"},
								Usage:   "New emacs command line to use for environment",
							},
							&cli.StringFlag{
								Name:    "config",
								Aliases: []string{"cfg"},
								Usage:   "Name of existing emacs configuration directory to use for environment",
							},
							&cli.StringFlag{
								Name:    "configdir",
								Aliases: []string{"cfgdir"},
								Usage:   "New emacs configuration directory to use for environment",
							},
							&cli.StringFlag{
								Name:    "description",
								Aliases: []string{"desc"},
								Usage:   "Description of the environment",
							},
						},
					},
					{
						Name:      "remove",
						Aliases:   []string{"rm"},
						Usage:     "Remove an existing environment from application state",
						Action:    removeEnvironment,
						Args:      true,
						ArgsUsage: "NAME",
					},
				},
			},
			{
				Name:    "command",
				Aliases: []string{"cmd"},
				Usage:   "Manage emacs command lines",
				Subcommands: []*cli.Command{
					{
						Name:    "list",
						Aliases: []string{"ls"},
						Usage:   "Display table of all emacs command lines in application state",
						Action:  listCommands,
					},
					{
						Name:      "add",
						Usage:     "Add a new emacs command line to application state",
						Action:    addCommand,
						Args:      true,
						ArgsUsage: "NAME CMD_LINE",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "description",
								Aliases: []string{"desc"},
								Usage:   "Description of the command line",
							},
						},
					},
					{
						Name:      "remove",
						Aliases:   []string{"rm"},
						Usage:     "Remove an existing emacs command line from application state",
						Action:    removeCommand,
						Args:      true,
						ArgsUsage: "NAME",
					},
				},
			},
			{
				Name:    "config",
				Aliases: []string{"cfg"},
				Usage:   "Manage emacs configuration directories in application state",
				Subcommands: []*cli.Command{
					{
						Name:    "list",
						Aliases: []string{"ls"},
						Usage:   "Display table of all emacs configuration directories in application state",
						Action:  listConfigs,
					},
					{
						Name:      "add",
						Usage:     "Add a new emacs configuration directory to application state",
						Action:    addConfig,
						Args:      true,
						ArgsUsage: "NAME DIR_PATH",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "description",
								Aliases: []string{"desc"},
								Usage:   "Description of the configuration directory",
							},
						},
					},
					{
						Name:      "remove",
						Aliases:   []string{"rm"},
						Usage:     "Remove an existing emacs configuration directory from application state",
						Action:    removeConfig,
						Args:      true,
						ArgsUsage: "NAME",
					},
				},
			},
			{
				Name:    "context",
				Aliases: []string{"ctx"},
				Usage:   "Manage active environment context in application state",
				Subcommands: []*cli.Command{
					{
						Name:   "get",
						Usage:  "Get the active environment context",
						Action: getContext,
					},
					{
						Name:      "set",
						Usage:     "Set the active environment context",
						Action:    setContext,
						Args:      true,
						ArgsUsage: "ENV",
					},
					{
						Name:   "clear",
						Usage:  "Clear the active environment context",
						Action: clearContext,
					},
				},
			},
			{
				Name:      "open",
				Aliases:   []string{"edit"},
				Usage:     "Open files in the desired emacs environment",
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

// listEnvironments prints a table of all environments in the state file.
func listEnvironments(_ *cli.Context) error {
	// Load the application state.
	appState, err := state.Load(statePath())
	if err != nil {
		return err
	}

	// If no environments are found, there's nothing else to do.
	if appState == nil || len(appState.Environments) == 0 {
		return nil
	}

	// Otherwise, print a pretty table of all environments.
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()
	tbl := table.New("Name", "Path")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)
	for name, path := range appState.Environments {
		tbl.AddRow(name, path)
	}

	tbl.Print()
	return nil

}

// addEnvironment adds a new environment to the state file.
func addEnvironment(c *cli.Context) error {
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

	// If is a dry run, there's nothing else to do.
	if dryRun {
		return nil
	}

	// Get the optional command line, configuration directory, and description from the flags.
	commandName := c.String("command")
	commandLine, ok := appState.CommandLines[commandName]
	if !ok {
		commandLine = util.DefaultEmacsCommandLine
	}

	configName := c.String("config")
	configPath, ok := appState.ConfigDirs[configName]
	if !ok {
		configPath = util.DefaultEmacsConfigDir
	}

	description := c.String("description")

	// Add the environment to the application state and save it back to the state file.
	if err := appState.AddEnvironment(name, commandLine, configPath, description); err != nil {
		return err
	}
	if err := state.Save(appState, statePath()); err != nil {
		return err
	}

	// Success!
	if verbose {
		fmt.Printf("added environment: %s\n", name)
	}
	return nil
}

// removeEnvironment removes an environment from the state file.
func removeEnvironment(c *cli.Context) error {
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

	// Find the environment in the application state.
	_, exists := appState.Environments[name]
	if !exists {
		return fmt.Errorf("environment %s not found", name)
	}

	// If is a dry run, there's nothing else to do.
	if dryRun {
		return nil
	}

	// Remove the environment from the application state and save it back to the state file.
	if err := appState.RemoveEnvironment(name); err != nil {
		return err
	}
	if err := state.Save(appState, statePath()); err != nil {
		return err
	}

	// Success!
	if verbose {
		fmt.Printf("removed environment: %s\n", name)
	}
	return nil
}

// listCommands prints a table of all commands in the state file.
func listCommands(_ *cli.Context) error {
	// Load the application state.
	appState, err := state.Load(statePath())
	if err != nil {
		return err
	}

	// If no commands are found, there's nothing else to do.
	if appState == nil || len(appState.CommandLines) == 0 {
		return nil
	}

	// Otherwise, print a pretty table of all commands.
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()
	tbl := table.New("Name", "Command")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)
	for name, command := range appState.CommandLines {
		tbl.AddRow(name, command)
	}

	tbl.Print()
	return nil
}

// addCommand adds a new command to the state file.
func addCommand(c *cli.Context) error {
	// Verify correct usage.
	if c.NArg() < 2 {
		return fmt.Errorf("expected minimum of 2 arguments, got %d", c.NArg())
	}
	name := c.Args().Get(0)
	command := c.Args().Tail()

	// Load the application state.
	appState, err := state.Load(statePath())
	if err != nil {
		return err
	}

	// If is a dry run, there's nothing else to do.
	if dryRun {
		return nil
	}

	// Add the command to the application state and save it back to the state file.
	if err := appState.AddCommandLine(name, command); err != nil {
		return err
	}
	if err := state.Save(appState, statePath()); err != nil {
		return err
	}

	// Success!
	if verbose {
		fmt.Printf("added command: %s\n", name)
	}
	return nil
}

// removeCommand removes a command from the state file.
func removeCommand(c *cli.Context) error {
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

	// Find the command in the application state.
	_, exists := appState.CommandLines[name]
	if !exists {
		return fmt.Errorf("command %s not found", name)
	}

	// If is a dry run, there's nothing else to do.
	if dryRun {
		return nil
	}

	// Remove the command from the application state and save it back to the state file.
	if err := appState.RemoveCommandLine(name); err != nil {
		return err
	}
	if err := state.Save(appState, statePath()); err != nil {
		return err
	}

	// Success!
	if verbose {
		fmt.Printf("removed command: %s\n", name)
	}
	return nil
}

// listConfigs prints a table of all configuration directories in the state file.
func listConfigs(_ *cli.Context) error {
	// Load the application state.
	appState, err := state.Load(statePath())
	if err != nil {
		return err
	}

	// If no configuration directories are found, there's nothing else to do.
	if appState == nil || len(appState.ConfigDirs) == 0 {
		return nil
	}

	// Otherwise, print a pretty table of all configuration directories.
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()
	tbl := table.New("Name", "Path")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)
	for name, path := range appState.ConfigDirs {
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
		if path, err = cache.AddRepo(cachePath(), name, path); err != nil {
			return err
		}
	}

	// Otherwise, add the configuration to the application state and save it back to the state file.
	if err := appState.AddConfigDir(name, path); err != nil {
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

	// Find the configuration directory in the application state.
	_, exists := appState.ConfigDirs[name]
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

	// Remove configuration directory from the application state and save it back to the state file.
	if err := appState.RemoveConfigDir(name); err != nil {
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

// showState prints the application state.
func showState(_ *cli.Context) error {
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

// showStatePath prints the path of the application state file.
func showStatePath(_ *cli.Context) error {
	fmt.Println(statePath())
	return nil
}

// getContext prints the active configuration context in the state file.
func getContext(_ *cli.Context) error {
	// Load the application state.
	appState, err := state.Load(statePath())
	if err != nil {
		return err
	}

	// Print the active context.
	fmt.Println(appState.Context)
	return nil
}

// setContext gets or sets the active configuration context in the state file.
func setContext(c *cli.Context) error {
	// Verify correct usage.
	if c.NArg() != 1 {
		return fmt.Errorf("expected 1 argument, got %d", c.NArg())
	}

	// Load the application state.
	appState, err := state.Load(statePath())
	if err != nil {
		return err
	}

	// If is a dry run, there's nothing else to do.
	if dryRun {
		return nil
	}

	// Otherwise, set the active context and save it back to the state file.
	appState.Context = c.Args().Get(0)
	return state.Save(appState, statePath())
}

// clearContext clears the active configuration context in the state file.
func clearContext(_ *cli.Context) error {
	// Load the application state.
	appState, err := state.Load(statePath())
	if err != nil {
		return err
	}

	// If is a dry run, there's nothing else to do.
	if dryRun {
		return nil
	}

	// Otherwise, clear the active context and save it back to the state file.
	appState.Context = ""
	return state.Save(appState, statePath())
}

// openEmacs opens emacs with the desired configuration and all provided arguments.
func openEmacs(c *cli.Context) error {
	// Load the application state.
	appState, err := state.Load(statePath())
	if err != nil {
		return err
	}

	// Ensure an active context is set.
	if context == "" {
		context = appState.Context
	}
	if context == "" {
		return noContextError
	}

	// Get the environment.
	env, ok := appState.Environments[context]
	if !ok {
		return fmt.Errorf("environment %s not found", context)
	}

	// Get the command line to execute.
	commandLine, ok := appState.CommandLines[env.CommandName]
	if !ok {
		return fmt.Errorf("command %s not found", env.CommandName)
	}

	// Get the config path to use.
	configDir, ok := appState.ConfigDirs[env.ConfigName]
	if !ok {
		return fmt.Errorf("configuration %s not found", env.ConfigName)
	}

	// Build the command line to execute.
	cmdline := strings.Split(commandLine, " ")
	cmdline = append(cmdline, "--init-directory", configDir)
	cmdline = append(cmdline, c.Args().Slice()...)

	// If is a dry run, print the command and return.
	if dryRun {
		fmt.Println(strings.Join(cmdline, " "))
		return nil
	}

	// Otherwise, execute the command.
	cmd := exec.Command(cmdline[0], cmdline[1:]...)
	return cmd.Run()
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
