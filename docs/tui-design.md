# Economist CLI: TUI Design Document

This document captures the design intent, research, and implementation plan for a delightful TUI reading experience, grounded in The Economist's official design system.

---

## 1. Readability Research

### Optimal Line Length

| Source | Characters Per Line |
|--------|---------------------|
| Ruder (Typographie) | **50-60** optimal |
| Baymard Institute | **50-75** acceptable |
| Web consensus | **45-80**, sweet spot ~66 |
| Print newspapers | **25-35** (narrow, non-scrolling) |

**Conclusion**: Target **66 characters** for article body text, max **72**.

### Multi-Column vs Single-Column

| Context | Recommendation |
|--------|----------------|
| Print | Multi-column works (visible simultaneously) |
| Screen/scrolling | **Single column** (avoids "return scrolling") |
| Short items (headlines) | Multi-column acceptable |

**Key insight**: Multi-column fails for scrollable content because readers must scroll back up to read column 2.

### TUI Width Strategy

```
80-char terminal (standard):
├─ Sidebar: 28 chars
└─ Content: ~50 chars (lower bound, acceptable)

120-char terminal (wide):
├─ Sidebar: 28 chars
└─ Content: MUST cap at ~70 chars + center

Reading mode (full-width, no sidebar):
└─ Content: Max 72 chars, centered with generous margins
```

```go
const (
    MinReadableWidth = 45
    MaxReadableWidth = 72
    IdealWidth       = 66
)

func readerContentWidth(terminalWidth int) int {
    available := terminalWidth - 8 // padding
    if available > MaxReadableWidth {
        return MaxReadableWidth
    }
    return max(available, MinReadableWidth)
}
```

---

## 2. Economist Design System

### Design Principles

1. **Less is more** — Be concise. Unburdened by decoration. Focus on the essential.
2. **Deliberate typography** — Strong typography with contrast. No excessive or arbitrary styles.
3. **Visual harmony** — Grid-based, restrained color, purposeful imagery.
4. **Clear wayfinding** — Well-structured paths. Obvious visual affordances.
5. **Intelligence and wit** — Clear but not dry. Thoughtful interactions.
6. **Recognisable consistency** — Same high standard across all mediums.

### Color Palette (Named After Cities)

| Role | Name | Hex | TUI Mapping |
|------|------|-----|-------------|
| **Brand** | Economist Red | `#E3120B` | ANSI 196 / True color |
| **Brand Light** | Economist Red 60 | `#F6423C` | Highlights |
| **Accent Primary** | Chicago 20 | `#141F52` | Deep blue (headers) |
| **Accent Primary** | Chicago 45 | `#2E45B8` | Blue accent |
| **Accent Primary** | Chicago 90 | `#D6DBF5` | Light blue bg |
| **Secondary** | Hong Kong 45 | `#1DC9A4` | Teal (success) |
| **Secondary** | Tokyo 45 | `#C91D42` | Rose (alerts/error) |
| **Tertiary** | Singapore 55 | `#F97A1F` | Orange |
| **Tertiary** | New York 55 | `#F9C31F` | Gold/Yellow (warning) |
| **Greyscale** | London 5 | `#0D0D0D` | Near black |
| **Greyscale** | London 10 | `#1A1A1A` | Dark bg |
| **Greyscale** | London 20 | `#333333` | Muted text |
| **Greyscale** | London 35 | `#595959` | Secondary text |
| **Greyscale** | London 70 | `#B3B3B3` | Disabled |
| **Greyscale** | London 85 | `#D9D9D9` | Borders |
| **Greyscale** | London 95 | `#F2F2F2` | Light bg |
| **Greyscale** | London 100 | `#FFFFFF` | White |
| **Canvas Warm** | Los Angeles 95 | `#F5F4EF` | Warm paper bg |
| **Canvas Cool** | Paris 95 | `#EFF5F5` | Cool paper bg |

### Typography

#### Typefaces

| Family | Use Case | TUI Equivalent |
|--------|----------|----------------|
| **Milo TE** | Body text, headlines | Default terminal font |
| **Milo SC TE** | Small caps | UPPERCASE with dimmed style |
| **Econ Sans OS** | Navigation, metadata, captions | Bold/highlighted text |
| **Econ Sans Condensed** | Charts, data | Compact layouts |

#### Modular Type Scale (Major Second 1.125)

```
Scale   Size(px)  rem      Usage in TUI
────────────────────────────────────────────
-3      11        0.702    —
-2      13        0.79     Footnotes (dim)
-1      14        0.889    Captions
 0      16        1.0      Body text (default)
 1      18        1.125    Subheadings
 2      20        1.266    Section titles
 3      23        1.424    Article subtitle
 4      26        1.602    Article title
 5      29        1.802    —
 6      32        2.027    Major headlines
 7+     36+       2.281+   Hero/display
```

#### Line Height

- **≤23px text**: 1.4 multiplier (more leading for readability)
- **≥23px text**: 1.2 multiplier (tighter for headlines)
- **TUI**: Use blank lines strategically for "leading"

