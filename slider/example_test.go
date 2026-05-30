package slider_test

import (
	"fmt"

	"github.com/chenhunghan/boba/slider"
)

func ExampleSlider() {
	s := slider.Slider{Min: 0, Max: 10, Value: 4, Width: 11}
	fmt.Println(s.Render())
	// Output: ────●──────
}
