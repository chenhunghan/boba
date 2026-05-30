package tab

import (
	"reflect"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// id is a stand-in for an app-defined tab key.
type id int

const (
	none id = iota
	a
	b
	c
)

// testStyle is a minimal Style — concrete colors don't matter for
// dimension and routing tests. Each tab in tests gets this style.
var testStyle = Style{
	Inactive:    lipgloss.NewStyle(),
	Hover:       lipgloss.NewStyle(),
	Active:      lipgloss.NewStyle(),
	Border:      lipgloss.NewStyle(),
	SelectedBar: lipgloss.NewStyle(),
}

// recModel is a tea.Model that records every Update call for test
// inspection. *recModel satisfies tea.Model; tests hold a pointer
// outside the Group so they can read the recorded calls after
// routing messages through the Group.
type recModel struct {
	initCmd     tea.Cmd
	view        string
	updateCount int
	lastMsg     tea.Msg
}

func (m *recModel) Init() tea.Cmd { return m.initCmd }
func (m *recModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.updateCount++
	m.lastMsg = msg
	return m, nil
}
func (m *recModel) View() string { return m.view }

// TestRenderHeaderDimensions verifies the header is exactly two
// rows of width visible cells, regardless of how many tabs fit.
func TestRenderHeaderDimensions(t *testing.T) {
	g := Group[id]{
		Tabs: []Tab[id]{
			{ID: a, Label: "A", Style: testStyle},
			{ID: b, Label: "BB", Style: testStyle, Closable: true},
		},
		Selected: a,
		Gap:      1,
	}
	for _, width := range []int{20, 50, 100} {
		out := g.RenderHeader(width, RenderState[id]{})
		rows := strings.Split(out, "\n")
		if len(rows) != 2 {
			t.Fatalf("width=%d: got %d rows, want 2", width, len(rows))
		}
		for i, r := range rows {
			if w := lipgloss.Width(r); w != width {
				t.Errorf("width=%d row %d: got %d cells, want %d", width, i, w, width)
			}
		}
	}
}

// TestHitTest covers the main routing cases: tab body, close glyph,
// gap between tabs, and out-of-bounds positions.
func TestHitTest(t *testing.T) {
	// Tab "A": no close, label 1 char → width 5
	//   layout: │ A │  (positions 0..4)
	// Tab "BB": closable, label 2 chars → width 8
	//   layout: │ B B   ×   │  (positions 0..7), close at 5
	g := Group[id]{
		Tabs: []Tab[id]{
			{ID: a, Label: "A", Style: testStyle},
			{ID: b, Label: "BB", Closable: true, Style: testStyle},
		},
		Gap: 1,
	}
	cases := []struct {
		name string
		x    int
		want Hit[id]
	}{
		{"before first tab", -1, Hit[id]{}},
		{"A left side", 0, Hit[id]{Found: true, ID: a}},
		{"A label", 2, Hit[id]{Found: true, ID: a}},
		{"A right side", 4, Hit[id]{Found: true, ID: a}},
		{"gap between A and B", 5, Hit[id]{}},
		{"B left side", 6, Hit[id]{Found: true, ID: b}},
		{"B label first char", 8, Hit[id]{Found: true, ID: b}},
		{"B close glyph", 11, Hit[id]{Found: true, ID: b, Close: true}},
		{"B right padding", 12, Hit[id]{Found: true, ID: b}},
		{"B right side", 13, Hit[id]{Found: true, ID: b}},
		{"past last tab", 14, Hit[id]{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// y=0 lands on the title row of the header; HitTest gates
			// y to [0, HeaderHeight) internally.
			got := g.HitTest(tc.x, 0)
			if got != tc.want {
				t.Errorf("HitTest(%d, 0) = %+v, want %+v", tc.x, got, tc.want)
			}
		})
	}
}

// TestApplyKeyDefaultBindings verifies "[" and "]" cycle through
// tabs. With Wrap=true, the cycle wraps at edges. ApplyKey itself
// does not gate on focus — callers do.
func TestApplyKeyDefaultBindings(t *testing.T) {
	mk := func(sel id, wrap bool) Group[id] {
		return Group[id]{
			Tabs: []Tab[id]{
				{ID: a, Label: "A", Style: testStyle},
				{ID: b, Label: "B", Style: testStyle},
				{ID: c, Label: "C", Style: testStyle},
			},
			Selected: sel,
			Wrap:     wrap,
		}
	}
	cases := []struct {
		name string
		from id
		key  string
		wrap bool
		want id
	}{
		{"] from A", a, "]", true, b},
		{"] from B", b, "]", true, c},
		{"] wraps C → A", c, "]", true, a},
		{"] clamps C → C without wrap", c, "]", false, c},
		{"[ from B", b, "[", true, a},
		{"[ wraps A → C", a, "[", true, c},
		{"[ clamps A → A without wrap", a, "[", false, a},
		{"unbound key no-op", a, "x", true, a},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := mk(tc.from, tc.wrap)
			got := g.ApplyKey(tc.key).Selected
			if got != tc.want {
				t.Errorf("from=%v key=%q wrap=%v: Selected=%v, want %v",
					tc.from, tc.key, tc.wrap, got, tc.want)
			}
		})
	}
}

