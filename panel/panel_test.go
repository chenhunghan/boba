package panel

import "testing"

// area is a stand-in for an app-defined panel/focus enum.
type area int

const (
	none area = iota
	hot
	nav
	main
	status
)

// TestHitTestSinglePanelRoot verifies that a tree consisting of a
// single Panel returns Found=true with content-local coords for any
// in-bounds query, including the corners. The Panel's own Size is
// ignored (no parent), so the panel always fills the screen rect.
func TestHitTestSinglePanelRoot(t *testing.T) {
	root := Panel[area]{ID: hot, Size: 999, Borders: Borders{Top: 1, Left: 2}}
	cases := []struct {
		name string
		x, y int
		want Hit[area]
		w, h int
	}{
		{"top-left in border", 0, 0, Hit[area]{Found: true, Panel: hot, LocalX: -2, LocalY: -1}, 10, 10},
		{"inside content", 5, 3, Hit[area]{Found: true, Panel: hot, LocalX: 3, LocalY: 2}, 10, 10},
		{"bottom-right corner", 9, 9, Hit[area]{Found: true, Panel: hot, LocalX: 7, LocalY: 8}, 10, 10},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := HitTest[area](root, tc.w, tc.h, tc.x, tc.y)
			if got != tc.want {
				t.Errorf("HitTest(%d,%d) = %+v, want %+v", tc.x, tc.y, got, tc.want)
			}
		})
	}
}

// TestHitTestOutOfBounds verifies the boundary contract: any query
// outside the (width, height) rect returns Found=false regardless of
// what the layout looks like.
func TestHitTestOutOfBounds(t *testing.T) {
	root := Panel[area]{ID: hot, Size: 0}
	cases := []struct {
		name string
		x, y int
	}{
		{"x negative", -1, 5},
		{"y negative", 5, -1},
		{"x at width", 10, 5},
		{"y at height", 5, 10},
		{"both negative", -1, -1},
		{"way past", 999, 999},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := HitTest[area](root, 10, 10, tc.x, tc.y)
			if got.Found {
				t.Errorf("HitTest(%d,%d) returned Found=true; want false", tc.x, tc.y)
			}
		})
	}
}

// TestHitTestHorizontalSplit verifies that a Horizontal split
// dispatches x correctly across fixed-Size and remainder children,
// with y passed through unchanged. Mirrors a typical content row.
func TestHitTestHorizontalSplit(t *testing.T) {
	root := Split[area]{
		Axis: Horizontal,
		Children: []Node[area]{
			Panel[area]{ID: hot, Size: 5},  // x=0..4
			Panel[area]{ID: nav, Size: 30}, // x=5..34
			Panel[area]{ID: main, Size: 0}, // x=35..(width-1) (remainder)
		},
	}
	width, height := 100, 20
	cases := []struct {
		name       string
		x, y       int
		wantPanel  area
		wantLocalX int
	}{
		{"first cell of hot", 0, 7, hot, 0},
		{"last cell of hot", 4, 7, hot, 4},
		{"first cell of nav", 5, 7, nav, 0},
		{"last cell of nav", 34, 7, nav, 29},
		{"first cell of main (remainder)", 35, 7, main, 0},
		{"middle of main", 60, 7, main, 25},
		{"last cell of main", 99, 7, main, 64},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := HitTest[area](root, width, height, tc.x, tc.y)
			if !got.Found {
				t.Fatalf("Found=false; want true")
			}
			if got.Panel != tc.wantPanel {
				t.Errorf("Panel = %v, want %v", got.Panel, tc.wantPanel)
			}
			if got.LocalX != tc.wantLocalX {
				t.Errorf("LocalX = %d, want %d", got.LocalX, tc.wantLocalX)
			}
			if got.LocalY != tc.y {
				t.Errorf("LocalY = %d, want %d (no vertical translation)", got.LocalY, tc.y)
			}
		})
	}
}

// TestHitTestVerticalSplit mirrors the horizontal test for the y
// axis. Verifies that x passes through unchanged.
func TestHitTestVerticalSplit(t *testing.T) {
	root := Split[area]{
		Axis: Vertical,
		Children: []Node[area]{
			Panel[area]{ID: main, Size: 0},   // y=0..(height-2) (remainder)
			Panel[area]{ID: status, Size: 1}, // y=height-1
		},
	}
	width, height := 100, 20
	cases := []struct {
		name       string
		x, y       int
		wantPanel  area
		wantLocalY int
	}{
		{"top row of main", 50, 0, main, 0},
		{"middle of main", 50, 10, main, 10},
		{"last row of main", 50, 18, main, 18},
		{"status row", 50, 19, status, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := HitTest[area](root, width, height, tc.x, tc.y)
			if !got.Found {
				t.Fatalf("Found=false; want true")
			}
			if got.Panel != tc.wantPanel {
				t.Errorf("Panel = %v, want %v", got.Panel, tc.wantPanel)
			}
			if got.LocalY != tc.wantLocalY {
				t.Errorf("LocalY = %d, want %d", got.LocalY, tc.wantLocalY)
			}
			if got.LocalX != tc.x {
				t.Errorf("LocalX = %d, want %d (no horizontal translation)", got.LocalX, tc.x)
			}
		})
	}
}

