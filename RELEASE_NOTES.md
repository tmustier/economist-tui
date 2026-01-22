# Release Notes

## v0.3.4

- **Demo mode**: add Suez, Hungary, Korea, Algeria, and Malaya archive fixtures.

## v0.3.3

- **Demo mode**: add Lincoln assassination and Stalin death fixtures, plus archive.org source links.

## v0.3.2

- **Demo mode**: update demo header to describe archive content.

## v0.3.1

- **Browse navigation**: keep list stable when moving within the visible window; page left/right preserves selection position.

## v0.3.0

- **Browse/reader**: support Ctrl+D to quit.

## v0.2.9

- **Reader layout**: increase multi-column minimum height to 24 lines.

## v0.2.8

- **Demo mode**: add historical fixtures (Sputnik, cybernetics, space travel, Kennedy election) and load demo metadata from a single index with flexible date parsing.

## v0.2.7

- **Reader layout**: prevent column text from bleeding into neighboring columns.

## v0.2.6

- **Reader layout**: multi-column mode now caps column count for short articles and centers the column block.
- **Reader UI**: columns hint now shows on/off state.

## v0.2.5

- **Reader layout**: multi-column mode now caps column width and adds extra columns on wide terminals.

## v0.2.4

- **UI foundation**: add screen host scaffolding and shared list renderer to support upcoming multi-screen views.
- **Browse refactor**: browse list rendering now uses shared list component for consistent layout.

## v0.2.3

- **Demo mode**: add `economist demo` for safe, offline screenshots with placeholder content.

## v0.2.2

- **Layout**: apply symmetric padding in browse and article views for centered columns.

## v0.2.1

- **Reader layout**: keep article headers full width when body is in two-column mode.
- Add tab/shift+tab section cycling in browse mode with updated help text.

## v0.2.0

Highlights

- **Full TUI browsing + reading**: interactive browse with search, paging, in‑TUI reading, back navigation, and column toggle.
- **Reader layout refresh**: centered readable widths, accent rules, header wrapping, consistent indentation, and improved footer/help layout.
- **Economist theme system**: dark/light detection (with override), shared rules/styles, and consistent color hierarchy across browse/read/headlines.
- **Performance & caching**: warm daemon, disk article cache (1h TTL), faster column toggles via cached render base.
- **Parsing upgrades**: overtitle/subtitle extraction improvements and cleaner article header handling.
- **Headline formatting**: aligned date column, compact dates on narrow terminals, dynamic line wrapping/ellipsis.
- **Project rename**: module/path and config migration from `economist-cli` → `economist-tui`.
- **Docs & design**: design doc added; README/SKILL streamlined.

## v0.1.0

- Initial public release.
