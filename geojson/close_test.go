package geojson

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/paulmach/orb"
)

func TestGeometryCloseRings(t *testing.T) {
	// outer ring open, hole closed, second hole open.
	data := []byte(`{
		"type": "Polygon",
		"coordinates": [
			[[0, 0], [4, 0], [4, 4], [0, 4]],
			[[1, 1], [2, 1], [2, 2], [1, 2], [1, 1]],
			[[3, 3], [3.5, 3], [3.5, 3.5], [3, 3.5]]
		]
	}`)

	g, err := UnmarshalGeometryClosed(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	p, ok := g.Geometry().(orb.Polygon)
	if !ok {
		t.Fatalf("expected orb.Polygon, got %T", g.Geometry())
	}

	expectedLens := []int{5, 5, 5}
	for i, want := range expectedLens {
		if len(p[i]) != want {
			t.Errorf("ring %d: expected length %d, got %d", i, want, len(p[i]))
		}
		if !p[i].Closed() {
			t.Errorf("ring %d: not closed", i)
		}
	}

	// already closed hole must not have a duplicated last point.
	if p[1][3] == p[1][4] && p[1][2] == p[1][3] {
		t.Errorf("closed hole: unexpected duplicated points: %v", p[1])
	}
	if p[1][4] != (orb.Point{1, 1}) {
		t.Errorf("closed hole: last point = %v, expected {1 1}", p[1][4])
	}

	// dimensions and hole order preserved.
	if p.Dimensions() != 2 {
		t.Errorf("dimensions = %d, expected 2", p.Dimensions())
	}
	if p[2][0] != (orb.Point{3, 3}) {
		t.Errorf("hole order changed: second hole starts at %v", p[2][0])
	}
}

func TestGeometryCloseRingsError(t *testing.T) {
	// the hole has only two distinct points.
	data := []byte(`{
		"type": "Polygon",
		"coordinates": [
			[[0, 0], [4, 0], [4, 4], [0, 4]],
			[[1, 1], [2, 2]]
		]
	}`)

	_, err := UnmarshalGeometryClosed(data)
	if err == nil {
		t.Fatal("expected an error for a hole with two distinct points")
	}

	var rce *orb.RingClosureError
	if !errors.As(err, &rce) {
		t.Fatalf("expected *orb.RingClosureError, got %T (%v)", err, err)
	}
	if rce.Ring != 1 {
		t.Errorf("error ring index = %d, expected 1 (the hole)", rce.Ring)
	}
	if rce.Distinct != 2 {
		t.Errorf("error distinct count = %d, expected 2", rce.Distinct)
	}
}

func TestGeometryCloseRingsCollection(t *testing.T) {
	data := []byte(`{
		"type": "GeometryCollection",
		"geometries": [
			{"type": "Point", "coordinates": [1, 2]},
			{"type": "Polygon", "coordinates": [[[0, 0], [1, 0], [1, 1], [0, 1]]]},
			{"type": "MultiPolygon", "coordinates": [[[[2, 2], [3, 2], [3, 3], [2, 3]]]]}
		]
	}`)

	g, err := UnmarshalGeometryClosed(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	c := g.Geometry().(orb.Collection)
	if !c[1].(orb.Polygon)[0].Closed() {
		t.Errorf("polygon in collection not closed")
	}
	if !c[2].(orb.MultiPolygon)[0][0].Closed() {
		t.Errorf("multipolygon in collection not closed")
	}
}

// TestUnmarshalGeometryClosedFromFile demonstrates the repair using a
// small fixture written to a temporary directory.
func TestUnmarshalGeometryClosedFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "unclosed.geojson")

	fixture := `{"type":"Polygon","coordinates":[[[0,0],[4,0],[4,4],[0,4]],[[1,1],[2,1],[2,2],[1,2],[1,1]]]}`
	if err := os.WriteFile(path, []byte(fixture), 0o644); err != nil {
		t.Fatalf("could not write fixture: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("could not read fixture: %v", err)
	}

	g, err := UnmarshalGeometryClosed(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	p := g.Geometry().(orb.Polygon)
	if len(p[0]) != 5 || len(p[1]) != 5 {
		t.Errorf("ring lengths = %d, %d; expected 5, 5", len(p[0]), len(p[1]))
	}
	if p[0][4] != (orb.Point{0, 0}) {
		t.Errorf("outer ring appended point = %v, expected {0 0}", p[0][4])
	}
	if p[1][4] != (orb.Point{1, 1}) {
		t.Errorf("closed hole last point = %v, expected {1 1}", p[1][4])
	}
}
