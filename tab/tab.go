// Package tab is a tabbed-navigation component for Bubble Tea apps.
// It owns the structural concerns — selection state, the keyboard
// step (`[` / `]`), header layout and hit-testing, dynamic add/close
// — and delegates content to a per-tab tea.Model. The visual is
// headless: each Tab carries its own Style.
//
// A Group is a value-type container. Caller stores it on their model
// and constructs/mutates it via the methods, which return updated
// copies. Nothing is hidden in package-level state.
//
// Sub-model integration is the load-bearing design choice: each Tab
// holds a tea.Model that gets its messages routed via UpdateActive
// (active tab only) or UpdateAll (every tab). AddTab returns the cmd
// from Model.Init() so initial side effects run when the tab opens.
//
// The one-line drop-in is Update (see its docs). The lower-level
// wiring it wraps, plus broadcast via UpdateAll on resize:
//
//	case tea.KeyMsg:
//	    if m.focus == focusMain {
//	        if m.tabs.IsBound(msg.String()) {
//	            m.tabs = m.tabs.ApplyKey(msg.String())
//	        } else {
//	            var cmd tea.Cmd
//	            m.tabs, cmd = m.tabs.UpdateActive(msg)
//	            return m, cmd
//	        }
//	    }
//	case tea.WindowSizeMsg:
//	    var cmd tea.Cmd
//	    m.tabs, cmd = m.tabs.UpdateAll(tab.SizeMsg{Width: w, Height: h})
//	    return m, cmd
//
// And rendering:
//
//	main := box.Box{ Body: m.tabs.Render(width-2), ... }
//
// For tabs whose content is just static text and doesn't need its own
// state, use tab.Static(s) as the Model — it's a no-op tea.Model that
// always renders s.
package tab

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// State is the visual state of one tab header.
type State int

const (
	// StateInactive: not selected, not under cursor.
	StateInactive State = iota
	// StateHover: cursor is over this tab's header.
	StateHover
	// StateActive: this tab is currently selected — its content shows.
	// Active is independent of whether the tab Group has keyboard focus,
	// because the user always needs to see *which* tab they're looking
	// at, regardless of focus.
	StateActive
)

// Style holds the per-state visual rules for a single tab. Border
// styling is separate from per-state styling so the tab's frame
// stays visually constant while only the title row content changes
// across states. SelectedBar replaces Border on the top edge of the
// selected (Active) tab — typically a thick colored line that says
// "this is the tab you're looking at."
type Style struct {
	// Per-state styles applied to the title row (icon + label + close).
	Inactive lipgloss.Style
	Hover    lipgloss.Style
	Active   lipgloss.Style

	// Border applies to the box-drawing chars of inactive tabs (top
	// edge + merge-row glyphs) and to the gap/trailing dashes that
	// fill the merge row outside any specific tab.
	Border lipgloss.Style

	// HoverBar replaces Border on the TOP edge of the hovered tab
	// only. The merge row (the transition line that joins the tab to
	// the content panel below) stays at Border — lighting it would
	// visually imply selection and conflict with the actually-active
	// tab. When HoverBar is the zero Style, the hovered top edge
	// renders with no styling — set it explicitly to match Border or
	// to a dedicated hover color.
	HoverBar lipgloss.Style

	// SelectedBar replaces Border on the top edge and merge row of
	// the active tab. Typically a heavy/colored bar matching the
	// panel's accent color.
	SelectedBar lipgloss.Style

	// Glyphs. Empty fields fall back to the defaults documented below.
	TopChar       string // default "─"
	TopActiveChar string // default "━"
	SideChar      string // default "│"
	TopLeftChar   string // default "┌"
	TopRightChar  string // default "┐"
	CloseChar     string // default "×"
}

// ForState returns the per-state title style for state.
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

func (s Style) topChar() string       { return orDefault(s.TopChar, "─") }
func (s Style) topActiveChar() string { return orDefault(s.TopActiveChar, "━") }
func (s Style) sideChar() string      { return orDefault(s.SideChar, "│") }
func (s Style) topLeftChar() string   { return orDefault(s.TopLeftChar, "┌") }
func (s Style) topRightChar() string  { return orDefault(s.TopRightChar, "┐") }
func (s Style) closeChar() string     { return orDefault(s.CloseChar, "×") }

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

