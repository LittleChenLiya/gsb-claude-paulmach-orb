package planar

import (
	"testing"

	"github.com/paulmach/orb"
)

func TestPolygonLocate(t *testing.T) {
	// A square outer ring with a square hole.
	outer := orb.Ring{{0, 0}, {4, 0}, {4, 4}, {0, 4}, {0, 0}}
	hole := orb.Ring{{1, 1}, {3, 1}, {3, 3}, {1, 3}, {1, 1}}
	poly := orb.Polygon{outer, hole}

	// Six points covering inside/outside/boundary for both the outer
	// ring and the hole.
	cases := []struct {
		name  string
		point orb.Point
		want  Location
	}{
		{name: "outer edge", point: orb.Point{2, 0}, want: Boundary},
		{name: "outside outer", point: orb.Point{2, -1}, want: Outside},
		{name: "inside polygon", point: orb.Point{0.5, 0.5}, want: Inside},
		{name: "hole edge", point: orb.Point{2, 1}, want: Boundary},
		{name: "inside hole", point: orb.Point{2, 2}, want: Outside},
		{name: "inside near hole", point: orb.Point{2, 0.5}, want: Inside},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := PolygonLocate(poly, tc.point); got != tc.want {
				t.Errorf("PolygonLocate() = %v, want %v", got, tc.want)
			}
		})
	}

	// Reversing the ring orientation must not change the result.
	reversed := poly.Clone()
	reversed[0].Reverse()
	reversed[1].Reverse()

	for _, tc := range cases {
		t.Run(tc.name+" reversed", func(t *testing.T) {
			if got := PolygonLocate(reversed, tc.point); got != tc.want {
				t.Errorf("PolygonLocate() reversed = %v, want %v", got, tc.want)
			}
		})
	}

	// The input geometry must not be modified.
	if !poly[0].Equal(outer) || !poly[1].Equal(hole) {
		t.Errorf("PolygonLocate modified the input geometry")
	}
}
