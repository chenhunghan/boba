package dialog_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/chenhunghan/boba/dialog"
)

func sample() dialog.Dialog {
	return dialog.Dialog{
		Title:   "Quit?",
		Body:    "Discard changes?",
		Buttons: []string{"OK", "Cancel"},
		Open:    true,
	}
}

func TestRenderShape(t *testing.T) {
	d := sample()
	out := d.Render()

	lines := strings.Split(out, "\n")
	if len(lines) != d.Height() {
		t.Fatalf("got %d lines, want %d", len(lines), d.Height())
	}
	for i, ln := range lines {
		if w := lipgloss.Width(ln); w != d.Width() {
			t.Fatalf("line %d width = %d, want %d", i, w, d.Width())
		}
	}
}

func TestWidthFitsWidestContent(t *testing.T) {
	d := dialog.Dialog{Title: "hi", Body: strings.Repeat("x", 40), Open: true}
	// 2 borders + 2 padding + body.
	if got, want := d.Width(), 40+4; got != want {
		t.Fatalf("Width() = %d, want %d", got, want)
	}
}

func TestNextClampsAtEnd(t *testing.T) {
	d := sample()
	d.Selected = 1
	d, _, _ = d.ApplyKey("right")
	if d.Selected != 1 {
		t.Fatalf("Selected = %d, want 1 (clamped)", d.Selected)
	}
}

func TestPrevClampsAtStart(t *testing.T) {
	d := sample()
	d, _, _ = d.ApplyKey("left")
	if d.Selected != 0 {
		t.Fatalf("Selected = %d, want 0 (clamped)", d.Selected)
	}
}

func TestNextMoves(t *testing.T) {
	d := sample()
	d, chosen, cancelled := d.ApplyKey("right")
	if d.Selected != 1 || chosen || cancelled {
		t.Fatalf("got Selected=%d chosen=%v cancelled=%v, want 1 false false", d.Selected, chosen, cancelled)
	}
}

func TestUpdateEnterChooses(t *testing.T) {
	d := sample()
	d.Selected = 1
	_, cmd := d.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter should emit a cmd")
	}
	msg, ok := cmd().(dialog.ChosenMsg)
	if !ok || msg.Index != 1 {
		t.Fatalf("got %#v, want dialog.ChosenMsg{Index: 1}", cmd())
	}
}

func TestUpdateEscCancels(t *testing.T) {
	d := sample()
	_, cmd := d.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("esc should emit a cmd")
	}
	if _, ok := cmd().(dialog.CancelledMsg); !ok {
		t.Fatalf("got %#v, want dialog.CancelledMsg", cmd())
	}
}

func TestUpdateClosedIsNoop(t *testing.T) {
	d := sample()
	d.Open = false
	d2, cmd := d.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil || d2.Selected != d.Selected {
		t.Fatal("a closed dialog should ignore keys")
	}
}

func TestUpdateIgnoresNonKey(t *testing.T) {
	d := sample()
	if _, cmd := d.Update(tea.MouseMsg{}); cmd != nil {
		t.Fatal("non-key messages should be ignored")
	}
}

func TestEnterWithoutButtonsDoesNotChoose(t *testing.T) {
	d := dialog.Dialog{Title: "x", Open: true}
	if _, cmd := d.Update(tea.KeyMsg{Type: tea.KeyEnter}); cmd != nil {
		t.Fatal("enter with no buttons should not emit ChosenMsg")
	}
}

func TestOverCentersAndPreservesScreen(t *testing.T) {
	d := sample()
	w, h := 40, 12
	bg := strings.TrimRight(strings.Repeat(strings.Repeat(".", w)+"\n", h), "\n")

	out := d.Over(bg, w, h)

	if got := strings.Count(out, "\n") + 1; got != h {
		t.Fatalf("Over line count = %d, want %d", got, h)
	}
	for i, ln := range strings.Split(out, "\n") {
		if got := lipgloss.Width(ln); got != w {
			t.Fatalf("Over line %d width = %d, want %d", i, got, w)
		}
	}
}

func TestOverClosedReturnsBg(t *testing.T) {
	d := sample()
	d.Open = false
	bg := "....."
	if out := d.Over(bg, 20, 5); out != bg {
		t.Fatalf("closed Over = %q, want %q", out, bg)
	}
}
