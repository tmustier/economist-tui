# Release Notes

## v0.2.1

- **Reader layout**: keep article headers full width when body is in two-column mode.

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
