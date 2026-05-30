// Package pins is a state machine for an ordered list of opaque
// string IDs with a selection cursor. The package owns the
// invariants — dedupe on Pin, clamp on Unpin, bounds on Move — so
// callers don't reinvent them at each mutation site.
//
// pins.List is purely behavioral: it knows nothing about what the
// IDs identify or how they're rendered. A consumer wraps a
// renderable component (typically button.Stack) around the list to
// display it. This headless split lets the same primitive back any
// UI for ordered, pinnable references — favorites, recents,
// shortcut bars — with the consumer providing the visual.
//
// Typical use:
//
//	var hot pins.List
//	hot.Pin("abc")
//	hot.Pin("def")
//	hot.Move(+1)              // selection moves to "def"
//	idx := hot.HoverIdx(y, 1) // y → index for hit-testing
package pins

// List is an ordered set of string IDs with a selection cursor.
// Methods mutate via pointer receiver. The zero value is a valid
// empty list.
type List struct {
	ids      []string
	selected int
}

// Pin appends id to the list if not already present. No-op if id
// is already pinned. Pinning never moves Selected — the user's
// current cursor position is preserved.
func (l *List) Pin(id string) {
	if l.Contains(id) {
		return
	}
	l.ids = append(l.ids, id)
}

// Unpin removes id from the list (no-op if absent) and clamps the
// selection cursor so it remains within bounds. After unpinning the
// last item, Selected returns -1.
func (l *List) Unpin(id string) {
	i := l.IndexOf(id)
	if i < 0 {
		return
	}
	l.ids = append(l.ids[:i], l.ids[i+1:]...)
	l.clampSelected()
}

// Contains reports whether id is currently in the list.
func (l *List) Contains(id string) bool {
	return l.IndexOf(id) >= 0
}

// IndexOf returns the position of id in the list, or -1 if not present.
func (l *List) IndexOf(id string) int {
	for i, x := range l.ids {
		if x == id {
			return i
		}
	}
	return -1
}

// Move shifts the selection cursor by delta (negative = toward
// index 0, positive = toward the end), clamping to [0, Len()).
// No-op when the list is empty.
func (l *List) Move(delta int) {
	if len(l.ids) == 0 {
		return
	}
	s := l.selected + delta
	if s < 0 {
		s = 0
	}
	if s >= len(l.ids) {
		s = len(l.ids) - 1
	}
	l.selected = s
}

// SetSelected jumps the cursor to idx, clamped to [0, Len()).
// Used by paths that select from a click rather than an arrow key.
// No-op semantics on empty lists (idx is clamped to 0 internally,
// Selected returns -1 by accessor convention).
func (l *List) SetSelected(idx int) {
	l.selected = idx
	l.clampSelected()
}

// Selected returns the current selection index, or -1 when the
// list is empty. When non-empty the value is always within [0, Len()).
func (l *List) Selected() int {
	if len(l.ids) == 0 {
		return -1
	}
	return l.selected
}

// Len returns the number of pinned IDs.
func (l *List) Len() int { return len(l.ids) }

// IDs returns the current list of IDs in display order. The
// returned slice is owned by List — callers must not mutate it.
func (l *List) IDs() []string { return l.ids }

// HoverIdx returns the index of the item at panel-local y when
// each item occupies itemHeight rows, or -1 if y is outside the
// list. itemHeight must be > 0; a non-positive value returns -1.
// Geometry math lives here so callers don't reinvent it (and so
// it stays consistent with whatever the consumer's renderer does).
func (l *List) HoverIdx(y, itemHeight int) int {
	if y < 0 || itemHeight <= 0 {
		return -1
	}
	idx := y / itemHeight
	if idx >= len(l.ids) {
		return -1
	}
	return idx
}

// clampSelected enforces 0 <= selected < len(ids). On an empty
// list, selected is forced to 0 internally; the Selected accessor
// returns -1 in that case to express "no valid selection."
func (l *List) clampSelected() {
	if len(l.ids) == 0 {
		l.selected = 0
		return
	}
	if l.selected < 0 {
		l.selected = 0
	}
	if l.selected >= len(l.ids) {
		l.selected = len(l.ids) - 1
	}
}
