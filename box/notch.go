package box

import "github.com/charmbracelet/lipgloss"

// NotchStyle bundles the lipgloss.Styles a Notch uses for its two
// slots. Each Notch carries its own NotchStyle, mirroring the
// per-instance styling pattern used elsewhere in this codebase
// (button.Button, navcard.Card, statusbar.Item) — the package owns
// layout and the consumer owns appearance.
type NotchStyle struct {
	Text  lipgloss.Style
	Badge lipgloss.Style
}

// Notch is a labeled cutout in a Box's top border. Despite the
// passing visual resemblance, this is *not* a navigation tab — it
// has no selection state, no view-switching, no click semantics.
// It's a purely-decorative inline label that breaks the top edge
// with a small notch shape.
//
// Visually, each Notch renders as ┐<Badge><Text>┌, where Badge is
// an optional accent (typically a keyboard-shortcut hint) drawn in
// the badge style, and Text is the main label drawn in the text
// style.
//
// Gap is the number of ─ dashes drawn before this notch's opening
// ┐ — i.e., the run of horizontal line that visually connects this
// notch to whatever sits to its left (the box's TL corner, the
// previous notch's closing ┌, or the trailing fill).
//
// Badge is just text, so callers pass whatever glyphs they want:
// glyph.Superscript(1) for `¹`, "*" for an asterisk, "▼" for a
// dropdown arrow, etc. Empty string means no badge.
type Notch struct {
	Text  string
	Gap   int
	Badge string
	Style NotchStyle
}
