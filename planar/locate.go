package planar

import "github.com/paulmach/orb"

// Location describes where a point lies relative to a polygon.
type Location int

// Possible locations of a point relative to a polygon.
const (
	// Outside means the point is not within the polygon.
	Outside Location = iota
	// Inside means the point is within the polygon but not on its boundary.
	Inside
	// Boundary means the point lies on one of the polygon's rings.
	Boundary
)

// String returns the name of the location.
func (l Location) String() string {
	switch l {
	case Inside:
		return "Inside"
	case Boundary:
		return "Boundary"
	default:
		return "Outside"
	}
}

// PolygonLocate classifies the point's location relative to the polygon as
// Inside, Outside or Boundary. The polygon is expected to be a single outer
// ring with at most one hole, made up of non-self-intersecting rings without
// repeated points. Ring orientation does not affect the result and the input
// geometry is not modified.
func PolygonLocate(p orb.Polygon, point orb.Point) Location {
	// A point on any ring edge (outer or hole) is on the boundary.
	for _, r := range p {
		if _, on := ringContains(r, point); on {
			return Boundary
		}
	}

	inside, _ := ringContains(p[0], point)
	if !inside {
		return Outside
	}

	// A point inside a hole is outside the polygon.
	for i := 1; i < len(p); i++ {
		if inside, _ := ringContains(p[i], point); inside {
			return Outside
		}
	}

	return Inside
}

// ringContains returns whether the point is inside the ring and whether it
// lies on the ring's boundary. Points on the boundary are considered in.
func ringContains(r orb.Ring, point orb.Point) (inside, onBoundary bool) {
	if !r.Bound().Contains(point) {
		return false, false
	}

	c, on := rayIntersect(point, r[0], r[len(r)-1])
	if on {
		return true, true
	}

	for i := 0; i < len(r)-1; i++ {
		inter, on := rayIntersect(point, r[i], r[i+1])
		if on {
			return true, true
		}

		if inter {
			c = !c
		}
	}

	return c, false
}
