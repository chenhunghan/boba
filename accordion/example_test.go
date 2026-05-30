package accordion_test

import (
	"fmt"

	"github.com/chenhunghan/boba/accordion"
)

func ExampleAccordion() {
	a := accordion.Accordion{
		Sections: []accordion.Section{
			{Title: "General", Body: "  display options"},
			{Title: "Network", Body: "  bandwidth limits"},
		},
		Expanded: []bool{true, false},
	}
	fmt.Println(a.Render())
	// Output:
	// ▼ General
	//   display options
	// ▶ Network
}
