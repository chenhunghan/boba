package statusbar_test

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/chenhunghan/boba/statusbar"
)

func ExampleBar() {
	bar := statusbar.Bar{
		Left:  []statusbar.Item{{Key: "1", Letter: "h", Text: "hot"}},
		Right: []statusbar.Item{{Key: "esc", Text: "back"}},
	}

	out := bar.Render(40)
	fmt.Println(lipgloss.Width(out))
	// Output: 40
}
