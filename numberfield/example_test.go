package numberfield_test

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/chenhunghan/boba/numberfield"
)

func ExampleNumber() {
	n := numberfield.Number{
		Min: 0, Max: 10, Step: 1, Focused: true,
		Style: numberfield.Style{Up: "+", Down: "-"},
	}.SetValue(3)

	// Press up twice: 3 -> 5.
	n, _ = n.Update(tea.KeyMsg{Type: tea.KeyUp})
	n, _ = n.Update(tea.KeyMsg{Type: tea.KeyUp})

	n.Focused = false // blur so the value renders as plain, unstyled text
	fmt.Println(n.View())
	// Output: 5+-
}
