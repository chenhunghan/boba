package toolbar_test

import (
	"fmt"

	"github.com/chenhunghan/boba/toolbar"
)

// Example renders a three-item toolbar with a one-cell gap. With the
// zero-value Style (no colors) each item is just its label flanked by a
// single padding space, so the plain output is deterministic.
func Example() {
	bar := toolbar.Toolbar{
		Items:    []string{"New", "Open", "Save"},
		Selected: 0,
		Hover:    -1,
		Gap:      1,
	}
	fmt.Printf("%q\n", bar.View())
	// Output:
	// " New   Open   Save "
}
