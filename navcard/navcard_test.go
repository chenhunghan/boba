package navcard

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/chenhunghan/boba/button"
)

// testCardStyle is a minimal Card Style. Concrete colors don't matter
// for dimension tests — we only care about row counts and visible widths.
var testCardStyle = func() Style {
	bg := lipgloss.NewStyle().Background(lipgloss.Color("237"))
	stateStyle := StateStyle{
		Bar:         lipgloss.NewStyle().Foreground(lipgloss.Color("99")),
		BarChar:     "▌",
		Fill:        bg,
		Title:       bg.Foreground(lipgloss.Color("231")).Bold(true),
		Subtitle:    bg.Foreground(lipgloss.Color("245")),
		Description: bg.Foreground(lipgloss.Color("245")),
	}
	return Style{
		Inactive: stateStyle,
		Hover:    stateStyle,
		Active:   stateStyle,
	}
}()

// testBtnStyle is the per-button style attached to inline action
// buttons in tests. Parallel to testCardStyle: each component carries
// its own.
var testBtnStyle = button.Style{
	Inactive: lipgloss.NewStyle().Background(lipgloss.Color("63")),
	Hover:    lipgloss.NewStyle().Background(lipgloss.Color("99")),
	Active:   lipgloss.NewStyle().Background(lipgloss.Color("33")),
}

// styledCard returns the card with testCardStyle applied; inline
// buttons get testBtnStyle. Keeps test rows readable.
func styledCard(c Card) Card {
	c.Style = testCardStyle
	for i := range c.Buttons {
		c.Buttons[i].Style = testBtnStyle
	}
	return c
}

// TestCardHeight verifies the height contract: 2 base rows of
// padding plus 1 row per filled slot. Buttons that fit in a single
// row contribute 1 row; wider button lists wrap and add more rows.
func TestCardHeight(t *testing.T) {
	cases := []struct {
		name  string
		card  Card
		width int
		want  int
	}{
		{"empty", Card{}, 30, 2},
		{"title only", Card{Title: "x"}, 30, 3},
		{"title + subtitle", Card{Title: "x", Subtitle: "y"}, 30, 4},
		{"all slots, 1-row buttons",
			Card{Title: "x", Subtitle: "y", Description: "z", Buttons: []button.Button{{Text: "a"}}},
			30, 6,
		},
		{"buttons wrap to 2 rows when narrow",
			// two buttons of width 6 each + 1 gap = 13 cells; available
			// at width=10 is 7 cells → second button wraps.
			Card{Buttons: []button.Button{{Text: "Edit"}, {Text: "Mute"}}},
			10, 4, // 2 padding + 0 slots + 2 button rows
		},
		{"buttons fit in one row when wide",
			Card{Buttons: []button.Button{{Text: "Edit"}, {Text: "Mute"}}},
			30, 3, // 2 padding + 1 button row
		},
		{"custom",
			Card{
				Custom:       func(int, State, Style) string { return "" },
				CustomHeight: 7,
			},
			30, 7,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.card.Height(tc.width); got != tc.want {
				t.Errorf("Height(%d) = %d, want %d", tc.width, got, tc.want)
			}
		})
	}
}

// TestRenderDimensions verifies the rendered output is exactly width
// cells per row and Card.Height(width) rows total.
func TestRenderDimensions(t *testing.T) {
	cards := []Card{
		styledCard(Card{Title: "A title"}),
		styledCard(Card{Title: "Another", Subtitle: "with subtitle"}),
		styledCard(Card{Title: "Big", Subtitle: "sub", Description: "desc", Buttons: []button.Button{{Text: "Edit"}}}),
	}
	width := 30
	for _, card := range cards {
		for _, state := range []State{StateInactive, StateHover, StateActive} {
			out := card.Render(state, width, -1)
			lines := strings.Split(out, "\n")
			if len(lines) != card.Height(width) {
				t.Errorf("state=%d card=%+v: got %d lines, want %d", state, card, len(lines), card.Height(width))
			}
			for i, line := range lines {
				if w := lipgloss.Width(line); w != width {
					t.Errorf("state=%d row %d: visible width %d, want %d (%q)", state, i, w, width, line)
				}
			}
		}
	}
}

// TestRenderWrappedButtons verifies that when buttons wrap, each
// button row is rendered at the same width and the total card height
// reflects the wrap.
func TestRenderWrappedButtons(t *testing.T) {
	card := styledCard(Card{
		Buttons: []button.Button{
			{Text: "Edit"}, {Text: "Mute"}, {Text: "Save"}, {Text: "Run"},
		},
	})
	width := 14 // available for buttons = 14 - 1 - 2 = 11; each btn ~6 cells
	out := card.Render(StateInactive, width, -1)
	lines := strings.Split(out, "\n")
	// Expected: 2 padding + 4 wrapped button rows? No, ~ 2 buttons fit per row.
	// "Edit"(6) + 1 gap + "Mute"(6) = 13 > 11 → 1 button per row → 4 button rows.
	wantH := card.Height(width)
	if len(lines) != wantH {
		t.Fatalf("got %d lines, want %d", len(lines), wantH)
	}
	for i, line := range lines {
		if w := lipgloss.Width(line); w != width {
			t.Errorf("row %d: width %d, want %d", i, w, width)
		}
	}
}