// TestApplyKeyNoTabs verifies ApplyKey is a no-op when there are no
// tabs to cycle through. (Focus gating is the caller's job — there's
// no "Active" field on the Group anymore.)
func TestApplyKeyNoTabs(t *testing.T) {
	g := Group[id]{Tabs: nil}
	if got := g.ApplyKey("]"); !reflect.DeepEqual(got, g) {
		t.Errorf("empty group ApplyKey should be no-op")
	}
}

// TestIsBound verifies the helper that lets callers decide whether
// to consume a key with ApplyKey or forward it to UpdateActive.
func TestIsBound(t *testing.T) {
	g := Group[id]{}
	if !g.IsBound("[") {
		t.Errorf("[ should be bound by default")
	}
	if !g.IsBound("]") {
		t.Errorf("] should be bound by default")
	}
	if g.IsBound("x") {
		t.Errorf("x should not be bound")
	}
	// Custom bindings.
	g.Next = []string{"tab"}
	g.Prev = []string{"shift+tab"}
	if g.IsBound("[") || g.IsBound("]") {
		t.Errorf("default keys should not be bound when custom Next/Prev set")
	}
	if !g.IsBound("tab") || !g.IsBound("shift+tab") {
		t.Errorf("custom Next/Prev keys should be bound")
	}
}

// TestAddTabFiresInit verifies AddTab returns the cmd from the new
// tab's Model.Init() so initial side effects run when the tab opens.
func TestAddTabFiresInit(t *testing.T) {
	sentinel := tea.Cmd(func() tea.Msg { return "init-fired" })
	model := &recModel{initCmd: sentinel}
	g := Group[id]{}
	g, cmd := g.AddTab(Tab[id]{ID: a, Label: "A", Model: model})

	if cmd == nil {
		t.Fatalf("AddTab should return Init's cmd, got nil")
	}
	if reflect.ValueOf(cmd).Pointer() != reflect.ValueOf(sentinel).Pointer() {
		t.Errorf("returned cmd is not the sentinel from Init")
	}
	if g.Selected != a {
		t.Errorf("AddTab should select the new tab; Selected=%v", g.Selected)
	}
	if len(g.Tabs) != 1 {
		t.Errorf("expected 1 tab after AddTab, got %d", len(g.Tabs))
	}
}

// TestAddTabDuplicateIsNoOp verifies that adding a tab with an
// existing ID doesn't duplicate or re-Init.
func TestAddTabDuplicateIsNoOp(t *testing.T) {
	g := Group[id]{}
	g, _ = g.AddTab(Tab[id]{ID: a, Label: "A", Model: Static("x")})
	g, cmd := g.AddTab(Tab[id]{ID: a, Label: "A again", Model: Static("y")})
	if cmd != nil {
		t.Errorf("duplicate AddTab should return nil cmd")
	}
	if len(g.Tabs) != 1 {
		t.Errorf("expected 1 tab after duplicate AddTab, got %d", len(g.Tabs))
	}
	if g.Tabs[0].Label != "A" {
		t.Errorf("original tab should be preserved; got Label=%q", g.Tabs[0].Label)
	}
}

// TestCloseTabReselectsNeighbor verifies that closing the active
// tab reselects an adjacent tab — the next tab when there is one
// to the right, or the previous tab when closing the rightmost.
func TestCloseTabReselectsNeighbor(t *testing.T) {
	mk := func(sel id) Group[id] {
		return Group[id]{
			Tabs: []Tab[id]{
				{ID: a, Label: "A"},
				{ID: b, Label: "B"},
				{ID: c, Label: "C"},
			},
			Selected: sel,
		}
	}

	// Close middle (B), with B selected → C (next slides in).
	g := mk(b).CloseTab(b)
	if g.Selected != c {
		t.Errorf("closing middle B (selected): Selected=%v, want %v", g.Selected, c)
	}
	if len(g.Tabs) != 2 {
		t.Errorf("expected 2 tabs after close, got %d", len(g.Tabs))
	}

	// Close last (C), with C selected → B (previous).
	g = mk(c).CloseTab(c)
	if g.Selected != b {
		t.Errorf("closing last C (selected): Selected=%v, want %v", g.Selected, b)
	}

	// Close non-selected tab — selection unchanged.
	g = mk(a).CloseTab(c)
	if g.Selected != a {
		t.Errorf("closing non-selected tab should not change Selected; got %v", g.Selected)
	}
	if len(g.Tabs) != 2 {
		t.Errorf("expected 2 tabs after non-selected close, got %d", len(g.Tabs))
	}
}

