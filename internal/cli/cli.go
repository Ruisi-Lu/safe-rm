package cli

import (
	"fmt"
	"strings"
)

// Options represents parsed command-line options
type Options struct {
	// Standard rm flags
	Force           bool     // -f, --force
	Interactive     bool     // -i
	InteractiveOnce bool     // -I
	Recursive       bool     // -r, -R, --recursive
	RemoveEmptyDirs bool     // -d, --dir
	Verbose         bool     // -v, --verbose
	PreserveRoot    bool     // --preserve-root (default true)
	NoPreserveRoot  bool     // --no-preserve-root
	Files           []string // Files/directories to remove

	// Safe-rm specific flags
	SafeList    bool   // --safe-list
	SafeRestore string // --safe-restore=PATH
	SafePurge   bool   // --safe-purge
	SafeEmpty   bool   // --safe-empty (empty entire trash)
	PurgeDays   int    // --purge-days=N (default 30)

	// Internal flags
	ExitClean bool // Set when --help or --version is used
}

// Parse parses command-line arguments and returns Options
func Parse(args []string) (*Options, error) {
	opts := &Options{
		PreserveRoot: true, // Default to preserve root
		PurgeDays:    30,   // Default purge days
	}

	i := 0
	for i < len(args) {
		arg := args[i]

		if arg == "--" {
			// Everything after -- is a file
			opts.Files = append(opts.Files, args[i+1:]...)
			break
		}

		if strings.HasPrefix(arg, "--") {
			// Long option
			if err := parseLongOption(opts, arg, args, &i); err != nil {
				return nil, err
			}
		} else if strings.HasPrefix(arg, "-") && len(arg) > 1 {
			// Short option(s)
			if err := parseShortOptions(opts, arg[1:]); err != nil {
				return nil, err
			}
		} else {
			// File argument
			opts.Files = append(opts.Files, arg)
		}
		i++
	}

	return opts, nil
}

func parseLongOption(opts *Options, arg string, args []string, i *int) error {
	// Handle --option=value format
	var value string
	if idx := strings.Index(arg, "="); idx != -1 {
		value = arg[idx+1:]
		arg = arg[:idx]
	}

	switch arg {
	case "--force":
		opts.Force = true
	case "--interactive":
		opts.Interactive = true
	case "--recursive":
		opts.Recursive = true
	case "--dir":
		opts.RemoveEmptyDirs = true
	case "--verbose":
		opts.Verbose = true
	case "--preserve-root":
		opts.PreserveRoot = true
		opts.NoPreserveRoot = false
	case "--no-preserve-root":
		opts.NoPreserveRoot = true
		opts.PreserveRoot = false
	case "--safe-list":
		opts.SafeList = true
	case "--safe-restore":
		if value == "" {
			return fmt.Errorf("--safe-restore requires a path argument")
		}
		opts.SafeRestore = value
	case "--safe-purge":
		opts.SafePurge = true
	case "--safe-empty":
		opts.SafeEmpty = true
	case "--purge-days":
		if value == "" {
			return fmt.Errorf("--purge-days requires a number argument")
		}
		var days int
		if _, err := fmt.Sscanf(value, "%d", &days); err != nil {
			return fmt.Errorf("--purge-days: invalid number: %s", value)
		}
		opts.PurgeDays = days
	case "--help":
		printHelp()
		opts.ExitClean = true
		return nil
	case "--version":
		fmt.Println("safe-rm version 1.0.0")
		opts.ExitClean = true
		return nil
	default:
		return fmt.Errorf("unrecognized option '%s'", arg)
	}

	return nil
}

func parseShortOptions(opts *Options, flags string) error {
	for _, flag := range flags {
		switch flag {
		case 'f':
			opts.Force = true
		case 'i':
			opts.Interactive = true
		case 'I':
			opts.InteractiveOnce = true
		case 'r', 'R':
			opts.Recursive = true
		case 'd':
			opts.RemoveEmptyDirs = true
		case 'v':
			opts.Verbose = true
		default:
			return fmt.Errorf("invalid option -- '%c'", flag)
		}
	}
	return nil
}

func printHelp() {
	help := `Usage: rm [OPTION]... [FILE]...
Remove (move to trash) the FILE(s).

This is safe-rm, a safer replacement for the standard rm command.
Instead of permanently deleting files, they are moved to a trash directory.

Standard options:
  -f, --force           ignore nonexistent files and arguments
  -i                    prompt before every removal
  -I                    prompt once before removing more than three files
  -r, -R, --recursive   remove directories and their contents recursively
  -d, --dir             remove empty directories
  -v, --verbose         explain what is being done
      --preserve-root   do not remove '/' (default)
      --no-preserve-root  do not treat '/' specially

Safe-rm options:
      --safe-list           list all items in the trash
      --safe-restore=PATH   restore a file from trash to its original location
      --safe-purge          purge old items from trash
      --purge-days=N        with --safe-purge, remove items older than N days (default 30)
      --safe-empty          permanently delete ALL items in trash (requires confirmation)

      --help     display this help and exit
      --version  output version information and exit

Protected paths (will require confirmation or be blocked):
  - Root directory (/) and top-level system directories
  - .git directories
  - Paths specified in ~/.config/safe-rm/config.yml

Environment variables:
  SAFERM_TRASH           Override trash directory location
  SAFERM_PROTECTED_PATHS Additional protected paths (colon-separated)

For more information, see: https://github.com/user/safe-rm
`
	fmt.Print(help)
}
