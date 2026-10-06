package geojson

import (
	"bytes"
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/paulmach/orb"
)

func loadDimensionFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatalf("unable to read fixture: %v", err)
	}
	return data
}

func TestUnmarshalGeometryWithDimension_valid3D(t *testing.T) {
	data := loadDimensionFixture(t, "dimension_valid_3d.geojson")
	original := append([]byte(nil), data...)

	g, err := UnmarshalGeometryWithDimension(data, Dimension3D)
	if err != nil {
		t.Fatalf("valid 3D input should parse: %v", err)
	}

	if !bytes.Equal(data, original) {
		t.Errorf("input data was modified")
	}

	if g.Type != "GeometryCollection" {
		t.Fatalf("expected GeometryCollection, got %v", g.Type)
	}

	// the declared dimension is uniform across nested collections,
	// so no height was silently dropped to make the input fit.
	p, ok := g.Geometries[0].Coordinates.(orb.Point)
	if !ok {
		t.Fatalf("expected orb.Point, got %T", g.Geometries[0].Coordinates)
	}
	if !p.Equal(orb.Point{100, 0}) {
		t.Errorf("incorrect point: %v", p)
	}

	nested := g.Geometries[1]
	if nested.Type != "GeometryCollection" {
		t.Fatalf("expected nested GeometryCollection, got %v", nested.Type)
	}
	ls, ok := nested.Geometries[0].Coordinates.(orb.LineString)
	if !ok {
		t.Fatalf("expected orb.LineString, got %T", nested.Geometries[0].Coordinates)
	}
	expected := orb.LineString{{100, 0}, {101, 1}}
	if !reflect.DeepEqual(ls, expected) {
		t.Errorf("incorrect linestring: %v", ls)
	}

	// the same input declared as 2D must be rejected, the height
	// must not be silently dropped.
	_, err = UnmarshalGeometryWithDimension(data, Dimension2D)
	var derr *DimensionError
	if !errors.As(err, &derr) {
		t.Fatalf("expected *DimensionError, got %T: %v", err, err)
	}
	if derr.Path != "geometries[0].coordinates" {
		t.Errorf("incorrect path: %v", derr.Path)
	}
	if derr.Expected != Dimension2D || derr.Got != 3 {
		t.Errorf("incorrect error: %v", derr)
	}
}

func TestUnmarshalGeometryWithDimension_mixedNestedRejected(t *testing.T) {
	data := loadDimensionFixture(t, "dimension_mixed_nested.geojson")
	original := append([]byte(nil), data...)

	g, err := UnmarshalGeometryWithDimension(data, Dimension3D)
	if err == nil {
		t.Fatal("expected dimension mismatch for mixed 2D/3D input")
	}
	if g != nil {
		t.Errorf("no geometry should be constructed on mismatch, got %v", g)
	}

	var derr *DimensionError
	if !errors.As(err, &derr) {
		t.Fatalf("expected *DimensionError, got %T: %v", err, err)
	}
	if derr.Path != "geometries[1].geometries[0].coordinates[1]" {
		t.Errorf("incorrect path: %v", derr.Path)
	}
	if derr.Expected != Dimension3D || derr.Got != 2 {
		t.Errorf("incorrect error: %v", derr)
	}

	if !bytes.Equal(data, original) {
		t.Errorf("input data was modified")
	}
}

