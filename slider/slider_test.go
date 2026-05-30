package slider_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/chenhunghan/boba/slider"
)

func TestRenderShape(t *testing.T) {
	s := slider.Slider{Min: 0, Max: 10, Value: 0, Width: 5}
	if got := s.Render(); got != "●────" {
		t.Fatalf("min value: got %q, want %q", got, "●────")
	}
	s.Value = 10
	if got := s.Render(); got != "────●" {
		t.Fatalf("max value: got %q, want %q", got, "────●")
	}
	s.Value = 5
	if got := s.Render(); got != "──●──" {
		t.Fatalf("mid value: got %q, want %q", got, "──●──")
	}
}

func TestRenderWidthMatchesWidth(t *testing.T) {
	s := slider.Slider{Min: 0, Max: 1, Value: 0.3, Width: 12}
	if got := len([]rune(s.Render())); got != 12 {
		t.Fatalf("rune width = %d, want 12", got)
	}
}

func TestRenderZeroWidthEmpty(t *testing.T) {
	if got := (slider.Slider{Width: 0}).Render(); got != "" {
		t.Fatalf("zero width should render empty, got %q", got)
	}
}

func TestApplyKeyAdjustsAndClamps(t *testing.T) {
	s := slider.Slider{Min: 0, Max: 10, Value: 5, Step: 2, Width: 10}
	s, changed := s.ApplyKey("right")
	if !changed || s.Value != 7 {
		t.Fatalf("right: got value %v changed=%v, want 7 true", s.Value, changed)
	}
	s.Value = 9
	s, _ = s.ApplyKey("right")
	if s.Value != 10 {
		t.Fatalf("right should clamp to Max, got %v", s.Value)
	}
	s, changed = s.ApplyKey("right")
	if changed || s.Value != 10 {
		t.Fatalf("right at Max should not change, got %v changed=%v", s.Value, changed)
	}
	s.Value = 1
	s, _ = s.ApplyKey("left")
	if s.Value != 0 {
		t.Fatalf("left should clamp to Min, got %v", s.Value)
	}
}

func TestApplyKeyUnboundIsNoop(t *testing.T) {
	s := slider.Slider{Min: 0, Max: 10, Value: 5, Step: 1, Width: 10}
	if got, changed := s.ApplyKey("x"); changed || got.Value != 5 {
		t.Fatalf("unbound key should be a no-op, got %v changed=%v", got.Value, changed)
	}
}

func TestUpdateOnlyWhenFocused(t *testing.T) {
	s := slider.Slider{Min: 0, Max: 10, Value: 5, Step: 1, Width: 10}
	s2, cmd := s.Update(tea.KeyMsg{Type: tea.KeyRight})
	if s2.Value != 5 || cmd != nil {
		t.Fatalf("unfocused slider must ignore keys, got %v cmd=%v", s2.Value, cmd)
	}

	s.Focused = true
	s3, cmd := s.Update(tea.KeyMsg{Type: tea.KeyRight})
	if s3.Value != 6 {
		t.Fatalf("focused right should step, got %v", s3.Value)
	}
	if msg, ok := cmd().(slider.ChangedMsg); !ok || msg.Value != 6 {
		t.Fatalf("got %#v, want slider.ChangedMsg{Value: 6}", cmd())
	}
}

func TestUpdateIgnoresNonKey(t *testing.T) {
	s := slider.Slider{Min: 0, Max: 10, Value: 5, Step: 1, Width: 10, Focused: true}
	if _, cmd := s.Update(tea.MouseMsg{}); cmd != nil {
		t.Fatal("non-key message should be ignored")
	}
}

func TestValueAtAcrossWidth(t *testing.T) {
	s := slider.Slider{Min: 0, Max: 100, Step: 1, Width: 11}
	cases := map[int]float64{0: 0, 5: 50, 10: 100, -3: 0, 99: 100}
	for x, want := range cases {
		if got := s.ValueAt(x); got != want {
			t.Fatalf("ValueAt(%d) = %v, want %v", x, got, want)
		}
	}
}

func TestValueAtSnapsToStep(t *testing.T) {
	s := slider.Slider{Min: 0, Max: 10, Step: 5, Width: 11}
	if got := s.ValueAt(3); got != 5 {
		t.Fatalf("ValueAt(3) with Step 5 = %v, want 5", got)
	}
}

func TestClickSetsValue(t *testing.T) {
	s := slider.Slider{Min: 0, Max: 100, Step: 1, Width: 11}
	s, cmd := s.ClickAt(10, 0)
	if s.Value != 100 {
		t.Fatalf("click at right end should set Max, got %v", s.Value)
	}
	if msg, ok := cmd().(slider.ChangedMsg); !ok || msg.Value != 100 {
		t.Fatalf("got %#v, want slider.ChangedMsg{Value: 100}", cmd())
	}
}

func TestClickOffRowIsNoop(t *testing.T) {
	s := slider.Slider{Min: 0, Max: 100, Step: 1, Width: 11}
	if _, cmd := s.ClickAt(3, 1); cmd != nil {
		t.Fatal("click off the row should be a no-op")
	}
	if _, cmd := s.ClickAt(99, 0); cmd != nil {
		t.Fatal("click past the track should be a no-op")
	}
}

func TestClickSameValueNoEvent(t *testing.T) {
	s := slider.Slider{Min: 0, Max: 100, Value: 100, Step: 1, Width: 11}
	if _, cmd := s.ClickAt(10, 0); cmd != nil {
		t.Fatal("clicking the current value should not emit an event")
	}
}
