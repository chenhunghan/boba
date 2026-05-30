package toggle_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/chenhunghan/boba/toggle"
)

func TestUpdatePressesWhenFocused(t *testing.T) {
	tg := toggle.Toggle{Label: "Auto", Focused: true}
	tg, cmd := tg.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !tg.Pressed {
		t.Fatal("enter should press a focused toggle")
	}
	if msg, ok := cmd().(toggle.ToggledMsg); !ok || !msg.Pressed {
		t.Fatalf("got %#v, want toggle.ToggledMsg{Pressed: true}", cmd())
	}
}

func TestUpdateIgnoredWhenNotFocused(t *testing.T) {
	tg := toggle.Toggle{Label: "Auto"}
	tg2, cmd := tg.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if tg2.Pressed || cmd != nil {
		t.Fatal("an unfocused toggle should ignore keys")
	}
}

func TestUpdateIgnoresNonKeyMsg(t *testing.T) {
	tg := toggle.Toggle{Label: "Auto", Focused: true}
	tg2, cmd := tg.Update(struct{}{})
	if tg2.Pressed || cmd != nil {
		t.Fatal("a non-key message should be a no-op")
	}
}

func TestCustomToggleKey(t *testing.T) {
	tg := toggle.Toggle{Label: "Auto", Focused: true, Toggle: []string{"t"}}
	if _, cmd := tg.Update(tea.KeyMsg{Type: tea.KeyEnter}); cmd != nil {
		t.Fatal("enter should not toggle when only 't' is bound")
	}
	tg, cmd := tg.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	if !tg.Pressed || cmd == nil {
		t.Fatal("the bound 't' key should press the toggle")
	}
}

func TestClickToggles(t *testing.T) {
	tg := toggle.Toggle{Label: "Auto", Pressed: true}
	tg, cmd := tg.ClickAt(1, 0)
	if tg.Pressed {
		t.Fatal("click should release a pressed toggle")
	}
	if msg, ok := cmd().(toggle.ToggledMsg); !ok || msg.Pressed {
		t.Fatalf("got %#v, want toggle.ToggledMsg{Pressed: false}", cmd())
	}
}

func TestClickOutsideIsNoop(t *testing.T) {
	tg := toggle.Toggle{Label: "Auto"}
	if _, cmd := tg.ClickAt(0, 1); cmd != nil {
		t.Fatal("a click off the row should be a no-op")
	}
	if _, cmd := tg.ClickAt(99, 0); cmd != nil {
		t.Fatal("a click past the label width should be a no-op")
	}
}

func TestWidthMatchesLabel(t *testing.T) {
	tg := toggle.Toggle{Label: "Auto"}
	if got := tg.Width(); got != 4 {
		t.Fatalf("Width() = %d, want 4", got)
	}
}