// TestStackRenderAndHitTestAgree verifies Render and HitTest walk
// the same heights and gaps — no drift possible.
func TestStackRenderAndHitTestAgree(t *testing.T) {
	stack := Stack{
		Cards: []Card{
			styledCard(Card{Title: "A"}),                      // height 3
			styledCard(Card{Title: "B", Subtitle: "subB"}),    // height 4
			styledCard(Card{Title: "C", Description: "desc"}), // height 4
		},
		Width: 30,
		Gap:   1,
	}
	out := stack.Render()
	rows := strings.Split(out, "\n")

	// Expected total: 3 + 1 + 4 + 1 + 4 = 13 rows.
	if got := len(rows); got != 13 {
		t.Fatalf("Render produced %d rows, want 13", got)
	}

	cases := []struct {
		y    int
		want int // expected Card index
		why  string
	}{
		{0, 0, "first row of card 0"},
		{2, 0, "last row of card 0"},
		{3, -1, "gap row after card 0"},
		{4, 1, "first row of card 1"},
		{7, 1, "last row of card 1"},
		{8, -1, "gap row after card 1"},
		{9, 2, "first row of card 2"},
		{12, 2, "last row of card 2"},
		{13, -1, "past the last card"},
		{-1, -1, "before the stack"},
	}
	for _, tc := range cases {
		// x=0 → over the card's bar (definitely not a button), so
		// HitTest's Button result will be -1 for all of these.
		if got := stack.HitTest(0, tc.y); got.Card != tc.want {
			t.Errorf("HitTest(0, %d).Card = %d, want %d (%s)", tc.y, got.Card, tc.want, tc.why)
		}
	}
}

// TestStackButtonHitTest verifies that clicks on inline buttons in a
// single-row layout return the right (Card, Button) pair.
func TestStackButtonHitTest(t *testing.T) {
	// Card with two inline buttons that fit in one row at width 30:
	//   row 0: top pad
	//   row 1: title
	//   row 2: button row → "  [ Edit ] [ Mute ]" (each button auto-sized)
	//   row 3: bottom pad
	stack := Stack{
		Cards: []Card{styledCard(Card{
			Title:   "thing",
			Buttons: []button.Button{{Text: "Edit"}, {Text: "Mute"}},
		})},
		Width: 30,
	}
	cases := []struct {
		x, y       int
		wantCard   int
		wantButton int
		why        string
	}{
		// Buttons are 6 cells each ("Edit" 4 + 2 padding), gap 1 cell.
		// x layout: bar(0), prefix(1,2), btn0(3..8), gap(9), btn1(10..15)
		{3, 2, 0, 0, "first cell of Edit"},
		{8, 2, 0, 0, "last cell of Edit"},
		{9, 2, 0, -1, "gap between buttons"},
		{10, 2, 0, 1, "first cell of Mute"},
		{15, 2, 0, 1, "last cell of Mute"},
		{16, 2, 0, -1, "past Mute, still in card body"},
		{1, 2, 0, -1, "card prefix area on button row"},
		{5, 1, 0, -1, "title row, not the button row"},
		{5, 0, 0, -1, "top-pad row"},
	}
	for _, tc := range cases {
		got := stack.HitTest(tc.x, tc.y)
		if got.Card != tc.wantCard || got.Button != tc.wantButton {
			t.Errorf("HitTest(%d, %d) = %+v, want Hit{%d, %d} (%s)",
				tc.x, tc.y, got, tc.wantCard, tc.wantButton, tc.why)
		}
	}
}

// TestStackButtonHitTestWrapped verifies that when buttons wrap
// across multiple rows, hit-testing still returns the right global
// button index.
func TestStackButtonHitTestWrapped(t *testing.T) {
	// Card with three buttons that wrap when width is narrow.
	// Width 14 → available = 11. "Edit"(6) + 1 + "Mute"(6) = 13 > 11
	// so each button takes its own row. 3 buttons → 3 wrap rows.
	// Card layout:
	//   row 0: top pad
	//   row 1: button row 0 → "Edit" (idx 0)
	//   row 2: button row 1 → "Mute" (idx 1)
	//   row 3: button row 2 → "Save" (idx 2)
	//   row 4: bottom pad
	stack := Stack{
		Cards: []Card{styledCard(Card{
			Buttons: []button.Button{{Text: "Edit"}, {Text: "Mute"}, {Text: "Save"}},
		})},
		Width: 14,
	}
	cases := []struct {
		x, y       int
		wantCard   int
		wantButton int
		why        string
	}{
		{3, 1, 0, 0, "row 0 (Edit)"},
		{3, 2, 0, 1, "row 1 (Mute)"},
		{3, 3, 0, 2, "row 2 (Save)"},
		{8, 1, 0, 0, "last cell of Edit on row 0"},
		{9, 1, 0, -1, "past Edit on row 0 (no wrap-mate, gap area)"},
		{1, 1, 0, -1, "prefix area on button row"},
		{3, 0, 0, -1, "top-pad row"},
		{3, 4, 0, -1, "bottom-pad row"},
	}
	for _, tc := range cases {
		got := stack.HitTest(tc.x, tc.y)
		if got.Card != tc.wantCard || got.Button != tc.wantButton {
			t.Errorf("HitTest(%d, %d) = %+v, want Hit{%d, %d} (%s)",
				tc.x, tc.y, got, tc.wantCard, tc.wantButton, tc.why)
		}
	}
}