### Grid System

| Breakpoint | Width | Columns | Gap | Gutter |
|------------|-------|---------|-----|--------|
| XS | <360px | 1 | 12px | 24px |
| SM | ≥360px | 4 | 12px | 24px |
| MD | ≥600px | 6 | 12px | 24px |
| LG | ≥960px | 12 | 16px | 32px |
| XL | ≥1280px | 12 | 16px | 32px |
| Max | 1432px | 12 | 16px | 32px |

**TUI Translation**:
- Sidebar: ~25-30 chars fixed
- Main content: Fluid, respects terminal width (capped at 72 for reading)
- Gutter between panes: 2-3 chars
- Internal padding: 1-2 chars

### Rules (Dividers)

| Type | Web CSS | TUI Rendering |
|------|---------|---------------|
| Default | 1px solid London-85 | `─` (light) |
| Emphasised | 1px solid London-5 | `─` (normal) |
| Heavy | 4px solid London-5 | `━` or `▀▀▀` |
| Accent | 4px solid Economist-Red | `━` in red |

### Interactions

- Web transition timing: `0.13s ease-in-out`
- TUI: Instant feedback, no animation needed

---

## 3. TUI Adaptation Strategy

### Visual Hierarchy Without Fonts

Since terminals use monospace fonts, we achieve hierarchy through:

1. **Color** — Economist Red for primary actions/selections
2. **Weight** — Bold for headlines, normal for body
3. **Decoration** — Dim/faint for secondary info
4. **Spacing** — Blank lines create "type scale" feeling
5. **Box characters** — Rules and borders for structure

### Component Mapping

| Web Component | TUI Equivalent |
|---------------|----------------|
| Badge | Uppercase + dim text |
| Section headline | Bold + rule below |
| Rule (accent) | `━━━` in Economist Red |
| Rule (default) | `───` in border color |
| Button (primary) | `[ Action ]` with red bg |
| Navigation link | Plain text, red when selected |
| Breadcrumb | `Section > Subsection` |

### Typography Hierarchy

| Element | Style | Example |
|---------|-------|---------|
| Section badge | `UPPERCASE` + muted | `LEADERS` |
| Headline | Bold | **The retreat of globalisation** |
| Subtitle | Italic/muted | *Why the world is becoming less connected* |
| Date | Faint | Jan 18th 2026 |
| Body | Normal | Regular paragraph text |
| Help | Faint | ↑/↓ navigate • q quit |

---

## 4. Layout Structure

### Browse View (Sidebar + Headlines)

```
┌─ The Economist ─────────────────────────────────────────────┐
│                                                             │
│ ┌─ Sections ────┐ ┌─ LEADERS ────────────────────────────┐ │
│ │               │ │ ─────────────────────────────────    │ │
│ │               │ │                                      │ │
│ │ ● Leaders     │ │ ▸ The retreat of globalisation       │ │
│ │   Briefing    │ │   Why the world is becoming less...  │ │
│ │   Business    │ │   Jan 18, 2026                       │ │
│ │   Finance     │ │                                      │ │
│ │   Britain     │ │   How to fix Britain's NHS           │ │
│ │   Europe      │ │   The health service needs reform    │ │
│ │   ...         │ │                                      │ │
│ │               │ │                                      │ │
│ └───────────────┘ └──────────────────────────────────────┘ │
│                                                             │
│ ↑/↓ navigate • enter read • tab switch • / search • q quit │
└─────────────────────────────────────────────────────────────┘
```

### Reading View (Centered, Max 72 chars)

```
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  (red)    │
│                                                             │
│  LEADERS                                                    │
│                                                             │
│  The retreat of globalisation                               │
│  ───────────────────────────────────────────────────────    │
│                                                             │
│  Why the world is becoming less connected—and what          │
│  to do about it                                             │
│                                                             │
│  Jan 18th 2026                                              │
│                                                             │
│  For decades, the world economy became steadily more        │
│  integrated. Goods, capital, people and ideas flowed        │
│  across borders with increasing ease. This period of        │
│  "hyperglobalisation" brought enormous benefits...          │
│                                                             │
│  [Content capped at 72 chars, centered if terminal wide]    │
│                                                             │
│                                                      23%    │
│                                                             │
│  ↑/↓ scroll • gg/G top/bottom • esc back • q quit          │
└─────────────────────────────────────────────────────────────┘
```

---

## 5. State Machine

```
┌─────────────────┐     enter      ┌──────────────────┐
│ FocusSidebar    │ ─────────────▶ │ FocusHeadlines   │
│ (section list)  │ ◀───────────── │ (article list)   │
└─────────────────┘   esc / ←      └──────────────────┘
                                        │
                                        │ enter
                                        ▼
                               ┌──────────────────┐
                    esc / q    │ FocusReader      │
                  ◀─────────── │ (article body)   │
                               └──────────────────┘

Navigation:
- Tab / ←/→: Switch between sidebar and headlines
- Enter: Drill down (section → headlines → article)
- Esc: Back up one level
- q: Quit (or go back if in reader)
- j/k or ↑/↓: Navigate within focused pane
- gg/G: Top/bottom (reader)
- /: Search (headlines)
```

