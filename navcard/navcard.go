// Package navcard renders headless "navigation card" cells: a
// list-row component with a thick left bar, optional title, subtitle,
// description, and inline action buttons.
//
// Like Package button, navcard is headless: the package owns layout
// and behavior (which slots render in what order, how text is padded,
// where buttons go) but each Card carries its own Style. Inline
// action Buttons in turn carry their own button.Style, so a card's
// styling decisions are independent of its inline buttons'. base-ui's
// pattern: separate component contract from component appearance, and
// compose appearance per instance.
//
// A typical card with all slots looks like:
//
//	▌
//	▌  Title in bold
//	▌  subtitle in dimmer text
//	▌  description on its own row
//	▌   [Edit] [Delete]
//	▌
//
// When a slot is empty (e.g., no Subtitle), its row is omitted, so
// cards with fewer slots are shorter. Buttons that don't all fit in
// one row (within the card's content width) wrap to additional rows
// below. Card.Height(width) reports the actual row count for the
// given card width — height is width-dependent because of wrapping.
//
// For cells that don't fit the title/subtitle/description shape, set
// Custom and the package delegates the entire body to your function.
package navcard

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/chenhunghan/boba/button"
)

// State mirrors button.State for cards: Inactive (default), Hover
// (cursor over the card), Active (selected + parent panel focused).
type State int

const (
	StateInactive State = iota
	StateHover
	StateActive
)

// StateStyle bundles the styles a card uses in one state. All five
// fg/bg pairs (Bar, Fill, Title, Subtitle, Description) should share
// a consistent background so the card reads as one uniform block.
//
// Bar is the lipgloss.Style applied to BarChar, the glyph drawn on
// the left edge of every row in the card. BarChar defaults to "│"
// when empty; "▌" gives a thicker, base-ui-like accent bar.
//
// Fill is the style for empty rows (top/bottom padding) and for the
// trailing whitespace after a text row's content.
type StateStyle struct {
	Bar         lipgloss.Style
	BarChar     string
	Fill        lipgloss.Style
	Title       lipgloss.Style
	Subtitle    lipgloss.Style
	Description lipgloss.Style
}

// Style holds one StateStyle per state. Inline buttons carry their
// own button.Style on each button.Button — this struct intentionally
// does not include button styling, mirroring the per-instance pattern.
type Style struct {
	Inactive StateStyle
	Hover    StateStyle
	Active   StateStyle
}

// ForState returns the StateStyle for the given state.
func (s Style) ForState(state State) StateStyle {
	switch state {
	case StateActive:
		return s.Active
	case StateHover:
		return s.Hover
	default:
		return s.Inactive
	}
}

// Card is a single nav cell. All slot fields are optional; empty
// slots are not rendered. Custom, if non-nil, takes over the body
// entirely — useful for cells that don't fit the slot shape. Style
// is owned by the card itself: each instance can have its own look.
type Card struct {
	Title        string
	Subtitle     string
	Description  string
	Buttons      []button.Button // left-aligned; may wrap onto multiple rows
	RightButtons []button.Button // right-aligned on the last button row

	// Custom replaces the default body renderer. Receives the inner
	// content width (the card's outer width minus 1 for the bar) and
	// returns a multi-line string. Each line should be exactly
	// `contentWidth` cells wide; the package prepends the bar.
	Custom func(contentWidth int, state State, style Style) string

	// CustomHeight reports how many rows Custom will produce. Required
	// when Custom is set so the parent can lay out variable-height
	// cards correctly.
	CustomHeight int

	Style Style
}

// rightButtonInset is how many cells of breathing room sit between
// the rightmost right-aligned button and the card's right edge.
// Gives right buttons a visual margin from whatever border the
// parent panel draws on the card's right side.
const rightButtonInset = 1

// buttonAreaWidth is how many cells are available for the button
// row(s) inside a card of the given outer width: outer width minus
// 1 for the bar minus 2 for the left indent.
func buttonAreaWidth(outerWidth int) int {
	return outerWidth - 1 - 2
}

// buttonOuterWidth is the rendered width of one button (auto-sizing
// to badge + text + 2 cells of padding) — mirrors button.Render's
// auto-width formula.
func buttonOuterWidth(b button.Button) int {
	return lipgloss.Width(b.Badge) + lipgloss.Width(b.Text) + 2
}

