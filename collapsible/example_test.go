package collapsible_test

import (
	"fmt"

	"github.com/chenhunghan/boba/collapsible"
)

func ExampleCollapsible() {
	c := collapsible.Collapsible{Title: "Details", Body: "first\nsecond", Expanded: true}
	fmt.Println(c.Render())
	// Output:
	// v Details
	// first
	// second
}
