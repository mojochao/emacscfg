// Package app provides the cli application using v2 of https://cli.urfave.org/.
package app

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/rodaine/table"
	"github.com/urfave/cli/v2"

	"github.com/mojochao/emacsctl/cache"
	"github.com/mojochao/emacsctl/config"
	"github.com/mojochao/emacsctl/errors"
	"github.com/mojochao/emacsctl/state"
	"github.com/mojochao/emacsctl/util"
)

// appDirFlag is the flag used to specify an alternate application directory
var appDirFlag = cli.StringFlag{
	Name:        "app-dir",
	Usage:       "Specify application directory",
	Destination: &config.AppDir,
	Value:       config.DefaultAppDir,
	EnvVars:     []string{"EMACSCFG_DIR"},
}

// dryRunFlag is the flag used to specify commands to be printed but not executed.
var dryRunFlag = cli.BoolFlag{
	Name:        "dry-run",
	Usage:       "Display the command that would be executed, but do not execute it",
	Destination: &config.DryRun,
}

// verboseFlag is the flag used to specify increased output.
var verboseFlag = cli.BoolFlag{
	Name:        "verbose",
	Aliases:     []string{"v"},
	Usage:       "Display verbose output",
	Destination: &config.Verbose,
}

// contextFlag is the flag used to provide name of an environment context to
// use instead of any active environment context found in the state.
var contextFlag = cli.StringFlag{
	Name:        "context",
	Aliases:     []string{"c"},
	Usage:       "Use a specific environment context",
	Destination: &config.Context,
}

// New creates a new cli application.
func New() *cli.App {
	return &cli.App{
		Name:        config.AppName,
		Usage:       "Manage multiple emacs environments",
		Description: config.AppDescription,
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
								Usage:   "Name of existing emacs command to use for environment",
							},
							&cli.StringFlag{
								Name:    "commandline",
								Aliases: []string{"cmdline"},
								Usage:   "New emacs command line to use for environment",
							},
							&cli.StringFlag{
								Name:    "config",
								Aliases: []string{"cfg"},
								Usage:   "Name of existing emacs configuration to use for environment",
							},
							&cli.StringFlag{
								Name:    "configdir",
								Aliases: []string{"cfgdir"},
								Usage:   "New emacs configuration directory path to use for environment",
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
						Usage:   "Display table of all emacs commands in application state",
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
						Usage:     "Remove an existing emacs command from application state",
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
						Usage:   "Display table of all emacs configurations in application state",
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
						Usage:     "Remove an existing emacs configuration from application state",
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
				Usage:  "Print application version",
				Action: showAppVersion,
			},
		},
	}
}

// listEnvironments prints a table of all environments in the state file.
func listEnvironments(_ *cli.Context) error {
	// Load the application state.
	appState, err := state.Load(config.StatePath())
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
	tbl := table.New("Name", "Command", "Config", "Description")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)
	for name, environment := range appState.Environments {
		tbl.AddRow(name, environment.CommandName, environment.ConfigName, environment.Description)
	}

	tbl.Print()
	return nil

}

// addEnvironment adds a new environment to the state file.
func addEnvironment(c *cli.Context) error {
	// Verify correct usage.
	if c.NArg() != 1 {
		return errors.UnexpectedNumArgsError{
			Expected: 1,
			Received: c.NArg(),
		}
	}
	name := c.Args().Get(0)

	// Load the application state.
	appState, err := state.Load(config.StatePath())
	if err != nil {
		return err
	}

	// If is a dry run, there's nothing else to do.
	if config.DryRun {
		return nil
	}

	// Get the optional command line, configuration directory, and description from the flags.
	commandName := c.String("command")
	if _, ok := appState.Commands[commandName]; !ok {
		return errors.CommandNotFoundError{Name: commandName}
	}

	configName := c.String("config")
	if _, ok := appState.Configs[configName]; !ok {
		return errors.ConfigNotFoundError{Name: configName}
	}

	description := c.String("description")
	if description == "" {
		description = "Not specified"
	}

	// Add the environment to the application state and save it back to the state file.
	if err := appState.AddEnvironment(name, commandName, configName, description); err != nil {
		return err
	}
	if err := state.Save(appState, config.StatePath()); err != nil {
		return err
	}

	// Success!
	if config.Verbose {
		fmt.Printf("added environment: %s\n", name)
	}
	return nil
}

