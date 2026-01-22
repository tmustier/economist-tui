# AGENTS.md

Guidelines for AI agents working on this project.

## Development Workflow

- Commit incrementally and often with descriptive messages
- For risky changes, consider creating a branch
- Write meaningful regression tests as you go

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
