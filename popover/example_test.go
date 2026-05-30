package popover_test

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/chenhunghan/boba/popover"
	"github.com/chenhunghan/boba/popup"
)

func ExamplePopover() {
	// A two-line body becomes a bordered box: widest row + side borders +
	// one cell of padding per side, plus top and bottom edges.
	p := popover.Popover{Content: "save?\n[y/n]", Open: true}
	out := p.Render()
	fmt.Println(lipgloss.Width(strings.SplitN(out, "\n", 2)[0]))
	fmt.Println(strings.Count(out, "\n") + 1)
	// Output:
	// 9
	// 4
}

func ExamplePopover_cancel() {
	p := popover.Popover{Content: "help", Open: true}
	p, cmd := p.Update(tea.KeyMsg{Type: tea.KeyEsc})
	_, closed := cmd().(popover.ClosedMsg)
	fmt.Println(p.Open, closed)
	// Output: false true
}

func ExamplePopover_Over() {
	// Placed Below a 4x1 anchor at (2,2): the panel's top-left lands at the
	// anchor's bottom-left, so its first row replaces background row 3.
	bg := strings.Repeat(".", 12) + "\n"
	bg = strings.Repeat(bg, 8)
	bg = strings.TrimSuffix(bg, "\n")

	p := popover.Popover{Content: "x", Open: true, X: 2, Y: 2, W: 4, H: 1, Placement: popup.Below}
	_, y := popup.Place(p.X, p.Y, p.W, p.H, p.Width(), p.Height(), 12, 8, popup.Below)

	out := p.Over(bg, 12, 8)
	rows := strings.Split(out, "\n")
	fmt.Println(y)
	fmt.Println(rows[1] == strings.Repeat(".", 12)) // untouched bg row
	fmt.Println(rows[y] == strings.Repeat(".", 12)) // panel row, changed
	// Output:
	// 3
	// true
	// false
}
