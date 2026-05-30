package togglegroup_test

import (
	"fmt"

	"github.com/chenhunghan/boba/togglegroup"
)

func ExampleGroup() {
	g := togglegroup.Group{
		Items:    []string{"Day", "Week", "Month"},
		Selected: 1,
		Gap:      1,
		Hover:    -1, // -1 = nothing under the cursor
	}
	fmt.Println(g.Render())

	// HitTest maps panel-local (x, y) back to an item index; x=9 lands on
	// "Month" in "Day Week Month".
	fmt.Println(g.HitTest(9, 0))
	// Output:
	// Day Week Month
	// 2
}