// TestRightButtons verifies that right-aligned buttons share the
// last left-button row, that hit-testing distinguishes left vs
// right buttons via global index, and that a card with only
// RightButtons gets one row anchored for them.
func TestRightButtons(t *testing.T) {
	width := 30
	t.Run("left + right share last row", func(t *testing.T) {
		card := styledCard(Card{
			Buttons:      []button.Button{{Text: "stop"}}, // 1 left button
			RightButtons: []button.Button{{Text: "⋯"}},    // 1 right button
		})
		// Height = 2 padding + 1 button row = 3
		if got, want := card.Height(width), 3; got != want {
			t.Errorf("Height = %d, want %d", got, want)
		}
		// Click on left button (global idx 0)
		stack := Stack{Cards: []Card{card}, Width: width}
		if got := stack.HitTest(3, 1); got.Button != 0 {
			t.Errorf("left button click: Button=%d, want 0", got.Button)
		}
		// "⋯" button width = 1 cell text + 2 padding = 3 cells.
		// rightStart = width - rightW - rightButtonInset = 30 - 3 - 1 = 26.
		// Right button occupies columns 26, 27, 28. Global idx 1.
		if got := stack.HitTest(26, 1); got.Button != 1 {
			t.Errorf("right button click: Button=%d, want 1", got.Button)
		}
		// Click in the spacer middle should hit nothing
		if got := stack.HitTest(15, 1); got.Button != -1 {
			t.Errorf("spacer click: Button=%d, want -1", got.Button)
		}
		// Click in the right inset (col 29) should also hit nothing.
		if got := stack.HitTest(29, 1); got.Button != -1 {
			t.Errorf("right-inset click: Button=%d, want -1", got.Button)
		}
	})

	t.Run("right only gets its own row", func(t *testing.T) {
		card := styledCard(Card{
			RightButtons: []button.Button{{Text: "⋯"}},
		})
		// Height = 2 padding + 1 row for right buttons = 3
		if got, want := card.Height(width), 3; got != want {
			t.Errorf("right-only Height = %d, want %d", got, want)
		}
		stack := Stack{Cards: []Card{card}, Width: width}
		// Right button at columns 26..28 (inset 1 from right edge);
		// global idx 0 (no left buttons).
		if got := stack.HitTest(26, 1); got.Button != 0 {
			t.Errorf("right-only button click: Button=%d, want 0", got.Button)
		}
	})
}

// TestRenderTruncatesOverflowingText verifies that a card with a
// title/subtitle/description longer than the card's contentWidth still
// renders rows of exactly `width` cells. Without truncation the row
// would overflow and break the surrounding panel's right border.
func TestRenderTruncatesOverflowingText(t *testing.T) {
	width := 23 // a narrow sidebar card width
	card := styledCard(Card{
		Title:        "● a long card title that overflows", // wider than width
		Subtitle:     "running with a very long subtitle",  // wider than width
		Description:  "a long description that overflows",  // wider than width
		Buttons:      []button.Button{{Text: "stop"}},
		RightButtons: []button.Button{{Text: "⋯"}},
	})
	out := card.Render(StateInactive, width, -1)
	for i, line := range strings.Split(out, "\n") {
		if w := lipgloss.Width(line); w != width {
			t.Errorf("row %d: visible width %d, want %d (%q)", i, w, width, line)
		}
	}
}

// TestCustomBody verifies the Custom hook replaces the default body
// renderer and is given the right contentWidth.
func TestCustomBody(t *testing.T) {
	gotWidth := 0
	card := styledCard(Card{
		CustomHeight: 3,
		Custom: func(contentWidth int, _ State, _ Style) string {
			gotWidth = contentWidth
			line := strings.Repeat("X", contentWidth)
			return line + "\n" + line + "\n" + line
		},
	})
	out := card.Render(StateInactive, 10, -1)
	if gotWidth != 9 {
		t.Errorf("Custom got contentWidth=%d, want 9 (= 10 - 1 for bar)", gotWidth)
	}
	if !strings.Contains(out, "XXXXXXXXX") {
		t.Errorf("Custom output not present in render: %q", out)
	}
}
