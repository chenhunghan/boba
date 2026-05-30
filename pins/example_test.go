package pins_test

import (
	"fmt"

	"github.com/chenhunghan/boba/pins"
)

func ExampleList() {
	var l pins.List

	l.Pin("a")
	l.Pin("b")
	l.Pin("a") // already pinned; ignored

	fmt.Println(l.IDs(), l.Len())
	// Output: [a b] 2
}
