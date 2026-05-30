// Package tooltip renders a headless text bubble anchored to a target rect
// and composited over a background. It owns layout — a bordered box around
// the content, placed beside the anchor and isolated so it does not inherit
// stray ANSI from the cells it covers — but every visual decision lives in a
// caller-supplied Style. There is no keyboard, hit-testing, or event: a
// tooltip is shown or hidden by the caller flipping Visible.
//
//	t := tooltip.Tooltip{
//	    Content:   "rename",
//	    Visible:   true,
//	    X:         12, Y: 4, W: 6, H: 1, // the anchor rect
//	    Placement: popup.Below,
//	    Style:     mystyle,
//	}
//	screen = t.Over(screen, screenW, screenH)
package tooltip

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/chenhunghan/boba/overlay"
	"github.com/chenhunghan/boba/popup"
)

// Style holds the caller's surface style plus the box-drawing glyphs. Surface
// is applied to every owned cell (border + content + padding); empty fields
// fall back to the standard box-drawing runes.
type Style struct {
	// Surface styles every cell of the bubble. Set its background so the
	// tooltip reads as a distinct surface over whatever it covers.
	Surface lipgloss.Style

	// Glyphs. Empty fields fall back to the box-drawing defaults.
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

// Tooltip is the bubble. Content is the text shown inside the border; multi-
// line content (split on "\n") is laid out one line per row. X, Y, W, H are
// the anchor rectangle the bubble is placed against; Placement is the
// preferred side. Visible gates both Render and Over.
type Tooltip struct {
	Content    string
	Visible    bool
	X, Y, W, H int
	Placement  popup.Placement
	Style      Style
}

func (t Tooltip) lines() []string { return strings.Split(t.Content, "\n") }

// contentWidth is the widest content line in cells.
func (t Tooltip) contentWidth() int {
	max := 0
	for _, ln := range t.lines() {
		if w := lipgloss.Width(ln); w > max {
			max = w
		}
	}
	return max
}

// Width is the bubble's outer width in cells (2 borders + content). Zero when
// not Visible.
func (t Tooltip) Width() int {
	if !t.Visible {
		return 0
	}
	return t.contentWidth() + 2
}

// Height is the bubble's outer height in rows (top border + content + bottom
// border). Zero when not Visible.
func (t Tooltip) Height() int {
	if !t.Visible {
		return 0
	}
	return len(t.lines()) + 2
}

// Render returns the bordered bubble sized Width x Height, or "" when not
// Visible.
//
// Each row begins with an explicit SGR reset ("\x1b[0m") so the bubble does
// not inherit lingering ANSI from cells it overlays; lipgloss's trailing
// reset after each styled piece isolates the remaining seams within a row.
func (t Tooltip) Render() string {
	if !t.Visible {
		return ""
	}
	const sgrReset = "\x1b[0m"
	st := t.Style.Surface
	innerW := t.contentWidth()

	var rows []string
	rows = append(rows, sgrReset+st.Render(
		t.Style.topLeft()+
			strings.Repeat(t.Style.horizontal(), innerW)+
			t.Style.topRight(),
	))
	for _, ln := range t.lines() {
		pad := innerW - lipgloss.Width(ln)
		if pad < 0 {
			pad = 0
		}
		rows = append(rows,
			sgrReset+st.Render(t.Style.vertical())+
				st.Render(ln+strings.Repeat(" ", pad))+
				st.Render(t.Style.vertical()),
		)
	}
	rows = append(rows, sgrReset+st.Render(
		t.Style.bottomLeft()+
			strings.Repeat(t.Style.horizontal(), innerW)+
			t.Style.bottomRight(),
	))
	return strings.Join(rows, "\n")
}

// View is an alias for Render, matching the Bubble Tea View() convention.
func (t Tooltip) View() string { return t.Render() }

// Over composites the bubble onto bg, placing it beside the anchor rect via
// the chosen Placement (flipping to stay on screen) and isolating it so it
// owns its ANSI envelope. Returns bg unchanged when not Visible.
func (t Tooltip) Over(bg string, screenW, screenH int) string {
	if !t.Visible {
		return bg
	}
	box := t.Render()
	x, y := popup.Place(t.X, t.Y, t.W, t.H, t.Width(), t.Height(), screenW, screenH, t.Placement)
	return overlay.Overlay(bg, popup.Isolate(box), x, y)
}
