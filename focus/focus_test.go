package focus

import (
	"reflect"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// area is a stand-in for an app-defined focus enum. focusZero is the
// "nothing focused" value — by convention the iota-zero of the type.
type area int

const (
	focusZero area = iota
	focusA
	focusB
	focusC
)

// Sentinel cmds used in place of tea.EnableMouseAllMotion /
// tea.DisableMouse during tests. tea.Cmd is a func type so values
// aren't directly == comparable; cmdEq compares by pointer identity.
var (
	enableSentinel  tea.Cmd = func() tea.Msg { return "enable" }
	disableSentinel tea.Cmd = func() tea.Msg { return "disable" }
)

// cmdEq reports whether two tea.Cmd values point at the same
// function. nil-vs-nil is true; nil-vs-non-nil is false.
func cmdEq(a, b tea.Cmd) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return reflect.ValueOf(a).Pointer() == reflect.ValueOf(b).Pointer()
}

// cfg is the shared test configuration. focusA needs mouse, focusB
// needs mouse, focusC does NOT need mouse, focusZero does NOT need
// mouse. Cycle ring is [A, B, C]; focusZero is outside the ring.
var cfg = Config[area]{
	KeyJumps: map[string]area{
		"1": focusA, "a": focusA,
		"2": focusB, "b": focusB,
		"3": focusC, "c": focusC,
	},
	CycleOrder:     []area{focusA, focusB, focusC},
	CycleNext:      []string{"right"},
	CyclePrev:      []string{"left"},
	Clear:          []string{"esc"},
	Zero:           focusZero,
	NeedsMouse:     func(f area) bool { return f == focusA || f == focusB },
	OnEnableMouse:  enableSentinel,
	OnDisableMouse: disableSentinel,
}

// TestApplyKey covers the full precedence ladder, the no-op fallback,
// and the capture-cmd transitions for key-driven focus changes.
func TestApplyKey(t *testing.T) {
	cases := []struct {
		name      string
		current   area
		key       string
		wantFocus area
		wantCmd   tea.Cmd
	}{
		// Clear (highest precedence).
		{"esc clears from A", focusA, "esc", focusZero, disableSentinel},
		{"esc clears from C (no capture change)", focusC, "esc", focusZero, nil},
		{"esc on already-zero (no-op cmd)", focusZero, "esc", focusZero, nil},

		// CycleNext.
		{"right cycles A→B (still mouse)", focusA, "right", focusB, nil},
		{"right cycles B→C (capture off)", focusB, "right", focusC, disableSentinel},
		{"right wraps C→A (capture on)", focusC, "right", focusA, enableSentinel},
		{"right from outside ring → first", focusZero, "right", focusA, enableSentinel},

		// CyclePrev.
		{"left cycles B→A", focusB, "left", focusA, nil},
		{"left wraps A→C (capture off)", focusA, "left", focusC, disableSentinel},
		{"left from outside ring → last", focusZero, "left", focusC, nil},

		// KeyJumps.
		{"1 jumps to A from zero (capture on)", focusZero, "1", focusA, enableSentinel},
		{"a jumps to A (alias key)", focusZero, "a", focusA, enableSentinel},
		{"3 jumps to C", focusZero, "3", focusC, nil},
		{"2 from C jumps to B (capture on)", focusC, "2", focusB, enableSentinel},

		// Unbound keys.
		{"x is unbound (focus unchanged, cmd nil)", focusA, "x", focusA, nil},
		{"unbound from zero is unchanged", focusZero, "shift+f9", focusZero, nil},

		// Same-target jump (no capture transition).
		{"jump to currently-focused panel (no cmd)", focusA, "1", focusA, nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := ApplyKey(tc.current, tc.key, cfg)
			if r.Focus != tc.wantFocus {
				t.Errorf("Focus = %v, want %v", r.Focus, tc.wantFocus)
			}
			if !cmdEq(r.Cmd, tc.wantCmd) {
				t.Errorf("Cmd identity mismatch (want %v, got %v)", tc.wantCmd, r.Cmd)
			}
		})
	}
}

// TestApplyClick covers the click → focus transition, including the
// no-cmd-when-target-equals-current case and the mouse-capture
// transitions in both directions.
func TestApplyClick(t *testing.T) {
	cases := []struct {
		name      string
		current   area
		clicked   area
		wantFocus area
		wantCmd   tea.Cmd
	}{
		{"click on focused panel (no cmd)", focusA, focusA, focusA, nil},
		{"click switches A → B (still mouse)", focusA, focusB, focusB, nil},
		{"click switches B → C (capture off)", focusB, focusC, focusC, disableSentinel},
		{"click from zero → A (capture on)", focusZero, focusA, focusA, enableSentinel},
		{"click from C → A (capture on)", focusC, focusA, focusA, enableSentinel},
		{"click on zero from non-zero", focusA, focusZero, focusZero, disableSentinel},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := ApplyClick(tc.current, tc.clicked, cfg)
			if r.Focus != tc.wantFocus {
				t.Errorf("Focus = %v, want %v", r.Focus, tc.wantFocus)
			}
			if !cmdEq(r.Cmd, tc.wantCmd) {
				t.Errorf("Cmd identity mismatch (want %v, got %v)", tc.wantCmd, r.Cmd)
			}
		})
	}
}

