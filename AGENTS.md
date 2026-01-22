# AGENTS.md

Guidelines for AI agents working on this project.

## Development Workflow

- Commit incrementally and often with descriptive messages
- For risky changes, consider creating a branch
- Write meaningful regression tests as you go
- When shipping user-facing changes, bump the version (cmd/root.go or build flags) and add release notes if appropriate
- When incrementing the version, update the Homebrew tap formula in `tmustier/homebrew-tap`
- When incrementing the version, create a GitHub release with the matching release-notes entry

## Codebase Orientation

- `cmd/`: CLI entry points and command wiring
- `internal/browse`: main TUI screens (browse + article reader)
- `internal/ui`: styling, layout, rendering, and shared UI helpers (LayoutSpec, BuildFooter, SelectHintLine)
- `internal/article`: article parsing and formatting
- `internal/rss`: section/feed metadata and parsing
- `internal/fetch` / `internal/browser`: network fetch + headless browser scraping
- `internal/cache`, `internal/config`, `internal/search`, `internal/logging`: supporting subsystems

## UI Layout Conventions

- Use `ui.LayoutSpec`/`ui.PageSize` for view height and list sizing instead of hard-coded reserved lines.
- Use `ui.BuildFooter` for footer construction and `ui.SelectHintLine` for responsive help text.
- Each screen should define a `layout.go` with its layout spec helpers (ex: `browseLayoutSpec`).

## TUI

Ensure:

- **UI/UX intuitiveness**: Navigation should be discoverable, hierarchy clear
- **UI/UX consistency**: Adhere to the overall design system; amend it only if critical; ensure user expectations (e.g., navigation patterns) are honored across all screens
- **Performance**: Especially for loading and rendering articles

Consistency is most important within the TUI, but also check user-facing outputs (e.g., `economist headlines`, `economist read` should be consistent with TUI patterns).

## Testing

- Write clean, meaningful, high-quality tests as you build functionality
- Include tests for:
  - **UI**: Test under 2-3 width assumptions (e.g., 80, 120, 40 cols)
  - **Performance**: Understand the impact of changes
  - **Golden records**: A few snapshot-style tests for key outputs
- When running locally, use available tools (browsing, tmux, etc.) for UI testing
