package orb

import (
	"errors"
	"fmt"
)

// ErrRingTooFewPoints is returned by the ring closing helpers when a
// non-empty ring of a polygon is defined by fewer than 3 distinct
// points. Such a ring cannot describe a valid area and is left
// untouched by the lossless closing operation.
var ErrRingTooFewPoints = errors.New("orb: ring has fewer than 3 distinct points")

// RingError locates a ring that could not be closed. For a Polygon the
// PolygonIndex is always 0 and RingIndex is the index of the ring in
// the polygon (0 is the outer ring, the others are holes). For a
// MultiPolygon PolygonIndex identifies the polygon and RingIndex the
// ring within that polygon.
type RingError struct {
	PolygonIndex int
	RingIndex    int
	Err          error
}

// Error implements the error interface.
func (e *RingError) Error() string {
	return fmt.Sprintf("orb: invalid ring at polygon %d, ring %d: %v",
		e.PolygonIndex, e.RingIndex, e.Err)
}

// Unwrap returns the wrapped error, allowing errors.Is and errors.As
// to find ErrRingTooFewPoints inside a RingError.
func (e *RingError) Unwrap() error {
	return e.Err
}

// Close performs a lossless closing of the rings of the polygon,
// in-place. A copy of the first point is appended to every non-empty
// ring that is not already closed; rings that are already closed are
// not touched, so the last point is never duplicated. The dimensions
// of every point and the order of the outer ring and the holes are
// preserved.
//
// Empty rings are ignored. If any non-empty ring is defined by fewer
// than 3 distinct points a *RingError locating the offending ring is
// returned and the polygon is not modified.
func (p Polygon) Close() error {
	if err := validatePolygonRings(p, 0); err != nil {
		return err
	}

	for i := range p {
		p[i] = closeRing(p[i])
	}

	return nil
}

// Close performs a lossless closing of the rings of every polygon of
// the multi-polygon, in-place. See Polygon.Close for the closing
// semantics. If any ring is defined by fewer than 3 distinct points a
// *RingError locating the offending ring is returned and none of the
// polygons are modified.
func (mp MultiPolygon) Close() error {
	for pi := range mp {
		if err := validatePolygonRings(mp[pi], pi); err != nil {
			return err
		}
	}

	for pi := range mp {
		for ri := range mp[pi] {
			mp[pi][ri] = closeRing(mp[pi][ri])
		}
	}

	return nil
}

// CloseRings performs a lossless closing of the polygon rings of the
// given geometry, in-place. Polygons and multi-polygons are closed
// directly and geometry collections are closed recursively. Every
// other geometry type is left unchanged. A *RingError is returned,
// without modifying the geometry, when a ring is defined by fewer
// than 3 distinct points.
func CloseRings(g Geometry) error {
	if err := ValidateRings(g); err != nil {
		return err
	}

	closeGeometryRings(g)
	return nil
}

// ValidateRings reports whether the polygon rings of the geometry can
// be losslessly closed. Polygons and multi-polygons are checked
// directly and geometry collections are checked recursively. It
// returns a *RingError locating the first non-empty ring that is
// defined by fewer than 3 distinct points; geometry types without
// polygon rings are always valid. The geometry is not modified.
func ValidateRings(g Geometry) error {
	switch g := g.(type) {
	case Polygon:
		return validatePolygonRings(g, 0)
	case MultiPolygon:
		for pi := range g {
			if err := validatePolygonRings(g[pi], pi); err != nil {
				return err
			}
		}
	case Collection:
		for _, c := range g {
			if err := ValidateRings(c); err != nil {
				return err
			}
		}
	}

	return nil
}

func closeGeometryRings(g Geometry) {
	switch g := g.(type) {
	case Polygon:
		for i := range g {
			g[i] = closeRing(g[i])
		}
	case MultiPolygon:
		for pi := range g {
			for ri := range g[pi] {
				g[pi][ri] = closeRing(g[pi][ri])
			}
		}
	case Collection:
		for _, c := range g {
			closeGeometryRings(c)
		}
	}
}

func validatePolygonRings(p Polygon, polygonIndex int) error {
	for ri, r := range p {
		if len(r) == 0 {
			// empty rings are ignored by the closing operation
			continue
		}

		if distinctPoints(r) < 3 {
			return &RingError{
				PolygonIndex: polygonIndex,
				RingIndex:    ri,
				Err:          ErrRingTooFewPoints,
			}
		}
	}

	return nil
}

// closeRing appends a copy of the first point to a non-empty ring that
// is not closed. Already closed, and empty, rings are left unchanged.
// The updated ring is returned because appending can allocate a new
// backing array.
func closeRing(r Ring) Ring {
	if len(r) == 0 {
		return r
	}

	if r[0] == r[len(r)-1] {
		return r
	}

	return append(r, r[0])
}

// distinctPoints counts how many unique points define the ring.
func distinctPoints(r Ring) int {
	seen := make(map[Point]struct{}, len(r))
	for _, p := range r {
		seen[p] = struct{}{}
	}

	return len(seen)
}
