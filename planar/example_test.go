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

func ExamplePolygonLocate() {
	// A square outer ring with a square hole.
	outer := orb.Ring{{0, 0}, {4, 0}, {4, 4}, {0, 4}, {0, 0}}
	hole := orb.Ring{{1, 1}, {3, 1}, {3, 3}, {1, 3}, {1, 1}}
	poly := orb.Polygon{outer, hole}

	// Six points covering inside/outside/boundary for both the outer
	// ring and the hole.
	points := []struct {
		name  string
		point orb.Point
	}{
		{"outer edge", orb.Point{2, 0}},
		{"outside outer", orb.Point{2, -1}},
		{"inside polygon", orb.Point{0.5, 0.5}},
		{"hole edge", orb.Point{2, 1}},
		{"inside hole", orb.Point{2, 2}},
		{"inside near hole", orb.Point{2, 0.5}},
	}

	for _, tc := range points {
		fmt.Printf("%s: %s\n", tc.name, planar.PolygonLocate(poly, tc.point))
	}

	// Output:
	// outer edge: Boundary
	// outside outer: Outside
	// inside polygon: Inside
	// hole edge: Boundary
	// inside hole: Outside
	// inside near hole: Inside
}
