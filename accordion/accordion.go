// Package accordion renders a headless vertical stack of independently
// collapsible sections: each section has a header (with an expand glyph)
// and a body shown only while the section is expanded. The package owns
// the layout, cursor movement, toggle behavior, and hit-testing; the
// caller owns every visual decision via a per-instance Style.
//
// An Accordion is a value type: store it on your model, route Bubble Tea
// messages through Update, and render with View.
//
//	case tea.KeyMsg:
//	    m.acc, cmd = m.acc.Update(msg)
//	    return m, cmd
//	case accordion.ToggledMsg:
//	    return m, nil
package accordion

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Style holds the caller's glyphs and per-state text styles. Expanded /
// Collapsed are the leading header glyphs (empty falls back to "▼" / "▶").
// Header / HeaderCursor style a header row when it is not / is under the
// cursor; Body styles a section's body block.
type Style struct {
	Expanded     string
	Collapsed    string
	Header       lipgloss.Style
	HeaderCursor lipgloss.Style
	Body         lipgloss.Style
}

func (s Style) expandedGlyph() string {
	if s.Expanded != "" {
		return s.Expanded
	}
	return "▼"
}

func (s Style) collapsedGlyph() string {
	if s.Collapsed != "" {
		return s.Collapsed
	}
	return "▶"
}

// Section is one entry: a header Title and a Body shown when expanded.
type Section struct {
	Title string
	Body  string
}

// Accordion is a stack of collapsible sections. Cursor is the index of the
// header the keyboard acts on; Expanded[i] reports whether section i shows
// its body. Focused reports whether the accordion owns keyboard input (the
// caller sets it); Update only responds when Focused.
type Accordion struct {
	Sections []Section
	Expanded []bool
	Cursor   int
	Focused  bool
	Style    Style

	// Up / Down move Cursor across headers; Toggle flips Expanded[Cursor].
	// nil falls back to up/k, down/j, and enter/space respectively.
	Up     []string
	Down   []string
	Toggle []string
}

// ToggledMsg is emitted, via the cmd from Update or ClickAt, when a section
// is toggled. Index is the section; Expanded is its new state.
type ToggledMsg struct {
	Index    int
	Expanded bool
}

// expanded reports whether section i is currently expanded, tolerating an
// Expanded slice shorter than Sections (a missing entry reads as false).
func (a Accordion) expanded(i int) bool {
	return i < len(a.Expanded) && a.Expanded[i]
}

func (a Accordion) headerLine(i int) string {
	st := a.Style.Header
	if i == a.Cursor {
		st = a.Style.HeaderCursor
	}
	glyph := a.Style.collapsedGlyph()
	if a.expanded(i) {
		glyph = a.Style.expandedGlyph()
	}
	return st.Render(glyph + " " + a.Sections[i].Title)
}

// Render draws the stack: each header on its own row, with the body of any
// expanded section beneath it. Body blocks may span multiple rows.
func (a Accordion) Render() string {
	var rows []string
	for i := range a.Sections {
		rows = append(rows, a.headerLine(i))
		if a.expanded(i) {
			rows = append(rows, a.Style.Body.Render(a.Sections[i].Body))
		}
	}
	return strings.Join(rows, "\n")
}

// View is an alias for Render, matching the Bubble Tea View() convention.
func (a Accordion) View() string { return a.Render() }

// rowsFor returns the number of terminal rows a section occupies: one
// header row plus, when expanded, the body's line count.
func (a Accordion) rowsFor(i int) int {
	n := 1
	if a.expanded(i) {
		n += lipgloss.Height(a.Style.Body.Render(a.Sections[i].Body))
	}
	return n
}

// HitTest returns the index of the section whose HEADER row sits at
// panel-local (x, y), and true. Clicks on a body block or outside the
// stack return (-1, false). It walks the same layout as Render so the two
// cannot drift: heights vary because expanded sections add body rows.
func (a Accordion) HitTest(x, y int) (int, bool) {
	if x < 0 || y < 0 {
		return -1, false
	}
	row := 0
	for i := range a.Sections {
		if y == row {
			return i, true
		}
		row += a.rowsFor(i)
	}
	return -1, false
}

// Update moves Cursor on Up/Down and toggles Expanded[Cursor] on a Toggle
// key, emitting ToggledMsg. It is a no-op unless Focused, and ignores
// non-key messages and empty stacks.
func (a Accordion) Update(msg tea.Msg) (Accordion, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok || !a.Focused || len(a.Sections) == 0 {
		return a, nil
	}
	switch k := key.String(); {
	case contains(orDefault(a.Up, "up", "k"), k):
		a.Cursor = clamp(a.Cursor-1, len(a.Sections))
	case contains(orDefault(a.Down, "down", "j"), k):
		a.Cursor = clamp(a.Cursor+1, len(a.Sections))
	case contains(orDefault(a.Toggle, "enter", " "), k):
		return a.toggle(a.Cursor)
	}
	return a, nil
}

// ClickAt toggles the section whose header is hit at panel-local (x, y),
// also moving Cursor there. A click on a body or a miss is a no-op.
func (a Accordion) ClickAt(x, y int) (Accordion, tea.Cmd) {
	i, ok := a.HitTest(x, y)
	if !ok {
		return a, nil
	}
	a.Cursor = i
	return a.toggle(i)
}

// toggle flips Expanded[i], growing the slice if needed, and returns a cmd
// carrying the resulting ToggledMsg.
func (a Accordion) toggle(i int) (Accordion, tea.Cmd) {
	exp := make([]bool, len(a.Sections))
	copy(exp, a.Expanded)
	exp[i] = !exp[i]
	a.Expanded = exp
	now := exp[i]
	return a, func() tea.Msg { return ToggledMsg{Index: i, Expanded: now} }
}

func clamp(i, n int) int {
	if i < 0 {
		return 0
	}
	if i >= n {
		return n - 1
	}
	return i
}

func orDefault(binding []string, def ...string) []string {
	if binding != nil {
		return binding
	}
	return def
}

func contains(xs []string, v string) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}