// Tab is a single tab in a Group. ID is the caller's identifier
// (must be unique within the group). Label is the visible title.
// Icon is an optional glyph rendered before the label (e.g., "🗎",
// "▶", "⬢"). Closable=true renders the close × and accepts close
// clicks. Model is the tea.Model whose View() supplies the tab's
// content area.
type Tab[ID comparable] struct {
	ID       ID
	Label    string
	Icon     string
	Style    Style
	Closable bool
	Model    tea.Model
}

// labelString builds the visible content of the tab — icon (with
// trailing space when present), label, and (when Closable) a space
// followed by the close glyph. This same string is embedded in the
// top edge row of the rendered tab.
func (t Tab[ID]) labelString() string {
	var b strings.Builder
	if t.Icon != "" {
		b.WriteString(t.Icon)
		b.WriteString(" ")
	}
	b.WriteString(t.Label)
	if t.Closable {
		b.WriteString(" ")
		b.WriteString(t.Style.closeChar())
	}
	return b.String()
}

// renderedWidth is the total width in cells of this tab's header.
//
//	┌─icon label ×─┐
//	^ ^ ^^^^^^^^^^ ^^
//	| | |          | top-right corner
//	| | label (icon + ' ' + label + ' ' + close)
//	| left dash before label
//	top-left corner
func (t Tab[ID]) renderedWidth() int {
	// 2 corner chars + 2 dashes flanking the label
	return 4 + lipgloss.Width(t.labelString())
}

// closeStart returns the offset within this tab where the close
// glyph starts, or -1 if the tab isn't closable.
func (t Tab[ID]) closeStart() int {
	if !t.Closable {
		return -1
	}
	closeW := lipgloss.Width(t.Style.closeChar())
	// Layout from the right: top-right corner(1), right dash(1),
	// close(closeW), ...
	return t.renderedWidth() - 1 - 1 - closeW
}

// render returns (topRow, mergeRow) for this tab in the given state.
//
// topRow is the top edge with the label embedded between the dashes:
//
//	┌─Output ×─┐
//
// mergeRow is the bottom transition. For the active tab it "opens"
// downward into the content area below; for inactive tabs it "closes"
// with a horizontal line that becomes part of the content's top edge:
//
//	active:    ┘                └      (open bottom)
//	inactive:  ┴────────────────┴      (line continues; junctions point up to ┌ and ┐)
//
// isFirst determines the leftmost glyph for inactive tabs: when
// nothing sits to the tab's left on the merge row, use a closed
// corner ('└') instead of a junction ('┴'). The active state always
// uses '┘' for its left glyph.
func (t Tab[ID]) render(state State, isFirst bool) (string, string) {
	style := t.Style
	labelStr := t.labelString()
	labelWidth := lipgloss.Width(labelStr)
	dash := style.topChar()

	// topStyle drives the tab's top edge; mergeStyle drives the
	// transition row that joins the tab to the content below. They
	// agree for inactive (both Border) and active (both SelectedBar),
	// but diverge for hover: only the top edge lights up so the
	// merge row continues to read as the content panel's top border
	// rather than implying selection.
	topStyle := style.Border
	mergeStyle := style.Border
	switch state {
	case StateActive:
		topStyle = style.SelectedBar
		mergeStyle = style.SelectedBar
	case StateHover:
		topStyle = style.HoverBar
	}
	titleStyle := style.ForState(state)

	// Row 0: ┌ ─ <label> ─ ┐
	topRow := topStyle.Render(style.topLeftChar()) +
		topStyle.Render(dash) +
		titleStyle.Render(labelStr) +
		topStyle.Render(dash) +
		topStyle.Render(style.topRightChar())

	// Row 1: bottom transition.
	interiorWidth := labelWidth + 2 // label + the two flanking dashes
	var leftCorner, middle, rightCorner string
	if state == StateActive {
		leftCorner = "┘"
		rightCorner = "└"
		middle = strings.Repeat(" ", interiorWidth)
	} else {
		if isFirst {
			leftCorner = "└"
		} else {
			leftCorner = "┴"
		}
		rightCorner = "┴" // line continues past the tab via gap/pad dashes
		middle = strings.Repeat(dash, interiorWidth)
	}
	mergeRow := mergeStyle.Render(leftCorner) +
		mergeStyle.Render(middle) +
		mergeStyle.Render(rightCorner)

	return topRow, mergeRow
}

