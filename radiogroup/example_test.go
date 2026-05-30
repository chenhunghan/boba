package radiogroup_test

import (
	"fmt"

	"github.com/chenhunghan/boba/radiogroup"
)

func ExampleRadioGroup() {
	r := radiogroup.RadioGroup{
		Options:  []string{"Low", "Medium", "High"},
		Selected: 1,
	}
	fmt.Println(r.Render())
	// Output:
	// ( ) Low
	// (*) Medium
	// ( ) High
}
