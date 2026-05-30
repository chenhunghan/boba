package scrollarea_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/chenhunghan/boba/scrollarea"
)

func lines(s string) []string { return strings.Split(s, "\n") }

const long = "l0\nl1\nl2\nl3\nl4\nl5\nl6\nl7\nl8\nl9"

func TestRenderShape(t *testing.T) {
	a := scrollarea.ScrollArea{Content: long, Height: 4, Width: 6}
	out := a.Render()
	rows := lines(out)
	if len(rows) != 4 {
		t.Fatalf("want 4 rows, got %d (%q)", len(rows), out)
	}
	for i, r := range rows {
		if w := lipgloss.Width(r); w != 6 {
			t.Fatalf("row %d width = %d, want 6 (%q)", i, w, r)
		}
	}
}

func TestRenderClipsToHeight(t *testing.T) {
	a := scrollarea.ScrollArea{Content: long, Height: 3, Width: 4}
	a.Scroll.Offset = 2
	rows := lines(a.Render())
	if len(rows) != 3 {
		t.Fatalf("want 3 rows, got %d", len(rows))
	}
	for i, want := range []string{"l2", "l3", "l4"} {
		if got := strings.TrimRight(rows[i], " │█"); got != want {
			t.Fatalf("row %d content = %q, want %q", i, got, want)
		}
	}
}

func TestRenderTruncatesWideLines(t *testing.T) {
	a := scrollarea.ScrollArea{Content: "abcdefgh\nx", Height: 2, Width: 5}
	rows := lines(a.Render())
	if len(rows) != 2 {
		t.Fatalf("a content line wider than the pane must not wrap: got %d rows", len(rows))
	}
	if !strings.HasPrefix(rows[0], "abcd") {
		t.Fatalf("row 0 should be truncated to the 4-cell pane, got %q", rows[0])
	}
}

func TestRenderPadsShortContent(t *testing.T) {
	a := scrollarea.ScrollArea{Content: "a\nb", Height: 4, Width: 5}
	rows := lines(a.Render())
	if len(rows) != 4 {
		t.Fatalf("want 4 rows even with 2 content lines, got %d", len(rows))
	}
	for i, r := range rows {
		if w := lipgloss.Width(r); w != 5 {
			t.Fatalf("row %d width = %d, want 5", i, w)
		}
	}
}

func TestThumbAtTopUsesThumbGlyph(t *testing.T) {
	a := scrollarea.ScrollArea{Content: long, Height: 4, Width: 6}
	rows := lines(a.Render())
	if last := lastCell(rows[0]); last != "█" {
		t.Fatalf("top row scrollbar = %q, want thumb █", last)
	}
	if last := lastCell(rows[3]); last != "│" {
		t.Fatalf("bottom row scrollbar = %q, want track │ when scrolled to top", last)
	}
}

func TestThumbMovesToBottomAtEnd(t *testing.T) {
	a := scrollarea.ScrollArea{Content: long, Height: 4, Width: 6}
	a.Scroll.Offset = a.MaxOffset()
	rows := lines(a.Render())
	if last := lastCell(rows[3]); last != "█" {
		t.Fatalf("bottom row scrollbar at end = %q, want thumb █", last)
	}
	if last := lastCell(rows[0]); last != "│" {
		t.Fatalf("top row scrollbar at end = %q, want track │", last)
	}
}

func TestThumbFillsBarWhenContentFits(t *testing.T) {
	a := scrollarea.ScrollArea{Content: "a\nb", Height: 4, Width: 3}
	for i, r := range lines(a.Render()) {
		if last := lastCell(r); last != "█" {
			t.Fatalf("row %d scrollbar = %q, want full thumb █ when all content fits", i, last)
		}
	}
}

func TestGlyphFallbacks(t *testing.T) {
	a := scrollarea.ScrollArea{Content: long, Height: 4, Width: 2}
	out := a.Render()
	if !strings.Contains(out, "█") || !strings.Contains(out, "│") {
		t.Fatalf("default glyphs █/│ missing: %q", out)
	}
}