// removeEnvironment removes an environment from the state file.
func removeEnvironment(c *cli.Context) error {
	// Verify correct usage.
	if c.NArg() != 1 {
		return errors.UnexpectedNumArgsError{Expected: 1, Received: c.NArg()}
	}
	name := c.Args().Get(0)

	// Load the application state.
	appState, err := state.Load(config.StatePath())
	if err != nil {
		return err
	}

	// Find the environment in the application state.
	if _, exists := appState.Environments[name]; !exists {
		return errors.EnvironmentNotFoundError{Name: name}
	}

	// If is a dry run, there's nothing else to do.
	if config.DryRun {
		return nil
	}

	// Remove the environment from the application state and save it back to the state file.
	if err := appState.RemoveEnvironment(name); err != nil {
		return err
	}
	if err := state.Save(appState, config.StatePath()); err != nil {
		return err
	}

	// Success!
	if config.Verbose {
		fmt.Printf("removed environment: %s\n", name)
	}
	return nil
}

// listCommands prints a table of all commands in the state file.
func listCommands(_ *cli.Context) error {
	// Load the application state.
	appState, err := state.Load(config.StatePath())
	if err != nil {
		return err
	}

	// If no commands are found, there's nothing else to do.
	if appState == nil || len(appState.Commands) == 0 {
		return nil
	}

	// Otherwise, print a pretty table of all commands.
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()
	tbl := table.New("Name", "Path", "Args", "Description")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)
	for name, command := range appState.Commands {
		tbl.AddRow(name, command.BinPath, strings.Join(command.BinArgs, " "), command.Description)
	}

	tbl.Print()
	return nil
}

// addCommand adds a new command to the state file.
func addCommand(c *cli.Context) error {
	// Verify correct usage.
	if c.NArg() < 2 {
		return errors.MinimumNumArgsError{Minimum: 2, Received: c.NArg()}
	}
	name := c.Args().Get(0)
	command := c.Args().Tail()
	description := c.String("description")

	// Load the application state.
	appState, err := state.Load(config.StatePath())
	if err != nil {
		return err
	}

	// If is a dry run, there's nothing else to do.
	if config.DryRun {
		return nil
	}

	// Add the command to the application state and save it back to the state file.
	if err := appState.AddCommand(name, command, description); err != nil {
		return err
	}
	if err := state.Save(appState, config.StatePath()); err != nil {
		return err
	}

	// Success!
	if config.Verbose {
		fmt.Printf("added command: %s\n", name)
	}
	return nil
}

// removeCommand removes a command from the state file.
func removeCommand(c *cli.Context) error {
	// Verify correct usage.
	if c.NArg() != 1 {
		return errors.UnexpectedNumArgsError{Expected: 1, Received: c.NArg()}
	}
	name := c.Args().Get(0)

	// Load the application state.
	appState, err := state.Load(config.StatePath())
	if err != nil {
		return err
	}

	// Find the command in the application state.
	if _, exists := appState.Commands[name]; !exists {
		return errors.CommandNotFoundError{Name: name}
	}

	// If is a dry run, there's nothing else to do.
	if config.DryRun {
		return nil
	}

	// Remove the command from the application state and save it back to the state file.
	if err := appState.RemoveCommand(name); err != nil {
		return err
	}
	if err := state.Save(appState, config.StatePath()); err != nil {
		return err
	}

	// Success!
	if config.Verbose {
		fmt.Printf("removed command: %s\n", name)
	}
	return nil
}

// listConfigs prints a table of all configuration directories in the state file.
func listConfigs(_ *cli.Context) error {
	// Load the application state.
	appState, err := state.Load(config.StatePath())
	if err != nil {
		return err
	}

	// If no configuration directories are found, there's nothing else to do.
	if appState == nil || len(appState.Configs) == 0 {
		return nil
	}

	// Otherwise, print a pretty table of all configuration directories.
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()
	tbl := table.New("Name", "Path", "Description")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)
	for name, cfg := range appState.Configs {
		tbl.AddRow(name, cfg.InitDir, cfg.Description)
	}

	tbl.Print()
	return nil
}

// addConfig adds a new configuration to the state file.
func addConfig(c *cli.Context) error {
	// Verify correct usage.
	if c.NArg() != 2 {
		return errors.UnexpectedNumArgsError{Expected: 2, Received: c.NArg()}
	}
	name := c.Args().Get(0)
	path := c.Args().Get(1)
	description := c.String("description")

	// Load the application state.
	appState, err := state.Load(config.StatePath())
	if err != nil {
		return err
	}

	// If is a dry run, there's nothing else to do.
	if config.DryRun {
		return nil
	}

	// If the path is a git URL, add the repository to the cache.
	if util.IsGitURL(path) {
		// Add the repository to the cache.
		url := path
		cacheDir := config.CachePath()
		if err := util.EnsureDir(cacheDir); err != nil {
			return err
		}
		if path, err = cache.AddRepo(cacheDir, name, url); err != nil {
			return err
		}
	}

	// Otherwise, add the configuration to the application state and save it back to the state file.
	if err := appState.AddConfig(name, path, description); err != nil {
		return err
	}
	if err := state.Save(appState, config.StatePath()); err != nil {
		return err
	}

	// Success!
	if config.Verbose {
		fmt.Printf("added configuration: %s\n", name)
	}
	return nil

}

