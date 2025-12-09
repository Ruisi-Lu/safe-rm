# safe-rm

A safer replacement for the Unix `rm` command that moves files to trash instead of permanently deleting them.

## Overview

`safe-rm` is designed to protect against accidental file deletion, whether caused by human error or automated systems such as AI agents issuing filesystem commands. Instead of permanently removing files, this tool moves them to a configurable trash directory while preserving their original paths for easy restoration.

### Why This Exists

Modern development workflows increasingly involve AI-powered tools and automation that may execute shell commands. A single misconfigured script or an AI agent hallucinating a dangerous `rm -rf` command can cause catastrophic data loss. This tool provides a safety net by:

- Moving files to trash instead of deleting them permanently
- Blocking dangerous operations like `rm -rf /` by default
- Protecting system directories, Git repositories, and user-defined paths
- Maintaining familiar `rm` command syntax and behavior

## Features

- **Drop-in replacement**: Use as `rm` with all standard flags (`-r`, `-f`, `-i`, `-v`, etc.)
- **Protected paths**: Built-in protection for system directories and `.git` folders
- **Trash-based deletion**: Files are moved to trash, not permanently deleted
- **Path preservation**: Original directory structure is maintained in trash
- **Metadata tracking**: Deletion timestamp and hostname recorded for each item
- **Easy restoration**: Restore files to their original location
- **Configurable**: Customize trash location and protected paths via config file or environment variables
- **AI-safe**: Designed to protect against automated deletion mistakes

## Installation

### Pre-compiled Binaries (Recommended)

