package focus_test

import (
	"fmt"

	"github.com/chenhunghan/boba/focus"
)

func ExampleApplyKey() {
	type area int
	const (
		none area = iota
		sidebar
		content
	)

	cfg := focus.Config[area]{
		KeyJumps: map[string]area{"1": sidebar, "2": content},
		Clear:    []string{"esc"},
		Zero:     none,
	}

	r := focus.ApplyKey(none, "2", cfg)
	fmt.Println(r.Focus == content)
	// Output: true
}