// TestHitTestNestedSplit verifies that nested Splits propagate origin
// translation and Borders subtraction correctly down the tree.
// Layout: vertical(horizontal(sidebar, content, main), status).
func TestHitTestNestedSplit(t *testing.T) {
	root := Split[area]{
		Axis: Vertical,
		Children: []Node[area]{
			Split[area]{
				Axis: Horizontal,
				Size: 0, // content fills remainder height
				Children: []Node[area]{
					Panel[area]{ID: hot, Size: 5, Borders: Borders{Top: 1, Left: 1}},
					Panel[area]{ID: nav, Size: 30, Borders: Borders{Top: 1, Left: 1}},
					Panel[area]{ID: main, Size: 0, Borders: Borders{Top: 1, Left: 1}},
				},
			},
			Panel[area]{ID: status, Size: 1},
		},
	}
	width, height := 100, 20

	cases := []struct {
		name                   string
		x, y                   int
		wantPanel              area
		wantLocalX, wantLocalY int
	}{
		// hot (col 0..4): Borders Top=1 Left=1 → (-1,-1) at the corners
		{"hot top-left border cell", 0, 0, hot, -1, -1},
		{"hot first content cell", 1, 1, hot, 0, 0},
		{"hot deep content", 4, 18, hot, 3, 17},
		// nav (col 5..34): Borders Top=1 Left=1
		{"nav top-left border", 5, 0, nav, -1, -1},
		{"nav first content cell", 6, 1, nav, 0, 0},
		{"nav last content cell", 34, 18, nav, 28, 17},
		// main (col 35..99): Borders Top=1 Left=1
		{"main top-left border", 35, 0, main, -1, -1},
		{"main first content cell", 36, 1, main, 0, 0},
		{"main far cell", 99, 18, main, 63, 17},
		// status (row 19): no borders
		{"status leftmost", 0, 19, status, 0, 0},
		{"status middle", 50, 19, status, 50, 0},
		{"status rightmost", 99, 19, status, 99, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := HitTest[area](root, width, height, tc.x, tc.y)
			if !got.Found {
				t.Fatalf("Found=false; want true")
			}
			want := Hit[area]{Found: true, Panel: tc.wantPanel, LocalX: tc.wantLocalX, LocalY: tc.wantLocalY}
			if got != want {
				t.Errorf("HitTest(%d,%d) = %+v, want %+v", tc.x, tc.y, got, want)
			}
		})
	}
}

// TestEmptySplit verifies that a Split with no Children is a black
// hole — every query inside its rect returns Found=false rather
// than crashing or returning a zero-value panel.
func TestEmptySplit(t *testing.T) {
	root := Split[area]{Axis: Horizontal, Children: nil}
	got := HitTest[area](root, 10, 10, 5, 5)
	if got.Found {
		t.Errorf("empty Split: Found=true; want false")
	}
}

// TestFixedSizesUnderfill verifies that when fixed Sizes don't
// account for the full parent extent and there's no remainder child,
// queries in the unfilled region return Found=false.
func TestFixedSizesUnderfill(t *testing.T) {
	root := Split[area]{
		Axis: Horizontal,
		Children: []Node[area]{
			Panel[area]{ID: hot, Size: 5}, // x=0..4
			Panel[area]{ID: nav, Size: 5}, // x=5..9
			// nothing covers x=10..99
		},
	}
	got := HitTest[area](root, 100, 10, 50, 5)
	if got.Found {
		t.Errorf("query in unfilled region: Found=true; want false")
	}
	// In-bounds query should still work.
	got = HitTest[area](root, 100, 10, 7, 5)
	if !got.Found || got.Panel != nav {
		t.Errorf("in-bounds query failed: %+v", got)
	}
}

// TestPanicOnMultipleRemainder verifies the documented invariant:
// at most one Size=0 child per Split. More than one is a structural
// bug and HitTest panics.
func TestPanicOnMultipleRemainder(t *testing.T) {
	root := Split[area]{
		Axis: Horizontal,
		Children: []Node[area]{
			Panel[area]{ID: hot, Size: 0},
			Panel[area]{ID: nav, Size: 0},
		},
	}
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic on multiple Size=0 children")
		}
	}()
	HitTest[area](root, 100, 10, 5, 5)
}

// TestSingleRemainderInSplit verifies that one Size=0 child takes
// the entire remainder regardless of where it sits in the children
// list (first, middle, last) — the index doesn't matter.
func TestSingleRemainderInSplit(t *testing.T) {
	cases := []struct {
		name string
		root Split[area]
		x    int
		want area
	}{
		{
			"remainder first",
			Split[area]{Axis: Horizontal, Children: []Node[area]{
				Panel[area]{ID: main, Size: 0},
				Panel[area]{ID: hot, Size: 5},
				Panel[area]{ID: nav, Size: 30},
			}},
			50, main, // remainder spans 0..64
		},
		{
			"remainder middle",
			Split[area]{Axis: Horizontal, Children: []Node[area]{
				Panel[area]{ID: hot, Size: 5},
				Panel[area]{ID: main, Size: 0},
				Panel[area]{ID: nav, Size: 30},
			}},
			40, main, // remainder spans 5..69
		},
		{
			"remainder last",
			Split[area]{Axis: Horizontal, Children: []Node[area]{
				Panel[area]{ID: hot, Size: 5},
				Panel[area]{ID: nav, Size: 30},
				Panel[area]{ID: main, Size: 0},
			}},
			60, main, // remainder spans 35..99
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := HitTest[area](tc.root, 100, 10, tc.x, 5)
			if !got.Found || got.Panel != tc.want {
				t.Errorf("got %+v, want Panel=%v", got, tc.want)
			}
		})
	}
}
