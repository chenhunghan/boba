package checkbox_test

import (
	"fmt"

	"github.com/chenhunghan/boba/checkbox"
)

func ExampleCheckbox() {
	c := checkbox.Checkbox{Label: "Accept", Checked: true}
	fmt.Println(c.Render())
	// Output: [x] Accept
}
