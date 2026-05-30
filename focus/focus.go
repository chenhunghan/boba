// Package focus is a pure-function state machine for the "which
// region owns the keyboard" question that every multi-panel TUI app
// has to answer. The package owns the rules; callers store the
// current focus value on their own model and call ApplyKey or
// ApplyClick on each input event.
//
// Focus values are caller-defined (typically an int enum), so
// callers get type-safe focus comparisons in their own code while the
// package operates uniformly via a comparable type parameter F.
//
// The package returns tea.Cmd values directly — it is intended for
// Bubble Tea applications. Mouse-capture commands are emitted only
// when the capture state actually changes (i.e., when
// NeedsMouse(old) != NeedsMouse(new)); otherwise Cmd is nil so the
// caller never re-enables an already-enabled capture.
//
// Typical usage in a Bubble Tea Update:
//
//	case tea.KeyMsg:
//	    r := focus.ApplyKey(m.focus, msg.String(), focusCfg)
//	    m.focus = r.Focus
//	    return m, r.Cmd
//	case tea.MouseMsg:
//	    if click {
//	        r := focus.ApplyClick(m.focus, panelAt(msg.X, msg.Y), focusCfg)
//	        m.focus = r.Focus
//	        return m, r.Cmd
//	    }
//
// And in Init:
//
//	return focus.InitCmd(m.focus, focusCfg)
package focus

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Config defines the rules for one focus state machine. F is the
// caller's focus type (any comparable, typically an int enum).
//
// All slice/map fields may be nil — a nil KeyJumps simply means no
// keys jump to anything; a nil CycleOrder means CycleNext/CyclePrev
// are no-ops; etc. Callers wire only the rules they actually want.
//
// Precedence within ApplyKey is fixed: Clear > CycleNext > CyclePrev
// > KeyJumps. Unbound keys produce a no-op result. Avoid placing the
// same key in two fields; the documented precedence resolves it but
// caller intent is usually clearer if the config is unambiguous.
type Config[F comparable] struct {
	// KeyJumps maps a key string (as returned by tea.KeyMsg.String) to
	// the focus value to jump to. e.g. {"1": focusSidebar, "h": focusSidebar}.
	KeyJumps map[string]F

	// CycleOrder is the ring of focus values traversed by CycleNext /
	// CyclePrev keys. Values not in this ring are skipped — when the
	// current focus is not in the ring, CycleNext lands on the first
	// element and CyclePrev on the last.
	CycleOrder []F

	// CycleNext / CyclePrev are the keys that step CycleOrder forward
	// / backward, e.g. ["right"] / ["left"].
	CycleNext []string
	CyclePrev []string

	// Clear is the set of keys that return focus to Zero (the "no
	// focus" value), e.g. ["esc"].
	Clear []string

	// Zero is the focus value representing "nothing focused". Returned
	// to on Clear-key presses. If the caller's focus type has no such
	// value, set Zero to a never-used sentinel and avoid Clear keys.
	Zero F

	// NeedsMouse reports whether terminal mouse capture should be
	// enabled while focus is f. May be nil — equivalent to "always
	// false" (mouse capture never enabled, OnEnableMouse never fires).
	NeedsMouse func(f F) bool

	// OnEnableMouse is the cmd emitted when transitioning from a
	// focus where NeedsMouse is false to one where it is true. nil
	// means use tea.EnableMouseAllMotion as the default.
	OnEnableMouse tea.Cmd

	// OnDisableMouse is the cmd emitted on the reverse transition.
	// nil means use tea.DisableMouse.
	OnDisableMouse tea.Cmd
}

// Result is the outcome of applying an event to a focus state.
//
// Focus is the (possibly unchanged) focus value the caller should
// store back onto their model.
//
// Cmd is non-nil only when terminal mouse-capture state has changed.
// When the new focus equals the old, or when NeedsMouse(old) ==
// NeedsMouse(new), Cmd is nil — the caller never re-emits capture
// commands needlessly.
type Result[F comparable] struct {
	Focus F
	Cmd   tea.Cmd
}

// ApplyKey processes a key press. Precedence (highest first):
//
//  1. Clear keys → Focus = cfg.Zero
//  2. CycleNext keys → step CycleOrder forward
//  3. CyclePrev keys → step CycleOrder backward
//  4. cfg.KeyJumps[key] → Focus = mapped value
//
// Unbound keys produce Result{Focus: current, Cmd: nil}.
func ApplyKey[F comparable](current F, key string, cfg Config[F]) Result[F] {
	if contains(cfg.Clear, key) {
		return transition(current, cfg.Zero, cfg)
	}
	if contains(cfg.CycleNext, key) {
		return transition(current, step(current, cfg.CycleOrder, +1), cfg)
	}
	if contains(cfg.CyclePrev, key) {
		return transition(current, step(current, cfg.CycleOrder, -1), cfg)
	}
	if next, ok := cfg.KeyJumps[key]; ok {
		return transition(current, next, cfg)
	}
	return Result[F]{Focus: current}
}

// ApplyClick processes a click on panel p. The new focus is p
// regardless of current. Capture cmd is derived from the transition
// the same way ApplyKey derives it.
func ApplyClick[F comparable](current F, p F, cfg Config[F]) Result[F] {
	return transition(current, p, cfg)
}

// InitCmd returns the right enable-mouse cmd if the initial focus
// needs mouse capture, else nil. Use in tea.Model.Init to put the
// terminal in the correct mode at startup.
func InitCmd[F comparable](initial F, cfg Config[F]) tea.Cmd {
	if cfg.NeedsMouse != nil && cfg.NeedsMouse(initial) {
		return enableCmd(cfg)
	}
	return nil
}

// transition computes a Result for current → next, deriving the
// capture cmd from NeedsMouse on either side. The cmd is non-nil
// only when the capture predicate flips.
func transition[F comparable](current, next F, cfg Config[F]) Result[F] {
	r := Result[F]{Focus: next}
	if cfg.NeedsMouse == nil {
		return r
	}
	wasMouse := cfg.NeedsMouse(current)
	willMouse := cfg.NeedsMouse(next)
	switch {
	case !wasMouse && willMouse:
		r.Cmd = enableCmd(cfg)
	case wasMouse && !willMouse:
		r.Cmd = disableCmd(cfg)
	}
	return r
}

// step returns the element of order one position from current in the
// given direction (+1 forward, -1 backward), wrapping at the ends.
// When current is not in order, +1 returns the first element and -1
// returns the last (preserving the convention used elsewhere in
// Example: a focus outside the cycle ring snaps to the appropriate
// end on the first cycle press).
func step[F comparable](current F, order []F, delta int) F {
	if len(order) == 0 {
		return current
	}
	for i, x := range order {
		if x == current {
			n := len(order)
			return order[(i+delta+n)%n]
		}
	}
	if delta > 0 {
		return order[0]
	}
	return order[len(order)-1]
}

// enableCmd / disableCmd resolve the configured cmd, falling back to
// the bubbletea defaults when the caller left them zero.
func enableCmd[F comparable](cfg Config[F]) tea.Cmd {
	if cfg.OnEnableMouse != nil {
		return cfg.OnEnableMouse
	}
	return tea.EnableMouseAllMotion
}

func disableCmd[F comparable](cfg Config[F]) tea.Cmd {
	if cfg.OnDisableMouse != nil {
		return cfg.OnDisableMouse
	}
	return tea.DisableMouse
}

// contains reports whether xs contains v. Linear scan — these slices
// are tiny (a handful of keys at most).
func contains[T comparable](xs []T, v T) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}
