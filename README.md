# Economist TUI

Responsive Terminal UI and CLI to browse and read The Economist with your subscription. Unofficial.
- `economist browse` starts the TUI with current headlines (requires login with `economist login`)
- `economist demo` starts the TUI with pre-loaded archived articles

## Screenshots (from `demo` mode)

<p> <img alt="Leaders" src="https://github.com/user-attachments/assets/67fe133b-6566-4e32-8544-614d7e97438c" width="38%"/>
<img alt="Article" src="https://github.com/user-attachments/assets/3b403b74-2af8-4398-8060-b420726defdd" width="60%"/> </p>



*Note: historical Economist articles have no subtitles; added in `economist demo` for illustration and convenience*

## Install

### Homebrew (macOS)

```bash
brew install tmustier/tap/economist-tui
```

### From source

```bash
git clone https://github.com/tmustier/economist-tui
cd economist-tui

# Build local binary
make
./economist --version

# Install to ~/bin (macOS codesign included)
make install
```

**Prereqs:** Go 1.25+, Chrome/Chromium (for login and article fetching).

## Quick start

```bash
# Login (one-time, for full articles)
economist login

# Demo mode (no login required)
economist demo

# Interactive Terminal UI
economist browse
economist browse finance

# Non-interactive
economist headlines leaders --json
economist read "https://www.economist.com/finance-and-economics/2026/01/19/article-slug" --raw
```

## Commands

- `login` — open browser to authenticate
- `browse [section]` — interactive TUI (defaults to Leaders)
  - `Enter` read article, `b` back, type to search
  - `c` toggle columns on/off, `Esc` clear, `q` quit
- `demo` — interactive TUI with demo content (no login required)
- `headlines [section]` — list headlines
  - `-n/--number`, `-s/--search`, `--json`, `--plain`
- `read [url|-]` — read full article (`--raw`, `--wrap`, `--columns`)
- `sections` — list sections

Global flags: `--version`, `--debug`, `--no-color`

## Configuration

Config + cookies: `~/.config/economist-tui/`
Cache: `~/.config/economist-tui/cache` (1h TTL)

## Notes

- RSS provides ~300 items per section (~10 months)
- Full articles require an active Economist subscription

## License

MIT
