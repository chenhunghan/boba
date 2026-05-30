package input_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/chenhunghan/boba/input"
)

func runes(s string) tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)} }

func TestInsertAtCursor(t *testing.T) {
	m := input.Model{Value: "ac", Cursor: 1, Focused: true}
	m, changed := m.ApplyKey(runes("b"))
	if !changed || m.Value != "abc" {
		t.Fatalf("got %q changed=%v, want %q true", m.Value, changed, "abc")
	}
	if m.Cursor != 2 {
		t.Fatalf("cursor = %d, want 2", m.Cursor)
	}
}

func TestInsertRespectsCharLimit(t *testing.T) {
	m := input.Model{Value: "ab", Cursor: 2, Focused: true, CharLimit: 2}
	m, changed := m.ApplyKey(runes("c"))
	if changed || m.Value != "ab" {
		t.Fatalf("got %q changed=%v, want %q false", m.Value, changed, "ab")
	}
}

func TestInsertTruncatesPasteToCharLimit(t *testing.T) {
	m := input.Model{Value: "a", Cursor: 1, Focused: true, CharLimit: 3}
	m, changed := m.ApplyKey(runes("xyz"))
	if !changed || m.Value != "axy" {
		t.Fatalf("got %q, want %q", m.Value, "axy")
	}
}

func TestBackspace(t *testing.T) {
	m := input.Model{Value: "abc", Cursor: 2, Focused: true}
	m, changed := m.ApplyKey(tea.KeyMsg{Type: tea.KeyBackspace})
	if !changed || m.Value != "ac" || m.Cursor != 1 {
		t.Fatalf("got %q cursor=%d, want %q 1", m.Value, m.Cursor, "ac")
	}
}

func TestBackspaceAtStartIsNoop(t *testing.T) {
	m := input.Model{Value: "abc", Cursor: 0, Focused: true}
	m, changed := m.ApplyKey(tea.KeyMsg{Type: tea.KeyBackspace})
	if changed || m.Value != "abc" {
		t.Fatalf("got %q changed=%v, want unchanged", m.Value, changed)
	}
}

func TestDelete(t *testing.T) {
	m := input.Model{Value: "abc", Cursor: 1, Focused: true}
	m, changed := m.ApplyKey(tea.KeyMsg{Type: tea.KeyDelete})
	if !changed || m.Value != "ac" || m.Cursor != 1 {
		t.Fatalf("got %q cursor=%d, want %q 1", m.Value, m.Cursor, "ac")
	}
}

func TestMovementKeys(t *testing.T) {
	m := input.Model{Value: "abc", Cursor: 1, Focused: true}
	for _, tc := range []struct {
		key  tea.KeyType
		want int
	}{
		{tea.KeyLeft, 0},
		{tea.KeyRight, 2},
		{tea.KeyHome, 0},
		{tea.KeyEnd, 3},
	} {
		got, changed := m.ApplyKey(tea.KeyMsg{Type: tc.key})
		if changed {
			t.Fatalf("%v: movement must not change Value", tc.key)
		}
		if got.Cursor != tc.want {
			t.Fatalf("%v: cursor = %d, want %d", tc.key, got.Cursor, tc.want)
		}
	}
}

func TestMovementClampsAtEdges(t *testing.T) {
	left, _ := input.Model{Value: "ab", Cursor: 0, Focused: true}.ApplyKey(tea.KeyMsg{Type: tea.KeyLeft})
	if left.Cursor != 0 {
		t.Fatalf("left at start: cursor = %d, want 0", left.Cursor)
	}
	right, _ := input.Model{Value: "ab", Cursor: 2, Focused: true}.ApplyKey(tea.KeyMsg{Type: tea.KeyRight})
	if right.Cursor != 2 {
		t.Fatalf("right at end: cursor = %d, want 2", right.Cursor)
	}
}

func TestUnicodeEditsByRune(t *testing.T) {
	m := input.Model{Value: "héllo", Cursor: 5, Focused: true}
	m, _ = m.ApplyKey(tea.KeyMsg{Type: tea.KeyBackspace})
	if m.Value != "héll" || m.Cursor != 4 {
		t.Fatalf("got %q cursor=%d, want %q 4", m.Value, m.Cursor, "héll")
	}
	m, _ = m.ApplyKey(tea.KeyMsg{Type: tea.KeyHome})
	m, _ = m.ApplyKey(tea.KeyMsg{Type: tea.KeyRight})
	m, _ = m.ApplyKey(tea.KeyMsg{Type: tea.KeyDelete})
	if m.Value != "hll" {
		t.Fatalf("got %q, want %q", m.Value, "hll")
	}
}

