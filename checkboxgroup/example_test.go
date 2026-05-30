package checkboxgroup_test

import (
	"fmt"

	"github.com/chenhunghan/boba/checkboxgroup"
)

func ExampleGroup() {
	g := checkboxgroup.Group{
		Options: []string{"Apples", "Bananas", "Cherries"},
		Checked: []bool{true, false, true},
	}
	fmt.Println(g.Render())
	// Output:
	// [x] Apples
	// [ ] Bananas
	// [x] Cherries
}
