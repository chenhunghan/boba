// Package statusbar renders a single-row, borderless status bar with
// optional left and right item groups.
//
// Each Item is a key+text pair like (`esc`, `back`); each Item carries
// its own ItemStyle, mirroring the per-instance styling pattern used
// in button, navcard, and box.Tab — the package owns layout, the
// consumer owns appearance. Items in the same group are space-
// separated; the left group hugs the left edge and the right group
// hugs the right edge, with the gap between them filled with spaces:
//
//	┐1/h hot ┐2/n nav ┐3/m main                        ┐focus main┌
//	└── left group ──┘                                  └── right ──┘
//
// Render is a pure function from (Bar, width) to a string of exactly
// width visible cells.
package statusbar

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// DefaultPadding is the number of blank cells inserted on the left
// and right edges of the status bar. Inner content fits in
// width - 2*DefaultPadding cells. Layout-only — there are no
// package-level Item style globals.
var DefaultPadding = 1

// ItemStyle bundles the lipgloss.Styles an Item uses for its two
// slots. Each Item carries its own ItemStyle. Letter highlighting
// reuses ItemStyle.Key as the highlight style.
type ItemStyle struct {
	Key  lipgloss.Style
	Text lipgloss.Style
}

// Item is a single key+text hint shown in a status bar group.
//
// Letter is an optional secondary shortcut: a substring of Text that
// will be rendered in the key style instead of the text style. This
// hints to the user that pressing that letter triggers the same
// action as Key. Example: Item{Key: "1", Letter: "h", Text: "hot"}
// renders as "1 hot" where "1" and "h" are bright (key style) and
// "ot" is muted (text style).
type Item struct {
	Key    string // shortcut glyph(s), e.g. "esc", "↑↓", "1"
	Text   string // human description, e.g. "back", "select", "hot"
	Letter string // optional letter within Text to highlight (case-sensitive)
	Style  ItemStyle
}

// render produces "<Key> <Text>" with the Item's own style, with
// Letter highlighted inline if set.
func (i Item) render() string {
	if i.Key == "" && i.Text == "" {
		return ""
	}
	if i.Text == "" {
		return i.Style.Key.Render(i.Key)
	}
	textRendered := i.renderText()
	if i.Key == "" {
		return textRendered
	}
	return i.Style.Key.Render(i.Key) + " " + textRendered
}

// renderText renders Item.Text with the optional Letter highlighted
// in the key style. Falls back to plain text style when Letter is
// empty or not found in Text.
func (i Item) renderText() string {
	if i.Letter == "" {
		return i.Style.Text.Render(i.Text)
	}
	idx := strings.Index(i.Text, i.Letter)
	if idx == -1 {
		return i.Style.Text.Render(i.Text)
	}
	before := i.Text[:idx]
	mid := i.Text[idx : idx+len(i.Letter)]
	after := i.Text[idx+len(i.Letter):]

	var b strings.Builder
	if before != "" {
		b.WriteString(i.Style.Text.Render(before))
	}
	b.WriteString(i.Style.Key.Render(mid))
	if after != "" {
		b.WriteString(i.Style.Text.Render(after))
	}
	return b.String()
}

// Bar is a single-row status indicator. Left items stack from the
// left edge, Right items hug the right edge.
type Bar struct {
	Left  []Item
	Right []Item
}

// Render fits the bar into width visible cells when the content fits;
// otherwise the bar may overflow. If left+right would collide, the
// right group is dropped (left content is more contextual). If left
// alone is wider than the inner area, it's emitted as-is — callers
// are expected to keep status bars to a reasonable item count.
//
// DefaultPadding cells are reserved on the left and right edges; the
// items render in the (width - 2*DefaultPadding) cells in between.
func (b Bar) Render(width int) string {
	pad := DefaultPadding
	if 2*pad >= width {
		pad = 0
	}
	innerWidth := width - 2*pad
	padStr := strings.Repeat(" ", pad)

	left := joinItems(b.Left)
	right := joinItems(b.Right)

	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)

	var inner string
	switch {
	case leftWidth >= innerWidth:
		// Left content alone exceeds the inner area — emit it whole
		// and let the bar overflow.
		inner = left
	case leftWidth+rightWidth+1 > innerWidth:
		// Left + right would collide; drop right, pad to inner width.
		inner = left + strings.Repeat(" ", innerWidth-leftWidth)
	default:
		fill := innerWidth - leftWidth - rightWidth
		inner = left + strings.Repeat(" ", fill) + right
	}

	return padStr + inner + padStr
}

// joinItems renders each item and joins them with two spaces, which
// reads better than one when the items are dense (key + text pairs).
func joinItems(items []Item) string {
	parts := make([]string, 0, len(items))
	for _, it := range items {
		parts = append(parts, it.render())
	}
	return strings.Join(parts, "  ")
}
