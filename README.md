# Economist CLI

A command-line tool to browse and read articles from The Economist.

## Features

- 📰 **Browse headlines** from any section via RSS
- 🔍 **Search** articles by keyword
- 📖 **Read full articles** in your terminal (requires subscription)
- 🎨 **Pretty rendering** with glamour markdown formatting

## Installation

### Prerequisites

- Go 1.25+
- Chrome/Chromium (for login and article fetching)

### Build from source

```bash
git clone https://github.com/tmustier/economist-cli
cd economist-cli

# Build and install (auto-signs on macOS)
make install

# Or manually:
go build -o economist .
codesign -s - economist  # macOS only
cp economist ~/bin/
```

## Quick Start

```bash
# Interactive browsing (TUI)
economist browse finance

# Browse headlines (non-interactive)
economist headlines leaders
economist headlines finance -n 5

# Search for topics
economist headlines business -s "AI"

# Login (one-time, for full articles)
economist login

# Read an article
economist read "https://www.economist.com/finance-and-economics/2026/01/19/article-slug"
```

## Commands

### Global flags

| Flag | Description |
|------|-------------|
| `--version` | Print version and exit |
| `--debug` | Dump raw HTML to a temp file for parser troubleshooting |
| `--no-color` | Disable color/styled output |

### `economist browse [section]`

Interactive TUI for browsing headlines and reading articles.

- `↑`/`↓` to navigate
- `←`/`→` to page
- Type to search (fuzzy filter, digits jump to item)
- `Enter` to read selected article
- `Esc` to clear search / quit
- `q` to quit

The browse command will opportunistically start the `economist serve` daemon in the background (it stays running until you stop it).

Requires an interactive terminal. For scripts, use `headlines --json`.

### `economist serve`

Run a local daemon that keeps a headless browser warm for faster reads. The daemon listens on a Unix socket in `~/.config/economist-cli/serve.sock`.

Run it manually:

```bash
economist serve
```

Check status or stop it:

```bash
economist serve --status
economist serve --stop
```


### `economist headlines [section]`

Fetch latest headlines from a section.

| Flag | Description |
|------|-------------|
| `-n, --number` | Number of headlines (default: 10) |
| `-s, --search` | Fuzzy search headlines (space-separated tokens) |
| `--json` | Output JSON array |
| `--plain` | Output `title<TAB>url` |

**Sections:** `leaders`, `briefing`, `finance`, `us`, `britain`, `europe`, `middle-east`, `asia`, `china`, `americas`, `business`, `tech`, `science`, `culture`, `graphic`, `world-this-week`

### `economist read [url|-]`

Fetch and display a full article. Use `-` or stdin to pass a URL.

| Flag | Description |
|------|-------------|
| `--raw` | Output plain markdown (no formatting) |
| `--wrap` | Wrap width for rendered output (0 = no wrap) |

Requires login for full content.

When `--debug` is set, the raw HTML is saved to a temp file, the path is printed to stderr, and step-by-step timing logs (with timestamps) are emitted.

### `economist login`

Opens a browser window for Economist login. Cookies are saved locally for future use.

### `economist sections`

List all available sections with aliases.

### `economist --version`

Print version and exit.

## Agent / Script Usage

The CLI is designed for both human and programmatic use:

```bash
# Get headlines as JSON
economist headlines finance --json

# Get first headline URL
economist headlines finance --json | jq -r '.[0].url'

# Read first headline
economist headlines finance --json | jq -r '.[0].url' | xargs economist read --raw

# Plain output for simple parsing
economist headlines finance --plain | head -1 | cut -f2

# With fzf for interactive selection
economist headlines finance --plain | fzf | cut -f2 | xargs economist read
```

## Configuration

Config and cookies stored in `~/.config/economist-cli/`

## How It Works

1. **Headlines**: Fetched from Economist RSS feeds (public, no auth needed)
2. **Articles**: Fetched via headless Chrome to bypass Cloudflare, using saved session cookies
3. **Login**: Opens visible Chrome, detects successful login, saves cookies

## Limitations

- RSS provides ~300 articles per section (~10 months history)
- No date filtering or advanced search on RSS
- Full articles require an active Economist subscription
- Article parsing may miss some content on non-standard layouts

## License

MIT
