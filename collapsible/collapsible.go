// Package collapsible renders a headless expand/collapse section: a header
// row (disclosure glyph plus a title) and a body shown only when expanded.
// The package owns the toggle behavior; styling and glyphs are the caller's.
package collapsible

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Style holds the caller's header/body text styles and the disclosure
// glyphs. OpenGlyph marks an expanded section, ClosedGlyph a collapsed one;
// empty falls back to "v" / ">".
type Style struct {
	Header      lipgloss.Style
	Body        lipgloss.Style
	OpenGlyph   string
	ClosedGlyph string
}

// Collapsible is a single expand/collapse section. Focused reports whether
// it owns keyboard input (the caller sets it); Update only responds when
// Focused. Body may span multiple lines and is rendered only when Expanded.
type Collapsible struct {
	Title    string
	Body     string
	Expanded bool
	Focused  bool
	Style    Style

	// Toggle are the keys that flip Expanded; nil falls back to enter/space.
	Toggle []string
}

// ToggledMsg is emitted, via the cmd from Update or ClickAt, when the
// section is toggled. Expanded is the new value.
type ToggledMsg struct{ Expanded bool }

func (c Collapsible) glyph() string {
	if c.Expanded {
		if c.Style.OpenGlyph != "" {
			return c.Style.OpenGlyph
		}
		return "v"
	}
	if c.Style.ClosedGlyph != "" {
		return c.Style.ClosedGlyph
	}
	return ">"
}

func (c Collapsible) header() string {
	if c.Title == "" {
		return c.glyph()
	}
	return c.glyph() + " " + c.Title
}

// Render draws the header row, followed by the body lines when Expanded.
func (c Collapsible) Render() string {
	header := c.Style.Header.Render(c.header())
	if !c.Expanded || c.Body == "" {
		return header
	}
	return header + "\n" + c.Style.Body.Render(c.Body)
}

// View is an alias for Render, matching the Bubble Tea View() convention.
func (c Collapsible) View() string { return c.Render() }

// HeaderWidth is the rendered width of the header row in cells — the caller
// needs it to route clicks (ClickAt takes panel-local coordinates).
func (c Collapsible) HeaderWidth() int { return lipgloss.Width(c.header()) }

// Lines is the total rendered height in rows: 1 for the header plus the body
// lines when Expanded.
func (c Collapsible) Lines() int {
	if !c.Expanded || c.Body == "" {
		return 1
	}
	return 1 + strings.Count(c.Body, "\n") + 1
}

// Update flips Expanded on a Toggle key and emits ToggledMsg. It is a no-op
// unless Focused, and ignores non-key messages.
func (c Collapsible) Update(msg tea.Msg) (Collapsible, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok || !c.Focused {
		return c, nil
	}
	toggle := c.Toggle
	if toggle == nil {
		toggle = []string{"enter", " "}
	}
	k := key.String()
	for _, t := range toggle {
		if k == t {
			return c.toggled()
		}
	}
	return c, nil
}

// ClickAt toggles Expanded when (x, y) lands on the header row (row 0).
// Clicks on body rows are a no-op.
func (c Collapsible) ClickAt(x, y int) (Collapsible, tea.Cmd) {
	if y != 0 || x < 0 || x >= c.HeaderWidth() {
		return c, nil
	}
	return c.toggled()
}

func (c Collapsible) toggled() (Collapsible, tea.Cmd) {
	c.Expanded = !c.Expanded
	expanded := c.Expanded
	return c, func() tea.Msg { return ToggledMsg{Expanded: expanded} }
}
