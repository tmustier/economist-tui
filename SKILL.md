---
name: economist-cli
description: Read articles from The Economist via CLI. Use when user wants to browse Economist headlines, search articles by topic, or read full article content in the terminal. Requires one-time browser login for full articles.
---

# Economist CLI

CLI tool to browse and read The Economist articles.

## Commands

```bash
# Interactive browse (TUI, human-only, supports / to search)
economist browse [section]

# Headlines (default section: leaders)
economist headlines [section] [-n count] [-s search] [--json|--plain]

# Read full article
economist read [url|-] [--raw] [--wrap N]

# Login (one-time, opens browser)
economist login

# List sections
economist sections
```

## Available Sections

`leaders`, `briefing`, `finance`, `us`, `britain`, `europe`, `middle-east`, `asia`, `china`, `americas`, `business`, `tech`, `science`, `culture`, `graphic`, `world-this-week`

## Global Flags

`--version`, `--debug`, `--no-color`

## Examples

```bash
# Get 5 finance headlines
economist headlines finance -n 5

# Search for China coverage (fuzzy tokens)
economist headlines finance -s "china"

# JSON output

economist headlines finance --json

# Read article (markdown output)
economist read "https://www.economist.com/..." --raw

# Pretty terminal rendering
economist read "https://www.economist.com/..." --wrap 100

# Debug HTML dump
economist --debug read "https://www.economist.com/..."

# Read URL from stdin
echo "https://www.economist.com/..." | economist read -
```

## Auth Flow

1. Run `economist login` (opens browser)
2. Log in to Economist account
3. Browser closes automatically when login detected
4. Cookies saved to `~/.config/economist-cli/`

## Agent Usage

```bash
# Get headlines as JSON (for parsing)
economist headlines finance --json

# Get first headline URL
economist headlines finance --json | jq -r '.[0].url'

# Read first headline
economist headlines finance --json | jq -r '.[0].url' | xargs economist read --raw

# Plain output (title<TAB>url)
economist headlines finance --plain
```

Note: `browse` requires a TTY and won't work in agent context. Use `headlines --json` instead.

## Notes

- Headlines via RSS: title, one-line description, date, URL (~300 items per section, ~10 months history)
- Full articles require login (bypasses Cloudflare via headless browser)
- Articles render as markdown with glamour formatting