// wrapButtons splits buttons into rows that each fit within the
// given available cell width, preserving order. Adjacent buttons in
// the same row are separated by `gap` cells. A button wider than
// `available` still occupies its own row (it overflows, but the
// row keeps it). Returns nil for an empty input or non-positive
// available width.
func wrapButtons(buttons []button.Button, available, gap int) [][]button.Button {
	if len(buttons) == 0 || available <= 0 {
		return nil
	}
	var rows [][]button.Button
	var current []button.Button
	used := 0
	for _, btn := range buttons {
		w := buttonOuterWidth(btn)
		next := used + w
		if len(current) > 0 {
			next += gap
		}
		if next > available && len(current) > 0 {
			rows = append(rows, current)
			current = []button.Button{btn}
			used = w
		} else {
			current = append(current, btn)
			used = next
		}
	}
	if len(current) > 0 {
		rows = append(rows, current)
	}
	return rows
}

// buttonRows is a small helper that returns the wrapped button rows
// for the given card-outer-width, or nil when the card has no
// buttons. Used by both rendering and hit-testing so the row count
// stays consistent.
func (c Card) buttonRows(width int) [][]button.Button {
	if len(c.Buttons) == 0 {
		return nil
	}
	return wrapButtons(c.Buttons, buttonAreaWidth(width), 1)
}

// Height reports the number of rows this card occupies when rendered
// at the given outer width. Height depends on width because the
// inline button list may wrap to multiple rows when it doesn't fit
// in the card's content area. When Custom is set, returns
// CustomHeight (caller's responsibility).
func (c Card) Height(width int) int {
	if c.Custom != nil {
		return c.CustomHeight
	}
	h := 2 // top + bottom padding
	if c.Title != "" {
		h++
	}
	if c.Subtitle != "" {
		h++
	}
	if c.Description != "" {
		h++
	}
	rows := c.buttonRows(width)
	h += len(rows)
	if len(rows) == 0 && len(c.RightButtons) > 0 {
		// Right-only buttons still need one row.
		h++
	}
	return h
}

// Render produces the card as a multi-line string of width × Height(width)
// visible cells. hoverButton is the index (into c.Buttons) of the
// currently hovered inline button, or -1 for no hover.
func (c Card) Render(state State, width, hoverButton int) string {
	s := c.Style.ForState(state)

	barChar := s.BarChar
	if barChar == "" {
		barChar = "│"
	}
	bar := s.Bar.Render(barChar)

	contentWidth := width - 1
	if contentWidth < 1 {
		contentWidth = 1
	}

	var bodyRows []string
	if c.Custom != nil {
		bodyRows = strings.Split(c.Custom(contentWidth, state, c.Style), "\n")
	} else {
		bodyRows = c.defaultBody(s, width, contentWidth, hoverButton)
	}

	for i, row := range bodyRows {
		bodyRows[i] = bar + row
	}
	return strings.Join(bodyRows, "\n")
}

