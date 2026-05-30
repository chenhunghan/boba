package meter_test

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/chenhunghan/boba/meter"
)

func TestRenderWidth(t *testing.T) {
	m := meter.Meter{Value: 3, Max: 10, Width: 8}
	if w := lipgloss.Width(m.Render()); w != 8 {
		t.Fatalf("width = %d, want 8", w)
	}
}

func TestFillFractionRounds(t *testing.T) {
	// 3/10 of 10 cells = 3 filled, 7 empty.
	m := meter.Meter{Value: 3, Max: 10, Width: 10, Style: meter.Style{FillChar: "#", EmptyChar: "."}}
	if got := m.Render(); got != "###......." {
		t.Fatalf("Render = %q, want %q", got, "###.......")
	}
}

func TestFullAndEmpty(t *testing.T) {
	full := meter.Meter{Value: 10, Max: 10, Width: 4, Style: meter.Style{FillChar: "#", EmptyChar: "."}}
	if got := full.Render(); got != "####" {
		t.Fatalf("full Render = %q, want %q", got, "####")
	}
	empty := meter.Meter{Value: 0, Max: 10, Width: 4, Style: meter.Style{FillChar: "#", EmptyChar: "."}}
	if got := empty.Render(); got != "...." {
		t.Fatalf("empty Render = %q, want %q", got, "....")
	}
}

func TestClampsOutOfRange(t *testing.T) {
	over := meter.Meter{Value: 99, Max: 10, Width: 4, Style: meter.Style{FillChar: "#"}}
	if got := over.Render(); got != "####" {
		t.Fatalf("over-range Render = %q, want %q", got, "####")
	}
	under := meter.Meter{Value: -5, Max: 10, Width: 4, Style: meter.Style{FillChar: "#", EmptyChar: "."}}
	if got := under.Render(); got != "...." {
		t.Fatalf("under-range Render = %q, want %q", got, "....")
	}
}

func TestMinOffsetsFill(t *testing.T) {
	// (75-50)/(100-50) = 0.5 of 8 cells = 4 filled.
	m := meter.Meter{Value: 75, Min: 50, Max: 100, Width: 8, Style: meter.Style{FillChar: "#", EmptyChar: "."}}
	if got := m.Render(); got != "####...." {
		t.Fatalf("Render = %q, want %q", got, "####....")
	}
}

// zoneStyle tags each zone's fill via a transform that survives any color
// profile (uppercasing the glyph), so the selected zone is observable even
// when the test environment strips ANSI color.
func zoneStyle(tag string) lipgloss.Style {
	return lipgloss.NewStyle().Transform(func(s string) string {
		return strings.ReplaceAll(s, "#", tag)
	})
}

func TestThresholdSelectsFillStyle(t *testing.T) {
	st := meter.Style{
		Normal:    zoneStyle("N"),
		LowFill:   zoneStyle("L"),
		HighFill:  zoneStyle("H"),
		FillChar:  "#",
		EmptyChar: ".",
	}
	cases := []struct {
		name string
		val  float64
		tag  string
	}{
		{"below low", 10, "L"},
		{"normal band", 50, "N"},
		{"at or above high", 90, "H"},
	}
	for _, c := range cases {
		m := meter.Meter{Value: c.val, Max: 100, Width: 10, Low: 25, High: 90, Style: st}
		got := m.Render()
		if !strings.Contains(got, c.tag) {
			t.Fatalf("%s: Render = %q, want fill tagged %q (zone selected)", c.name, got, c.tag)
		}
		for _, other := range []string{"N", "L", "H"} {
			if other != c.tag && strings.Contains(got, other) {
				t.Fatalf("%s: Render = %q, leaked %q from the wrong zone", c.name, got, other)
			}
		}
	}
}

func TestUnsetThresholdsUseNormal(t *testing.T) {
	st := meter.Style{
		Normal:   zoneStyle("N"),
		LowFill:  zoneStyle("L"),
		HighFill: zoneStyle("H"),
		FillChar: "#",
	}
	// No Low/High set: a mid-range value uses Normal, never LowFill/HighFill.
	m := meter.Meter{Value: 20, Max: 100, Width: 10, Style: st}
	got := m.Render()
	if !strings.Contains(got, "N") {
		t.Fatalf("Render = %q, want Normal-styled fill", got)
	}
	if strings.ContainsAny(got, "LH") {
		t.Fatalf("Render = %q, must not use Low/HighFill when thresholds are unset", got)
	}
}

func TestZeroWidthEmpty(t *testing.T) {
	if got := (meter.Meter{Value: 5, Max: 10}).Render(); got != "" {
		t.Fatalf("zero-width Render = %q, want empty", got)
	}
}

func TestZeroSpanEmpty(t *testing.T) {
	m := meter.Meter{Value: 5, Min: 10, Max: 10, Width: 4, Style: meter.Style{FillChar: "#", EmptyChar: "."}}
	if got := m.Render(); got != "...." {
		t.Fatalf("zero-span Render = %q, want %q", got, "....")
	}
}

func TestViewMatchesRender(t *testing.T) {
	m := meter.Meter{Value: 4, Max: 10, Width: 6, Style: meter.Style{FillChar: "#", EmptyChar: "."}}
	if m.View() != m.Render() {
		t.Fatalf("View %q != Render %q", m.View(), m.Render())
	}
}
