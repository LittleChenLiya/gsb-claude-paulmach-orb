package geojson

import "github.com/paulmach/orb"

// CloseRings losslessly closes every unclosed ring of the Polygon and
// MultiPolygon geometries held by this Geometry, including geometries
// nested in a GeometryCollection. Only unclosed non-empty rings get
// their first point appended; already closed rings are left untouched.
// Ring/hole order and all dimensions are preserved.
//
// If a ring has fewer than 3 distinct points an *orb.RingClosureError
// locating the ring is returned and the geometry is left unmodified.
func (g *Geometry) CloseRings() error {
	switch c := g.Coordinates.(type) {
	case orb.Polygon:
		closed, err := c.Close()
		if err != nil {
			return err
		}
		g.Coordinates = closed
	case orb.MultiPolygon:
		closed, err := c.Close()
		if err != nil {
			return err
		}
		g.Coordinates = closed
	}

	for _, child := range g.Geometries {
		if err := child.CloseRings(); err != nil {
			return err
		}
	}

	return nil
}

// UnmarshalGeometryClosed decodes the JSON data into a GeoJSON geometry
// and losslessly closes any unclosed polygon rings in it. It behaves
// like UnmarshalGeometry followed by Geometry.CloseRings.
func UnmarshalGeometryClosed(data []byte) (*Geometry, error) {
	g, err := UnmarshalGeometry(data)
	if err != nil {
		return nil, err
	}

	if err := g.CloseRings(); err != nil {
		return nil, err
	}

	return g, nil
}
