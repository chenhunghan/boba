// Package popover is a headless anchored floating panel: a bordered box of
// arbitrary content placed next to an anchor rectangle and composited over a
// background. It owns behavior — placement, sizing, ANSI isolation, and
// cancel-to-close — but every visual decision lives in a caller-supplied
// Style. Placement and overlay are delegated to the popup and overlay
// primitives, so the same composition rules apply as the rest of boba.
//
// Store a Popover on your model, route messages through Update, and composite
// it onto your view with Over:
//
//	m.help, cmd = m.help.Update(msg)
//	...
//	return m.help.Over(base, screenW, screenH)
package popover

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/chenhunghan/boba/overlay"
	"github.com/chenhunghan/boba/popup"
)

// Style holds the caller's surface and border chrome. Surface is applied to
// every owned cell (border + content + padding) so the panel reads as a
// distinct surface; its background, when set, fills the box uniformly. Border
// styles the box-drawing runes. Glyph fields fall back to box-drawing
// defaults when empty.
type Style struct {
	Surface lipgloss.Style
	Border  lipgloss.Style

	TopLeft, TopRight, BottomLeft, BottomRight string
	Horizontal, Vertical                       string
}

func orDefault(s, def string) string {
	if s != "" {
		return s
	}
	return def
}

func (s Style) topLeft() string     { return orDefault(s.TopLeft, "┌") }
func (s Style) topRight() string    { return orDefault(s.TopRight, "┐") }
func (s Style) bottomLeft() string  { return orDefault(s.BottomLeft, "└") }
func (s Style) bottomRight() string { return orDefault(s.BottomRight, "┘") }
func (s Style) horizontal() string  { return orDefault(s.Horizontal, "─") }
func (s Style) vertical() string    { return orDefault(s.Vertical, "│") }

// Popover is an anchored floating panel. Content is the body (newlines split
// rows). Open gates rendering and input. X, Y, W, H are the anchor rectangle
// in screen coordinates that Placement positions the panel against. Cancel
// are the keys that close the panel; nil falls back to ["esc"].
type Popover struct {
	Content   string
	Open      bool
	X, Y      int
	W, H      int
	Placement popup.Placement
	Cancel    []string
	Style     Style
}

// ClosedMsg is emitted, via the cmd from Update, when a Cancel key dismisses
// the panel.
type ClosedMsg struct{}

func (p Popover) lines() []string {
	return strings.Split(p.Content, "\n")
}

// Width returns the panel's outer width in cells: the widest content row plus
// side borders and one cell of padding per side.
func (p Popover) Width() int {
	max := 0
	for _, ln := range p.lines() {
		if w := lipgloss.Width(ln); w > max {
			max = w
		}
	}
	return max + 4
}

// Height returns the panel's outer height in cells: the content rows plus the
// top and bottom borders.
func (p Popover) Height() int {
	return len(p.lines()) + 2
}

// Render returns the bordered panel sized Width x Height, or "" when closed.
// Each row begins with an SGR reset so the panel owns its ANSI envelope and
// does not inherit lingering state from whatever Over composites it onto; the
// output carries no positioning.
func (p Popover) Render() string {
	if !p.Open {
		return ""
	}
	const sgrReset = "\x1b[0m"
	innerW := p.Width() - 2
	bodyW := innerW - 2

	border := withBackground(p.Style.Border, p.Style.Surface)
	surface := p.Style.Surface

	rows := make([]string, 0, p.Height())
	rows = append(rows, sgrReset+border.Render(
		p.Style.topLeft()+strings.Repeat(p.Style.horizontal(), innerW)+p.Style.topRight(),
	))
	for _, ln := range p.lines() {
		pad := bodyW - lipgloss.Width(ln)
		if pad < 0 {
			pad = 0
		}
		content := " " + ln + strings.Repeat(" ", pad) + " "
		rows = append(rows,
			sgrReset+border.Render(p.Style.vertical())+
				surface.Render(content)+
				border.Render(p.Style.vertical()),
		)
	}
	rows = append(rows, sgrReset+border.Render(
		p.Style.bottomLeft()+strings.Repeat(p.Style.horizontal(), innerW)+p.Style.bottomRight(),
	))
	return strings.Join(rows, "\n")
}

// View is an alias for Render, matching the Bubble Tea View() convention.
func (p Popover) View() string { return p.Render() }

// Over composites the panel onto bg and returns the result, or bg unchanged
// when closed. The panel is placed against the anchor rect via popup.Place
// (flipping when it would overflow), isolated so it cannot inherit bg's ANSI,
// and spliced in with overlay.Overlay.
func (p Popover) Over(bg string, screenW, screenH int) string {
	if !p.Open {
		return bg
	}
	panel := p.Render()
	x, y := popup.Place(p.X, p.Y, p.W, p.H, p.Width(), p.Height(), screenW, screenH, p.Placement)
	return overlay.Overlay(bg, popup.Isolate(panel), x, y)
}

// Update closes the panel on a Cancel key and emits ClosedMsg. It is a no-op
// while closed, and ignores non-key messages.
func (p Popover) Update(msg tea.Msg) (Popover, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok || !p.Open {
		return p, nil
	}
	cancel := p.Cancel
	if cancel == nil {
		cancel = []string{"esc"}
	}
	k := key.String()
	for _, c := range cancel {
		if k == c {
			p.Open = false
			return p, func() tea.Msg { return ClosedMsg{} }
		}
	}
	return p, nil
}

// withBackground copies src's background onto style when src sets one, so the
// border shares the surface's fill; otherwise style is returned unchanged.
func withBackground(style, src lipgloss.Style) lipgloss.Style {
	if bg := src.GetBackground(); bg != nil {
		return style.Background(bg)
	}
	return style
}
