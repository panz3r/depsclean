# depsclean

> A fast, interactive CLI/TUI for discovering and removing dependency directories (`node_modules`, `.venv`, `vendor`, and more) to reclaim disk space.
>
> Inspired by [npkill](https://github.com/voidcosmos/npkill) and [bunkill](https://github.com/codingstark-dev/bunkill).

---

## Features

- **Streaming discovery** — results appear immediately as the scan walks your filesystem
- **Async enrichment** — size, age, and package metadata load in the background without blocking navigation
- **Full-screen TUI** — keyboard-driven list with compact/details row modes, search, sort cycling, and a details panel
- **Multi-select and range selection** — delete single items, a selected set, or use `delete-all` for batch removal
- **Dry-run mode** — simulate any deletion safely before committing
- **Multiple scan profiles** — Node, Python, Rust, Go, PHP, Ruby, Java, or "all"
- **JSON / NDJSON output** — machine-friendly output for scripting and pipelines
- **Safe deletion guards** — basename allowlist, absolute-path enforcement, and empty-targets guard prevent accidental deletion

---

## Installation

### Pre-built Binaries (Recommended)

Download the latest pre-built binary for your platform from the [GitHub Releases](https://github.com/panz3r/depsclean/releases) page.

#### Linux/macOS

```bash
# Download and install (replace $VERSION with the latest version)
curl -L -o depsclean "https://github.com/panz3r/depsclean/releases/download/v$VERSION/depsclean_linux_amd64"

# Make executable
chmod +x depsclean

# Move to PATH (optional)
sudo mv depsclean /usr/local/bin/
```

#### macOS note

If you install the binary manually on macOS, remove the quarantine attribute before first use:

```bash
xattr -dr com.apple.quarantine /usr/local/bin/depsclean
```

#### Windows

1. Download `depsclean_windows_amd64.exe` from the [GitHub Releases](https://github.com/panz3r/depsclean/releases) page
2. Rename it to `depsclean.exe` if you want
3. Place it in a directory on your `PATH`, or run it directly

#### Platform-specific downloads

- **Linux AMD64**: `depsclean_linux_amd64`
- **Linux ARM64**: `depsclean_linux_arm64`
- **macOS Intel**: `depsclean_macos_intel`
- **macOS Apple Silicon**: `depsclean_macos_arm64`
- **Windows AMD64**: `depsclean_windows_amd64.exe`
- **Windows ARM64**: `depsclean_windows_arm64.exe`

### From source (requires Go 1.26.3+)

```sh
go install github.com/panz3r/depsclean/cmd/depsclean@latest
```

---

## How it works

1. `depsclean` (or `depsclean ui`) launches the interactive TUI and starts scanning from the current directory (or `--root`).
2. Discovered target directories stream into the list as they are found.
3. A background worker pool enriches each row with its on-disk size, last-modified date, and package metadata (package name/version, package manager).
4. You navigate, search, sort, and delete interactively — or exit and use `depsclean scan` for non-interactive output.

Deletion is safe by design:
- The path must be absolute.
- The directory's basename must match an entry in the active profile's target list.
- The directory must exist and be a directory (not a file).
- `--dry-run` skips the actual `os.RemoveAll` call and reports success without touching the filesystem.

---

## Configuration

`depsclean` looks for a configuration file in two locations (first match wins):

1. `.depsclean.json` in the current working directory
2. `~/.config/depsclean/config.json`

You can also pass an explicit path with `depsclean scan --config /path/to/config.json`.

### Config file format

```json
{
  "version": 1,
  "profile": "node",
  "targets": ["node_modules"],
  "excludes": ["vendor", ".git"],
  "skip_hidden": true,
  "max_depth": 10,
  "dry_run": false,
  "output_format": "text"
}
```

| Field           | Type       | Default         | Description                                              |
|-----------------|------------|-----------------|----------------------------------------------------------|
| `version`       | `int`      | —               | Config file version (informational, currently unused)   |
| `profile`       | `string`   | `"node"`        | Built-in scan profile (see `depsclean profiles`)          |
| `targets`       | `[]string` | `["node_modules"]` | Directory names to treat as targets                  |
| `excludes`      | `[]string` | `[]`            | Glob patterns to exclude (matched on full path or basename) |
| `skip_hidden`   | `bool`     | `true`          | Skip directories whose name starts with `.`             |
| `max_depth`     | `int`      | `10`            | Maximum directory depth to scan (`0` = unlimited)       |
| `dry_run`       | `bool`     | `false`         | Simulate deletions without removing any files           |
| `output_format` | `string`   | `"text"`        | Output format for `scan`: `text`, `json`, or `ndjson`   |

**Precedence order (highest wins):** CLI flags › config file › built-in defaults.

---

## CLI usage

```
depsclean [flags]                  Launch the interactive TUI (default command)
depsclean ui [flags]               Launch the interactive TUI explicitly
depsclean scan [flags]             Scan and print results (non-interactive)
depsclean delete-all [flags]       Delete all discovered directories (batch)
depsclean profiles                 List built-in scan profiles
depsclean version [--check-update] Print version; optionally check for updates
```

### Global / shared flags

| Flag             | Default        | Description                                         |
|------------------|----------------|-----------------------------------------------------|
| `--root`         | `.`            | Root directory to scan                              |
| `--profile`      | `node`         | Built-in profile (`node`, `python`, `rust`, etc.)  |
| `--exclude`      | _(none)_       | Glob pattern to exclude (repeatable)                |
| `--skip-hidden`  | `true`         | Skip hidden directories (name starts with `.`)      |
| `--max-depth`    | `10`           | Maximum scan depth (`0` = unlimited)                |
| `--dry-run`      | `false`        | Simulate deletions without removing files           |

### `depsclean scan` additional flags

| Flag        | Default  | Description                                            |
|-------------|----------|--------------------------------------------------------|
| `--config`  | _(auto)_ | Explicit config file path (bypasses auto-detection)    |
| `--format`  | `text`   | Output format: `text`, `json`, or `ndjson`             |

### `depsclean delete-all` additional flags

| Flag      | Default | Description                                              |
|-----------|---------|----------------------------------------------------------|
| `--yes`   | `false` | Confirm actual deletion (required unless `--dry-run`)    |

### Examples

```sh
# Launch TUI in the current directory
depsclean

# Scan ~/workspace with the Python profile
depsclean scan --root ~/workspace --profile python

# Stream NDJSON to a file
depsclean scan --root ~/projects --format ndjson > results.ndjson

# Dry-run delete-all to preview what would be removed
depsclean delete-all --dry-run --root ~/workspace

# Actually delete all found node_modules (requires --yes)
depsclean delete-all --yes --root ~/workspace
```

---

## Built-in profiles

| Profile  | Targets                                                          |
|----------|------------------------------------------------------------------|
| `node`   | `node_modules`                                                   |
| `python` | `.venv`, `venv`, `__pycache__`, `.pytest_cache`, `.mypy_cache`, `dist`, `build`, `*.egg-info` |
| `rust`   | `target`                                                         |
| `go`     | `vendor`                                                         |
| `php`    | `vendor`                                                         |
| `ruby`   | `vendor/bundle`, `.bundle`                                       |
| `java`   | `target`, `.gradle`, `build`                                     |
| `all`    | Union of all the above                                           |

Run `depsclean profiles` to see the full list with descriptions.

---

## TUI key bindings

| Key(s)             | Action                                              |
|--------------------|-----------------------------------------------------|
| `↑` / `k`          | Move cursor up                                      |
| `↓` / `j`          | Move cursor down                                    |
| `PgUp` / `Ctrl+B`  | Page up                                             |
| `PgDn` / `Ctrl+F`  | Page down                                           |
| `Home` / `g`       | Jump to first item                                  |
| `End` / `G`        | Jump to last item                                   |
| `/`                | Enter search/filter mode                            |
| `Esc`              | Exit search mode / clear selection / reset range    |
| `s`                | Cycle sort mode (size↓ → size↑ → name↑ → path↑ → newest → oldest) |
| `d`                | Toggle compact / details row mode                   |
| `Enter`            | Toggle details panel for the focused item           |
| `Space`            | Toggle selection of the focused item                |
| `a`                | Select all visible eligible items (toggle)          |
| `r`                | Set range anchor (first press) / select range (second press) |
| `x` / `Delete`     | Delete the focused item                             |
| `X`                | Delete all selected items                           |
| `o`                | Open project folder in the system file explorer     |
| `q` / `Ctrl+C`     | Quit                                                |

---

## Credits

- [npkill](https://github.com/voidcosmos/npkill) — inspiration for streaming discovery and safe deletion
- [bunkill](https://github.com/codingstark-dev/bunkill) — inspiration for richer metadata and sort/filter ergonomics
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — terminal styling
- [Cobra](https://github.com/spf13/cobra) — CLI framework

---

## License

[MPL-2.0 License](https://www.mozilla.org/en-US/MPL/2.0/)