// defaultBody composes the body when Custom is not set. Each returned
// row is exactly contentWidth visible cells. Buttons that don't all
// fit in one row are wrapped onto subsequent rows.
func (c Card) defaultBody(s StateStyle, width, contentWidth, hoverButton int) []string {
	var rows []string

	rows = append(rows, fillRow(s.Fill, contentWidth)) // top padding

	if c.Title != "" {
		rows = append(rows, textRow(s.Title, s.Fill, "  "+c.Title, contentWidth))
	}
	if c.Subtitle != "" {
		rows = append(rows, textRow(s.Subtitle, s.Fill, "  "+c.Subtitle, contentWidth))
	}
	if c.Description != "" {
		rows = append(rows, textRow(s.Description, s.Fill, "  "+c.Description, contentWidth))
	}

	btnRows := c.buttonRows(width)
	// When there are only right buttons (no left), still render
	// one button row so they have a place to sit.
	if len(btnRows) == 0 && len(c.RightButtons) > 0 {
		btnRows = append(btnRows, nil)
	}

	// Total left-button count — the global index of the first
	// right button equals this.
	leftCount := 0
	for _, r := range btnRows {
		leftCount += len(r)
	}

	btnIdx := 0
	for i, group := range btnRows {
		isLastRow := i == len(btnRows)-1
		// Find the hovered LEFT-button index for this row.
		rowHover := -1
		if hoverButton >= btnIdx && hoverButton < btnIdx+len(group) {
			rowHover = hoverButton - btnIdx
		}
		leftStr := button.HorizontalRow(group, rowHover, 1, 1)
		leftWidth := lipgloss.Width(leftStr)
		prefix := s.Fill.Render("  ")

		if isLastRow && len(c.RightButtons) > 0 {
			// Right buttons join the last row, right-aligned, with
			// rightButtonInset cells of breathing room before the
			// card's right edge.
			rightHover := -1
			if hoverButton >= leftCount && hoverButton < leftCount+len(c.RightButtons) {
				rightHover = hoverButton - leftCount
			}
			rightStr := button.HorizontalRow(c.RightButtons, rightHover, 1, 1)
			rightWidth := lipgloss.Width(rightStr)
			spacer := contentWidth - 2 - leftWidth - rightWidth - rightButtonInset
			if spacer < 1 {
				spacer = 1
			}
			rows = append(rows,
				prefix+leftStr+
					s.Fill.Render(strings.Repeat(" ", spacer))+
					rightStr+
					s.Fill.Render(strings.Repeat(" ", rightButtonInset)),
			)
		} else {
			pad := contentWidth - 2 - leftWidth
			if pad < 0 {
				pad = 0
			}
			rows = append(rows, prefix+leftStr+s.Fill.Render(strings.Repeat(" ", pad)))
		}
		btnIdx += len(group)
	}

	rows = append(rows, fillRow(s.Fill, contentWidth)) // bottom padding
	return rows
}

// firstButtonRowOffset returns the row index (within this card)
// where the first button row is rendered, or -1 if there are no
// buttons (left or right). Layout matches defaultBody: 1 top-pad
// row, then optional title, subtitle, description rows, then the
// button rows, then bottom-pad.
func (c Card) firstButtonRowOffset() int {
	if len(c.Buttons) == 0 && len(c.RightButtons) == 0 {
		return -1
	}
	row := 1 // skip top padding
	if c.Title != "" {
		row++
	}
	if c.Subtitle != "" {
		row++
	}
	if c.Description != "" {
		row++
	}
	return row
}

// buttonHitTest returns the global button index at panel-local
// (x, yInCard) for a card rendered at the given outer width, or -1
// if (x, yInCard) is not on any button. The global index counts
// left buttons first (across wrapped rows) then right buttons:
//
//   - 0 .. len(Buttons)-1   → c.Buttons[i]
//   - len(Buttons) ..       → c.RightButtons[i-len(Buttons)]
//
// Right buttons sit on the LAST left-button row, right-aligned.
// When there are no left buttons but right buttons exist, an
// implicit empty last row is used to anchor them.
//
// Coordinate layout within a card:
//
//	x=0:               bar character
//	x=1..2:            2-cell left padding (the "  " prefix)
//	x=3..:             left button row starts here (gap=1)
//	x=width-rightW..:  right buttons sit here (right-aligned)
//
// Each inline button auto-sizes to badge+text+2 cells.
func (c Card) buttonHitTest(x, yInCard, width int) int {
	first := c.firstButtonRowOffset()
	if first < 0 {
		return -1
	}
	rows := c.buttonRows(width)
	if len(rows) == 0 && len(c.RightButtons) > 0 {
		rows = append(rows, nil) // anchor row for right-only layout
	}
	rowIdx := yInCard - first
	if rowIdx < 0 || rowIdx >= len(rows) {
		return -1
	}
	const startX = 3 // bar + 2-cell prefix
	const buttonGap = 1

	// Index of the first button on this row (counting only left).
	btnIdx := 0
	for i := 0; i < rowIdx; i++ {
		btnIdx += len(rows[i])
	}
	// Try left buttons on this row.
	cursor := startX
	for _, btn := range rows[rowIdx] {
		w := buttonOuterWidth(btn)
		if x >= cursor && x < cursor+w {
			return btnIdx
		}
		cursor += w + buttonGap
		btnIdx++
	}
	// Right buttons live on the last row only.
	isLastRow := rowIdx == len(rows)-1
	if !isLastRow || len(c.RightButtons) == 0 {
		return -1
	}
	leftCount := 0
	for _, r := range rows {
		leftCount += len(r)
	}
	// Compute right buttons' visible width and starting x. Right
	// buttons sit `rightButtonInset` cells in from the card's right
	// edge — matches the layout in defaultBody.
	rightW := 0
	for i, btn := range c.RightButtons {
		if i > 0 {
			rightW += buttonGap
		}
		rightW += buttonOuterWidth(btn)
	}
	rightStart := width - rightW - rightButtonInset
	cursor = rightStart
	rIdx := leftCount
	for _, btn := range c.RightButtons {
		w := buttonOuterWidth(btn)
		if x >= cursor && x < cursor+w {
			return rIdx
		}
		cursor += w + buttonGap
		rIdx++
	}
	return -1
}

