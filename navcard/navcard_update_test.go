package navcard_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/chenhunghan/boba/button"
	"github.com/chenhunghan/boba/navcard"
)

func navStack() navcard.Stack {
	return navcard.Stack{
		Cards: []navcard.Card{
			{Title: "First", Buttons: []button.Button{{Text: "Go"}}},
			{Title: "Second"},
		},
		Width:       20,
		Hover:       -1,
		HoverButton: -1,
		Active:      true,
	}
}

func TestUpdateMovesAndActivatesCard(t *testing.T) {
	s, cmd := navStack().Update(tea.KeyMsg{Type: tea.KeyDown})
	if s.Selected != 1 {
		t.Fatalf("down: Selected=%d, want 1", s.Selected)
	}
	if cmd != nil {
		t.Fatal("a move should not emit a cmd")
	}

	_, cmd = navStack().Update(tea.KeyMsg{Type: tea.KeyEnter})
	if msg, ok := cmd().(navcard.CardActivatedMsg); !ok || msg.Card != 0 {
		t.Fatalf("got %#v, want navcard.CardActivatedMsg{Card: 0}", cmd())
	}
}

func TestUpdateInactiveIsNoop(t *testing.T) {
	s := navStack()
	s.Active = false
	if s2, cmd := s.Update(tea.KeyMsg{Type: tea.KeyDown}); cmd != nil || s2.Selected != s.Selected {
		t.Fatal("an inactive stack should ignore keys")
	}
}

func TestClickInlineButtonAndCardBody(t *testing.T) {
	s := navStack()

	_, cmd := s.ClickAt(4, 2)
	if msg, ok := cmd().(navcard.ButtonActivatedMsg); !ok || msg.Card != 0 || msg.Button != 0 {
		t.Fatalf("got %#v, want navcard.ButtonActivatedMsg{Card: 0, Button: 0}", cmd())
	}

	s2, cmd := s.ClickAt(1, 5)
	if s2.Selected != 1 {
		t.Fatalf("card-body click: Selected=%d, want 1", s2.Selected)
	}
	if msg, ok := cmd().(navcard.CardActivatedMsg); !ok || msg.Card != 1 {
		t.Fatalf("got %#v, want navcard.CardActivatedMsg{Card: 1}", cmd())
	}
}

func TestHoverAtSetsAndClears(t *testing.T) {
	s := navStack().HoverAt(4, 2)
	if s.Hover != 0 || s.HoverButton != 0 {
		t.Fatalf("Hover=%d HoverButton=%d, want 0 0", s.Hover, s.HoverButton)
	}
	if s = s.HoverAt(0, 99); s.Hover != -1 || s.HoverButton != -1 {
		t.Fatalf("Hover=%d HoverButton=%d, want -1 -1", s.Hover, s.HoverButton)
	}
}
