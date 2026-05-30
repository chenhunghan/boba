package scroll_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/chenhunghan/boba/scroll"
)

const content = "l0\nl1\nl2\nl3\nl4"

func TestRenderClipsToHeight(t *testing.T) {
	s := scroll.Scroll{Height: 2, Offset: 1}
	if got, want := s.Render(content), "l1\nl2"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestRenderClampsOffsetWithoutMutating(t *testing.T) {
	s := scroll.Scroll{Height: 2, Offset: 99}
	if got, want := s.Render(content), "l3\nl4"; got != want {
		t.Fatalf("over-range offset should clamp to the last window: got %q, want %q", got, want)
	}
	if s.Offset != 99 {
		t.Fatalf("Render must not mutate Offset, got %d", s.Offset)
	}
}

func TestRenderZeroHeightIsEmpty(t *testing.T) {
	s := scroll.Scroll{Height: 0}
	if got := s.Render(content); got != "" {
		t.Fatalf("zero height should render empty, got %q", got)
	}
}

func TestMaxOffset(t *testing.T) {
	if got, want := (scroll.Scroll{Height: 2}).MaxOffset(content), 3; got != want {
		t.Fatalf("got %d, want %d", got, want)
	}
	if got := (scroll.Scroll{Height: 10}).MaxOffset(content); got != 0 {
		t.Fatalf("height past content should floor MaxOffset at 0, got %d", got)
	}
}

func TestApplyKeyMovesAndClamps(t *testing.T) {
	s := scroll.Scroll{Height: 2}
	if got := s.ApplyKey("down", content).Offset; got != 1 {
		t.Fatalf("down: got %d, want 1", got)
	}
	if got := s.ApplyKey("up", content).Offset; got != 0 {
		t.Fatalf("up at top should clamp to 0, got %d", got)
	}
	if got := s.ApplyKey("end", content).Offset; got != 3 {
		t.Fatalf("end: got %d, want MaxOffset 3", got)
	}
	atEnd := scroll.Scroll{Height: 2, Offset: 3}
	if got := atEnd.ApplyKey("pgdown", content).Offset; got != 3 {
		t.Fatalf("pgdown at end should clamp to MaxOffset, got %d", got)
	}
	if got := atEnd.ApplyKey("home", content).Offset; got != 0 {
		t.Fatalf("home: got %d, want 0", got)
	}
}

func TestApplyKeyUnknownIsNoop(t *testing.T) {
	s := scroll.Scroll{Height: 2, Offset: 1}
	if got := s.ApplyKey("x", content).Offset; got != 1 {
		t.Fatalf("unknown key should not move offset, got %d", got)
	}
}

func TestUpdateRespectsFocus(t *testing.T) {
	blurred := scroll.Scroll{Height: 2}
	if s2, cmd := blurred.Update(tea.KeyMsg{Type: tea.KeyDown}, content); s2.Offset != 0 || cmd != nil {
		t.Fatalf("unfocused Update should be a no-op, got offset %d cmd %v", s2.Offset, cmd)
	}
	focused := scroll.Scroll{Height: 2, Focused: true}
	s2, cmd := focused.Update(tea.KeyMsg{Type: tea.KeyDown}, content)
	if s2.Offset != 1 {
		t.Fatalf("focused down: got %d, want 1", s2.Offset)
	}
	if cmd != nil {
		t.Fatal("scroll emits no events; cmd must be nil")
	}
}

func TestUpdateIgnoresNonKey(t *testing.T) {
	s := scroll.Scroll{Height: 2, Offset: 1, Focused: true}
	if s2, cmd := s.Update(tea.MouseMsg{}, content); s2.Offset != 1 || cmd != nil {
		t.Fatalf("non-key msg should be ignored, got offset %d cmd %v", s2.Offset, cmd)
	}
}

func TestCustomBindings(t *testing.T) {
	s := scroll.Scroll{Height: 2, Down: []string{"n"}}
	if got := s.ApplyKey("n", content).Offset; got != 1 {
		t.Fatalf("custom down key: got %d, want 1", got)
	}
	if got := s.ApplyKey("down", content).Offset; got != 0 {
		t.Fatalf("custom binding should replace the default, got %d", got)
	}
}

func TestRenderLineCount(t *testing.T) {
	s := scroll.Scroll{Height: 3, Offset: 0}
	if got := strings.Count(s.Render(content), "\n") + 1; got != 3 {
		t.Fatalf("expected 3 visible lines, got %d", got)
	}
}