// Hit is the result of a coordinate query against a Stack. Card is
// the index of the card under the cursor (-1 if none — past the
// stack or in a gap). Button is the index of the inline button
// within that card if the cursor is on a button position
// (-1 otherwise, including when Card is also -1).
type Hit struct {
	Card   int
	Button int
}

// Stack is a vertical list of cards with optional blank rows between
// them. Cards may have variable heights via Card.Height(width); both
// Render and HitTest walk the same height-summing logic so they
// cannot drift out of sync.
//
// Width is the outer width per card. It must be set; HitTest needs
// it because card height is width-dependent (button rows wrap),
// and Render needs it for the same reason. Storing width on the
// Stack keeps the HitTest signature uniform with the rest of the
// design system: HitTest(x, y int) — same shape as button.Stack
// and tab.Group.
type Stack struct {
	Cards       []Card
	Width       int  // outer width per card; required for hit-testing and rendering
	Gap         int  // blank rows between consecutive cards
	Selected    int  // index of the chosen card (visible only when Active)
	Hover       int  // hovered card index; -1 = none
	HoverButton int  // hovered inline-button index within Hover'd card; -1 = none
	Active      bool // whether the parent panel currently has focus

	// Keyboard bindings for Update (nil → sensible defaults): Up/Down
	// move Selected, Confirm activates the selected card; Wrap cycles.
	Up, Down, Confirm []string
	Wrap              bool
}

// Render produces the stack as a multi-line string of Width cells
// per row. Each card is rendered with the State derived from Active,
// Selected, and Hover (priority: Active > Hover > Inactive). Only
// the hovered card receives a non-negative HoverButton; others get -1.
// The Stack does not own card styling — each Card carries its own.
func (s Stack) Render() string {
	parts := make([]string, 0, len(s.Cards)*2)
	for i, card := range s.Cards {
		if i > 0 && s.Gap > 0 {
			parts = append(parts, strings.Repeat("\n", s.Gap-1))
		}
		state := StateInactive
		switch {
		case s.Active && i == s.Selected:
			state = StateActive
		case i == s.Hover:
			state = StateHover
		}
		hoverBtn := -1
		if i == s.Hover {
			hoverBtn = s.HoverButton
		}
		parts = append(parts, card.Render(state, s.Width, hoverBtn))
	}
	return strings.Join(parts, "\n")
}

// HitTest returns a Hit identifying the card and inline-button (if
// any) at panel-local (x, y). The Stack's own Width drives per-card
// heights, so HitTest computes the same wrap as Render. The (x, y)
// coordinates are in the stack's own coordinate space; the caller
// is responsible for translating from app-global coordinates.
func (s Stack) HitTest(x, y int) Hit {
	if y < 0 {
		return Hit{Card: -1, Button: -1}
	}
	cursor := 0
	for i, card := range s.Cards {
		h := card.Height(s.Width)
		if y < cursor+h {
			return Hit{Card: i, Button: card.buttonHitTest(x, y-cursor, s.Width)}
		}
		cursor += h
		if i < len(s.Cards)-1 {
			cursor += s.Gap
			if y < cursor {
				return Hit{Card: -1, Button: -1} // in a gap
			}
		}
	}
	return Hit{Card: -1, Button: -1}
}

