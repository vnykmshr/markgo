// Package main provides the unified CLI entry point for all MarkGo commands.
package main

import (
	"fmt"
	"os"

	initcmd "github.com/vnykmshr/markgo/internal/commands/init"
	"github.com/vnykmshr/markgo/internal/commands/new"
	"github.com/vnykmshr/markgo/internal/commands/serve"
	"github.com/vnykmshr/markgo/internal/constants"
)

func main() {
	// If no arguments provided, default to serve
	if len(os.Args) < 2 {
		serve.Run(os.Args)
		return
	}

	command := os.Args[1]

	// Handle help and version flags
	if command == "-h" || command == "--help" || command == "help" {
		showHelp()
		return
	}

	if command == "-v" || command == "--version" || command == "version" {
		fmt.Printf("MarkGo %s\n", constants.AppVersion)
		if constants.GitCommit != "unknown" {
			fmt.Printf("  commit: %s\n  built:  %s\n", constants.GitCommit, constants.BuildTime)
		}
		return
	}

	// Route to appropriate subcommand
	// Remove the subcommand from args for the subcommand handler
	subArgs := append([]string{os.Args[0]}, os.Args[2:]...)

	switch command {
	case "serve", "server", "start":
		serve.Run(subArgs)
	case "init", "initialize":
		initcmd.Run(subArgs)
	case "new", "create", "article":
		new.Run(subArgs)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		showHelp()
		os.Exit(1)
	}
}

func showHelp() {
	fmt.Printf(`MarkGo %s - File-based blog engine

USAGE:
    markgo [command] [flags]

COMMANDS:
    serve       Start the blog server (default)
    init        Initialize a new blog
    new         Create a new article
    version     Show version information
    help        Show this help message

ALIASES:
    server, start    → serve
    initialize       → init
    create, article  → new

Run 'markgo [command] --help' for more information on a specific command.

EXAMPLES:
    markgo                          # Start server (default command)
    markgo serve                    # Start server explicitly
    markgo serve --help             # Show server options
    markgo init --quick             # Quick blog initialization
    markgo new --title "Hello"      # Create new article

For more information, visit: https://github.com/vnykmshr/markgo
`, constants.AppVersion)
}
