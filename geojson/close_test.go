package geojson

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/paulmach/orb"
)

func TestCloseRings_Polygon(t *testing.T) {
	// outer ring open, one hole closed, one hole open.
	input := orb.Polygon{
		{{0, 0}, {10, 0}, {10, 10}, {0, 10}},     // open outer
		{{2, 2}, {4, 2}, {4, 4}, {2, 4}, {2, 2}}, // closed hole
		{{6, 6}, {8, 6}, {8, 8}},                 // open hole
	}
	backup := input.Clone()

	if err := CloseRings(input); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(input[0]) != 5 || len(input[1]) != 5 || len(input[2]) != 4 {
		t.Errorf("unexpected ring lengths: %d, %d, %d",
			len(input[0]), len(input[1]), len(input[2]))
	}

	for i, r := range input {
		if !r.Closed() {
			t.Errorf("ring %d is not closed", i)
		}
	}

	if input[1][len(input[1])-1] != backup[1][len(backup[1])-1] {
		t.Error("already closed ring was modified")
	}

	// original coordinates unchanged and holes kept in order.
	if input[1][0] != backup[1][0] || input[2][0] != backup[2][0] {
		t.Error("hole order changed")
	}
	for ri := range backup {
		for pi := 0; pi < len(backup[ri]); pi++ {
			if input[ri][pi] != backup[ri][pi] {
				t.Errorf("ring %d point %d changed", ri, pi)
			}
		}
	}
}