func TestUpdateNoopWhenUnfocused(t *testing.T) {
	m := input.Model{Value: "a", Cursor: 1}
	got, cmd := m.Update(runes("b"))
	if got.Value != "a" || cmd != nil {
		t.Fatal("unfocused field must ignore keys")
	}
}

func TestUpdateIgnoresNonKeyMsg(t *testing.T) {
	m := input.Model{Value: "a", Cursor: 1, Focused: true}
	if _, cmd := m.Update(struct{}{}); cmd != nil {
		t.Fatal("non-key message must be a no-op")
	}
}

func TestUpdateEmitsChangedMsg(t *testing.T) {
	m := input.Model{Value: "a", Cursor: 1, Focused: true}
	_, cmd := m.Update(runes("b"))
	if cmd == nil {
		t.Fatal("expected a cmd carrying ChangedMsg")
	}
	if msg, ok := cmd().(input.ChangedMsg); !ok || msg.Value != "ab" {
		t.Fatalf("got %#v, want input.ChangedMsg{Value: \"ab\"}", cmd())
	}
}

func TestUpdateNoCmdWhenUnchanged(t *testing.T) {
	m := input.Model{Value: "ab", Cursor: 1, Focused: true}
	if _, cmd := m.Update(tea.KeyMsg{Type: tea.KeyLeft}); cmd != nil {
		t.Fatal("cursor-only movement must not emit ChangedMsg")
	}
}

func TestSetValueTruncatesAndClamps(t *testing.T) {
	m := input.Model{Cursor: 9, CharLimit: 3}.SetValue("abcdef")
	if m.Value != "abc" {
		t.Fatalf("value = %q, want %q", m.Value, "abc")
	}
	if m.Cursor != 3 {
		t.Fatalf("cursor = %d, want 3", m.Cursor)
	}
}

func TestRenderPlaceholderWhenEmptyUnfocused(t *testing.T) {
	m := input.Model{Placeholder: "type…"}
	if got := m.Render(); got != "type…" {
		t.Fatalf("render = %q, want %q", got, "type…")
	}
}

func TestRenderNoPlaceholderWhenFocused(t *testing.T) {
	m := input.Model{Placeholder: "type…", Focused: true}
	// Focused + empty draws the trailing cursor block, not the placeholder.
	if got := lipgloss.Width(m.Render()); got != 1 {
		t.Fatalf("focused empty width = %d, want 1 (cursor block)", got)
	}
}

func TestRenderPlainTextUnfocused(t *testing.T) {
	m := input.Model{Value: "hello"}
	if got := m.Render(); got != "hello" {
		t.Fatalf("render = %q, want %q", got, "hello")
	}
}

func TestRenderWindowsToWidth(t *testing.T) {
	// Cursor at end of a string longer than Width: window shows the tail
	// runes plus the trailing block, totaling exactly Width cells.
	m := input.Model{Value: "abcdef", Cursor: 6, Focused: true, Width: 3}
	if got := lipgloss.Width(m.Render()); got != 3 {
		t.Fatalf("windowed width = %d, want 3", got)
	}
}

func TestRenderWindowKeepsCursorInView(t *testing.T) {
	cur := lipgloss.NewStyle().Reverse(true)
	m := input.Model{Value: "abcdef", Cursor: 0, Focused: true, Width: 3, Style: input.Style{Cursor: cur}}
	// Cursor at index 0 must be the first visible cell ("a"), styled.
	if got := m.Render(); got != cur.Render("a")+"bc" {
		t.Fatalf("render = %q, want cursor a + bc", got)
	}
}

func TestRenderFullWhenWidthZero(t *testing.T) {
	m := input.Model{Value: "abc", Focused: true, Width: 0}
	// All runes plus the trailing cursor block.
	if got := lipgloss.Width(m.Render()); got != 4 {
		t.Fatalf("width = %d, want 4", got)
	}
}

func TestViewMatchesRender(t *testing.T) {
	m := input.Model{Value: "hello"}
	if m.View() != m.Render() {
		t.Fatal("View must alias Render")
	}
}
