# boba

[![Go Reference](https://pkg.go.dev/badge/github.com/chenhunghan/boba.svg)](https://pkg.go.dev/github.com/chenhunghan/boba)
[![CI](https://github.com/chenhunghan/boba/actions/workflows/ci.yml/badge.svg)](https://github.com/chenhunghan/boba/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/chenhunghan/boba)](https://goreportcard.com/report/github.com/chenhunghan/boba)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

**Unstyled, composable TUI components for [Bubble Tea](https://github.com/charmbracelet/bubbletea).**

boba is a *headless* component library: each package owns **behavior**
— layout, hit-testing, focus, selection state — and emits plain
strings, while **you** own the **style** (the [Lip Gloss](https://github.com/charmbracelet/lipgloss)
colors and theme). Think of it as the headless counterpart to
[`charmbracelet/bubbles`](https://github.com/charmbracelet/bubbles):
where bubbles ships opinionated, styled widgets, boba ships the
primitives and lets you bring your own look.

> **Status: `v0` — the API is unstable** and may change between minor
> versions until `v1.0.0`.

## Install

```sh
go get github.com/chenhunghan/boba@latest
```

## Why headless?

- **Bring your own style.** No package ships a theme or a color
  default — an unstyled component renders with the terminal's
  defaults, and every color is a choice you make at the call site.
- **Composable primitives.** `focus`, `panel`, `overlay`, `popup`, and
  `scroll` are framework-level building blocks; the rest are components
  (inputs, controls, overlays, containers) you arrange inside them.
- **Keyboard-first.** Focus cycling, jumps, and hit-testing are
  first-class, with a uniform `HitTest(x, y)` shape across components.

## Components

**Inputs & controls**

| Package | What it is |
| --- | --- |
| [`button`](button) | Buttons + stacks with hover / active / selected state |
| [`input`](input) | Single-line text field with cursor + editing |
| [`checkbox`](checkbox) | Labeled checkbox, toggled by key or click |
| [`switch`](switch) | On/off switch |
| [`toggle`](toggle) | Pressable toggle button |
| [`togglegroup`](togglegroup) | Single-select row of toggles |
| [`radiogroup`](radiogroup) | Single-select radio list |
| [`checkboxgroup`](checkboxgroup) | Multi-select checkbox list |
| [`slider`](slider) | Horizontal value slider |
| [`numberfield`](numberfield) | Numeric field with steppers |
| [`selectbox`](selectbox) | Select / dropdown with a popup list |

**Display & feedback**

| Package | What it is |
| --- | --- |
| [`progress`](progress) | Progress bar |
| [`meter`](meter) | Gauge / meter with thresholds |
| [`separator`](separator) | Horizontal / vertical divider |
| [`statusbar`](statusbar) | Key/label status line |
| [`glyph`](glyph) | Sub/superscript numeral helpers |
| [`tooltip`](tooltip) | Anchored text tooltip |

**Containers, navigation & overlays**

| Package | What it is |
| --- | --- |
| [`box`](box) | Bordered region with labeled notches |
| [`navcard`](navcard) | List/navigation card with inline actions |
| [`tab`](tab) | Dynamic tab group; each tab holds a `tea.Model` |
| [`accordion`](accordion) | Stack of independently collapsible sections |
| [`collapsible`](collapsible) | Single expand/collapse section |
| [`toolbar`](toolbar) | Horizontal action bar |
| [`menu`](menu) | Popup menu with keyboard + mouse |
| [`popover`](popover) | Anchored floating panel |
| [`dialog`](dialog) | Centered modal dialog |
| [`scrollarea`](scrollarea) | Scrollable viewport with scrollbar |
| [`pins`](pins) | Ordered, de-duplicated pinnable list |

## Primitives

| Package | What it is |
| --- | --- |
| [`focus`](focus) | Pure-function focus state machine (jumps, ring cycling, esc, click) |
| [`panel`](panel) | Layout tree (`Split` / `Panel`) + screen→local coords + hit-testing |
| [`overlay`](overlay) | ANSI-aware compositing of one rendered string over another |
| [`popup`](popup) | Placement (anchor / center with edge-flipping) + ANSI isolation |
| [`scroll`](scroll) | Vertical viewport over multi-line content |

## Quick start

```go
package main

import (
	"fmt"

	"github.com/chenhunghan/boba/box"
)

func main() {
	b := box.Box{
		LeftNotches:  []box.Notch{{Text: "status"}},
		RightNotches: []box.Notch{{Text: "1", Badge: "q"}}, // Badge = shortcut hint
		Body:         "ready.",
	}
	// Render is pure: (width, height) -> a string of exactly width×height cells.
	fmt.Println(b.Render(30, 4))
}
```

Styling is yours to supply: set `BorderColor` / `FillColor` on the box,
`Style` on each notch, and the `Style` struct on stateful components
like `button.Stack` and `tab.Group`. Every package ships runnable
[examples](https://pkg.go.dev/github.com/chenhunghan/boba) on
pkg.go.dev. Runnable apps in this repo: a 3-column service dashboard that
composes most of the library ([`examples/dashboard`](examples/dashboard)),
a minimal example ([`examples/demo`](examples/demo)), and a component
gallery (`go run ./cmd/storybook`).

## Interactive components

Stateful components (`menu`, `button`, `navcard`, `tab`) offer a
bubbles-style drop-in: route messages through `Update`, render with
`View`, and handle the events they emit. Keyboard wiring is one line:

```go
case tea.KeyMsg, tea.MouseMsg:
    m.menu, cmd = m.menu.Update(msg)   // menu owns its anchor → handles mouse too
    return m, cmd
case menu.ChosenMsg[Action]:
    return apply(m, msg.ID), nil
```

Components positioned by a parent (`button`, `navcard`, `tab`) take mouse
via `ClickAt` / `HoverAt` (tab: `HoverState`), fed the local coordinates
from `panel.HitTest`. The lower-level pure API (`ApplyKey` / `HitTest` /
`Render`) stays available underneath.

## Contributing

Issues and PRs welcome — see [CONTRIBUTING.md](CONTRIBUTING.md). PRs are
squash-merged and **PR titles must follow [Conventional Commits](https://www.conventionalcommits.org/)**
(a CI check enforces it) — that's what drives automated releases.

## License

[MIT](LICENSE)
