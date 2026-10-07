package orb

import "fmt"

// RingClosureError is returned when a ring cannot be losslessly closed
// because it contains fewer than 3 distinct points. It locates the
// offending ring within the polygon (and multipolygon, if applicable).
type RingClosureError struct {
	// Polygon is the index of the polygon within a multipolygon.
	// It is -1 when the ring belongs to a single polygon.
	Polygon int
	// Ring is the index of the ring within its polygon.
	// The outer ring is 0, holes start at 1.
	Ring int
	// Distinct is the number of distinct points found in the ring.
	Distinct int
}

func (e *RingClosureError) Error() string {
	if e.Polygon >= 0 {
		return fmt.Sprintf("orb: polygon %d ring %d has only %d distinct point(s), at least 3 required to close",
			e.Polygon, e.Ring, e.Distinct)
	}

	return fmt.Sprintf("orb: ring %d has only %d distinct point(s), at least 3 required to close",
		e.Ring, e.Distinct)
}

// Close returns a copy of the polygon with every ring closed.
// Only unclosed non-empty rings get their first point appended;
// already closed rings are kept as is and their last point is NOT
// duplicated. Ring order (outer ring first, then holes) and all
// point dimensions are preserved.
//
// The repair is lossless: the original polygon is never modified.
// If a non-empty ring has fewer than 3 distinct points a
// *RingClosureError locating the ring is returned and the original
// geometry is left untouched.
func (p Polygon) Close() (Polygon, error) {
	// validate everything first so the caller's geometry is never
	// partially processed when an error occurs.
	for i, r := range p {
		if err := validateClosable(r, -1, i); err != nil {
			return nil, err
		}
	}

	closed := make(Polygon, len(p))
	for i, r := range p {
		closed[i] = closeRing(r)
	}

	return closed, nil
}

// Close returns a copy of the multipolygon with every ring of every
// polygon closed. See Polygon.Close for the exact semantics. The
// returned error, if any, locates the offending polygon and ring.
func (mp MultiPolygon) Close() (MultiPolygon, error) {
	for i, p := range mp {
		for j, r := range p {
			if err := validateClosable(r, i, j); err != nil {
				return nil, err
			}
		}
	}

	closed := make(MultiPolygon, len(mp))
	for i, p := range mp {
		cp := make(Polygon, len(p))
		for j, r := range p {
			cp[j] = closeRing(r)
		}
		closed[i] = cp
	}

	return closed, nil
}

// validateClosable returns a *RingClosureError if the ring is non-empty
// and has fewer than 3 distinct points. Empty rings are left alone.
func validateClosable(r Ring, polygonIndex, ringIndex int) error {
	if len(r) == 0 {
		return nil
	}

	distinct := make(map[Point]struct{}, len(r))
	for _, p := range r {
		distinct[p] = struct{}{}
	}

	if len(distinct) < 3 {
		return &RingClosureError{Polygon: polygonIndex, Ring: ringIndex, Distinct: len(distinct)}
	}

	return nil
}

// closeRing appends the first point to unclosed non-empty rings.
// Closed and empty rings are returned unchanged.
func closeRing(r Ring) Ring {
	if len(r) == 0 || r.Closed() {
		return r
	}

	// allocate a fresh slice so the original ring's backing
	// array is never written to.
	nr := make(Ring, len(r)+1)
	copy(nr, r)
	nr[len(r)] = r[0]

	return nr
}
