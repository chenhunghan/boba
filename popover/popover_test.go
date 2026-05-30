package popover_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/chenhunghan/boba/overlay"
	"github.com/chenhunghan/boba/popover"
	"github.com/chenhunghan/boba/popup"
)

func TestRenderClosedIsEmpty(t *testing.T) {
	p := popover.Popover{Content: "hi"}
	if p.Render() != "" {
		t.Fatalf("closed popover should render empty, got %q", p.Render())
	}
}

func TestRenderShape(t *testing.T) {
	p := popover.Popover{Content: "ab\nlonger", Open: true}
	// Width: widest row "longer" (6) + 2 borders + 2 padding = 10.
	if got, want := p.Width(), 10; got != want {
		t.Fatalf("Width = %d, want %d", got, want)
	}
	// Height: 2 content rows + top + bottom = 4.
	if got, want := p.Height(), 4; got != want {
		t.Fatalf("Height = %d, want %d", got, want)
	}
	lines := strings.Split(p.Render(), "\n")
	if len(lines) != p.Height() {
		t.Fatalf("rendered %d lines, want %d", len(lines), p.Height())
	}
	for i, ln := range lines {
		if w := lipgloss.Width(ln); w != p.Width() {
			t.Fatalf("line %d width = %d, want %d", i, w, p.Width())
		}
	}
}

func TestRenderIsolatesEveryRow(t *testing.T) {
	p := popover.Popover{Content: "x\ny", Open: true}
	for i, ln := range strings.Split(p.Render(), "\n") {
		if !strings.HasPrefix(ln, "\x1b[0m") {
			t.Fatalf("row %d missing leading SGR reset: %q", i, ln)
		}
	}
}

func TestRenderCustomGlyphs(t *testing.T) {
	p := popover.Popover{
		Content: "x",
		Open:    true,
		Style: popover.Style{
			TopLeft: "+", TopRight: "+",
			BottomLeft: "+", BottomRight: "+",
			Horizontal: "-", Vertical: "|",
		},
	}
	top := strings.Split(p.Render(), "\n")[0]
	if !strings.Contains(top, "+---+") {
		t.Fatalf("top edge should use custom glyphs, got %q", top)
	}
}

func TestUpdateCancelCloses(t *testing.T) {
	p := popover.Popover{Content: "hi", Open: true}
	p, cmd := p.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if p.Open {
		t.Fatal("esc should close an open popover")
	}
	if cmd == nil {
		t.Fatal("closing should return a cmd")
	}
	if _, ok := cmd().(popover.ClosedMsg); !ok {
		t.Fatalf("got %#v, want popover.ClosedMsg", cmd())
	}
}

func TestUpdateCustomCancel(t *testing.T) {
	p := popover.Popover{Content: "hi", Open: true, Cancel: []string{"q"}}
	// esc is no longer a cancel key once Cancel is set explicitly.
	if p2, cmd := p.Update(tea.KeyMsg{Type: tea.KeyEsc}); !p2.Open || cmd != nil {
		t.Fatal("esc should be ignored when Cancel does not include it")
	}
	p, cmd := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if p.Open || cmd == nil {
		t.Fatal("q should close when it is the configured Cancel key")
	}
}

func TestUpdateClosedIsNoop(t *testing.T) {
	p := popover.Popover{Content: "hi"}
	if p2, cmd := p.Update(tea.KeyMsg{Type: tea.KeyEsc}); p2.Open || cmd != nil {
		t.Fatal("a closed popover should ignore keys")
	}
}

func TestUpdateIgnoresNonCancelKey(t *testing.T) {
	p := popover.Popover{Content: "hi", Open: true}
	if p2, cmd := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")}); !p2.Open || cmd != nil {
		t.Fatal("a non-cancel key should be a no-op")
	}
}

func TestUpdateIgnoresNonKeyMsg(t *testing.T) {
	p := popover.Popover{Content: "hi", Open: true}
	if p2, cmd := p.Update(tea.WindowSizeMsg{Width: 10}); !p2.Open || cmd != nil {
		t.Fatal("a non-key message should be a no-op")
	}
}

func TestOverClosedReturnsBackground(t *testing.T) {
	bg := "line one\nline two"
	p := popover.Popover{Content: "hi"}
	if got := p.Over(bg, 40, 20); got != bg {
		t.Fatalf("closed Over should return bg unchanged, got %q", got)
	}
}

func TestOverCompositesAtPlacedPosition(t *testing.T) {
	bg := blank(40, 20)
	p := popover.Popover{Content: "hi", Open: true, X: 2, Y: 2, W: 4, H: 1, Placement: popup.Below}

	x, y := popup.Place(p.X, p.Y, p.W, p.H, p.Width(), p.Height(), 40, 20, popup.Below)
	want := overlay.Overlay(bg, popup.Isolate(p.Render()), x, y)
	if got := p.Over(bg, 40, 20); got != want {
		t.Fatal("Over should match popup.Place + Isolate + Overlay composition")
	}
	// The placed row should differ from the untouched background row.
	gotRow := strings.Split(p.Over(bg, 40, 20), "\n")[y]
	if gotRow == strings.Split(bg, "\n")[y] {
		t.Fatal("panel did not land on the expected row")
	}
}

func blank(w, h int) string {
	rows := make([]string, h)
	for i := range rows {
		rows[i] = strings.Repeat(" ", w)
	}
	return strings.Join(rows, "\n")
}
