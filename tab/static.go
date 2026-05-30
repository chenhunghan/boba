package tab

import tea "github.com/charmbracelet/bubbletea"

// staticModel is a no-op tea.Model that always renders a fixed
// string. Used by Static for trivial tab content.
type staticModel string

func (m staticModel) Init() tea.Cmd                       { return nil }
func (m staticModel) Update(tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (m staticModel) View() string                        { return string(m) }

// Static returns a tea.Model that always renders s and ignores
// all messages. Useful for tabs whose content is fixed text and
// doesn't need its own state — pass it as the Tab.Model field
// instead of writing a one-off no-op model.
func Static(s string) tea.Model {
	return staticModel(s)
}