// Hit is the result of a HitTest query. Found is false when x lies
// outside any tab (in a gap, before the first, or past the last).
// When Found, ID identifies the tab; Close is true when x lands
// specifically on the close glyph (caller should call CloseTab),
// false for any other position on the tab header (caller should
// call ApplyClick to select it).
type Hit[ID comparable] struct {
	Found bool
	ID    ID
	Close bool
}

// SizeMsg is the canonical message for telling a tab's Model how
// big its content area is. Sub-models that lay out based on size
// should handle this message and store the dimensions for View.
// Caller dispatches it via Group.UpdateAll from their parent's
// tea.WindowSizeMsg handler.
type SizeMsg struct {
	Width, Height int
}

// Group is the tabbed-navigation widget. Caller stores it on their
// model and constructs/mutates via methods. Selected is persistent
// state (which tab the group is currently showing); Tabs is
// structural; Next/Prev/Wrap/Gap are configuration.
//
// Per-render dynamic state — what's hovered, whether the group has
// focus — does NOT live on Group. It is passed to Render and
// ApplyKey as a RenderState / focused parameter, so the contract is
// explicit at every call site instead of stashed in mutable fields.
type Group[ID comparable] struct {
	Tabs []Tab[ID]
	Gap  int // cells between adjacent tab headers (default 0)

	Selected ID

	Next []string // step forward; nil → ["]"]
	Prev []string // step backward; nil → ["["]
	Wrap bool     // wrap at edges
}

// RenderState carries the per-render dynamic state needed to draw
// the tab header. Selected is persistent and lives on Group; Hover
// is derived (typically from the current mouse position) and varies
// per render, so it's passed in fresh each time. HasHover
// disambiguates "no hover" from "hover on the zero-value ID."
type RenderState[ID comparable] struct {
	Hover    ID
	HasHover bool
}

// HeaderHeight returns the number of rows the tab header occupies.
// Currently always 2 (top edge + title row).
func (g Group[ID]) HeaderHeight() int { return 2 }

// stateOf computes the visual state for the tab with the given ID,
// given the per-render hover state.
func (g Group[ID]) stateOf(id ID, state RenderState[ID]) State {
	if id == g.Selected {
		return StateActive
	}
	if state.HasHover && id == state.Hover {
		return StateHover
	}
	return StateInactive
}

// RenderHeader produces the two-row tab header strip, fitted to
// width. The first row holds the tab pills (top edges with embedded
// labels); the second row is the merge row that file-folder-attaches
// the tabs to whatever sits below — gap and trailing space on this
// row are filled with the dash glyph so the line is continuous.
// Tabs that don't fit are dropped from the right.
//
// state supplies per-render hover; pass the zero RenderState to
// render with no tab hovered.
func (g Group[ID]) RenderHeader(width int, state RenderState[ID]) string {
	if width < 1 {
		return ""
	}

	// Style and dash glyph for the merge-line continuations (gaps
	// between tabs, trailing pad to fill width). Use the first tab's
	// Border style if available; otherwise no styling.
	var lineStyle lipgloss.Style
	dash := "─"
	if len(g.Tabs) > 0 {
		lineStyle = g.Tabs[0].Style.Border
		dash = g.Tabs[0].Style.topChar()
	}

	var topRow, mergeRow strings.Builder
	used := 0
	for i, t := range g.Tabs {
		if i > 0 && g.Gap > 0 {
			topRow.WriteString(strings.Repeat(" ", g.Gap))
			mergeRow.WriteString(lineStyle.Render(strings.Repeat(dash, g.Gap)))
			used += g.Gap
		}
		tw := t.renderedWidth()
		if used+tw > width {
			break
		}
		st := g.stateOf(t.ID, state)
		top, merge := t.render(st, i == 0)
		topRow.WriteString(top)
		mergeRow.WriteString(merge)
		used += tw
	}
	if used < width {
		pad := width - used
		topRow.WriteString(strings.Repeat(" ", pad))
		mergeRow.WriteString(lineStyle.Render(strings.Repeat(dash, pad)))
	}
	return topRow.String() + "\n" + mergeRow.String()
}

// RenderContent returns the active tab's Model.View(). Empty string
// when no tab is selected or the active tab has no Model.
func (g Group[ID]) RenderContent() string {
	for _, t := range g.Tabs {
		if t.ID == g.Selected {
			if t.Model != nil {
				return t.Model.View()
			}
			return ""
		}
	}
	return ""
}

