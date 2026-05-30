package progress_test

import (
	"fmt"

	"github.com/chenhunghan/boba/progress"
)

func ExampleProgress() {
	p := progress.Progress{
		Value: 3, Max: 10, Width: 10,
		Style: progress.Style{FilledChar: "#", EmptyChar: "."},
	}
	fmt.Println(p.Render())
	// Output: ###.......
}
