---
name: economist-cli
description: Read articles from The Economist via CLI. Use when user wants to browse Economist headlines, search articles by topic, or read full article content in the terminal. Requires one-time browser login for full articles.
---

# Economist CLI

CLI tool to browse and read The Economist articles.

## Commands

```bash
# Headlines (default section: leaders)
economist headlines [section] [-n count] [-s search]

# Read full article
economist read [url|-] [--raw]

# Debug (dump HTML to temp file)
economist --debug read <url>

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

# Search for China coverage
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

## Notes

- Headlines via RSS: title, one-line description, date, URL (~300 items per section, ~10 months history)
- Full articles require login (bypasses Cloudflare via headless browser)
- Articles render as markdown with glamour formatting
