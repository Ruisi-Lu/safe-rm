package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/user/safe-rm/internal/cli"
	"github.com/user/safe-rm/internal/config"
	"github.com/user/safe-rm/internal/protect"
	"github.com/user/safe-rm/internal/restore"
	"github.com/user/safe-rm/internal/trash"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "safe-rm: warning: failed to load config: %v\n", err)
		cfg = config.Default()
	}

	opts, err := cli.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "safe-rm: %v\n", err)
		os.Exit(1)
	}

	// Handle --help and --version (already printed, just exit cleanly)
	if opts.ExitClean {
		return
	}

	// Handle special safe-rm subcommands
	switch {
	case opts.SafeList:
		if err := restore.List(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "safe-rm: %v\n", err)
			os.Exit(1)
		}
		return
	case opts.SafeRestore != "":
		if err := restore.Restore(cfg, opts.SafeRestore); err != nil {
			fmt.Fprintf(os.Stderr, "safe-rm: %v\n", err)
			os.Exit(1)
		}
		return
	case opts.SafePurge:
		if err := restore.Purge(cfg, opts.PurgeDays); err != nil {
			fmt.Fprintf(os.Stderr, "safe-rm: %v\n", err)
			os.Exit(1)
		}
		return
	case opts.SafeEmpty:
		if err := restore.Empty(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "safe-rm: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// No files specified
	if len(opts.Files) == 0 {
		if !opts.Force {
			fmt.Fprintln(os.Stderr, "safe-rm: missing operand")
			os.Exit(1)
		}
		return
	}

	// Process each file/directory
	exitCode := 0
	for _, path := range opts.Files {
		if err := processPath(cfg, opts, path); err != nil {
			fmt.Fprintf(os.Stderr, "safe-rm: cannot remove '%s': %v\n", path, err)
			exitCode = 1
			if !opts.Force {
				continue
			}
		}
	}

	os.Exit(exitCode)
}

func processPath(cfg *config.Config, opts *cli.Options, path string) error {
	// Get absolute path for protection checking
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	// Check file/directory existence
	info, err := os.Lstat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			if opts.Force {
				return nil // -f ignores nonexistent files
			}
			return fmt.Errorf("No such file or directory")
		}
		return err
	}

	// Check if it's a directory without -r flag
	if info.IsDir() && !opts.Recursive {
		if opts.RemoveEmptyDirs {
			// -d flag: try to remove empty directory
			entries, err := os.ReadDir(absPath)
			if err != nil {
				return err
			}
			if len(entries) > 0 {
				return fmt.Errorf("Directory not empty")
			}
		} else {
			return fmt.Errorf("Is a directory")
		}
	}

	// Check protection rules
	status := protect.Check(cfg, absPath, opts.Recursive)
	if status.Protected {
		if cfg.ProtectedBehavior == "block" {
			return fmt.Errorf("BLOCKED: %s\n  Reason: %s\n  This path is protected and cannot be removed.", absPath, status.Reason)
		}

		// Require confirmation
		if !opts.Force {
			fmt.Fprintf(os.Stderr, "WARNING: You are about to remove a protected path!\n")
			fmt.Fprintf(os.Stderr, "  Path: %s\n", absPath)
			fmt.Fprintf(os.Stderr, "  Reason: %s\n", status.Reason)
			fmt.Fprintf(os.Stderr, "Type 'yes I am sure' to confirm: ")

			var response string
			fmt.Scanln(&response)
			if response != "yes I am sure" {
				return fmt.Errorf("aborted by user")
			}
		} else {
			// Even with -f, block protected paths unless explicitly confirmed
			return fmt.Errorf("BLOCKED: %s is protected (%s). Use interactive mode to confirm.", absPath, status.Reason)
		}
	}

	// Interactive mode (-i)
	if opts.Interactive && !opts.Force {
		fmt.Fprintf(os.Stderr, "remove '%s'? ", path)
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "yes" {
			return nil
		}
	}

	// Move to trash instead of permanent deletion
	trashPath, err := trash.Move(cfg, absPath)
	if err != nil {
		return fmt.Errorf("failed to move to trash: %v", err)
	}

	if opts.Verbose {
		fmt.Printf("removed '%s' (moved to trash: %s)\n", path, trashPath)
	}

	return nil
}
