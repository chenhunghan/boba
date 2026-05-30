package menu_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/chenhunghan/boba/menu"
)

func openGroup() menu.Group[string] {
	return menu.Group[string]{
		Items: []menu.Item[string]{{ID: "a", Label: "A"}, {ID: "b", Label: "B"}},
		Open:  true,
		Hover: 1,
	}
}

func TestUpdateConfirmEmitsChosen(t *testing.T) {
	g, cmd := openGroup().Update(tea.KeyMsg{Type: tea.KeyEnter})
	if g.Open {
		t.Error("menu should close after confirm")
	}
	if cmd == nil {
		t.Fatal("confirm should return a cmd")
	}
	chosen, ok := cmd().(menu.ChosenMsg[string])
	if !ok || chosen.ID != "b" {
		t.Fatalf("got %#v, want menu.ChosenMsg[string]{ID: \"b\"}", cmd())
	}
}

func TestUpdateCancelEmitsCancelled(t *testing.T) {
	g, cmd := openGroup().Update(tea.KeyMsg{Type: tea.KeyEsc})
	if g.Open {
		t.Error("menu should close after cancel")
	}
	if cmd == nil {
		t.Fatal("cancel should return a cmd")
	}
	if _, ok := cmd().(menu.CancelledMsg); !ok {
		t.Fatalf("got %#v, want menu.CancelledMsg", cmd())
	}
}

func TestUpdateClosedIsNoop(t *testing.T) {
	var g menu.Group[string]
	if _, cmd := g.Update(tea.KeyMsg{Type: tea.KeyEnter}); cmd != nil {
		t.Error("closed menu should emit no cmd")
	}
}