// TestCloseTabEmptiesGroup verifies the zero-tab edge case: when
// the last tab is closed, Selected is reset to the zero ID.
func TestCloseTabEmptiesGroup(t *testing.T) {
	g := Group[id]{
		Tabs:     []Tab[id]{{ID: a}},
		Selected: a,
	}
	g = g.CloseTab(a)
	if len(g.Tabs) != 0 {
		t.Errorf("expected empty Tabs, got %d", len(g.Tabs))
	}
	if g.Selected != none {
		t.Errorf("Selected should reset to zero ID; got %v", g.Selected)
	}
}

// TestUpdateActiveOnlyForwardsToSelected verifies that UpdateActive
// only invokes the active tab's Model.Update.
func TestUpdateActiveOnlyForwardsToSelected(t *testing.T) {
	mA := &recModel{}
	mB := &recModel{}
	g := Group[id]{
		Tabs: []Tab[id]{
			{ID: a, Model: mA},
			{ID: b, Model: mB},
		},
		Selected: a,
	}
	g, _ = g.UpdateActive("hello")
	if mA.updateCount != 1 {
		t.Errorf("active tab's Update should fire once; got %d", mA.updateCount)
	}
	if mB.updateCount != 0 {
		t.Errorf("inactive tab's Update should not fire; got %d", mB.updateCount)
	}
	if mA.lastMsg != "hello" {
		t.Errorf("active model received wrong msg: %v", mA.lastMsg)
	}
}

// TestUpdateAllForwardsToAll verifies UpdateAll routes to every tab.
func TestUpdateAllForwardsToAll(t *testing.T) {
	mA := &recModel{}
	mB := &recModel{}
	g := Group[id]{
		Tabs: []Tab[id]{
			{ID: a, Model: mA},
			{ID: b, Model: mB},
		},
		Selected: a,
	}
	g, _ = g.UpdateAll(SizeMsg{Width: 80, Height: 24})
	if mA.updateCount != 1 || mB.updateCount != 1 {
		t.Errorf("all tabs should receive UpdateAll msg; counts: %d, %d",
			mA.updateCount, mB.updateCount)
	}
	if mA.lastMsg != (SizeMsg{Width: 80, Height: 24}) {
		t.Errorf("model A got wrong msg: %v", mA.lastMsg)
	}
}

// TestFind verifies the lookup helper used by callers to decide
// whether to AddTab a new one or ApplyClick an existing one.
func TestFind(t *testing.T) {
	g := Group[id]{
		Tabs: []Tab[id]{
			{ID: a, Label: "A"},
			{ID: b, Label: "B"},
		},
	}
	if got, ok := g.Find(b); !ok || got.Label != "B" {
		t.Errorf("Find(b) = %+v, %v; want existing tab", got, ok)
	}
	if _, ok := g.Find(c); ok {
		t.Errorf("Find(c) should return false for nonexistent ID")
	}
}

// TestStatic verifies the no-op tea.Model factory: View returns the
// string, Init returns nil, Update returns the same model + nil.
func TestStatic(t *testing.T) {
	m := Static("hello world")
	if m.View() != "hello world" {
		t.Errorf("Static.View = %q, want %q", m.View(), "hello world")
	}
	if m.Init() != nil {
		t.Errorf("Static.Init should return nil")
	}
	m2, cmd := m.Update("anything")
	if cmd != nil {
		t.Errorf("Static.Update should return nil cmd")
	}
	if m2.View() != m.View() {
		t.Errorf("Static.Update should return an equivalent model")
	}
}

// TestRenderContentEmpty verifies rendering with no Selected or no
// Model returns empty without panicking.
func TestRenderContentEmpty(t *testing.T) {
	cases := []struct {
		name string
		g    Group[id]
	}{
		{"no tabs", Group[id]{}},
		{"selected tab has nil Model", Group[id]{
			Tabs:     []Tab[id]{{ID: a, Model: nil}},
			Selected: a,
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.g.RenderContent(); got != "" {
				t.Errorf("expected empty content, got %q", got)
			}
		})
	}
}
