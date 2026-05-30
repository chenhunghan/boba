package input_test

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/chenhunghan/boba/input"
)

func ExampleModel() {
	m := input.Model{Placeholder: "name", Focused: true}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Ada")})
	m.Focused = false // blur so Render emits plain, unstyled text
	fmt.Println(m.View())
	// Output: Ada
}
