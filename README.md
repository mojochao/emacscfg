# emacscfg - A simple CLI to manage and use multiple emacs configurations

## Features

- listing all managed configurations
- adding a managed configuration by name
- removing a managed configuration by name
- setting the active managed configuration by name
- switching between managed configurations by name
- getting the path of a managed configuration
- starting emacs with the default or named configuration

## Requirements

- emacs 29.1 or later
- git (optional, required if you want to clone a configuration from a git repository URL)

## Installation

```text
go install github.com/mojochao/emacscfg@latest
```

## Usage

Display help information on all commands and options with the `help` subcommand:

```text
$ emacscfg help
NAME:
   emacscfg - Manage multiple emacs configuration profiles

USAGE:
   emacscfg [global options] command [command options] 

COMMANDS:
   state         Display application state
   context, ctx  Get or set the active configuration context in application state
   list, ls      Display table of all configurations in application state
   add           Add a new configuration to application state
   remove, rm    Remove a configuration from application state
   path, dir     Print the path of the configuration directory
   open          Open files in emacs with the desired configuration
   version       Print the version of the application
   help, h       Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --app-dir value  Specify application directory (default: "~/.config/emacscfg") [$EMACSCFG_DIR]
   --dry-run        Display the command that would be executed, but do not execute it (default: false)
   --verbose, -v    Display verbose output (default: false)
   --help, -h       show help
```

The `help` subcommand can also be used to display help information for a specific subcommand:

```text
$ emacscfg add help
```

List all managed configurations with the `list` subcommand:

```text
$ emacscfg list
```

Add a managed configuration with the `add` subcommand:

```text
$ emacscfg add my-config /path/to/my/emacs-config
```

If you pass a URL as the configuration path, the `add` subcommand will clone
the repository to the application repositories cache directory and set the
configuration path to the location of the cloned repository.

```text
$ emacscfg add my-de https://github.com/mojochao/myde.el
```

Remove a managed configuration with the `remove` subcommand:

```text
$ emacscfg remove my-config
```

Get the active managed configuration context with the `context` subcommand:

```text
$ emacscfg context
```

Set the active managed configuration context with a configuration name argument:

```text
$ emacscfg context my-config
```

Get the path of the active managed configuration context with the `path` subcommand:

```text
$ emacscfg path
```

Get the path of any managed configuration by adding the `--context` flag:

```text
$ emacscfg path --context my-config
```

Open emacs with the active managed configuration with the `open` subcommand:

```text
$ emacscfg open my-file-1 my-file-2
```

Edit files in emacs with any managed configuration by adding the `--context` flag:

```text
$ emacscfg open --context my-config my-file-1 my-file-2
```

That's all folks!
