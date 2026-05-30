package tooltip_test

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/chenhunghan/boba/popup"
	"github.com/chenhunghan/boba/tooltip"
)

func ExampleTooltip_Render() {
	t := tooltip.Tooltip{
		Content: "hi",
		Visible: true,
		Style:   tooltip.Style{Surface: lipgloss.NewStyle()},
	}
	out := strings.ReplaceAll(t.Render(), "\x1b[0m", "")
	fmt.Println(out)
	// Output:
	// ┌──┐
	// │hi│
	// └──┘
}

func ExampleTooltip_Over() {
	// A 6x6 dotted screen; a tooltip anchored at (1,1) placed below it.
	bg := strings.TrimRight(strings.Repeat(strings.Repeat(".", 6)+"\n", 6), "\n")
	t := tooltip.Tooltip{
		Content: "x",
		Visible: true,
		X:       1, Y: 1, W: 1, H: 1,
		Placement: popup.Below,
		Style:     tooltip.Style{Surface: lipgloss.NewStyle()},
	}
	out := t.Over(bg, 6, 6)
	fmt.Println(strings.Count(out, "\n") + 1) // line count unchanged
	// Output: 6
}
