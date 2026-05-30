// Package togglegroup renders a headless single-select toggle row: a line
// of mutually-exclusive toggle buttons laid out left-to-right, where exactly
// one item is active. It owns behavior — selection state, keyboard stepping,
// hit-testing, hover — but every visual decision lives in caller-supplied
// styles.
//
// A Group is a value type: store it on your model, route Bubble Tea messages
// through Update, and render with View. Reach for ApplyKey / HitTest when you
// want the pure core directly.
package togglegroup

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// State is the visual state of one toggle item. Inactive is the default;
// Hover indicates the cursor is over the item; Active marks the selected
// item.
type State int

const (
	StateInactive State = iota
	StateHover
	StateActive
)

// Style holds one lipgloss.Style per State. The styles are expected to
// fully define the look of an item; the package applies the per-state
// variant as-is to each item's text.
type Style struct {
	Inactive lipgloss.Style
	Hover    lipgloss.Style
	Active   lipgloss.Style
}

// ForState returns the lipgloss.Style for the given state.
func (s Style) ForState(state State) lipgloss.Style {
	switch state {
	case StateActive:
		return s.Active
	case StateHover:
		return s.Hover
	default:
		return s.Inactive
	}
}

// Group is a horizontal row of mutually-exclusive toggles. Selected is the
// index of the active item. Focused reports whether the group owns keyboard
// input (the caller sets it); Update only responds when Focused. Gap is the
// number of blank cells between items. Hover is the index under the cursor,
// or -1 for none.
type Group struct {
	Items    []string
	Selected int
	Focused  bool
	Gap      int
	Hover    int
	Style    Style

	// Key bindings (defaults applied internally when nil): Left/Right
	// move Selected; Wrap cycles at the ends.
	Left  []string
	Right []string
	Wrap  bool
}

// ChangedMsg is emitted, via the cmd from Update or ClickAt, when Selected
// changes. Selected is the new index.
type ChangedMsg struct{ Selected int }

func (g Group) stateOf(i int) State {
	switch {
	case i == g.Selected:
		return StateActive
	case i == g.Hover:
		return StateHover
	default:
		return StateInactive
	}
}

// Render draws the toggles on a single row, the selected one styled Active.
func (g Group) Render() string {
	if len(g.Items) == 0 {
		return ""
	}
	parts := make([]string, 0, len(g.Items)*2)
	gap := ""
	if g.Gap > 0 {
		gap = strings.Repeat(" ", g.Gap)
	}
	for i, item := range g.Items {
		if i > 0 && gap != "" {
			parts = append(parts, gap)
		}
		parts = append(parts, g.Style.ForState(g.stateOf(i)).Render(item))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

// View is an alias for Render, matching the Bubble Tea View() convention.
func (g Group) View() string { return g.Render() }

// HitTest returns the index of the item at panel-local (x, y), or -1 when
// the coordinate misses the row (wrong row, a gap, or past the ends).
func (g Group) HitTest(x, y int) int {
	if y != 0 || x < 0 || len(g.Items) == 0 {
		return -1
	}
	col := 0
	for i, item := range g.Items {
		if i > 0 {
			col += g.Gap
		}
		w := lipgloss.Width(item)
		if x >= col && x < col+w {
			return i
		}
		col += w
	}
	return -1
}

// ApplyKey is the pure key handler Update wraps. On a Left/Right key it moves
// Selected and returns the new Group plus a cmd carrying ChangedMsg; other
// keys are a no-op (nil cmd). It does not check Focused — Update does.
func (g Group) ApplyKey(key string) (Group, tea.Cmd) {
	if len(g.Items) == 0 {
		return g, nil
	}
	switch {
	case contains(g.keys(g.Left, "left", "h"), key):
		return g.moveTo(step(g.Selected, -1, len(g.Items), g.Wrap))
	case contains(g.keys(g.Right, "right", "l"), key):
		return g.moveTo(step(g.Selected, +1, len(g.Items), g.Wrap))
	}
	return g, nil
}

// Update routes a key message: Left/Right move Selected and emit ChangedMsg.
// It is a no-op unless Focused, and ignores non-key messages — mouse goes
// through ClickAt / HoverAt, which take panel-local coordinates.
//
//	case tea.KeyMsg:
//	    m.toggles, cmd = m.toggles.Update(msg)
//	    return m, cmd
//	case togglegroup.ChangedMsg:
//	    return apply(m, msg.Selected), nil
func (g Group) Update(msg tea.Msg) (Group, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok || !g.Focused {
		return g, nil
	}
	return g.ApplyKey(key.String())
}

// ClickAt selects the item at panel-local (x, y), emitting ChangedMsg when
// the selection changes; a miss or a click on the already-selected item is a
// no-op.
func (g Group) ClickAt(x, y int) (Group, tea.Cmd) {
	if idx := g.HitTest(x, y); idx >= 0 {
		return g.moveTo(idx)
	}
	return g, nil
}

// HoverAt sets Hover from panel-local (x, y), clearing it to -1 when the
// coordinate misses the row.
func (g Group) HoverAt(x, y int) Group {
	g.Hover = g.HitTest(x, y)
	return g
}

func (g Group) moveTo(idx int) (Group, tea.Cmd) {
	if idx == g.Selected {
		return g, nil
	}
	g.Selected = idx
	return g, func() tea.Msg { return ChangedMsg{Selected: idx} }
}

func (g Group) keys(binding []string, def ...string) []string {
	if binding != nil {
		return binding
	}
	return def
}

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