// TestInitCmd verifies the startup helper emits the configured
// enable cmd iff the initial focus needs mouse.
func TestInitCmd(t *testing.T) {
	cases := []struct {
		name    string
		initial area
		wantCmd tea.Cmd
	}{
		{"initial needs mouse", focusA, enableSentinel},
		{"initial does not need mouse", focusC, nil},
		{"initial zero (no mouse)", focusZero, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := InitCmd(tc.initial, cfg)
			if !cmdEq(got, tc.wantCmd) {
				t.Errorf("InitCmd identity mismatch (want %v, got %v)", tc.wantCmd, got)
			}
		})
	}
}

// TestPrecedence pins the documented ordering: Clear > CycleNext >
// CyclePrev > KeyJumps. We rig a config where one key appears in
// multiple roles and verify the higher-precedence one wins.
func TestPrecedence(t *testing.T) {
	// Same key "x" registered as Clear, CycleNext, CyclePrev, and
	// a KeyJump. Only Clear should fire.
	collide := Config[area]{
		KeyJumps:   map[string]area{"x": focusB},
		CycleOrder: []area{focusA, focusB, focusC},
		CycleNext:  []string{"x"},
		CyclePrev:  []string{"x"},
		Clear:      []string{"x"},
		Zero:       focusZero,
		NeedsMouse: func(area) bool { return false },
	}
	r := ApplyKey(focusA, "x", collide)
	if r.Focus != focusZero {
		t.Errorf("Clear should win precedence; got %v", r.Focus)
	}

	// With Clear removed, CycleNext should win over CyclePrev and KeyJumps.
	collide.Clear = nil
	r = ApplyKey(focusA, "x", collide)
	if r.Focus != focusB { // step(focusA, [A,B,C], +1) == focusB
		t.Errorf("CycleNext should win after Clear removed; got %v", r.Focus)
	}

	// With CycleNext also removed, CyclePrev wins over KeyJumps.
	collide.CycleNext = nil
	r = ApplyKey(focusA, "x", collide)
	if r.Focus != focusC { // step(focusA, [A,B,C], -1) wraps to focusC
		t.Errorf("CyclePrev should win after Clear+CycleNext removed; got %v", r.Focus)
	}

	// Finally, only KeyJumps remains.
	collide.CyclePrev = nil
	r = ApplyKey(focusA, "x", collide)
	if r.Focus != focusB {
		t.Errorf("KeyJumps should fire when no higher-precedence rule matches; got %v", r.Focus)
	}
}

// TestNilNeedsMouse verifies a nil NeedsMouse config never emits a
// capture cmd — useful for apps that don't use mouse at all.
func TestNilNeedsMouse(t *testing.T) {
	bare := Config[area]{
		KeyJumps:   map[string]area{"1": focusA},
		CycleOrder: []area{focusA, focusB},
		CycleNext:  []string{"right"},
		Clear:      []string{"esc"},
		Zero:       focusZero,
	}
	for _, tc := range []struct {
		name    string
		current area
		key     string
	}{
		{"jump", focusZero, "1"},
		{"cycle", focusA, "right"},
		{"clear", focusA, "esc"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := ApplyKey(tc.current, tc.key, bare)
			if r.Cmd != nil {
				t.Errorf("Cmd must be nil when NeedsMouse is nil, got %v", r.Cmd)
			}
		})
	}
	if got := InitCmd(focusA, bare); got != nil {
		t.Errorf("InitCmd must be nil when NeedsMouse is nil, got %v", got)
	}
}

// TestEmptyCycleOrder verifies cycle keys no-op when CycleOrder is
// empty (rather than panicking on a zero-length slice).
func TestEmptyCycleOrder(t *testing.T) {
	bare := Config[area]{
		CycleNext: []string{"right"},
		CyclePrev: []string{"left"},
		Zero:      focusZero,
	}
	for _, key := range []string{"right", "left"} {
		r := ApplyKey(focusA, key, bare)
		if r.Focus != focusA {
			t.Errorf("empty CycleOrder + %q: focus changed to %v, want %v", key, r.Focus, focusA)
		}
	}
}

// TestDefaultMouseCmds verifies that when OnEnableMouse /
// OnDisableMouse are zero, the package falls back to bubbletea's
// own EnableMouseAllMotion / DisableMouse rather than emitting nil.
func TestDefaultMouseCmds(t *testing.T) {
	bare := Config[area]{
		KeyJumps:   map[string]area{"1": focusA},
		Zero:       focusZero,
		NeedsMouse: func(f area) bool { return f == focusA },
	}
	r := ApplyKey(focusZero, "1", bare)
	// Compare by pointer to tea.EnableMouseAllMotion specifically.
	want := reflect.ValueOf(tea.EnableMouseAllMotion).Pointer()
	if r.Cmd == nil || reflect.ValueOf(r.Cmd).Pointer() != want {
		t.Errorf("default OnEnableMouse should be tea.EnableMouseAllMotion")
	}

	// Reverse direction.
	bare.KeyJumps["x"] = focusZero
	r = ApplyKey(focusA, "x", bare)
	want = reflect.ValueOf(tea.DisableMouse).Pointer()
	if r.Cmd == nil || reflect.ValueOf(r.Cmd).Pointer() != want {
		t.Errorf("default OnDisableMouse should be tea.DisableMouse")
	}
}