// Render is a convenience that joins RenderHeader and RenderContent
// vertically. Use the two methods separately when you need to
// compose them differently (e.g., put something between them).
func (g Group[ID]) Render(width int, state RenderState[ID]) string {
	header := g.RenderHeader(width, state)
	content := g.RenderContent()
	if content == "" {
		return header
	}
	return header + "\n" + content
}

// HitTest takes panel-local (x, y) and returns the hit info. Returns
// a not-found Hit when y is outside the header rows; otherwise
// performs the same x-based routing as before. Signature parity with
// button.Stack.HitTest and navcard.Stack.HitTest — every hit-testable
// surface in the design system takes (x, y int).
func (g Group[ID]) HitTest(x, y int) Hit[ID] {
	if y < 0 || y >= g.HeaderHeight() {
		return Hit[ID]{}
	}
	cursor := 0
	for i, t := range g.Tabs {
		if i > 0 && g.Gap > 0 {
			cursor += g.Gap
		}
		tw := t.renderedWidth()
		if x >= cursor && x < cursor+tw {
			rel := x - cursor
			if t.Closable {
				cs := t.closeStart()
				cw := lipgloss.Width(t.Style.closeChar())
				if rel >= cs && rel < cs+cw {
					return Hit[ID]{Found: true, ID: t.ID, Close: true}
				}
			}
			return Hit[ID]{Found: true, ID: t.ID}
		}
		cursor += tw
	}
	return Hit[ID]{}
}

// IsBound reports whether key is bound to a tab-cycling action
// (Next, Prev). Caller can use this to decide whether to consume
// the key with ApplyKey or forward it to the active tab's Model.
// Doesn't check Active — caller is responsible for gating on focus.
func (g Group[ID]) IsBound(key string) bool {
	next := g.Next
	if next == nil {
		next = []string{"]"}
	}
	prev := g.Prev
	if prev == nil {
		prev = []string{"["}
	}
	return contains(next, key) || contains(prev, key)
}

// ApplyKey processes a key. Caller must gate on focus before
// calling — ApplyKey does not check whether the group has keyboard
// ownership; it just routes the key through. When key is in Next
// (default "]") or Prev (default "[") and there's at least one tab,
// advances/retreats Selected. Otherwise returns g unchanged.
func (g Group[ID]) ApplyKey(key string) Group[ID] {
	if len(g.Tabs) == 0 {
		return g
	}
	next := g.Next
	if next == nil {
		next = []string{"]"}
	}
	prev := g.Prev
	if prev == nil {
		prev = []string{"["}
	}
	if contains(next, key) {
		g.Selected = step(g, +1)
		return g
	}
	if contains(prev, key) {
		g.Selected = step(g, -1)
		return g
	}
	return g
}

func step[ID comparable](g Group[ID], delta int) ID {
	n := len(g.Tabs)
	idx := -1
	for i, t := range g.Tabs {
		if t.ID == g.Selected {
			idx = i
			break
		}
	}
	if idx < 0 {
		// Selected not in Tabs (shouldn't happen normally).
		if delta > 0 {
			return g.Tabs[0].ID
		}
		return g.Tabs[n-1].ID
	}
	next := idx + delta
	if g.Wrap {
		next = ((next % n) + n) % n
	} else {
		if next < 0 {
			next = 0
		}
		if next >= n {
			next = n - 1
		}
	}
	return g.Tabs[next].ID
}

// ApplyClick selects the tab with id and returns the updated Group.
// No-op when id doesn't match any tab.
func (g Group[ID]) ApplyClick(id ID) Group[ID] {
	for _, t := range g.Tabs {
		if t.ID == id {
			g.Selected = id
			return g
		}
	}
	return g
}

// AddTab appends t and selects it. Returns the updated Group and
// the cmd from t.Model.Init() (or nil when Model is nil). No-op
// when a tab with t.ID already exists; caller should call
// ApplyClick to focus the existing tab in that case.
func (g Group[ID]) AddTab(t Tab[ID]) (Group[ID], tea.Cmd) {
	for _, existing := range g.Tabs {
		if existing.ID == t.ID {
			return g, nil
		}
	}
	g.Tabs = append(g.Tabs, t)
	g.Selected = t.ID
	if t.Model == nil {
		return g, nil
	}
	return g, t.Model.Init()
}

