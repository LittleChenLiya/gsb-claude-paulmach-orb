package geojson

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/paulmach/orb"
)

func loadDimensionFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatalf("could not read fixture %s: %v", name, err)
	}
	return bytes.TrimSpace(data)
}

func TestUnmarshalGeometryWithDimension_valid3D(t *testing.T) {
	data := loadDimensionFixture(t, "dimension_linestring_3d.geojson")
	original := append([]byte(nil), data...)

	g, err := UnmarshalGeometryWithDimension(data, Dimension3D)
	if err != nil {
		t.Fatalf("valid 3D geometry should build: %v", err)
	}

	ls, ok := g.Coordinates.(orb.LineString)
	if !ok {
		t.Fatalf("expected orb.LineString, got %T", g.Coordinates)
	}
	if len(ls) != 2 {
		t.Fatalf("expected 2 points, got %d", len(ls))
	}
	if ls[0] != (orb.Point{-87.6, 41.8}) || ls[1] != (orb.Point{-87.5, 41.9}) {
		t.Errorf("2D ordinates mismatch: %v", ls)
	}

	// the input must be untouched ...
	if !bytes.Equal(data, original) {
		t.Errorf("input data was modified")
	}

	// ... and the heights must still be there: a valid 3D input keeps
	// its height instead of having it silently dropped.
	var raw struct {
		Coordinates [][]float64 `json:"coordinates"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("could not decode raw coordinates: %v", err)
	}
	heights := []float64{180.5, 195.25}
	for i, pos := range raw.Coordinates {
		if len(pos) != 3 {
			t.Fatalf("position %d lost its height: %v", i, pos)
		}
		if pos[2] != heights[i] {
			t.Errorf("position %d height = %v, expected %v", i, pos[2], heights[i])
		}
	}
}

func TestUnmarshalGeometryWithDimension_rejectsHeightWhen2D(t *testing.T) {
	data := loadDimensionFixture(t, "dimension_linestring_3d.geojson")
	original := append([]byte(nil), data...)

	// a height must not be silently dropped when 2D is declared.
	_, err := UnmarshalGeometryWithDimension(data, Dimension2D)
	if err == nil {
		t.Fatal("expected a dimension error, got nil")
	}

	var derr *DimensionError
	if !errors.As(err, &derr) {
		t.Fatalf("expected *DimensionError, got %T: %v", err, err)
	}
	if derr.Path != "coordinates[0]" {
		t.Errorf("wrong path: %q", derr.Path)
	}
	if derr.Declared != Dimension2D || derr.Actual != 3 {
		t.Errorf("declared/actual = %v/%v, expected 2/3", derr.Declared, derr.Actual)
	}

	if !bytes.Equal(data, original) {
		t.Errorf("input data was modified")
	}

	// normal parsing is unchanged: it still accepts the same input.
	if _, err := UnmarshalGeometry(data); err != nil {
		t.Errorf("UnmarshalGeometry behavior changed: %v", err)
	}
}

func TestUnmarshalGeometryWithDimension_nestedMixedRejected(t *testing.T) {
	data := loadDimensionFixture(t, "dimension_collection_mixed.geojson")
	original := append([]byte(nil), data...)

	// declared 3D: the 2D point in the first geometry is the bad node.
	_, err := UnmarshalGeometryWithDimension(data, Dimension3D)
	var derr *DimensionError
	if !errors.As(err, &derr) {
		t.Fatalf("expected *DimensionError, got %T: %v", err, err)
	}
	if derr.Path != "geometries[0].coordinates" {
		t.Errorf("wrong path: %q", derr.Path)
	}
	if derr.Declared != Dimension3D || derr.Actual != 2 {
		t.Errorf("declared/actual = %v/%v, expected 3/2", derr.Declared, derr.Actual)
	}

	// declared 2D: the 3D point inside the nested collection is the bad node.
	_, err = UnmarshalGeometryWithDimension(data, Dimension2D)
	if !errors.As(err, &derr) {
		t.Fatalf("expected *DimensionError, got %T: %v", err, err)
	}
	if derr.Path != "geometries[1].geometries[0].coordinates" {
		t.Errorf("wrong path: %q", derr.Path)
	}
	if derr.Declared != Dimension2D || derr.Actual != 3 {
		t.Errorf("declared/actual = %v/%v, expected 2/3", derr.Declared, derr.Actual)
	}

	// rejected input must not be modified.
	if !bytes.Equal(data, original) {
		t.Errorf("input data was modified")
	}
}

func TestUnmarshalGeometryWithDimension_valid2D(t *testing.T) {
	data := []byte(`{"type":"Polygon","coordinates":[[[0,0],[1,0],[1,1],[0,0]]]}`)

	g, err := UnmarshalGeometryWithDimension(data, Dimension2D)
	if err != nil {
		t.Fatalf("valid 2D geometry should build: %v", err)
	}
	if _, ok := g.Coordinates.(orb.Polygon); !ok {
		t.Fatalf("expected orb.Polygon, got %T", g.Coordinates)
	}
}

func TestUnmarshalGeometryWithDimension_unsupportedDimension(t *testing.T) {
	data := []byte(`{"type":"Point","coordinates":[1,2]}`)
	if _, err := UnmarshalGeometryWithDimension(data, Dimension(4)); err == nil {
		t.Fatal("expected an error for an unsupported dimension")
	}
}

func TestUnmarshalGeometryWithDimension_noExtraDiagnostics(t *testing.T) {
	// empty rings, unclosed rings and coordinate values are not
	// diagnosed by the dimension entry.
	cases := []string{
		`{"type":"Polygon","coordinates":[[]]}`,                  // empty ring
		`{"type":"LineString","coordinates":[[0,0],[999,0]]}`,    // out of range coordinate
		`{"type":"Polygon","coordinates":[[[0,0],[1,0],[1,1]]]}`, // unclosed ring
	}

	for _, c := range cases {
		if _, err := UnmarshalGeometryWithDimension([]byte(c), Dimension2D); err != nil {
			t.Errorf("%s: unexpected diagnostic: %v", c, err)
		}
	}
}
