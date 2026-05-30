// Package menu is a headless context-menu / dropdown widget. It owns
// behavior — open/closed state, item layout, hit-testing, keyboard
// navigation — but every visual decision lives in caller-supplied
// styles. Mirrors the per-instance pattern used by button, navcard,
// tab, focus, and panel: same shape, same composition rules.
//
// A Group is a value type: store it on your model, route Bubble Tea
// messages through Update, and render with View. Prefer Update for
// drop-in wiring; reach for ApplyKey / ApplyClick when you'd rather
// handle the Outcome synchronously.
//
// Lower-level wiring (Update is the one-line drop-in; see its docs):
//
//	case tea.KeyMsg:
//	    if m.menu.Open {
//	        out := m.menu.ApplyKey(msg.String())
//	        m.menu = out.Group
//	        if out.Confirmed { m = handleAction(m, out.Chosen) }
//	        return m, nil
//	    }
//	case tea.MouseMsg:
//	    if m.menu.Open && msg.Action == tea.MouseActionPress {
//	        out := m.menu.ApplyClick(msg.X, msg.Y)
//	        m.menu = out.Group
//	        if out.Confirmed { m = handleAction(m, out.Chosen) }
//	        return m, nil
//	    }
package menu

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// State is the visual state of one menu item.
type State int

const (
	// StateInactive: not under cursor / not keyboard-selected.
	StateInactive State = iota
	// StateHover: under cursor or keyboard-selected (lit up).
	StateHover
	// StateDisabled: not selectable; rendered but doesn't respond.
	StateDisabled
)

// Style holds the per-state styles for a menu's items plus the
// chrome (border + glyphs). Items render as one row each; the
// border wraps them in a 1-cell-thick box.
//
// Surface, when set, paints every owned cell (border, items, padding)
// with that background color — useful when the popup should pop as a
// distinct surface. When Surface is empty (zero value), the menu
// renders against the terminal's default background; either way each
// row starts with an SGR reset so the popup cannot inherit lingering
// ANSI state from whatever the overlay primitive places it on top of.
type Style struct {
	// Per-state item styles.
	Inactive lipgloss.Style
	Hover    lipgloss.Style
	Disabled lipgloss.Style

	// Border style applied to every box-drawing rune.
	Border lipgloss.Style

	// Surface is the background color painted onto every cell of the
	// popup's rectangle (border + items + padding). Empty = no fill.
	Surface lipgloss.Color

	// Glyphs. Empty fields fall back to the box-drawing defaults.
	TopLeft, TopRight, BottomLeft, BottomRight string
	Horizontal, Vertical                       string
}

// withSurface returns s with Surface applied as the background, when
// Surface is set. Overrides any bg the caller put on s — Surface owns
// the popup's bg uniformly. When Surface is empty, returns s unchanged.
func (s Style) withSurface(style lipgloss.Style) lipgloss.Style {
	if s.Surface == "" {
		return style
	}
	return style.Background(s.Surface)
}