func TestUnmarshalGeometryWithDimension_types(t *testing.T) {
	cases := []struct {
		name string
		data string
		dim  Dimension
		path string // expected DimensionError path, empty means valid
	}{
		{"point 2d", `{"type":"Point","coordinates":[1,2]}`, Dimension2D, ""},
		{"point 3d", `{"type":"Point","coordinates":[1,2,3]}`, Dimension3D, ""},
		{"point height not dropped", `{"type":"Point","coordinates":[1,2,3]}`, Dimension2D, "coordinates"},
		{"point missing height", `{"type":"Point","coordinates":[1,2]}`, Dimension3D, "coordinates"},
		{"multi point", `{"type":"MultiPoint","coordinates":[[1,2],[3,4,5]]}`, Dimension2D, "coordinates[1]"},
		{"linestring", `{"type":"LineString","coordinates":[[1,2],[3,4,5]]}`, Dimension2D, "coordinates[1]"},
		{"linestring 3d", `{"type":"LineString","coordinates":[[1,2,7],[3,4,5]]}`, Dimension3D, ""},
		{"multi linestring", `{"type":"MultiLineString","coordinates":[[[1,2]],[[3,4],[5,6,7]]]}`, Dimension2D, "coordinates[1][1]"},
		{"polygon", `{"type":"Polygon","coordinates":[[[0,0],[1,0],[1,1],[0,0,9]]]}`, Dimension2D, "coordinates[0][3]"},
		{"multi polygon", `{"type":"MultiPolygon","coordinates":[[[[0,0,1],[1,0,1],[0,0]]]]}`, Dimension3D, "coordinates[0][0][2]"},
		{"collection", `{"type":"GeometryCollection","geometries":[{"type":"Point","coordinates":[1,2,3]}]}`, Dimension2D, "geometries[0].coordinates"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := []byte(tc.data)
			original := append([]byte(nil), data...)

			g, err := UnmarshalGeometryWithDimension(data, tc.dim)
			if tc.path == "" {
				if err != nil {
					t.Fatalf("expected valid input: %v", err)
				}
				if g == nil {
					t.Fatal("expected a geometry")
				}
			} else {
				if err == nil {
					t.Fatal("expected dimension mismatch")
				}
				if g != nil {
					t.Errorf("no geometry should be constructed on mismatch, got %v", g)
				}
				var derr *DimensionError
				if !errors.As(err, &derr) {
					t.Fatalf("expected *DimensionError, got %T: %v", err, err)
				}
				if derr.Path != tc.path {
					t.Errorf("incorrect path: expected %v, got %v", tc.path, derr.Path)
				}
			}

			if !bytes.Equal(data, original) {
				t.Errorf("input data was modified")
			}
		})
	}
}

func TestUnmarshalGeometryWithDimension_regularParsingUnchanged(t *testing.T) {
	cases := []string{
		`{"type":"NotAThing","coordinates":[1,2]}`,               // unknown type
		`{"type":"Point","coordinates":["a","b"]}`,               // non numeric position
		`{"type":"Polygon","coordinates":[[]]}`,                  // empty ring
		`{"type":"Polygon","coordinates":[[[0,0],[1,0],[1,1]]]}`, // unclosed ring
		`{"type":"LineString","coordinates":[]}`,                 // no positions
		`{"type":"GeometryCollection","geometries":[]}`,          // empty collection
		`{"type":"Point","coordinates":[1,2],"extra":[[1,2,3]]}`, // foreign members ignored
	}

	for _, c := range cases {
		data := []byte(c)
		g1, err1 := UnmarshalGeometryWithDimension(data, Dimension2D)
		g2, err2 := UnmarshalGeometry(data)

		if (err1 == nil) != (err2 == nil) {
			t.Errorf("%v: dimension entry error %v, regular parse error %v", c, err1, err2)
			continue
		}
		if err1 == nil && !reflect.DeepEqual(g1, g2) {
			t.Errorf("%v: dimension entry %v, regular parse %v", c, g1, g2)
		}
	}

	// unknown types are still reported by regular parsing
	_, err := UnmarshalGeometryWithDimension([]byte(`{"type":"NotAThing"}`), Dimension2D)
	if !errors.Is(err, ErrInvalidGeometry) {
		t.Errorf("expected ErrInvalidGeometry, got %v", err)
	}

	// unsupported declared dimension
	_, err = UnmarshalGeometryWithDimension([]byte(`{"type":"Point","coordinates":[1,2]}`), Dimension(4))
	if err == nil {
		t.Error("expected error for unsupported dimension")
	}
	var derr *DimensionError
	if errors.As(err, &derr) {
		t.Errorf("unsupported dimension should not be a DimensionError: %v", err)
	}
}
