package pins

import (
	"reflect"
	"testing"
)

// TestPinDeduplicates verifies that pinning the same ID twice
// leaves the list unchanged — no duplicate entry, no reordering.
func TestPinDeduplicates(t *testing.T) {
	var l List
	l.Pin("a")
	l.Pin("b")
	l.Pin("a") // duplicate
	if got, want := l.IDs(), []string{"a", "b"}; !reflect.DeepEqual(got, want) {
		t.Errorf("ids = %v, want %v", got, want)
	}
	if l.Len() != 2 {
		t.Errorf("Len = %d, want 2", l.Len())
	}
}

// TestUnpinAbsentNoop confirms Unpin on a missing ID is harmless.
func TestUnpinAbsentNoop(t *testing.T) {
	var l List
	l.Pin("a")
	l.Unpin("z")
	if got, want := l.IDs(), []string{"a"}; !reflect.DeepEqual(got, want) {
		t.Errorf("ids = %v, want %v", got, want)
	}
}

// TestUnpinClampsSelected verifies that the selection cursor stays
// within bounds after a removal — and goes to -1 when the list
// becomes empty.
func TestUnpinClampsSelected(t *testing.T) {
	var l List
	l.Pin("a")
	l.Pin("b")
	l.Pin("c")
	l.SetSelected(2) // cursor on "c"
	l.Unpin("c")     // removes the selected item
	if got := l.Selected(); got != 1 {
		t.Errorf("Selected after removing last = %d, want 1", got)
	}
	l.Unpin("b")
	l.Unpin("a")
	if got := l.Selected(); got != -1 {
		t.Errorf("Selected on empty = %d, want -1", got)
	}
}

// TestUnpinKeepsSelectedWhenAhead verifies that removing an item
// at or after the cursor leaves the cursor where it points (the
// cursor index isn't shifted for removals to its right).
func TestUnpinKeepsSelectedWhenAhead(t *testing.T) {
	var l List
	l.Pin("a")
	l.Pin("b")
	l.Pin("c")
	l.SetSelected(0) // cursor on "a"
	l.Unpin("c")     // remove from the right of the cursor
	if got, want := l.Selected(), 0; got != want {
		t.Errorf("Selected = %d, want %d (cursor should be unchanged)", got, want)
	}
}

// TestMoveBounds checks that Move clamps to [0, Len()).
func TestMoveBounds(t *testing.T) {
	var l List
	l.Pin("a")
	l.Pin("b")
	l.Pin("c")
	// Selected starts at 0.
	l.Move(-5)
	if got, want := l.Selected(), 0; got != want {
		t.Errorf("Move(-5) Selected = %d, want %d", got, want)
	}
	l.Move(+10)
	if got, want := l.Selected(), 2; got != want {
		t.Errorf("Move(+10) Selected = %d, want %d", got, want)
	}
}

// TestMoveOnEmpty confirms moving an empty list is harmless and
// Selected stays at -1.
func TestMoveOnEmpty(t *testing.T) {
	var l List
	l.Move(+1)
	if got := l.Selected(); got != -1 {
		t.Errorf("Selected on empty after Move = %d, want -1", got)
	}
}

// TestSetSelectedClamps confirms a too-large or negative idx is
// clamped to the valid range.
func TestSetSelectedClamps(t *testing.T) {
	var l List
	l.Pin("a")
	l.Pin("b")
	l.SetSelected(99)
	if got, want := l.Selected(), 1; got != want {
		t.Errorf("SetSelected(99) Selected = %d, want %d", got, want)
	}
	l.SetSelected(-3)
	if got, want := l.Selected(), 0; got != want {
		t.Errorf("SetSelected(-3) Selected = %d, want %d", got, want)
	}
}

// TestHoverIdx covers boundary conditions on the y-coord → index
// projection that consumers use during hit-testing.
func TestHoverIdx(t *testing.T) {
	var l List
	l.Pin("a")
	l.Pin("b")
	l.Pin("c")
	cases := []struct {
		y, itemHeight, want int
	}{
		{-1, 1, -1}, // before stack
		{0, 1, 0},   // first row → first item
		{1, 1, 1},   // second row → second item
		{2, 1, 2},   // third row → third item
		{3, 1, -1},  // past the last item
		{0, 2, 0},   // taller items: first row is item 0
		{1, 2, 0},   // taller items: second row is also item 0
		{2, 2, 1},   // taller items: third row is item 1
		{0, 0, -1},  // invalid itemHeight
		{0, -1, -1}, // invalid itemHeight
	}
	for _, tc := range cases {
		if got := l.HoverIdx(tc.y, tc.itemHeight); got != tc.want {
			t.Errorf("HoverIdx(%d, %d) = %d, want %d", tc.y, tc.itemHeight, got, tc.want)
		}
	}
}

// TestContainsIndexOf checks the lookup helpers agree with each
// other and with the underlying slice.
func TestContainsIndexOf(t *testing.T) {
	var l List
	l.Pin("a")
	l.Pin("b")
	if !l.Contains("a") {
		t.Error("Contains(a) = false, want true")
	}
	if l.Contains("z") {
		t.Error("Contains(z) = true, want false")
	}
	if got := l.IndexOf("b"); got != 1 {
		t.Errorf("IndexOf(b) = %d, want 1", got)
	}
	if got := l.IndexOf("z"); got != -1 {
		t.Errorf("IndexOf(z) = %d, want -1", got)
	}
}