// ForState returns the lipgloss.Style for the given state.
func (s Style) ForState(state State) lipgloss.Style {
	switch state {
	case StateHover:
		return s.Hover
	case StateDisabled:
		return s.Disabled
	default:
		return s.Inactive
	}
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

func (s Style) topLeft() string     { return orDefault(s.TopLeft, "┌") }
func (s Style) topRight() string    { return orDefault(s.TopRight, "┐") }
func (s Style) bottomLeft() string  { return orDefault(s.BottomLeft, "└") }
func (s Style) bottomRight() string { return orDefault(s.BottomRight, "┘") }
func (s Style) horizontal() string  { return orDefault(s.Horizontal, "─") }
func (s Style) vertical() string    { return orDefault(s.Vertical, "│") }

// Item is one entry in a menu. ID is the caller-defined identifier
// returned via Outcome.Chosen when the user confirms this item.
// Disabled items render but cannot be confirmed.
type Item[ID comparable] struct {
	ID       ID
	Label    string
	Disabled bool
}

// Group is the menu widget. State (Open, Hover, AnchorX, AnchorY)
// is mutated by caller in response to events; Items + Style + the
// keymap fields are set up once per open.
//
// Hover is the index (into Items) of the currently highlighted
// item, or -1 for no highlight. ApplyKey moves Hover; ApplyClick at
// an item position sets Hover to that index momentarily before
// firing the confirm.
type Group[ID comparable] struct {
	Items   []Item[ID]
	Open    bool
	Hover   int
	AnchorX int // top-left x in screen coords
	AnchorY int // top-left y in screen coords
	Style   Style

	// Key bindings (defaults applied internally when nil).
	Up      []string // step Hover backward; default ["up", "k"]
	Down    []string // step Hover forward; default ["down", "j"]
	Confirm []string // confirm Hover; default ["enter"]
	Cancel  []string // cancel; default ["esc"]
	Wrap    bool     // wrap Hover at edges
}

// Outcome is the result of applying an event to the menu. Group is
// the (possibly updated) Group state. Confirmed/Cancelled are
// mutually exclusive flags — at most one is true. Chosen is valid
// only when Confirmed is true.
type Outcome[ID comparable] struct {
	Group     Group[ID]
	Confirmed bool
	Cancelled bool
	Chosen    ID
}

// Width returns the menu's outer width in cells (border + content).
// Zero when Items is empty.
func (g Group[ID]) Width() int {
	if len(g.Items) == 0 {
		return 0
	}
	max := 0
	for _, it := range g.Items {
		w := lipgloss.Width(it.Label)
		if w > max {
			max = w
		}
	}
	// 2 borders + 2 padding + label
	return max + 4
}

// Height returns the menu's outer height in cells (border + items).
// Zero when Items is empty.
func (g Group[ID]) Height() int {
	if len(g.Items) == 0 {
		return 0
	}
	return len(g.Items) + 2 // top + items + bottom
}

// Render returns the menu as a multi-line string sized Width x
// Height, or "" when closed. The output does NOT include any
// positioning — callers are responsible for placing the rendered
// string at AnchorX/AnchorY in their own composition.
//
// Each row of the output begins with an explicit SGR reset
// ("\x1b[0m"). This is what guarantees the menu doesn't inherit
// lingering ANSI state from the cells it overlays — without the
// leading reset, an open background SGR in the underlying row would
// bleed through into the menu's first piece (the border + first
// pad cell), tinting that part of the popup with whatever color was
// active in the bg row at the cut point. lipgloss already emits a
// trailing reset after each styled piece, so subsequent pieces in
// the same row are isolated from each other; the per-row leading
// reset closes the only remaining seam (prefix → first piece).
func (g Group[ID]) Render() string {
	if !g.Open || len(g.Items) == 0 {
		return ""
	}
	w := g.Width()
	innerW := w - 2 // exclude side borders

	const sgrReset = "\x1b[0m"
	border := g.Style.withSurface(g.Style.Border)

	// Top edge
	var rows []string
	rows = append(rows, sgrReset+border.Render(
		g.Style.topLeft()+
			strings.Repeat(g.Style.horizontal(), innerW)+
			g.Style.topRight(),
	))

	// Item rows
	for i, it := range g.Items {
		st := StateInactive
		if it.Disabled {
			st = StateDisabled
		} else if i == g.Hover {
			st = StateHover
		}
		itemStyle := g.Style.withSurface(g.Style.ForState(st))
		// Layout inside the row: " label<padding> "
		labelWidth := lipgloss.Width(it.Label)
		pad := innerW - 2 - labelWidth // 2 = left/right 1-cell padding
		if pad < 0 {
			pad = 0
		}
		content := " " + it.Label + strings.Repeat(" ", pad) + " "
		rows = append(rows,
			sgrReset+border.Render(g.Style.vertical())+
				itemStyle.Render(content)+
				border.Render(g.Style.vertical()),
		)
	}

	// Bottom edge
	rows = append(rows, sgrReset+border.Render(
		g.Style.bottomLeft()+
			strings.Repeat(g.Style.horizontal(), innerW)+
			g.Style.bottomRight(),
	))

	return strings.Join(rows, "\n")
}

// Inside reports whether (x, y) — in screen coordinates — falls
// inside the menu's outer rectangle. Always false when closed.
func (g Group[ID]) Inside(x, y int) bool {
	if !g.Open {
		return false
	}
	return x >= g.AnchorX &&
		x < g.AnchorX+g.Width() &&
		y >= g.AnchorY &&
		y < g.AnchorY+g.Height()
}

// HitTest returns the item at screen coords (x, y), or zero+false if
// the click is outside the menu or on the border / padding rows.
func (g Group[ID]) HitTest(x, y int) (ID, bool) {
	var zero ID
	if !g.Inside(x, y) {
		return zero, false
	}
	// y - AnchorY: 0 = top border, 1..len(Items) = items, last = bottom border
	rowIdx := y - g.AnchorY - 1
	if rowIdx < 0 || rowIdx >= len(g.Items) {
		return zero, false
	}
	// x must be inside the content area (not on side borders)
	relX := x - g.AnchorX
	if relX <= 0 || relX >= g.Width()-1 {
		return zero, false
	}
	return g.Items[rowIdx].ID, true
}

// ApplyKey processes a key press. No-op when closed. Outcome:
//   - Up/Down: moves Hover (skipping disabled items); Confirmed=false
//   - Confirm: if Hover is on an enabled item, Confirmed=true and
//     Chosen=item.ID; sets Open=false
//   - Cancel: sets Open=false, Cancelled=true
//   - Other: no change
func (g Group[ID]) ApplyKey(key string) Outcome[ID] {
	if !g.Open {
		return Outcome[ID]{Group: g}
	}
	up := g.Up
	if up == nil {
		up = []string{"up", "k"}
	}
	down := g.Down
	if down == nil {
		down = []string{"down", "j"}
	}
	confirm := g.Confirm
	if confirm == nil {
		confirm = []string{"enter"}
	}
	cancel := g.Cancel
	if cancel == nil {
		cancel = []string{"esc"}
	}

	switch {
	case contains(cancel, key):
		g.Open = false
		return Outcome[ID]{Group: g, Cancelled: true}
	case contains(confirm, key):
		if g.Hover >= 0 && g.Hover < len(g.Items) && !g.Items[g.Hover].Disabled {
			id := g.Items[g.Hover].ID
			g.Open = false
			return Outcome[ID]{Group: g, Confirmed: true, Chosen: id}
		}
		// Hover is invalid or on disabled item — no-op
		return Outcome[ID]{Group: g}
	case contains(up, key):
		g.Hover = stepHover(g.Items, g.Hover, -1, g.Wrap)
		return Outcome[ID]{Group: g}
	case contains(down, key):
		g.Hover = stepHover(g.Items, g.Hover, +1, g.Wrap)
		return Outcome[ID]{Group: g}
	}
	return Outcome[ID]{Group: g}
}

// HoverAt updates Hover from a mouse position in screen coords.
// When (x, y) lands on an item, sets Hover to that index; when it
// lands on the border/padding (or outside the menu entirely),
// leaves Hover unchanged. Use from the parent's tea.MouseMsg
// handler to track motion-driven hover. No-op when closed.
func (g Group[ID]) HoverAt(x, y int) Group[ID] {
	if !g.Open {
		return g
	}
	id, ok := g.HitTest(x, y)
	if !ok {
		return g
	}
	for i, it := range g.Items {
		if it.ID == id {
			g.Hover = i
			return g
		}
	}
	return g
}

// ApplyClick processes a click at screen coords. No-op when closed.
// Outcome:
//   - Click on an enabled item: Confirmed=true, Chosen=item.ID, Open=false
//   - Click on a disabled item: no-op (menu stays open)
//   - Click on border / padding inside menu: Hover updated if relevant; menu stays open
//   - Click outside menu: Open=false, Cancelled=true
func (g Group[ID]) ApplyClick(x, y int) Outcome[ID] {
	if !g.Open {
		return Outcome[ID]{Group: g}
	}
	if !g.Inside(x, y) {
		g.Open = false
		return Outcome[ID]{Group: g, Cancelled: true}
	}
	id, ok := g.HitTest(x, y)
	if !ok {
		// Inside menu but on border / padding — no-op.
		return Outcome[ID]{Group: g}
	}
	// Find item index for the chosen ID.
	for i, it := range g.Items {
		if it.ID == id {
			if it.Disabled {
				return Outcome[ID]{Group: g}
			}
			g.Hover = i
			g.Open = false
			return Outcome[ID]{Group: g, Confirmed: true, Chosen: id}
		}
	}
	return Outcome[ID]{Group: g}
}

// ChosenMsg is emitted, via the cmd returned from Update, when the user
// confirms an item. Handle it in your own Update to act on the choice.
type ChosenMsg[ID comparable] struct{ ID ID }

// CancelledMsg is emitted, via the cmd returned from Update, when the
// menu is dismissed (Cancel key, or a click outside).
type CancelledMsg struct{}

// Update routes a Bubble Tea message to the menu and returns the new
// Group plus a cmd carrying any semantic event (ChosenMsg / CancelledMsg);
// the cmd is nil when nothing notable happened. It is the one-line,
// drop-in alternative to ApplyKey / ApplyClick:
//
//	case tea.KeyMsg, tea.MouseMsg:
//	    m.menu, cmd = m.menu.Update(msg)
//	    return m, cmd
//	case menu.ChosenMsg[Action]:
//	    return handleAction(m, msg.ID), nil
//
// Update is a no-op while the menu is closed. Because a Group owns its
// screen position (AnchorX/AnchorY), it resolves both key and mouse
// events with no coordinate wiring from the caller. Reach for ApplyKey /
// ApplyClick when you'd rather inspect the Outcome synchronously.
func (g Group[ID]) Update(msg tea.Msg) (Group[ID], tea.Cmd) {
	if !g.Open {
		return g, nil
	}
	switch m := msg.(type) {
	case tea.KeyMsg:
		return outcomeCmd(g.ApplyKey(m.String()))
	case tea.MouseMsg:
		switch m.Action {
		case tea.MouseActionMotion:
			return g.HoverAt(m.X, m.Y), nil
		case tea.MouseActionPress:
			return outcomeCmd(g.ApplyClick(m.X, m.Y))
		}
	}
	return g, nil
}

func outcomeCmd[ID comparable](o Outcome[ID]) (Group[ID], tea.Cmd) {
	switch {
	case o.Confirmed:
		id := o.Chosen
		return o.Group, func() tea.Msg { return ChosenMsg[ID]{ID: id} }
	case o.Cancelled:
		return o.Group, func() tea.Msg { return CancelledMsg{} }
	}
	return o.Group, nil
}

// View renders the menu — an alias for Render that matches the View()
// convention of Bubble Tea models and charmbracelet/bubbles components.
func (g Group[ID]) View() string { return g.Render() }

// stepHover returns the next index in items moving by delta, skipping
// disabled items. When wrap is true, wraps at the edges; otherwise
// clamps. Returns -1 when there are no enabled items.
func stepHover[ID comparable](items []Item[ID], current, delta int, wrap bool) int {
	n := len(items)
	if n == 0 {
		return -1
	}
	// Find the first enabled item if current is invalid.
	idx := current
	if idx < 0 || idx >= n {
		idx = -delta // start from the appropriate end
	}
	for step := 0; step < n; step++ {
		idx += delta
		if wrap {
			idx = ((idx % n) + n) % n
		} else {
			if idx < 0 {
				idx = 0
			}
			if idx >= n {
				idx = n - 1
			}
		}
		if !items[idx].Disabled {
			return idx
		}
		if !wrap && (idx == 0 || idx == n-1) {
			// Hit the edge with no enabled item — stay
			return current
		}
	}
	return current
}

func contains[T comparable](xs []T, v T) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}