---

## 6. Implementation

### 6.1 Theme (`internal/ui/theme.go`)

```go
package ui

import "github.com/charmbracelet/lipgloss"

// Economist Design System colors
var (
    // Brand
    EconomistRed   = lipgloss.Color("#E3120B")
    EconomistRed60 = lipgloss.Color("#F6423C")

    // Accent (Chicago blues)
    Chicago20 = lipgloss.Color("#141F52")
    Chicago45 = lipgloss.Color("#2E45B8")
    Chicago90 = lipgloss.Color("#D6DBF5")

    // Secondary
    HongKong45 = lipgloss.Color("#1DC9A4") // Teal/success
    Tokyo45    = lipgloss.Color("#C91D42") // Rose/error

    // Tertiary
    Singapore55 = lipgloss.Color("#F97A1F") // Orange
    NewYork55   = lipgloss.Color("#F9C31F") // Gold/warning

    // Greyscale (London)
    London5   = lipgloss.Color("#0D0D0D")
    London10  = lipgloss.Color("#1A1A1A")
    London20  = lipgloss.Color("#333333")
    London35  = lipgloss.Color("#595959")
    London70  = lipgloss.Color("#B3B3B3")
    London85  = lipgloss.Color("#D9D9D9")
    London95  = lipgloss.Color("#F2F2F2")
    London100 = lipgloss.Color("#FFFFFF")

    // Canvas
    LosAngeles95 = lipgloss.Color("#F5F4EF") // Warm paper
    Paris95      = lipgloss.Color("#EFF5F5") // Cool paper
)

// Theme defines semantic color assignments
// (see the implementation for full details)
```

### 6.2 Styles (`internal/ui/styles.go`)

```go
package ui

import "github.com/charmbracelet/lipgloss"

// Styles defines semantic styles for the TUI.
// (see the implementation for full details)
```

### 6.3 Rules (`internal/ui/rules.go`)

```go
package ui

import "strings"

const (
    RuleLight  = "─"
    RuleHeavy  = "━"
    RuleDouble = "═"
    RuleDotted = "┄"
)

func DrawRule(width int, char string, style lipgloss.Style) string {
    return style.Render(strings.Repeat(char, width))
}
```

### 6.4 File Structure

```
internal/
├── ui/
│   ├── theme.go      # Color palette + Theme struct
│   ├── styles.go     # Computed styles from theme
│   ├── rules.go      # Rule/divider helpers
│   ├── headlines.go  # Headline layout (existing, enhanced)
│   └── terminal.go   # Terminal size helpers (existing)
│
├── tui/
│   ├── model.go      # Main TUI model
│   ├── browse.go     # Browse view rendering
│   ├── reader.go     # Reader view rendering
│   ├── commands.go   # Async commands (fetch)
│   └── keys.go       # Key bindings
│
└── ... (existing packages)
```

### 6.5 Migration Path

1. **Add theme.go** — No breaking changes
2. **Update styles.go** — Backwards compatible, add new styles
3. **Create tui/ package** — New code, doesn't touch existing browse.go
4. **Wire up in cmd/browse.go** — Replace model with tui.Model
5. **Test & iterate** — Ensure reading flow works smoothly

---

## 7. Key Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Line length | Max 72 chars | Research-backed optimal readability |
| Article layout | Single column | Avoid return-scrolling problem |
| Wide terminals | Center + cap | Prevent eye strain from long saccades |
| Color accent | Economist Red only | "Less is more" principle |
| Multi-column headlines | Defer | Nice-to-have, adds navigation complexity |
| Help bar | Always visible | Clear wayfinding |
| Vim keys | Support j/k/gg/G | Power user friendly |

---

## 8. Open Questions

1. **Dark mode detection**: Auto-detect terminal background or manual `--dark` flag?
2. **Image handling**: Skip, placeholder text, or kitty/sixel protocol?
3. **Offline caching**: How aggressive? Show stale data with indicator?
4. **Search scope**: Current section only, or all cached headlines?

---

## 9. What "Delightful" Means

Applying the Economist principles to TUI:

1. **Less is more**: No unnecessary chrome. Content-first.
2. **Deliberate typography**: Clear hierarchy via color/weight, not decoration.
3. **Visual harmony**: Consistent spacing, restrained palette.
4. **Clear wayfinding**: Always know where you are, how to go back.
5. **Intelligence and wit**: The red accent *is* the wit — unmistakably Economist.
6. **Recognisable consistency**: Someone familiar with economist.com should feel at home.

---

## References

- Economist Design System: https://design-system.economist.com/
- Design Tokens: https://design-system.economist.com/developer-resources/design-system/design-tokens
- Baymard (line length): https://baymard.com/blog/line-length-readability
- UX research (multi-column): https://ux.stackexchange.com/questions/39480/multi-column-articles
