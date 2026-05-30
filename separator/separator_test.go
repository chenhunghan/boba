package separator_test

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/chenhunghan/boba/separator"
)

func TestHorizontalWidth(t *testing.T) {
	s := separator.Separator{Length: 6}
	if w := lipgloss.Width(s.Render()); w != 6 {
		t.Fatalf("width = %d, want 6", w)
	}
}

func TestVerticalRows(t *testing.T) {
	s := separator.Separator{Orientation: separator.Vertical, Length: 3}
	if n := strings.Count(s.Render(), "\n") + 1; n != 3 {
		t.Fatalf("rows = %d, want 3", n)
	}
}

func TestZeroLengthEmpty(t *testing.T) {
	if got := (separator.Separator{}).Render(); got != "" {
		t.Fatalf("zero-length Render = %q, want empty", got)
	}
}

func TestCustomChar(t *testing.T) {
	s := separator.Separator{Length: 4, Char: "="}
	if got := s.Render(); got != "====" {
		t.Fatalf("Render = %q, want \"====\"", got)
	}
}
