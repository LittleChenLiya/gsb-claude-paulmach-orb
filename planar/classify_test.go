package planar

import (
	"testing"

	"github.com/paulmach/orb"
)

// squareWithHole is a 4x4 outer square with a 2x2 hole in the middle.
// The outer ring is counter-clockwise and the hole is clockwise.
func squareWithHole() orb.Polygon {
	return orb.Polygon{
		{{0, 0}, {0, 4}, {4, 4}, {4, 0}, {0, 0}},
		{{1, 1}, {1, 3}, {3, 3}, {3, 1}, {1, 1}},
	}
}

var classifyCases = []struct {
	name     string
	point    orb.Point
	expected PointPosition
}{
	{name: "interior of the square", point: orb.Point{0.5, 0.5}, expected: Interior},
	{name: "interior between outer and hole", point: orb.Point{3.5, 2}, expected: Interior},
	{name: "inside the hole", point: orb.Point{2, 2}, expected: Exterior},
	{name: "outside the square", point: orb.Point{5, 2}, expected: Exterior},
	{name: "on the outer edge", point: orb.Point{0, 2}, expected: Boundary},
	{name: "on the hole edge", point: orb.Point{1, 2}, expected: Boundary},
}

func TestPolygonPointPosition(t *testing.T) {
	polygon := squareWithHole()

	// Six points covering all the required cases.
	for _, tc := range classifyCases {
		t.Run(tc.name, func(t *testing.T) {
			result := PolygonPointPosition(polygon, tc.point)
			if result != tc.expected {
				t.Errorf("point %v: got %d, expected %d", tc.point, result, tc.expected)
			}
		})
	}
}

func TestPolygonPointPositionOrientation(t *testing.T) {
	// The classification must not depend on how the rings are oriented.
	// Both rings are reversed relative to the original fixture.
	for _, reverseHole := range []bool{false, true} {
		polygon := squareWithHole()
		polygon[0].Reverse()
		if reverseHole {
			polygon[1].Reverse()
		}

		for _, tc := range classifyCases {
			result := PolygonPointPosition(polygon, tc.point)
			if result != tc.expected {
				t.Errorf("reverseHole=%v point %v: got %d, expected %d",
					reverseHole, tc.point, result, tc.expected)
			}
		}
	}

	// The input polygon must not be modified by the classification.
	input := squareWithHole()
	for _, tc := range classifyCases {
		PolygonPointPosition(input, tc.point)
	}

	if want := squareWithHole(); !input.Equal(want) {
		t.Errorf("polygon was modified during classification")
	}
}
