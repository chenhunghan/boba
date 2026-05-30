package selectbox_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/chenhunghan/boba/selectbox"
)

func sample() selectbox.SelectBox {
	return selectbox.SelectBox{
		Options: []string{"Low", "Medium", "High"},
		W:       10,
		H:       1,
	}
}

func key(s string) tea.KeyMsg {
	switch s {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case " ":
		return tea.KeyMsg{Type: tea.KeySpace}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func TestRenderClosedShowsValueAndChevron(t *testing.T) {
	s := sample()
	s.Selected = 1
	out := s.Render()
	if !strings.Contains(out, "Medium") {
		t.Errorf("closed control should show the selected value, got %q", out)
	}
	if !strings.Contains(out, "▾") {
		t.Errorf("closed control should show the default chevron, got %q", out)
	}
}

func TestRenderClosedWidth(t *testing.T) {
	s := sample()
	s.Selected = 0
	if w := lipgloss.Width(s.Render()); w != s.W {
		t.Errorf("closed control width = %d, want W=%d", w, s.W)
	}
}

func TestChevronFallbackOverridable(t *testing.T) {
	s := sample()
	s.Style.Chevron = "v"
	if strings.Contains(s.Render(), "▾") {
		t.Errorf("custom chevron should override the default, got %q", s.Render())
	}
	if !strings.Contains(s.Render(), "v") {
		t.Errorf("custom chevron not rendered, got %q", s.Render())
	}
}

func TestApplyKeyOpensWhenClosed(t *testing.T) {
	for _, k := range []string{"enter", " ", "down"} {
		s := sample()
		s.Selected = 2
		s, _ = s.ApplyKey(k)
		if !s.Open {
			t.Errorf("%q should open a closed select", k)
		}
		if s.Highlight != 2 {
			t.Errorf("%q should highlight the selected index, got %d", k, s.Highlight)
		}
	}
}

func TestApplyKeyNavigatesWhenOpen(t *testing.T) {
	s := sample()
	s.Open = true
	s.Highlight = 0

	s, _ = s.ApplyKey("down")
	if s.Highlight != 1 {
		t.Errorf("down: Highlight=%d, want 1", s.Highlight)
	}
	s, _ = s.ApplyKey("down")
	s, _ = s.ApplyKey("down") // clamp at last
	if s.Highlight != 2 {
		t.Errorf("down clamp: Highlight=%d, want 2", s.Highlight)
	}
	s, _ = s.ApplyKey("up")
	if s.Highlight != 1 {
		t.Errorf("up: Highlight=%d, want 1", s.Highlight)
	}
}

func TestApplyKeyConfirmSelectsAndCloses(t *testing.T) {
	s := sample()
	s.Open = true
	s.Highlight = 2
	s, confirmed := s.ApplyKey("enter")
	if !confirmed {
		t.Fatal("enter on a highlighted option should confirm")
	}
	if s.Selected != 2 {
		t.Errorf("Selected=%d, want 2", s.Selected)
	}
	if s.Open {
		t.Error("Open should be false after confirm")
	}
}

func TestApplyKeyCancelCloses(t *testing.T) {
	s := sample()
	s.Open = true
	s.Highlight = 1
	s.Selected = 0
	s, confirmed := s.ApplyKey("esc")
	if confirmed {
		t.Error("esc should not confirm")
	}
	if s.Open {
		t.Error("esc should close the list")
	}
	if s.Selected != 0 {
		t.Errorf("esc should not change Selected, got %d", s.Selected)
	}
}

func TestUpdateIgnoredWhenNotFocused(t *testing.T) {
	s := sample()
	s2, cmd := s.Update(key("down"))
	if s2.Open || cmd != nil {
		t.Error("an unfocused select should ignore keys")
	}
}

func TestUpdateEmitsChangedOnConfirm(t *testing.T) {
	s := sample()
	s.Focused = true
	s.Open = true
	s.Highlight = 1
	s, cmd := s.Update(key("enter"))
	if cmd == nil {
		t.Fatal("confirm should return a cmd")
	}
	msg, ok := cmd().(selectbox.ChangedMsg)
	if !ok || msg.Selected != 1 {
		t.Fatalf("got %#v, want selectbox.ChangedMsg{Selected: 1}", cmd())
	}
	if s.Open {
		t.Error("Open should be false after confirm")
	}
}

func TestUpdateOpenEmitsNoMessage(t *testing.T) {
	s := sample()
	s.Focused = true
	s, cmd := s.Update(key("down"))
	if !s.Open {
		t.Error("down should open the list")
	}
	if cmd != nil {
		t.Error("opening the list should not emit a message")
	}
}

func TestClickTogglesClosedControl(t *testing.T) {
	s := sample()
	s, cmd := s.ClickAt(0, 0)
	if !s.Open || cmd != nil {
		t.Error("clicking the closed control should open it without a message")
	}
	s, cmd = s.ClickAt(0, 0)
	if s.Open || cmd != nil {
		t.Error("clicking the control again should close it without a message")
	}
}

func TestClickSelectsOption(t *testing.T) {
	s := sample()
	s.Open = true
	s, cmd := s.ClickAt(0, 2) // option index 1 (row y=2 -> idx 1)
	if cmd == nil {
		t.Fatal("clicking an option should emit a cmd")
	}
	msg, ok := cmd().(selectbox.ChangedMsg)
	if !ok || msg.Selected != 1 {
		t.Fatalf("got %#v, want selectbox.ChangedMsg{Selected: 1}", cmd())
	}
	if s.Selected != 1 {
		t.Errorf("Selected=%d, want 1", s.Selected)
	}
	if s.Open {
		t.Error("Open should be false after selecting an option")
	}
}

func TestClickOutsideIsNoop(t *testing.T) {
	s := sample()
	s.Open = true
	if _, cmd := s.ClickAt(0, 99); cmd != nil {
		t.Error("a click below the list should be a no-op")
	}
	if _, cmd := s.ClickAt(-1, 0); cmd != nil {
		t.Error("a click left of the control should be a no-op")
	}
}

func TestHitTest(t *testing.T) {
	s := sample()
	s.Open = true
	cases := []struct {
		name    string
		x, y    int
		wantRow int
		wantOk  bool
	}{
		{"closed control", 0, 0, -1, true},
		{"option 0", 0, 1, 0, true},
		{"option 2", 0, 3, 2, true},
		{"below list", 0, 4, -1, false},
		{"left of control", -1, 0, -1, false},
		{"right of list", s.W, 1, -1, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			row, ok := s.HitTest(tc.x, tc.y)
			if row != tc.wantRow || ok != tc.wantOk {
				t.Errorf("HitTest(%d,%d) = (%d, %v), want (%d, %v)", tc.x, tc.y, row, ok, tc.wantRow, tc.wantOk)
			}
		})
	}
}

func TestHitTestRowsClosedWhenNotOpen(t *testing.T) {
	s := sample() // closed
	if _, ok := s.HitTest(0, 1); ok {
		t.Error("list rows must not be hittable while closed")
	}
}

func TestOpenViewClosedReturnsBackgroundUnchanged(t *testing.T) {
	s := sample()
	bg := "background"
	if got := s.OpenView(bg, 40, 10); got != bg {
		t.Errorf("closed OpenView must return bg unchanged, got %q", got)
	}
}

func TestOpenViewCompositesListRows(t *testing.T) {
	s := sample()
	s.Open = true
	s.X, s.Y = 0, 0
	bg := strings.Join([]string{
		strings.Repeat(".", 20),
		strings.Repeat(".", 20),
		strings.Repeat(".", 20),
		strings.Repeat(".", 20),
	}, "\n")
	out := s.OpenView(bg, 20, 4)
	for _, opt := range s.Options {
		if !strings.Contains(out, opt) {
			t.Errorf("OpenView output missing option %q:\n%s", opt, out)
		}
	}
	if lines := strings.Count(out, "\n"); lines != 3 {
		t.Errorf("OpenView should preserve the 4-line background, got %d newlines", lines)
	}
}

func TestRenderListWidthUniform(t *testing.T) {
	s := selectbox.SelectBox{
		Options: []string{"a", "bbbbbbbb", "cc"},
		Open:    true,
		W:       4,
	}
	// Each list row, overlaid onto a blank background, must be the same
	// width = the widest option (which exceeds W here).
	want := lipgloss.Width("bbbbbbbb")
	bg := strings.Join(make([]string, len(s.Options)+1), "\n")
	out := s.OpenView(bg, 40, len(s.Options)+1)
	for i, row := range strings.Split(out, "\n")[:len(s.Options)] {
		if w := lipgloss.Width(row); w != want {
			t.Errorf("list row %d width = %d, want %d (%q)", i, w, want, row)
		}
	}
}