Download the latest pre-compiled binary from the [GitHub Releases](https://github.com/Ruisi-Lu/safe-rm/releases) page:

```bash
# Linux (amd64)
curl -L https://github.com/Ruisi-Lu/safe-rm/releases/latest/download/rm-linux-amd64 -o rm
chmod +x rm

# Linux (arm64)
curl -L https://github.com/Ruisi-Lu/safe-rm/releases/latest/download/rm-linux-arm64 -o rm
chmod +x rm

# macOS (Apple Silicon)
curl -L https://github.com/Ruisi-Lu/safe-rm/releases/latest/download/rm-darwin-arm64 -o rm
chmod +x rm

# macOS (Intel)
curl -L https://github.com/Ruisi-Lu/safe-rm/releases/latest/download/rm-darwin-amd64 -o rm
chmod +x rm
```

Then move the binary to your PATH:

```bash
mv rm ~/.local/bin/rm
```

### Building from Source

```bash
git clone https://github.com/Ruisi-Lu/safe-rm.git
cd safe-rm
go build -o rm ./cmd/rm
```

### Installing as System rm

**Option 1: PATH Priority (Recommended)**

Add the safe-rm binary to a directory that appears before `/usr/bin` in your PATH:

```bash
# Create local bin directory
mkdir -p ~/.local/bin

# Build and install
go build -o ~/.local/bin/rm ./cmd/rm

# Ensure ~/.local/bin is first in PATH (add to ~/.bashrc or ~/.zshrc)
export PATH="$HOME/.local/bin:$PATH"
```

**Option 2: Replace System rm (Advanced)**

Back up the original binary and replace it:

```bash
# Backup original rm
sudo mv /usr/bin/rm /usr/bin/rm.real

# Install safe-rm as rm
sudo go build -o /usr/bin/rm ./cmd/rm
sudo chmod +x /usr/bin/rm
```

To access the original rm when needed:

```bash
/usr/bin/rm.real -rf /tmp/unwanted
```

## Usage

### Basic Usage

`safe-rm` accepts all standard `rm` flags:

```bash
# Remove a file (moves to trash)
rm file.txt

# Remove a directory recursively
rm -r directory/

# Force remove without prompts
rm -f file.txt

# Verbose output
rm -v file.txt

# Combined flags
rm -rf directory/
```

### Safe-rm Specific Commands

```bash
# List all items in trash
rm --safe-list

# Restore a file to its original location
rm --safe-restore=/home/user/documents/file.txt

# Purge items older than 30 days (default)
rm --safe-purge

# Purge items older than 7 days
rm --safe-purge --purge-days=7

# Permanently delete ALL items in trash (requires confirmation)
rm --safe-empty
```

### Protected Path Behavior

When attempting to delete a protected path:

```bash
$ rm -rf /home
safe-rm: cannot remove '/home': BLOCKED: /home
  Reason: System directory is protected: /home
  This path is protected and cannot be removed.

$ rm -rf .git
safe-rm: cannot remove '.git': BLOCKED: /path/to/.git
  Reason: .git directory or repository root is protected
  This path is protected and cannot be removed.
```

With `protected_behavior: confirm` in config, you can confirm dangerous operations:

```
WARNING: You are about to remove a protected path!
  Path: /home/user/.git
  Reason: .git directory or repository root is protected
Type 'yes I am sure' to confirm: 
```

## Configuration

### Configuration File

Create `~/.config/safe-rm/config.yml`:

```yaml
# Trash directory location
trash_dir: ~/.local/share/safe-rm/trash

# Auto-purge items older than this many days
retention_days: 30

# Additional protected paths (glob patterns)
protected_paths:
  - ~/.ssh/*
  - ~/Documents/**
  - /etc/passwd

# Behavior for protected paths: "block" or "confirm"
protected_behavior: confirm

# Show detailed warnings
verbose_warnings: true
```

### Environment Variables

Environment variables take precedence over config file settings:

| Variable | Description | Example |
|----------|-------------|---------|
| `SAFERM_TRASH` | Trash directory path | `/var/trash/safe-rm` |
| `SAFERM_PROTECTED_PATHS` | Additional protected paths (colon-separated) | `/data/important:/backup` |
| `SAFERM_RETENTION_DAYS` | Retention period in days | `7` |
| `SAFERM_PROTECTED_BEHAVIOR` | `block` or `confirm` | `block` |

### Protected Paths

The following paths are protected by default:

- Root directory: `/`
- System directories: `/bin`, `/boot`, `/dev`, `/etc`, `/home`, `/lib`, `/lib64`, `/opt`, `/proc`, `/root`, `/run`, `/sbin`, `/srv`, `/sys`, `/tmp`, `/usr`, `/var`
- Any `.git` directory

## Trash Structure

Files are moved to trash preserving their original path:

```
~/.local/share/safe-rm/trash/
  <hostname>/
    home/
      user/
        documents/
          file.txt
          file.txt.saferm-meta
```

Each trashed item has a corresponding `.saferm-meta` file:

```json
{
  "original_path": "/home/user/documents/file.txt",
  "deleted_at": "2025-12-10T03:15:00+08:00",
  "hostname": "myhost",
  "is_directory": false
}
```

## Restoring Files

### Using safe-rm

```bash
# List trashed items
rm --safe-list

# Restore a specific file
rm --safe-restore=/home/user/documents/file.txt
```

### Manual Restoration

Files can also be restored manually by copying from the trash directory:

```bash
# Find the trashed file
ls ~/.local/share/safe-rm/trash/$(hostname)/home/user/documents/

# Copy back to original location
cp ~/.local/share/safe-rm/trash/$(hostname)/home/user/documents/file.txt \
   /home/user/documents/file.txt
```

## AI Automation Safety

This tool is specifically designed to protect against AI-driven automation mistakes. When integrating AI tools with live shells or CI/CD pipelines, consider:

1. **Install safe-rm as the default rm**: Ensure all automated commands use the safe version
2. **Set protected_behavior to block**: Prevent any confirmation bypass in automated contexts
3. **Define critical paths**: Add project-specific protected paths for important data
4. **Monitor trash usage**: Set up alerts if trash grows unexpectedly large

Example CI/CD integration:

```yaml
# In your CI/CD pipeline
- name: Install safe-rm
  run: |
    go install github.com/Ruisi-Lu/safe-rm/cmd/rm@latest
    export PATH="$(go env GOPATH)/bin:$PATH"
    
- name: Set protected behavior
  run: |
    export SAFERM_PROTECTED_BEHAVIOR=block
    export SAFERM_PROTECTED_PATHS="${PWD}/.git:${PWD}/src"
```

## Development

### Running Tests

```bash
go test ./...
```

### Building

```bash
# Build for current platform
go build -o rm ./cmd/rm

# Build for Linux
GOOS=linux GOARCH=amd64 go build -o rm-linux-amd64 ./cmd/rm
```

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
