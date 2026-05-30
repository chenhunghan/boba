package swtch_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	swtch "github.com/chenhunghan/boba/switch"
)

func TestRenderDefaultGlyphs(t *testing.T) {
	if got := (swtch.Switch{Label: "Wi-Fi"}).Render(); got != "[off] Wi-Fi" {
		t.Fatalf("off render = %q, want %q", got, "[off] Wi-Fi")
	}
	if got := (swtch.Switch{On: true, Label: "Wi-Fi"}).Render(); got != "[on] Wi-Fi" {
		t.Fatalf("on render = %q, want %q", got, "[on] Wi-Fi")
	}
}

func TestRenderNoLabelIsGlyphOnly(t *testing.T) {
	s := swtch.Switch{On: true}
	if got := s.Render(); got != "[on]" {
		t.Fatalf("render = %q, want %q", got, "[on]")
	}
	if got, want := s.Width(), len("[on]"); got != want {
		t.Fatalf("width = %d, want %d", got, want)
	}
}

func TestGlyphFallbackOverride(t *testing.T) {
	s := swtch.Switch{Style: swtch.Style{OnGlyph: "◉", OffGlyph: "◯"}}
	if got := s.Render(); got != "◯" {
		t.Fatalf("off glyph = %q, want %q", got, "◯")
	}
	s.On = true
	if got := s.Render(); got != "◉" {
		t.Fatalf("on glyph = %q, want %q", got, "◉")
	}
}

func TestUpdateTogglesWhenFocused(t *testing.T) {
	s := swtch.Switch{Label: "Wi-Fi", Focused: true}
	s, cmd := s.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !s.On {
		t.Fatal("enter should turn a focused switch on")
	}
	if msg, ok := cmd().(swtch.ToggledMsg); !ok || !msg.On {
		t.Fatalf("got %#v, want swtch.ToggledMsg{On: true}", cmd())
	}
}

func TestUpdateSpaceToggles(t *testing.T) {
	s := swtch.Switch{Focused: true}
	s, cmd := s.Update(tea.KeyMsg{Type: tea.KeySpace})
	if !s.On || cmd == nil {
		t.Fatal("space should toggle a focused switch")
	}
}

func TestUpdateCustomKeyOnly(t *testing.T) {
	s := swtch.Switch{Focused: true, Toggle: []string{"x"}}
	if s2, cmd := s.Update(tea.KeyMsg{Type: tea.KeyEnter}); s2.On || cmd != nil {
		t.Fatal("enter should be ignored when Toggle is custom")
	}
	s3, cmd := s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if !s3.On || cmd == nil {
		t.Fatal("custom key x should toggle")
	}
}

func TestUpdateIgnoredWhenNotFocused(t *testing.T) {
	s := swtch.Switch{Label: "Wi-Fi"}
	if s2, cmd := s.Update(tea.KeyMsg{Type: tea.KeyEnter}); s2.On || cmd != nil {
		t.Fatal("an unfocused switch should ignore keys")
	}
}

func TestUpdateIgnoresNonKeyMsg(t *testing.T) {
	s := swtch.Switch{Focused: true}
	if s2, cmd := s.Update(tea.MouseMsg{}); s2.On || cmd != nil {
		t.Fatal("non-key messages should be ignored")
	}
}

func TestClickTogglesOnRow(t *testing.T) {
	s := swtch.Switch{On: true, Label: "Wi-Fi"}
	s, cmd := s.ClickAt(1, 0)
	if s.On {
		t.Fatal("click should turn the switch off")
	}
	if msg, ok := cmd().(swtch.ToggledMsg); !ok || msg.On {
		t.Fatalf("got %#v, want swtch.ToggledMsg{On: false}", cmd())
	}
}

func TestClickOffRowIsNoop(t *testing.T) {
	s := swtch.Switch{Label: "Wi-Fi"}
	if _, cmd := s.ClickAt(0, 1); cmd != nil {
		t.Fatal("a click off the row should be a no-op")
	}
	if _, cmd := s.ClickAt(s.Width(), 0); cmd != nil {
		t.Fatal("a click past the right edge should be a no-op")
	}
	if _, cmd := s.ClickAt(-1, 0); cmd != nil {
		t.Fatal("a click left of the row should be a no-op")
	}
}
