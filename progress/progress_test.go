package progress_test

import (
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/chenhunghan/boba/progress"
)

func TestRenderWidth(t *testing.T) {
	p := progress.Progress{Value: 3, Max: 10, Width: 20}
	if w := lipgloss.Width(p.Render()); w != 20 {
		t.Fatalf("width = %d, want 20", w)
	}
}

func TestHalfFilled(t *testing.T) {
	p := progress.Progress{Value: 5, Max: 10, Width: 10}
	if got := p.Render(); got != "█████     " {
		t.Fatalf("Render = %q, want %q", got, "█████     ")
	}
}

func TestEmptyAtZero(t *testing.T) {
	p := progress.Progress{Value: 0, Max: 10, Width: 4}
	if got := p.Render(); got != "    " {
		t.Fatalf("Render = %q, want four spaces", got)
	}
}

func TestFullAtMax(t *testing.T) {
	p := progress.Progress{Value: 10, Max: 10, Width: 4}
	if got := p.Render(); got != "████" {
		t.Fatalf("Render = %q, want %q", got, "████")
	}
}

func TestClampsOverMax(t *testing.T) {
	p := progress.Progress{Value: 99, Max: 10, Width: 4}
	if got := p.Render(); got != "████" {
		t.Fatalf("over-max Render = %q, want full bar", got)
	}
}

func TestClampsNegative(t *testing.T) {
	p := progress.Progress{Value: -5, Max: 10, Width: 4}
	if got := p.Render(); got != "    " {
		t.Fatalf("negative Render = %q, want empty bar", got)
	}
}

func TestZeroMaxEmpty(t *testing.T) {
	p := progress.Progress{Value: 5, Max: 0, Width: 4}
	if got := p.Render(); got != "    " {
		t.Fatalf("zero-Max Render = %q, want empty bar", got)
	}
}

func TestZeroWidthEmpty(t *testing.T) {
	if got := (progress.Progress{Value: 5, Max: 10}).Render(); got != "" {
		t.Fatalf("zero-width Render = %q, want empty", got)
	}
}

func TestCustomGlyphs(t *testing.T) {
	p := progress.Progress{
		Value: 5, Max: 10, Width: 4,
		Style: progress.Style{FilledChar: "#", EmptyChar: "-"},
	}
	if got := p.Render(); got != "##--" {
		t.Fatalf("Render = %q, want %q", got, "##--")
	}
}

func TestViewMatchesRender(t *testing.T) {
	p := progress.Progress{Value: 7, Max: 10, Width: 8}
	if p.View() != p.Render() {
		t.Fatal("View != Render")
	}
}
