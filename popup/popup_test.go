package popup_test

import (
	"strings"
	"testing"

	"github.com/chenhunghan/boba/popup"
)

func TestPlaceBelowFits(t *testing.T) {
	// Anchor at (2,2) size 4x1; content 4x3; roomy 40x20 screen.
	x, y := popup.Place(2, 2, 4, 1, 4, 3, 40, 20, popup.Below)
	if x != 2 || y != 3 {
		t.Fatalf("Below: got (%d,%d), want (2,3)", x, y)
	}
}

func TestPlaceBelowFlipsToAbove(t *testing.T) {
	// Anchor near the bottom: below would overflow, so flip above.
	x, y := popup.Place(2, 18, 4, 1, 4, 3, 40, 20, popup.Below)
	if x != 2 || y != 15 {
		t.Fatalf("Below->Above: got (%d,%d), want (2,15)", x, y)
	}
}

func TestPlaceAboveFlipsToBelow(t *testing.T) {
	// Anchor at the top: above would overflow (negative), so flip below.
	x, y := popup.Place(2, 0, 4, 1, 4, 3, 40, 20, popup.Above)
	if x != 2 || y != 1 {
		t.Fatalf("Above->Below: got (%d,%d), want (2,1)", x, y)
	}
}

func TestPlaceRightFits(t *testing.T) {
	x, y := popup.Place(2, 2, 4, 1, 6, 3, 40, 20, popup.Right)
	if x != 6 || y != 2 {
		t.Fatalf("Right: got (%d,%d), want (6,2)", x, y)
	}
}

func TestPlaceRightFlipsToLeft(t *testing.T) {
	// Anchor flush against the right edge: right overflows, flip left.
	x, y := popup.Place(36, 2, 4, 1, 6, 3, 40, 20, popup.Right)
	if x != 30 || y != 2 {
		t.Fatalf("Right->Left: got (%d,%d), want (30,2)", x, y)
	}
}

func TestPlaceLeftFlipsToRight(t *testing.T) {
	// Anchor at the left edge: left overflows (negative), flip right.
	x, y := popup.Place(0, 2, 4, 1, 6, 3, 40, 20, popup.Left)
	if x != 4 || y != 2 {
		t.Fatalf("Left->Right: got (%d,%d), want (4,2)", x, y)
	}
}

func TestPlaceClampsOntoScreen(t *testing.T) {
	// Below an anchor at the right edge: x would be 38 but a 6-wide box
	// must clamp to 34 so it stays within [0,40).
	x, y := popup.Place(38, 2, 4, 1, 6, 2, 40, 20, popup.Below)
	if x != 34 {
		t.Fatalf("clamp x: got %d, want 34", x)
	}
	if y != 3 {
		t.Fatalf("clamp y: got %d, want 3", y)
	}
}

func TestPlaceOversizedPinsToOrigin(t *testing.T) {
	// Content wider/taller than the screen pins to (0,0).
	x, y := popup.Place(2, 2, 4, 1, 50, 30, 40, 20, popup.Below)
	if x != 0 || y != 0 {
		t.Fatalf("oversized: got (%d,%d), want (0,0)", x, y)
	}
}

func TestCenter(t *testing.T) {
	x, y := popup.Center(10, 4, 40, 20)
	if x != 15 || y != 8 {
		t.Fatalf("Center: got (%d,%d), want (15,8)", x, y)
	}
}

func TestCenterOversizedPinsToOrigin(t *testing.T) {
	x, y := popup.Center(60, 40, 40, 20)
	if x != 0 || y != 0 {
		t.Fatalf("Center oversized: got (%d,%d), want (0,0)", x, y)
	}
}

func TestIsolatePrefixesEveryLine(t *testing.T) {
	const reset = "\x1b[0m"
	out := popup.Isolate("a\nb\nc")
	lines := strings.Split(out, "\n")
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3", len(lines))
	}
	for i, line := range lines {
		if !strings.HasPrefix(line, reset) {
			t.Fatalf("line %d not prefixed with SGR reset: %q", i, line)
		}
	}
	if out != reset+"a\n"+reset+"b\n"+reset+"c" {
		t.Fatalf("unexpected isolate output: %q", out)
	}
}

func TestIsolateSingleLine(t *testing.T) {
	out := popup.Isolate("solo")
	if out != "\x1b[0m"+"solo" {
		t.Fatalf("got %q", out)
	}
	if strings.Contains(out, "\n") {
		t.Fatal("single-line input should not gain a newline")
	}
}
