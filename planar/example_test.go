package planar_test

import (
	"fmt"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/planar"
)

func ExampleArea() {
	// +
	// |\
	// | \
	// |  \
	// +---+

	r := orb.Ring{{0, 0}, {3, 0}, {0, 4}, {0, 0}}
	a := planar.Area(r)

	fmt.Println(a)
	// Output:
	// 6
}

func ExampleDistance() {
	d := planar.Distance(orb.Point{0, 0}, orb.Point{3, 4})

	fmt.Println(d)
	// Output:
	// 5
}

func ExamplePolygonPointPosition() {
	// square outer ring with a square hole in the middle
	p := orb.Polygon{
		{{0, 0}, {0, 4}, {4, 4}, {4, 0}, {0, 0}},
		{{1, 1}, {1, 3}, {3, 3}, {3, 1}, {1, 1}},
	}

	for _, point := range []orb.Point{
		{0.5, 0.5}, // interior
		{2, 2},     // inside the hole, so exterior
		{0, 2},     // on the outer edge
	} {
		fmt.Println(point, planar.PolygonPointPosition(p, point))
	}

	// reversing the rings does not change the result
	p[0].Reverse()
	p[1].Reverse()
	fmt.Println(planar.PolygonPointPosition(p, orb.Point{2, 2}))
	// Output:
	// [0.5 0.5] 2
	// [2 2] 0
	// [0 2] 1
	// 0
}

func ExampleLength() {
	// +
	// |\
	// | \
	// |  \
	// +---+

	r := orb.Ring{{0, 0}, {3, 0}, {0, 4}, {0, 0}}
	l := planar.Length(r)

	fmt.Println(l)
	// Output:
	// 12
}
