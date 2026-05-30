package panel_test

import (
	"fmt"

	"github.com/chenhunghan/boba/panel"
)

func ExampleHitTest() {
	type area int
	const (
		sidebar area = iota
		main
	)

	root := panel.Split[area]{
		Axis: panel.Horizontal,
		Children: []panel.Node[area]{
			panel.Panel[area]{ID: sidebar, Size: 5},
			panel.Panel[area]{ID: main, Size: 0},
		},
	}

	hit := panel.HitTest[area](root, 30, 10, 12, 3)
	fmt.Println(hit.Found, hit.Panel == main, hit.LocalX, hit.LocalY)
	// Output: true true 7 3
}