func TestGeometry_CloseRings_Mixed(t *testing.T) {
	raw := []byte(`{
		"type": "Polygon",
		"coordinates": [
			[[0,0],[10,0],[10,10],[0,10]],
			[[2,2],[4,2],[4,4],[2,4],[2,2]],
			[[6,6],[8,6],[8,8]]
		]
	}`)

	g, err := UnmarshalGeometry(raw)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	polygon := g.Coordinates.(orb.Polygon)
	if err := g.CloseRings(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	closed := g.Coordinates.(orb.Polygon)
	if len(closed[0]) != 5 || len(closed[1]) != 5 || len(closed[2]) != 4 {
		t.Errorf("unexpected ring lengths after CloseRings: %d, %d, %d",
			len(closed[0]), len(closed[1]), len(closed[2]))
	}

	// unmarshalled coordinates were untouched apart from the appended points.
	for ri := range polygon {
		for pi := range polygon[ri] {
			if closed[ri][pi] != polygon[ri][pi] {
				t.Errorf("ring %d point %d modified", ri, pi)
			}
		}
	}

	// marshalled output is valid closed GeoJSON.
	out, err := g.MarshalJSON()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	roundTrip, err := UnmarshalGeometry(out)
	if err != nil {
		t.Fatalf("roundtrip: %v", err)
	}
	rt := roundTrip.Coordinates.(orb.Polygon)
	for i, r := range rt {
		if !r.Closed() {
			t.Errorf("ring %d not closed after roundtrip", i)
		}
	}
}

func TestGeometry_CloseRings_TwoPointHoleErrorIndex(t *testing.T) {
	g := &Geometry{
		Type: "Polygon",
		Coordinates: orb.Polygon{
			{{0, 0}, {10, 0}, {10, 10}, {0, 10}}, // valid open outer
			{{2, 2}, {3, 3}},                     // hole with 2 distinct points
		},
	}
	original := g.Coordinates.(orb.Polygon).Clone()

	err := g.CloseRings()
	if err == nil {
		t.Fatal("expected an error, got nil")
	}

	var ringErr *orb.RingError
	if !errors.As(err, &ringErr) {
		t.Fatalf("expected *orb.RingError, got %T: %v", err, err)
	}
	if !errors.Is(err, orb.ErrRingTooFewPoints) {
		t.Errorf("error should wrap orb.ErrRingTooFewPoints: %v", err)
	}
	if ringErr.PolygonIndex != 0 {
		t.Errorf("polygon index: %d != 0", ringErr.PolygonIndex)
	}
	if ringErr.RingIndex != 1 {
		t.Errorf("ring index: %d != 1 (the hole)", ringErr.RingIndex)
	}

	got := g.Coordinates.(orb.Polygon)
	if !got.Equal(original) {
		t.Errorf("geometry modified on error: got %v want %v", got, original)
	}
}

func TestGeometry_CloseRings_Collection(t *testing.T) {
	g := &Geometry{
		Type: "GeometryCollection",
		Geometries: []*Geometry{
			{
				Type:        "Polygon",
				Coordinates: orb.Polygon{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}},
			},
			{
				Type: "GeometryCollection",
				Geometries: []*Geometry{
					{
						Type:        "Point",
						Coordinates: orb.Point{1, 2},
					},
				},
			},
		},
	}

	if err := g.CloseRings(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	p := g.Geometries[0].Coordinates.(orb.Polygon)
	if !p[0].Closed() || len(p[0]) != 5 {
		t.Errorf("nested polygon not closed: %v", p[0])
	}

	// nested geometry collection structure preserved.
	inner := g.Geometries[1]
	if inner.Type != "GeometryCollection" || len(inner.Geometries) != 1 {
		t.Error("nested geometry collection structure lost")
	}
	if inner.Geometries[0].Coordinates.(orb.Point) != (orb.Point{1, 2}) {
		t.Error("point in nested collection modified")
	}
}

func TestGeometry_CloseRings_CollectionErrorIsAtomic(t *testing.T) {
	g := &Geometry{
		Type: "GeometryCollection",
		Geometries: []*Geometry{
			{
				Type:        "Polygon",
				Coordinates: orb.Polygon{{{0, 0}, {10, 0}, {10, 10}, {0, 10}}},
			},
			{
				Type:        "Polygon",
				Coordinates: orb.Polygon{{{5, 5}, {6, 6}}}, // invalid
			},
		},
	}
	firstBackup := g.Geometries[0].Coordinates.(orb.Polygon).Clone()

	err := g.CloseRings()
	if err == nil {
		t.Fatal("expected an error, got nil")
	}

	var ringErr *orb.RingError
	if !errors.As(err, &ringErr) {
		t.Fatalf("expected *orb.RingError, got %T", err)
	}
	if ringErr.PolygonIndex != 0 || ringErr.RingIndex != 0 {
		t.Errorf("unexpected error location: %+v", ringErr)
	}

	first := g.Geometries[0].Coordinates.(orb.Polygon)
	if !first.Equal(firstBackup) {
		t.Errorf("earlier collection member modified on error: %v", first)
	}
}

// TestCloseRings_TempFileFixture exercises the lossless repair through a
// small GeoJSON file written to the per-test temporary directory. It is
// the end-to-end acceptance fixture: mixed open/closed outer ring and
// holes are repaired, coordinates stay intact and dimensions stay 2d.
func TestCloseRings_TempFileFixture(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "polygon.geojson")

	input := `{"type":"Polygon","coordinates":[
		[[0,0],[10,0],[10,10],[0,10]],
		[[2,2],[4,2],[4,4],[2,4],[2,2]],
		[[6,6],[8,6],[8,8]]
	]}`
	if err := os.WriteFile(path, []byte(input), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	g, err := UnmarshalGeometry(data)
	if err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}

	before := g.Coordinates.(orb.Polygon).Clone()
	if before.Dimensions() != 2 {
		t.Fatalf("dimensions before: %d", before.Dimensions())
	}

	if err := g.CloseRings(); err != nil {
		t.Fatalf("close rings: %v", err)
	}

	after := g.Coordinates.(orb.Polygon)
	if after.Dimensions() != 2 {
		t.Errorf("dimensions after: %d != 2", after.Dimensions())
	}
	if len(after[0]) != 5 || len(after[1]) != 5 || len(after[2]) != 4 {
		t.Fatalf("unexpected ring lengths: %d, %d, %d",
			len(after[0]), len(after[1]), len(after[2]))
	}

	for ri := range before {
		for pi := range before[ri] {
			if after[ri][pi] != before[ri][pi] {
				t.Fatalf("ring %d point %d changed", ri, pi)
			}
		}
		if !after[ri].Closed() {
			t.Fatalf("ring %d not closed", ri)
		}
	}

	repaired, err := g.MarshalJSON()
	if err != nil {
		t.Fatalf("marshal repaired: %v", err)
	}
	if err := os.WriteFile(path, repaired, 0o600); err != nil {
		t.Fatalf("write repaired fixture: %v", err)
	}

	t.Logf("repaired fixture written to %s: %s", path, repaired)
}