// CardActivatedMsg is emitted — via the cmd from Update or ClickAt —
// when a card is activated (the Confirm key, or a click on the card
// body). Card is the activated card index.
type CardActivatedMsg struct{ Card int }

// ButtonActivatedMsg is emitted — via the cmd from ClickAt — when an
// inline action button inside a card is clicked. Card is the card
// index; Button is the inline-button index within it (left buttons
// first, then right buttons), matching Hit.Button.
type ButtonActivatedMsg struct{ Card, Button int }

// Update routes a key message to the stack: Up/Down move Selected and
// Confirm activates the selected card (emitting CardActivatedMsg). It
// is a no-op unless Active, and ignores non-key messages — mouse goes
// through ClickAt / HoverAt, which take panel-local coordinates. Inline
// buttons are activated by clicking, not by keyboard.
//
//	case tea.KeyMsg:
//	    m.nav, cmd = m.nav.Update(msg)
//	    return m, cmd
//	case navcard.CardActivatedMsg:
//	    return open(m, msg.Card), nil
func (s Stack) Update(msg tea.Msg) (Stack, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok || !s.Active || len(s.Cards) == 0 {
		return s, nil
	}
	switch k := key.String(); {
	case contains(keysOr(s.Up, "up", "k"), k):
		s.Selected = step(s.Selected, -1, len(s.Cards), s.Wrap)
	case contains(keysOr(s.Down, "down", "j"), k):
		s.Selected = step(s.Selected, +1, len(s.Cards), s.Wrap)
	case contains(keysOr(s.Confirm, "enter"), k):
		return s, fire(CardActivatedMsg{Card: s.Selected})
	}
	return s, nil
}

// ClickAt applies a click at panel-local (x, y) — typically the LocalX
// / LocalY from panel.HitTest. A hit on an inline button emits
// ButtonActivatedMsg; a hit on a card body selects that card and emits
// CardActivatedMsg; a miss is a no-op.
func (s Stack) ClickAt(x, y int) (Stack, tea.Cmd) {
	hit := s.HitTest(x, y)
	switch {
	case hit.Card < 0:
		return s, nil
	case hit.Button >= 0:
		return s, fire(ButtonActivatedMsg(hit))
	default:
		s.Selected = hit.Card
		return s, fire(CardActivatedMsg{Card: hit.Card})
	}
}

// HoverAt sets Hover and HoverButton from panel-local (x, y), clearing
// them to -1 when the coordinate misses the stack.
func (s Stack) HoverAt(x, y int) Stack {
	hit := s.HitTest(x, y)
	s.Hover = hit.Card
	s.HoverButton = hit.Button
	return s
}

// View renders the stack — an alias for Render that matches the View()
// convention of Bubble Tea models and charmbracelet/bubbles components.
func (s Stack) View() string { return s.Render() }

func keysOr(binding []string, def ...string) []string {
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

// fillRow returns `width` cells of fillStyle.
func fillRow(fillStyle lipgloss.Style, width int) string {
	return fillStyle.Render(strings.Repeat(" ", width))
}

// textRow renders `text` with textStyle and pads to `width` with
// fillStyle. Both styles must use the same background for the row
// to look uniform. Text wider than `width` is truncated with an
// ellipsis so the row never overflows — overflow would corrupt the
// surrounding panel border (box.Render byte-truncates wide rows).
func textRow(textStyle, fillStyle lipgloss.Style, text string, width int) string {
	text = truncateToWidth(text, width)
	visible := lipgloss.Width(text)
	pad := width - visible
	if pad < 0 {
		pad = 0
	}
	return textStyle.Render(text) + fillStyle.Render(strings.Repeat(" ", pad))
}

// truncateToWidth shortens s rune-wise so its visible width is at
// most `width` cells. When truncation occurs, the last visible cell
// is replaced with "…" so the cut is signalled. Returns "" for
// non-positive widths.
func truncateToWidth(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= width {
		return s
	}
	if width == 1 {
		return "…"
	}
	var b strings.Builder
	cells := 0
	target := width - 1 // leave one cell for the ellipsis
	for _, r := range s {
		rw := lipgloss.Width(string(r))
		if cells+rw > target {
			break
		}
		b.WriteRune(r)
		cells += rw
	}
	b.WriteRune('…')
	return b.String()
}
