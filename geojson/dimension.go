package geojson

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Dimension declares the number of ordinates every position of a
// geometry must have. It is used by UnmarshalGeometryWithDimension
// to validate a geometry before it is built.
type Dimension int

const (
	// Dimension2D requires every position to have exactly 2 ordinates,
	// e.g. longitude and latitude.
	Dimension2D Dimension = 2

	// Dimension3D requires every position to have exactly 3 ordinates,
	// e.g. longitude, latitude and height.
	Dimension3D Dimension = 3
)

// DimensionError is returned when a position in the input does not
// have the number of ordinates required by the declared dimension.
type DimensionError struct {
	Path     string    // path of the offending node, e.g. "geometries[1].coordinates[0]"
	Declared Dimension // the dimension declared for the geometry
	Actual   int       // the number of ordinates the position actually has
}

func (e *DimensionError) Error() string {
	return fmt.Sprintf("geojson: dimension mismatch at %s: declared %dD but position has %d ordinates",
		e.Path, int(e.Declared), e.Actual)
}

// UnmarshalGeometryWithDimension decodes the JSON data into a GeoJSON
// geometry like UnmarshalGeometry, but every position in the input —
// including positions inside nested geometry collections — must have
// exactly the number of ordinates declared by dim.
//
// The whole input is validated before any geometry is constructed: if
// a position does not match the declared dimension, a *DimensionError
// pointing at the offending node is returned and no geometry is built.
// The input data is never modified. In particular a position carrying
// a height (3 ordinates) is rejected when dim is Dimension2D instead
// of being silently truncated to 2D, and a valid 3D input is accepted
// as is with its heights left untouched in the input data.
//
// Only the dimension of positions is checked. No diagnostics about
// empty rings, unclosed rings or coordinate values are added, and the
// behavior of UnmarshalGeometry itself is unchanged.
func UnmarshalGeometryWithDimension(data []byte, dim Dimension) (*Geometry, error) {
	if dim != Dimension2D && dim != Dimension3D {
		return nil, fmt.Errorf("geojson: unsupported dimension %d", int(dim))
	}

	if err := validateGeometryDimension(data, dim, ""); err != nil {
		return nil, err
	}

	return UnmarshalGeometry(data)
}

// dimensionGeometry mirrors the GeoJSON geometry structure just enough
// to walk the coordinate tree without constructing anything.
type dimensionGeometry struct {
	Coordinates json.RawMessage   `json:"coordinates"`
	Geometries  []json.RawMessage `json:"geometries"`
}

func validateGeometryDimension(data json.RawMessage, dim Dimension, path string) error {
	g := dimensionGeometry{}
	if err := json.Unmarshal(data, &g); err != nil {
		return err
	}

	// nested collections follow the same rule.
	for i, sub := range g.Geometries {
		err := validateGeometryDimension(sub, dim, joinPath(path, fmt.Sprintf("geometries[%d]", i)))
		if err != nil {
			return err
		}
	}

	if len(g.Coordinates) > 0 {
		return validatePositionsDimension(g.Coordinates, dim, joinPath(path, "coordinates"))
	}

	return nil
}

func validatePositionsDimension(data json.RawMessage, dim Dimension, path string) error {
	var elems []json.RawMessage
	if err := json.Unmarshal(data, &elems); err != nil {
		return err
	}

	if len(elems) == 0 {
		// An empty array carries no positions. Empty rings and other
		// coordinate diagnostics are intentionally out of scope.
		return nil
	}

	if first := bytes.TrimSpace(elems[0]); len(first) > 0 && first[0] == '[' {
		for i, e := range elems {
			err := validatePositionsDimension(e, dim, fmt.Sprintf("%s[%d]", path, i))
			if err != nil {
				return err
			}
		}
		return nil
	}

	// a leaf array of numbers is a single position
	if len(elems) != int(dim) {
		return &DimensionError{Path: path, Declared: dim, Actual: len(elems)}
	}

	return nil
}

func joinPath(parent, elem string) string {
	if parent == "" {
		return elem
	}
	return parent + "." + elem
}