func TestCustomGlyphs(t *testing.T) {
	a := scrollarea.ScrollArea{
		Content: long, Height: 4, Width: 2,
		Style: scrollarea.Style{BarChar: ".", ThumbChar: "#"},
	}
	out := a.Render()
	if !strings.Contains(out, "#") || !strings.Contains(out, ".") {
		t.Fatalf("custom glyphs #/. missing: %q", out)
	}
	if strings.ContainsAny(out, "█│") {
		t.Fatalf("default glyphs should be overridden: %q", out)
	}
}

func TestWidthOneIsScrollbarOnly(t *testing.T) {
	a := scrollarea.ScrollArea{Content: long, Height: 3, Width: 1}
	for i, r := range lines(a.Render()) {
		if w := lipgloss.Width(r); w != 1 {
			t.Fatalf("row %d width = %d, want 1 (scrollbar only)", i, w)
		}
	}
}

func TestZeroDimensionsRenderEmpty(t *testing.T) {
	if got := (scrollarea.ScrollArea{Content: long, Height: 0, Width: 6}).Render(); got != "" {
		t.Fatalf("zero height should render empty, got %q", got)
	}
	if got := (scrollarea.ScrollArea{Content: long, Height: 4, Width: 0}).Render(); got != "" {
		t.Fatalf("zero width should render empty, got %q", got)
	}
}

func TestMaxOffsetAndOffset(t *testing.T) {
	a := scrollarea.ScrollArea{Content: long, Height: 4}
	if got := a.MaxOffset(); got != 6 {
		t.Fatalf("MaxOffset = %d, want 6", got)
	}
	a.Scroll.Offset = 3
	if got := a.Offset(); got != 3 {
		t.Fatalf("Offset = %d, want 3", got)
	}
}

func TestUpdateForwardsKeysWhenFocused(t *testing.T) {
	a := scrollarea.ScrollArea{Content: long, Height: 4, Width: 6, Focused: true}
	a2, cmd := a.Update(tea.KeyMsg{Type: tea.KeyDown})
	if a2.Offset() != 1 {
		t.Fatalf("focused down should move offset to 1, got %d", a2.Offset())
	}
	if cmd != nil {
		t.Fatal("scrollarea emits no events; cmd must be nil")
	}
}

func TestUpdateNoopWhenBlurred(t *testing.T) {
	a := scrollarea.ScrollArea{Content: long, Height: 4, Width: 6}
	a2, cmd := a.Update(tea.KeyMsg{Type: tea.KeyDown})
	if a2.Offset() != 0 || cmd != nil {
		t.Fatalf("blurred Update should be a no-op, got offset %d cmd %v", a2.Offset(), cmd)
	}
}

func TestUpdateIgnoresNonKey(t *testing.T) {
	a := scrollarea.ScrollArea{Content: long, Height: 4, Width: 6, Focused: true}
	a.Scroll.Offset = 2
	a2, cmd := a.Update(tea.MouseMsg{})
	if a2.Offset() != 2 || cmd != nil {
		t.Fatalf("non-key msg should be ignored, got offset %d cmd %v", a2.Offset(), cmd)
	}
}

func TestUpdateRespectsCustomBindings(t *testing.T) {
	a := scrollarea.ScrollArea{Content: long, Height: 4, Width: 6, Focused: true}
	a.Scroll.Down = []string{"n"}
	a2, _ := a.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if a2.Offset() != 1 {
		t.Fatalf("custom down key 'n' should move offset to 1, got %d", a2.Offset())
	}
	a3, _ := a.Update(tea.KeyMsg{Type: tea.KeyDown})
	if a3.Offset() != 0 {
		t.Fatalf("default down should be replaced by custom binding, got %d", a3.Offset())
	}
}

func TestViewMatchesRender(t *testing.T) {
	a := scrollarea.ScrollArea{Content: long, Height: 4, Width: 6}
	if a.View() != a.Render() {
		t.Fatal("View must alias Render")
	}
}

func lastCell(row string) string {
	r := []rune(row)
	if len(r) == 0 {
		return ""
	}
	return string(r[len(r)-1])
}