// removeConfig removes a configuration from the state file.
func removeConfig(c *cli.Context) error {
	// Verify correct usage.
	if c.NArg() != 1 {
		return errors.UnexpectedNumArgsError{Expected: 1, Received: c.NArg()}
	}
	name := c.Args().Get(0)

	// Load the application state.
	appState, err := state.Load(config.StatePath())
	if err != nil {
		return err
	}

	// Find the config in the application state.
	if _, exists := appState.Configs[name]; !exists {
		return errors.ConfigNotFoundError{Name: name}
	}

	// If is a dry run, there's nothing else to do.
	if config.DryRun {
		return nil
	}

	// Otherwise, remove any cached repository from the filesystem.
	cacheDir := config.CachePath()
	if cache.IsCached(cacheDir, name) {
		if err := cache.RemoveRepo(cacheDir, name); err != nil {
			return err
		}
	}

	// Remove config from the application state and save it back to the state file.
	if err := appState.RemoveConfig(name); err != nil {
		return err
	}
	if err := state.Save(appState, config.StatePath()); err != nil {
		return err
	}

	// Success!
	if config.Verbose {
		fmt.Printf("removed configuration: %s\n", name)
	}
	return nil
}

// showState prints the application state.
func showState(_ *cli.Context) error {
	// Load the application state and print it to stdout.
	appState, err := state.Load(config.StatePath())
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
	fmt.Println(config.StatePath())
	return nil
}

// getContext prints the active configuration context in the state file.
func getContext(_ *cli.Context) error {
	// Load the application state.
	appState, err := state.Load(config.StatePath())
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
		return errors.UnexpectedNumArgsError{Expected: 1, Received: c.NArg()}
	}

	// Load the application state.
	appState, err := state.Load(config.StatePath())
	if err != nil {
		return err
	}

	// If is a dry run, there's nothing else to do.
	if config.DryRun {
		return nil
	}

	// Otherwise, set the active context and save it back to the state file.
	appState.Context = c.Args().Get(0)
	return state.Save(appState, config.StatePath())
}

// clearContext clears the active configuration context in the state file.
func clearContext(_ *cli.Context) error {
	// Load the application state.
	appState, err := state.Load(config.StatePath())
	if err != nil {
		return err
	}

	// If is a dry run, there's nothing else to do.
	if config.DryRun {
		return nil
	}

	// Otherwise, clear the active context and save it back to the state file.
	appState.Context = ""
	return state.Save(appState, config.StatePath())
}

// openEmacs opens emacs with the desired configuration and all provided arguments.
func openEmacs(_ *cli.Context) error {
	// Load the application state.
	appState, err := state.Load(config.StatePath())
	if err != nil {
		return err
	}

	// Ensure an active context is set.
	context := config.Context
	if context == "" {
		context = appState.Context
	}
	if context == "" {
		return errors.NoContextError
	}

	// Get the environment.
	env, ok := appState.Environments[context]
	if !ok {
		return errors.EnvironmentNotFoundError{Name: context}
	}

	// Get the command to use.
	cmd, ok := appState.Commands[env.CommandName]
	if !ok {
		return errors.CommandNotFoundError{Name: env.CommandName}
	}

	// Get the config to use.
	cfg, ok := appState.Configs[env.ConfigName]
	if !ok {
		return errors.ConfigNotFoundError{Name: env.ConfigName}
	}

	// Build the command line to execute.
	cmdLine := cmd.CommandLine(cfg.InitDir)

	// If is a dry run, print the command line and return.
	if config.DryRun {
		fmt.Println(strings.Join(cmdLine, " "))
		return nil
	}

	// Otherwise, execute the command.
	return exec.Command(cmdLine[0], cmdLine[1:]...).Run()
}

// showAppVersion prints the version of the application set at build time by
// the `go build -ldflags "-X github.com/mojochao/emacsctl/app.version=0.10.0" -o emacsctl .` command.
var version string

// showAppVersion prints the version of the application.
func showAppVersion(_ *cli.Context) error {
	fmt.Printf("emacsctl version %s\n", version)
	if !config.Verbose {
		return nil
	}

	buildInfo := util.GetBuildInfo()
	if buildInfo == nil {
		return nil
	}

	for key, value := range buildInfo {
		fmt.Printf("%s: %s\n", key, value)
	}
	return nil
}
