package accordion_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/chenhunghan/boba/accordion"
)

func sample() accordion.Accordion {
	return accordion.Accordion{
		Sections: []accordion.Section{
			{Title: "One", Body: "body one"},
			{Title: "Two", Body: "line a\nline b"},
			{Title: "Three", Body: "body three"},
		},
		Expanded: []bool{false, false, false},
		Focused:  true,
	}
}

func TestRenderCollapsedShowsOnlyHeaders(t *testing.T) {
	a := sample()
	rows := strings.Split(a.Render(), "\n")
	if len(rows) != 3 {
		t.Fatalf("collapsed: got %d rows, want 3 (headers only)", len(rows))
	}
	for i, want := range []string{"One", "Two", "Three"} {
		if !strings.Contains(rows[i], want) {
			t.Errorf("row %d = %q, want it to contain %q", i, rows[i], want)
		}
	}
}

func TestRenderExpandedShowsBody(t *testing.T) {
	a := sample()
	a.Expanded[1] = true // Two has a 2-line body
	rows := strings.Split(a.Render(), "\n")
	// header One, header Two, 2 body lines, header Three = 5 rows
	if len(rows) != 5 {
		t.Fatalf("got %d rows, want 5", len(rows))
	}
	if !strings.Contains(rows[2], "line a") || !strings.Contains(rows[3], "line b") {
		t.Errorf("expanded body not rendered under its header: %q", rows)
	}
}

func TestGlyphFallbacks(t *testing.T) {
	a := sample()
	a.Expanded[0] = true
	rows := strings.Split(a.Render(), "\n")
	if !strings.HasPrefix(rows[0], "▼") {
		t.Errorf("expanded header should lead with ▼, got %q", rows[0])
	}
	// row[1] is the body of section 0; row[2] is the collapsed header Two.
	if !strings.HasPrefix(rows[2], "▶") {
		t.Errorf("collapsed header should lead with ▶, got %q", rows[2])
	}
}

func TestUpdateMovesCursor(t *testing.T) {
	a := sample()
	a, _ = a.Update(tea.KeyMsg{Type: tea.KeyDown})
	if a.Cursor != 1 {
		t.Fatalf("down: Cursor=%d, want 1", a.Cursor)
	}
	a, _ = a.Update(tea.KeyMsg{Type: tea.KeyUp})
	if a.Cursor != 0 {
		t.Fatalf("up: Cursor=%d, want 0", a.Cursor)
	}
}

func TestUpdateCursorClampsAtEnds(t *testing.T) {
	a := sample()
	a, _ = a.Update(tea.KeyMsg{Type: tea.KeyUp})
	if a.Cursor != 0 {
		t.Errorf("up at top: Cursor=%d, want 0 (no wrap)", a.Cursor)
	}
	a.Cursor = 2
	a, _ = a.Update(tea.KeyMsg{Type: tea.KeyDown})
	if a.Cursor != 2 {
		t.Errorf("down at bottom: Cursor=%d, want 2 (no wrap)", a.Cursor)
	}
}

func TestUpdateTogglesAndEmits(t *testing.T) {
	a := sample()
	a.Cursor = 1
	a, cmd := a.Update(tea.KeyMsg{Type: tea.KeySpace})
	if !a.Expanded[1] {
		t.Fatal("space should expand the section at Cursor")
	}
	msg, ok := cmd().(accordion.ToggledMsg)
	if !ok || msg.Index != 1 || !msg.Expanded {
		t.Fatalf("got %#v, want ToggledMsg{Index:1, Expanded:true}", cmd())
	}
}

func TestUpdateIgnoredWhenNotFocused(t *testing.T) {
	a := sample()
	a.Focused = false
	a2, cmd := a.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if a2.Expanded[0] || cmd != nil {
		t.Fatal("an unfocused accordion must ignore keys")
	}
}

// TestToggleDoesNotMutateOriginal pins the value-type contract: toggling
// returns a new Accordion without aliasing the caller's Expanded slice.
func TestToggleDoesNotMutateOriginal(t *testing.T) {
	a := sample()
	_, _ = a.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if a.Expanded[0] {
		t.Fatal("Update must not mutate the receiver's Expanded slice")
	}
}

// TestHitTestVariableHeights is the core invariant: HitTest walks the same
// layout as Render, so an expanded section's body shifts later headers down.
func TestHitTestVariableHeights(t *testing.T) {
	a := sample()
	a.Expanded[0] = true // section 0 body is 1 line, pushing Two/Three down
	// Layout rows: 0=One header, 1=One body, 2=Two header, 3=Three header.
	cases := []struct {
		y    int
		want int
		ok   bool
	}{
		{0, 0, true},   // One header
		{1, -1, false}, // One body
		{2, 1, true},   // Two header (shifted by the body)
		{3, 2, true},   // Three header
		{4, -1, false}, // past the end
	}
	for _, tc := range cases {
		got, ok := a.HitTest(0, tc.y)
		if got != tc.want || ok != tc.ok {
			t.Errorf("HitTest(0,%d) = (%d,%v), want (%d,%v)", tc.y, got, ok, tc.want, tc.ok)
		}
	}
}

func TestClickTogglesHeader(t *testing.T) {
	a := sample()
	a.Expanded[0] = true
	// Two's header is on row 2 because section 0's body occupies row 1.
	a, cmd := a.ClickAt(0, 2)
	if !a.Expanded[1] {
		t.Fatal("click on Two header should expand it")
	}
	if a.Cursor != 1 {
		t.Errorf("click should move Cursor to 1, got %d", a.Cursor)
	}
	if msg, ok := cmd().(accordion.ToggledMsg); !ok || msg.Index != 1 {
		t.Fatalf("got %#v, want ToggledMsg{Index:1}", cmd())
	}
}

func TestClickOnBodyIsNoop(t *testing.T) {
	a := sample()
	a.Expanded[0] = true
	if _, cmd := a.ClickAt(0, 1); cmd != nil { // row 1 = One's body
		t.Fatal("a click on a body row must be a no-op")
	}
}

func TestClickOutsideIsNoop(t *testing.T) {
	a := sample()
	if _, cmd := a.ClickAt(0, 99); cmd != nil {
		t.Fatal("a click below the stack must be a no-op")
	}
}

func TestEmptyAccordionIsSafe(t *testing.T) {
	var a accordion.Accordion
	a.Focused = true
	if got := a.Render(); got != "" {
		t.Errorf("empty Render = %q, want \"\"", got)
	}
	if _, cmd := a.Update(tea.KeyMsg{Type: tea.KeyEnter}); cmd != nil {
		t.Fatal("empty Update must be a no-op")
	}
	if i, ok := a.HitTest(0, 0); ok || i != -1 {
		t.Fatalf("empty HitTest = (%d,%v), want (-1,false)", i, ok)
	}
}