// CloseTab removes the tab with id and returns the updated Group.
// If the closed tab was the selected one, the next tab becomes
// selected (or the previous one if it was the last in the list).
// When the group becomes empty, Selected is reset to the zero ID.
// No-op when id doesn't match any tab.
func (g Group[ID]) CloseTab(id ID) Group[ID] {
	idx := -1
	for i, t := range g.Tabs {
		if t.ID == id {
			idx = i
			break
		}
	}
	if idx < 0 {
		return g
	}
	g.Tabs = append(g.Tabs[:idx], g.Tabs[idx+1:]...)
	if g.Selected != id {
		return g
	}
	if len(g.Tabs) == 0 {
		var zero ID
		g.Selected = zero
		return g
	}
	if idx >= len(g.Tabs) {
		g.Selected = g.Tabs[len(g.Tabs)-1].ID // was last → previous
	} else {
		g.Selected = g.Tabs[idx].ID // next slid into the slot
	}
	return g
}

// Find returns the Tab with the given ID and true, or zero+false.
func (g Group[ID]) Find(id ID) (Tab[ID], bool) {
	for _, t := range g.Tabs {
		if t.ID == id {
			return t, true
		}
	}
	var zero Tab[ID]
	return zero, false
}

// UpdateActive forwards msg to the active tab's Model and returns
// the updated Group with the new sub-model state. No-op when no
// tab is selected or the active tab has no Model.
func (g Group[ID]) UpdateActive(msg tea.Msg) (Group[ID], tea.Cmd) {
	for i := range g.Tabs {
		if g.Tabs[i].ID != g.Selected {
			continue
		}
		if g.Tabs[i].Model == nil {
			return g, nil
		}
		newModel, cmd := g.Tabs[i].Model.Update(msg)
		g.Tabs[i].Model = newModel
		return g, cmd
	}
	return g, nil
}

// UpdateAll forwards msg to every tab's Model and returns the
// updated Group with all sub-models updated, batching their cmds
// via tea.Batch. Use for messages every tab needs (SizeMsg, theme
// changes, periodic ticks).
func (g Group[ID]) UpdateAll(msg tea.Msg) (Group[ID], tea.Cmd) {
	var cmds []tea.Cmd
	for i := range g.Tabs {
		if g.Tabs[i].Model == nil {
			continue
		}
		newModel, cmd := g.Tabs[i].Model.Update(msg)
		g.Tabs[i].Model = newModel
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	if len(cmds) == 0 {
		return g, nil
	}
	return g, tea.Batch(cmds...)
}

// Update is the one-line drop-in for routing a message to the group: a
// key bound to tab-cycling (Next/Prev) advances Selected; every other
// message is forwarded to the active tab's Model via UpdateActive, whose
// cmd is returned. Like ApplyKey it does not gate on focus — route a
// message here only when the tab group owns the keyboard. Broadcast
// messages (e.g. SizeMsg on resize) still go through UpdateAll.
//
//	case tea.KeyMsg:
//	    if m.focus == focusMain {
//	        m.tabs, cmd = m.tabs.Update(msg)
//	        return m, cmd
//	    }
func (g Group[ID]) Update(msg tea.Msg) (Group[ID], tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok && g.IsBound(key.String()) {
		return g.ApplyKey(key.String()), nil
	}
	return g.UpdateActive(msg)
}

// ClickAt applies a click at panel-local (x, y) — typically the LocalX /
// LocalY from panel.HitTest. A click on a tab's close glyph closes it; a
// click elsewhere on a tab header selects it; a miss is a no-op.
func (g Group[ID]) ClickAt(x, y int) (Group[ID], tea.Cmd) {
	switch hit := g.HitTest(x, y); {
	case !hit.Found:
		return g, nil
	case hit.Close:
		return g.CloseTab(hit.ID), nil
	default:
		return g.ApplyClick(hit.ID), nil
	}
}

// HoverState derives the per-render RenderState from a panel-local mouse
// position — the tab analog of the other components' HoverAt. Hover is
// per-render state (not stored on Group), so pass the result to Render /
// RenderHeader rather than storing it. A miss yields the zero
// RenderState (no hover).
func (g Group[ID]) HoverState(x, y int) RenderState[ID] {
	if hit := g.HitTest(x, y); hit.Found {
		return RenderState[ID]{Hover: hit.ID, HasHover: true}
	}
	return RenderState[ID]{}
}

func contains[T comparable](xs []T, v T) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}
