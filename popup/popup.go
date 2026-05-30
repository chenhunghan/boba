// Package popup provides pure placement helpers for overlaid content: where
// to put a floating box relative to an anchor (flipping when it would run off
// screen), how to center it, and how to keep it from inheriting stray ANSI
// from whatever it covers. It is a primitive — just functions, no widget
// struct — so callers compose it with their own overlay/positioning code.
package popup

import "strings"

// Placement is the preferred side of the anchor to put the content on. When
// that side would overflow the screen, Place flips to the opposite side.
type Placement int

const (
	// Below puts the content under the anchor (its left edge aligned).
	Below Placement = iota
	// Above puts the content over the anchor.
	Above
	// Right puts the content to the anchor's right (top edges aligned).
	Right
	// Left puts the content to the anchor's left.
	Left
)

// Place returns the screen top-left (x, y) at which to render a contentW ×
// contentH box for the given Placement relative to the anchor rectangle
// (anchorX, anchorY, anchorW, anchorH).
//
// The preferred side is used unless the box would overflow that edge of the
// screen, in which case Place flips to the opposite side. The result is then
// clamped so the box stays within [0, screen) on each axis when it fits;
// content larger than the screen is pinned to the top-left.
func Place(anchorX, anchorY, anchorW, anchorH, contentW, contentH, screenW, screenH int, p Placement) (int, int) {
	var x, y int
	switch p {
	case Above:
		x = anchorX
		y = flip(anchorY-contentH, anchorY+anchorH, contentH, screenH)
	case Right:
		x = flip(anchorX+anchorW, anchorX-contentW, contentW, screenW)
		y = anchorY
	case Left:
		x = flip(anchorX-contentW, anchorX+anchorW, contentW, screenW)
		y = anchorY
	default: // Below
		x = anchorX
		y = flip(anchorY+anchorH, anchorY-contentH, contentH, screenH)
	}
	return clamp(x, contentW, screenW), clamp(y, contentH, screenH)
}

// Center returns the top-left at which a contentW × contentH box is centered
// in a screenW × screenH screen, clamped so a too-large box pins to (0, 0).
func Center(contentW, contentH, screenW, screenH int) (int, int) {
	return clamp((screenW-contentW)/2, contentW, screenW),
		clamp((screenH-contentH)/2, contentH, screenH)
}

// Isolate prefixes every line of content with an SGR reset so an overlaid
// block does not inherit lingering ANSI (colors, attributes) from the cells
// it is painted on top of. It does not otherwise alter the content.
func Isolate(content string) string {
	const sgrReset = "\x1b[0m"
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lines[i] = sgrReset + line
	}
	return strings.Join(lines, "\n")
}

// flip returns pref unless placing size cells there overflows [0, limit), in
// which case it returns the alternative position on the opposite side.
func flip(pref, alt, size, limit int) int {
	if pref < 0 || pref+size > limit {
		return alt
	}
	return pref
}

// clamp keeps a top-left coordinate within [0, limit-size] so a size-cell box
// stays on screen; a box larger than the screen is pinned to 0.
func clamp(pos, size, limit int) int {
	if pos < 0 {
		return 0
	}
	if max := limit - size; pos > max {
		if max < 0 {
			return 0
		}
		return max
	}
	return pos
}
