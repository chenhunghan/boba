package menu_test

import (
	"fmt"

	"github.com/chenhunghan/boba/menu"
)

func ExampleGroup() {
	g := menu.Group[string]{
		Items: []menu.Item[string]{
			{ID: "copy", Label: "Copy"},
			{ID: "paste", Label: "Paste"},
		},
		Open:  true,
		Hover: 0,
	}

	o := g.ApplyKey("down")
	o = o.Group.ApplyKey("enter")

	fmt.Println(o.Confirmed, o.Chosen)
	// Output: true paste
}
