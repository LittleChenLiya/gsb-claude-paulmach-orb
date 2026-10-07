package geojson

import "github.com/paulmach/orb"

// CloseRings performs a lossless closing of the polygon rings in the
// geometry, in-place. For Polygon and MultiPolygon coordinates the
// first point is appended to every non-empty ring that is not already
// closed; already closed rings are left untouched, all point
// dimensions and the ring/hole order are preserved. Geometry
// collections are closed recursively and every other geometry type is
// left unchanged.
//
// If a non-empty ring is defined by fewer than 3 distinct points an
// *orb.RingError locating the offending ring is returned and the
// geometry is not modified.
func CloseRings(g orb.Geometry) error {
	return orb.CloseRings(g)
}

// CloseRings performs a lossless closing of the polygon rings held by
// the GeoJSON geometry, in-place. Polygons, multi-polygons and
// geometry collections, including nested collections, are supported;
// every other geometry type is left unchanged. See the package level
// CloseRings function for the closing semantics. An *orb.RingError is
// returned, without modifying the geometry, when a ring is defined by
// fewer than 3 distinct points.
func (g *Geometry) CloseRings() error {
	if err := validateGeometryRings(g); err != nil {
		return err
	}

	closeGeometryRings(g)
	return nil
}

func validateGeometryRings(g *Geometry) error {
	if g.Coordinates != nil {
		return orb.ValidateRings(g.Coordinates)
	}

	for _, child := range g.Geometries {
		if err := validateGeometryRings(child); err != nil {
			return err
		}
	}

	return nil
}

func closeGeometryRings(g *Geometry) {
	if g.Coordinates != nil {
		// All rings were validated up front, so the error is guaranteed
		// nil. Re-assign because closing can allocate new backing arrays.
		_ = orb.CloseRings(g.Coordinates)
		return
	}

	for _, child := range g.Geometries {
		closeGeometryRings(child)
	}
}
