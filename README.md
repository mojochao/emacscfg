# emacsctl - A simple CLI to manage and use multiple emacs commands and configurations

## Features

- managing available emacs commands (emacs binary and command line options)
- managing available emacs configurations (emacs config directories)
- managing available emacs environments (a combination of a command and a configuration)
- managing the active environment context
- opening files in the desired emacs environment

## Requirements

- emacs 29.1 or later
- git (optional, required if you want to clone a configuration from a git repository URL)

## Installation

```text
go install github.com/mojochao/emacsctl@latest
```

## Usage

Display help information on all commands and options with the `help` subcommand
or the `-h` or `--help` global options:

```text
$ emacsctl help
NAME:
   emacsctl - Manage multiple emacs environments

USAGE:
   emacsctl [global options] command [command options] 

COMMANDS:
   state             Display application state
   environment, env  Manage emacs environments
   command, cmd      Manage emacs command lines
   config, cfg       Manage emacs configuration directories in application state
   context, ctx      Manage active environment context in application state
   open, edit        Open files in the desired emacs environment
   version           Print the version of the application
   help, h           Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --app-dir value  Specify application directory (default: "~/.config/emacsctl") [$EMACSCFG_DIR]
   --dry-run        Display the command that would be executed, but do not execute it (default: false)
   --verbose, -v    Display verbose output (default: false)
   --help, -h       show help
```

This can also be used to display help information for a specific subcommand:

```text
$ emacsctl environment add help
```

List all environments with the `environment list` subcommand:

```text
$ emacsctl env list
```

There will be none initially, so let's add one.

Add a managed configuration with the `add` subcommand:

```text
$ emacsctl add my-config /path/to/my/emacs-config
```

If you pass a URL as the configuration path, the `add` subcommand will clone
the repository to the application repositories cache directory and set the
configuration path to the location of the cloned repository.

```text
$ emacsctl add my-de https://github.com/mojochao/myde.el
```

Remove a managed configuration with the `remove` subcommand:

```text
$ emacsctl remove my-config
```

Get the active managed configuration context with the `context` subcommand:

```text
$ emacsctl context
```

Set the active managed configuration context with a configuration name argument:

```text
$ emacsctl context my-config
```

Get the path of the active managed configuration context with the `path` subcommand:

```text
$ emacsctl path
```

Get the path of any managed configuration by adding the `--context` flag:

```text
$ emacsctl path --context my-config
```

Open emacs with the active managed configuration with the `open` subcommand:

```text
$ emacsctl open my-file-1 my-file-2
```

Edit files in emacs with any managed configuration by adding the `--context` flag:

```text
$ emacsctl open --context my-config my-file-1 my-file-2
```

That's all folks!
