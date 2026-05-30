package separator_test

import (
	"fmt"

	"github.com/chenhunghan/boba/separator"
)

func ExampleSeparator() {
	s := separator.Separator{Length: 5}
	fmt.Println(s.Render())
	// Output: ─────
}
