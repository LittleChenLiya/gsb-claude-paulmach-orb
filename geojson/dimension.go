package geojson

import (
	"encoding/json"
	"fmt"
)

// Dimension is the number of ordinates every position in a geometry
// must have. It is used to declare a uniform coordinate dimension
// when building a geometry from JSON.
type Dimension int

// Supported coordinate dimensions.
const (
	Dimension2D Dimension = 2
	Dimension3D Dimension = 3
)

// A DimensionError is returned when a position in the input does not
// have the declared number of ordinates. Path locates the offending
// position within the document, e.g. "coordinates[0][2]" for a polygon
// or "geometries[1].coordinates[3]" inside a geometry collection.
type DimensionError struct {
	Path     string
	Expected Dimension
	Got      int
}

func (e *DimensionError) Error() string {
	return fmt.Sprintf("geojson: dimension mismatch at %s: expected %d ordinates, got %d",
		e.Path, int(e.Expected), e.Got)
}

// UnmarshalGeometryWithDimension decodes the JSON data into a GeoJSON
// geometry like UnmarshalGeometry, but first verifies that every
// position in the document has exactly the declared number of
// ordinates. Nested geometry collections are checked against the same
// rule. If any position does not match, a *DimensionError is returned
// and no geometry is constructed, so a height is never silently
// dropped to make the input fit. The input is never modified and
// regular parsing is unchanged; only the dimension of positions is
// checked, nothing else about the coordinates.
func UnmarshalGeometryWithDimension(data []byte, dim Dimension) (*Geometry, error) {
	if dim != Dimension2D && dim != Dimension3D {
		return nil, fmt.Errorf("geojson: unsupported dimension %d", int(dim))
	}

	var doc any
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}

	if err := checkDimension(doc, dim, ""); err != nil {
		return nil, err
	}

	return UnmarshalGeometry(data)
}

// checkDimension verifies all positions of a single decoded geometry
// object against the declared dimension. path is the location of the
// object within the document, e.g. "geometries[1]." for a member of a
// geometry collection. Anything that is not a plain geometry object is
// left for regular parsing to diagnose.
func checkDimension(node any, dim Dimension, path string) error {
	obj, ok := node.(map[string]any)
	if !ok {
		return nil
	}

	typ, _ := obj["type"].(string)
	if typ == "GeometryCollection" {
		geometries, _ := obj["geometries"].([]any)
		for i, g := range geometries {
			err := checkDimension(g, dim, fmt.Sprintf("%sgeometries[%d].", path, i))
			if err != nil {
				return err
			}
		}
		return nil
	}

	depth, ok := coordinateDepth(typ)
	if !ok {
		return nil // unknown type, regular parsing will report it
	}

	return checkCoordinates(obj["coordinates"], depth, dim, path+"coordinates")
}

// coordinateDepth returns the number of array levels wrapping a single
// position in the coordinates of the given GeoJSON geometry type.
func coordinateDepth(typ string) (int, bool) {
	switch typ {
	case "Point":
		return 0, true
	case "MultiPoint", "LineString":
		return 1, true
	case "MultiLineString", "Polygon":
		return 2, true
	case "MultiPolygon":
		return 3, true
	}
	return 0, false
}

func checkCoordinates(node any, depth int, dim Dimension, path string) error {
	arr, ok := node.([]any)
	if !ok {
		return nil // not an array, regular parsing will report it
	}

	if depth == 0 {
		return checkPosition(arr, dim, path)
	}

	for i, e := range arr {
		err := checkCoordinates(e, depth-1, dim, fmt.Sprintf("%s[%d]", path, i))
		if err != nil {
			return err
		}
	}
	return nil
}

func checkPosition(pos []any, dim Dimension, path string) error {
	if len(pos) == int(dim) {
		return nil
	}

	for _, o := range pos {
		if _, ok := o.(float64); !ok {
			return nil // not a plain position, regular parsing will report it
		}
	}

	return &DimensionError{Path: path, Expected: dim, Got: len(pos)}
}
