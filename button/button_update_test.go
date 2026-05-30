package button_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/chenhunghan/boba/button"
)

func stack() button.Stack {
	return button.Stack{
		Buttons:    []button.Button{{Text: "A"}, {Text: "B", Trailing: "×"}},
		Width:      8,
		ItemHeight: 1,
		Hover:      -1,
		Active:     true,
	}
}

func TestUpdateMovesAndActivates(t *testing.T) {
	s, cmd := stack().Update(tea.KeyMsg{Type: tea.KeyDown})
	if s.Selected != 1 {
		t.Fatalf("down: Selected=%d, want 1", s.Selected)
	}
	if cmd != nil {
		t.Fatal("a move should not emit a cmd")
	}

	_, cmd = s.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("confirm should emit a cmd")
	}
	if msg, ok := cmd().(button.ActivatedMsg); !ok || msg.Index != 1 {
		t.Fatalf("got %#v, want button.ActivatedMsg{Index: 1}", cmd())
	}
}

func TestUpdateInactiveIsNoop(t *testing.T) {
	s := stack()
	s.Active = false
	s2, cmd := s.Update(tea.KeyMsg{Type: tea.KeyDown})
	if cmd != nil || s2.Selected != s.Selected {
		t.Fatal("an inactive stack should ignore keys")
	}
}

func TestClickBodyAndTrailing(t *testing.T) {
	s := stack()

	s2, cmd := s.ClickAt(0, 1)
	if s2.Selected != 1 {
		t.Fatalf("click body: Selected=%d, want 1", s2.Selected)
	}
	if msg, ok := cmd().(button.ActivatedMsg); !ok || msg.Index != 1 {
		t.Fatalf("got %#v, want button.ActivatedMsg{Index: 1}", cmd())
	}

	_, cmd = s.ClickAt(7, 1)
	if msg, ok := cmd().(button.TrailingMsg); !ok || msg.Index != 1 {
		t.Fatalf("got %#v, want button.TrailingMsg{Index: 1}", cmd())
	}
}

func TestHoverAtSetsAndClears(t *testing.T) {
	s := stack().HoverAt(0, 0)
	if s.Hover != 0 {
		t.Fatalf("Hover=%d, want 0", s.Hover)
	}
	if s = s.HoverAt(0, 99); s.Hover != -1 {
		t.Fatalf("Hover=%d, want -1", s.Hover)
	}
}
