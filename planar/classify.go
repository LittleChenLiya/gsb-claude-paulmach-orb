package planar

import (
	"github.com/paulmach/orb"
)

// PointPosition describes where a point is located relative to a polygon.
type PointPosition int

const (
	// Exterior means the point is outside of the polygon,
	// including points that fall inside a hole.
	Exterior PointPosition = iota
	// Boundary means the point lies exactly on an edge of the
	// outer ring or one of the holes.
	Boundary
	// Interior means the point is inside of the outer ring and
	// outside of all the holes.
	Interior
)

// PolygonPointPosition classifies a point as interior, on the boundary or
// exterior relative to the polygon. The first ring of the polygon is treated
// as the outer ring and the remaining rings, up to one hole, as its holes.
//
// The direction (clockwise or counter-clockwise) of the rings does not affect
// the result. Neither the polygon nor the point are modified.
//
// Only simple rings without self-intersections or repeated points are
// supported. Coordinates are compared in a planar/Euclidean context, so no
// geodesic or geographic measurements are performed.
func PolygonPointPosition(p orb.Polygon, point orb.Point) PointPosition {
	if len(p) == 0 || len(p[0]) == 0 {
		return Exterior
	}

	inside, on := ringPointPosition(p[0], point)
	if !inside && !on {
		return Exterior
	}

	if on {
		return Boundary
	}

	// Inside the outer ring: a point inside or on the boundary of a hole
	// is not part of the interior.
	for i := 1; i < len(p); i++ {
		holeInside, holeOn := ringPointPosition(p[i], point)
		if holeOn {
			return Boundary
		}

		if holeInside {
			return Exterior
		}
	}

	return Interior
}

// ringPointPosition reports whether the point is strictly inside the ring
// or exactly on its boundary. Points on the boundary are not reported as
// inside. The ring and the point are not modified.
func ringPointPosition(r orb.Ring, point orb.Point) (inside, onBoundary bool) {
	if !r.Bound().Contains(point) {
		return false, false
	}

	c, on := rayIntersect(point, r[0], r[len(r)-1])
	if on {
		return false, true
	}

	for i := 0; i < len(r)-1; i++ {
		inter, segOn := rayIntersect(point, r[i], r[i+1])
		if segOn {
			return false, true
		}

		if inter {
			c = !c
		}
	}

	return c, false
}
