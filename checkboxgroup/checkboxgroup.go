// Package checkboxgroup renders a headless vertical multi-select list: one
// indicator glyph plus a label per row, with a movable cursor. The package
// owns navigation, toggling, and hit-testing; styling and glyphs are the
// caller's.
//
// A Group is a value type: store it on your model, route messages through
// Update, and render with View. Reach for ApplyKey / HitTest when you'd
// rather drive the pure core directly.
//
//	case tea.KeyMsg:
//	    m.opts, cmd = m.opts.Update(msg)
//	    return m, cmd
//	case checkboxgroup.ChangedMsg:
//	    m.selected = msg.Checked
//	    return m, nil
package checkboxgroup

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Style holds the caller's glyphs and per-state row styles. Checked /
// Unchecked are the indicator strings; empty falls back to "[x]" / "[ ]".
// Normal styles an unhighlighted row; Cursor styles the row under the
// cursor (applied only when the group is Focused).
type Style struct {
	Checked   string
	Unchecked string
	Normal    lipgloss.Style
	Cursor    lipgloss.Style
}

// Group is a vertical list of checkboxes. Checked is parallel to Options;
// a shorter or nil Checked reads as unchecked past its end. Cursor is the
// highlighted row. Focused reports whether the group owns keyboard input
// (the caller sets it); Update only responds when Focused.
type Group struct {
	Options []string
	Checked []bool
	Cursor  int
	Focused bool
	Style   Style

	// Up / Down move the cursor; Toggle flips the cursor row. Nil falls
	// back to up/k, down/j, and enter/space respectively.
	Up, Down, Toggle []string
}

// ChangedMsg is emitted, via the cmd from Update or ClickAt, when a row
// toggles. Checked is a fresh copy of the new checked state.
type ChangedMsg struct{ Checked []bool }

func (g Group) checkedAt(i int) bool {
	return i < len(g.Checked) && g.Checked[i]
}

func (g Group) glyph(checked bool) string {
	if checked {
		if g.Style.Checked != "" {
			return g.Style.Checked
		}
		return "[x]"
	}
	if g.Style.Unchecked != "" {
		return g.Style.Unchecked
	}
	return "[ ]"
}

func (g Group) row(i int) string {
	text := g.glyph(g.checkedAt(i))
	if g.Options[i] != "" {
		text += " " + g.Options[i]
	}
	return text
}

// Render draws one row per option, the indicator glyph followed by the
// label. The cursor row uses Style.Cursor only when Focused; every other
// row uses Style.Normal. Returns "" when there are no options.
func (g Group) Render() string {
	if len(g.Options) == 0 {
		return ""
	}
	rows := make([]string, len(g.Options))
	for i := range g.Options {
		st := g.Style.Normal
		if g.Focused && i == g.Cursor {
			st = g.Style.Cursor
		}
		rows[i] = st.Render(g.row(i))
	}
	return strings.Join(rows, "\n")
}

// View is an alias for Render, matching the Bubble Tea View() convention.
func (g Group) View() string { return g.Render() }

// HitTest returns the row index at panel-local (x, y), or -1 when the
// coordinate misses the list. x is bounded by each row's own width so
// clicks past a short label do not register.
func (g Group) HitTest(x, y int) int {
	if y < 0 || y >= len(g.Options) || x < 0 {
		return -1
	}
	if x >= lipgloss.Width(g.row(y)) {
		return -1
	}
	return y
}

// ApplyKey is the pure key handler Update wraps: it moves the cursor or
// toggles the cursor row, returning the new Group and whether a toggle
// occurred. It does not check Focused — Update gates that.
func (g Group) ApplyKey(key string) (Group, bool) {
	switch {
	case contains(g.keys(g.Up, "up", "k"), key):
		g.Cursor = clamp(g.Cursor-1, len(g.Options))
	case contains(g.keys(g.Down, "down", "j"), key):
		g.Cursor = clamp(g.Cursor+1, len(g.Options))
	case contains(g.keys(g.Toggle, "enter", " "), key):
		return g.toggle(g.Cursor)
	}
	return g, false
}

// Update moves the cursor or toggles the cursor row, emitting ChangedMsg
// on a toggle. It is a no-op unless Focused, and ignores non-key messages
// — pointer input goes through ClickAt with panel-local coordinates.
func (g Group) Update(msg tea.Msg) (Group, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok || !g.Focused {
		return g, nil
	}
	g, changed := g.ApplyKey(key.String())
	if changed {
		return g, g.changedCmd()
	}
	return g, nil
}

// ClickAt toggles the row at panel-local (x, y), moving the cursor there
// and emitting ChangedMsg. A miss is a no-op.
func (g Group) ClickAt(x, y int) (Group, tea.Cmd) {
	i := g.HitTest(x, y)
	if i < 0 {
		return g, nil
	}
	g.Cursor = i
	g, _ = g.toggle(i)
	return g, g.changedCmd()
}

func (g Group) toggle(i int) (Group, bool) {
	if i < 0 || i >= len(g.Options) {
		return g, false
	}
	checked := make([]bool, len(g.Options))
	for j := range checked {
		checked[j] = g.checkedAt(j)
	}
	checked[i] = !checked[i]
	g.Checked = checked
	return g, true
}

func (g Group) changedCmd() tea.Cmd {
	checked := make([]bool, len(g.Checked))
	copy(checked, g.Checked)
	return func() tea.Msg { return ChangedMsg{Checked: checked} }
}

func (g Group) keys(binding []string, def ...string) []string {
	if binding != nil {
		return binding
	}
	return def
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

func contains(xs []string, v string) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}
