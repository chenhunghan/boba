// Package toolbar is a headless horizontal bar of action items. It owns
// behavior — selection, keyboard left/right movement, hit-testing, and
// hover — and emits plain strings; the caller owns styling via a
// per-instance button.Style. Each item renders as an auto-sized
// button.Button, laid out left-to-right with Gap blank cells between
// them, so the toolbar reuses button's rendering and width math.
//
// A Toolbar is a value type: store it on your model, route messages
// through Update, and render with View. For pointer input feed the
// panel-local coordinates from panel.HitTest into ClickAt / HoverAt.
//
//	bar := toolbar.Toolbar{Items: []string{"New", "Open", "Save"}, Hover: -1, Focused: true, Style: mystyle}
//	out := bar.View()
//
//	case tea.KeyMsg:
//	    m.bar, cmd = m.bar.Update(msg)
//	    return m, cmd
//	case toolbar.ActivatedMsg:
//	    return run(m, msg.Index), nil
package toolbar

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/chenhunghan/boba/button"
)

// Style is the per-state look of a toolbar item, reusing button.Style:
// Inactive for idle items, Hover for the item under the cursor, and
// Active for the selected item while the toolbar is Focused.
type Style = button.Style

// Toolbar is a horizontal bar of action items. Items are the labels;
// Selected is the index moved by the arrow keys and shown in the Active
// state when Focused. Hover is the index under the cursor (-1 for none).
// Gap is the blank cells between adjacent items. Focused reports whether
// the toolbar owns keyboard input — Update only responds when it's set.
type Toolbar struct {
	Items    []string
	Selected int
	Hover    int
	Focused  bool
	Gap      int
	Style    Style

	// Key bindings (defaults applied internally when nil): Left/Right
	// move Selected, Confirm activates it; Wrap cycles at the ends.
	Left, Right, Confirm []string
	Wrap                 bool
}

// ActivatedMsg is emitted — via the cmd from Update or ClickAt — when an
// item is activated (the Confirm key while Focused, or a click). Index is
// the activated item.
type ActivatedMsg struct{ Index int }

// itemWidth is the rendered width of item i: its label plus button's
// 1-cell padding on each side. Render and HitTest share this so their
// layouts cannot drift.
func (t Toolbar) itemWidth(i int) int {
	return lipgloss.Width(t.Items[i]) + 2
}

func (t Toolbar) stateOf(i int) button.State {
	switch {
	case t.Focused && i == t.Selected:
		return button.StateActive
	case i == t.Hover:
		return button.StateHover
	default:
		return button.StateInactive
	}
}

// Render draws the items left-to-right on a single row, each in the
// state derived from Focused, Selected, and Hover (priority: Active >
// Hover > Inactive). Returns "" when there are no items.
func (t Toolbar) Render() string {
	if len(t.Items) == 0 {
		return ""
	}
	parts := make([]string, 0, len(t.Items)*2)
	for i, item := range t.Items {
		if i > 0 && t.Gap > 0 {
			parts = append(parts, strings.Repeat(" ", t.Gap))
		}
		btn := button.Button{Text: item, Style: t.Style}
		parts = append(parts, btn.Render(t.stateOf(i), 0, 1))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

// View is an alias for Render, matching the Bubble Tea View() convention.
func (t Toolbar) View() string { return t.Render() }

// HitTest returns the item index at panel-local (x, y), or -1 when the
// coordinate misses the single row or lands in a gap. Coordinates are
// toolbar-local — the caller translates from app-global ones.
func (t Toolbar) HitTest(x, y int) int {
	if y != 0 || x < 0 {
		return -1
	}
	cursor := 0
	for i := range t.Items {
		if i > 0 && t.Gap > 0 {
			cursor += t.Gap
		}
		w := t.itemWidth(i)
		if x >= cursor && x < cursor+w {
			return i
		}
		cursor += w
	}
	return -1
}

// ApplyKey moves Selected on a Left/Right key and reports the activated
// index on a Confirm key. The caller gates on focus; ApplyKey routes the
// key through and returns the updated Toolbar, the activated index, and
// whether a Confirm fired.
func (t Toolbar) ApplyKey(key string) (Toolbar, int, bool) {
	if len(t.Items) == 0 {
		return t, 0, false
	}
	switch {
	case contains(t.keys(t.Left, "left", "h"), key):
		t.Selected = step(t.Selected, -1, len(t.Items), t.Wrap)
	case contains(t.keys(t.Right, "right", "l"), key):
		t.Selected = step(t.Selected, +1, len(t.Items), t.Wrap)
	case contains(t.keys(t.Confirm, "enter"), key):
		return t, t.Selected, true
	}
	return t, 0, false
}

// Update routes a key message: Left/Right move Selected and Confirm
// activates it (emitting ActivatedMsg). It is a no-op unless Focused, and
// ignores non-key messages — mouse goes through ClickAt / HoverAt.
func (t Toolbar) Update(msg tea.Msg) (Toolbar, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok || !t.Focused {
		return t, nil
	}
	next, idx, fired := t.ApplyKey(key.String())
	if fired {
		return next, fire(ActivatedMsg{Index: idx})
	}
	return next, nil
}

// ClickAt applies a click at panel-local (x, y). A hit on an item selects
// it and emits ActivatedMsg; a miss is a no-op.
func (t Toolbar) ClickAt(x, y int) (Toolbar, tea.Cmd) {
	idx := t.HitTest(x, y)
	if idx < 0 {
		return t, nil
	}
	t.Selected = idx
	return t, fire(ActivatedMsg{Index: idx})
}

// HoverAt sets Hover from panel-local (x, y), clearing it to -1 when the
// coordinate misses the toolbar.
func (t Toolbar) HoverAt(x, y int) Toolbar {
	t.Hover = t.HitTest(x, y)
	return t
}

func (t Toolbar) keys(binding []string, def ...string) []string {
	if binding != nil {
		return binding
	}
	return def
}

func fire(msg tea.Msg) tea.Cmd { return func() tea.Msg { return msg } }

func step(cur, delta, n int, wrap bool) int {
	if n == 0 {
		return cur
	}
	next := cur + delta
	if wrap {
		return ((next % n) + n) % n
	}
	if next < 0 {
		return 0
	}
	if next >= n {
		return n - 1
	}
	return next
}

func contains(xs []string, v string) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}
