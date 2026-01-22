package browse

const (
	browseHelpText         = "↑/↓ navigate • ←/→ page • enter read • type to search • esc clear • q quit"
	articleHelpFormat      = "b back • c columns (%s) • ↑/↓ scroll • pgup/pgdn • q quit"
	articleLoadingHelp     = "b back • q quit"
	browseTitleLines       = 2
	browseSubtitleLines    = 2
	browseHeaderLines      = 4
	browseFooterLines      = 4
	browseFooterPadding    = 1
	browseFooterGapLines   = 0
	browseReservedLines    = browseHeaderLines + browseFooterLines + browseFooterPadding + browseFooterGapLines
	browseMinVisibleLines  = 5
	browseItemHeight       = browseTitleLines + browseSubtitleLines + 1
	articleFooterLines     = 4
	articleFooterPadding   = 1
	articleFooterGapLines  = 0
	articleReservedLines   = articleFooterLines + articleFooterPadding + articleFooterGapLines
	articleMinVisibleLines = 5
)
