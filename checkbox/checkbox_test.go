package checkbox_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/chenhunghan/boba/checkbox"
)

func TestUpdateTogglesWhenFocused(t *testing.T) {
	c := checkbox.Checkbox{Label: "Accept", Focused: true}
	c, cmd := c.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !c.Checked {
		t.Fatal("enter should check a focused checkbox")
	}
	if msg, ok := cmd().(checkbox.ToggledMsg); !ok || !msg.Checked {
		t.Fatalf("got %#v, want checkbox.ToggledMsg{Checked: true}", cmd())
	}
}

func TestUpdateIgnoredWhenNotFocused(t *testing.T) {
	c := checkbox.Checkbox{Label: "Accept"}
	c2, cmd := c.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if c2.Checked || cmd != nil {
		t.Fatal("an unfocused checkbox should ignore keys")
	}
}

func TestClickToggles(t *testing.T) {
	c := checkbox.Checkbox{Label: "Accept", Checked: true}
	c, cmd := c.ClickAt(1, 0)
	if c.Checked {
		t.Fatal("click should uncheck")
	}
	if msg, ok := cmd().(checkbox.ToggledMsg); !ok || msg.Checked {
		t.Fatalf("got %#v, want checkbox.ToggledMsg{Checked: false}", cmd())
	}
}

func TestClickOutsideIsNoop(t *testing.T) {
	c := checkbox.Checkbox{Label: "Accept"}
	if _, cmd := c.ClickAt(0, 1); cmd != nil {
		t.Fatal("a click off the row should be a no-op")
	}
}
