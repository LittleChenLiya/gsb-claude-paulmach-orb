package orb

import (
	"errors"
	"testing"
)

func TestPolygon_Close(t *testing.T) {
	// outer ring is open, first hole is already closed, second hole is open.
	original := Polygon{
		{{0, 0}, {10, 0}, {10, 10}, {0, 10}},     // open outer
		{{2, 2}, {3, 2}, {3, 3}, {2, 3}, {2, 2}}, // already closed hole
		{{6, 6}, {8, 6}, {8, 8}},                 // open hole
	}

	backup := original.Clone()

	if err := original.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// lengths: open rings get exactly one appended point, closed ring unchanged.
	expectedLengths := []int{5, 5, 4}
	for i, r := range original {
		if len(r) != expectedLengths[i] {
			t.Errorf("ring %d: length %d != %d", i, len(r), expectedLengths[i])
		}
		if !r.Closed() {
			t.Errorf("ring %d not closed after Close", i)
		}
	}

	// the already closed ring must not have its last point duplicated.
	if len(original[1]) != len(backup[1]) {
		t.Errorf("closed ring length changed: %d != %d", len(original[1]), len(backup[1]))
	}

	// the appended point is a copy of the first point.
	if original[0][4] != backup[0][0] {
		t.Errorf("outer: appended point %v != first point %v", original[0][4], backup[0][0])
	}
	if original[2][3] != backup[2][0] {
		t.Errorf("hole: appended point %v != first point %v", original[2][3], backup[2][0])
	}

	// original vertex coordinates and dimensions preserved.
	for ri := range backup {
		for pi := 0; pi < len(backup[ri]); pi++ {
			if original[ri][pi] != backup[ri][pi] {
				t.Errorf("ring %d point %d changed: %v != %v",
					ri, pi, original[ri][pi], backup[ri][pi])
			}
			if len(original[ri][pi]) != len(backup[ri][pi]) {
				t.Errorf("ring %d point %d dimensions changed", ri, pi)
			}
		}
	}

	// hole order preserved.
	if original[1][0] != backup[1][0] || original[2][0] != backup[2][0] {
		t.Error("hole order changed")
	}
}

func TestPolygon_Close_Dimensions(t *testing.T) {
	// every position keeps both its ordinates and the polygon remains 2d.
	p := Polygon{
		{{0, 0}, {10, 0}, {10, 10}, {0, 10}},
	}

	if p.Dimensions() != 2 {
		t.Fatalf("polygon dimension: %d != 2", p.Dimensions())
	}

	if err := p.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p.Dimensions() != 2 {
		t.Errorf("polygon dimension changed: %d != 2", p.Dimensions())
	}

	r := p[0]
	last := r[len(r)-1]
	if len(last) != 2 {
		t.Fatalf("position ordinates not preserved: got %d", len(last))
	}
	if last[0] != 0 || last[1] != 0 {
		t.Errorf("appended point does not copy the first point: %v", last)
	}
}

func TestPolygon_Close_EmptyRing(t *testing.T) {
	p := Polygon{
		{{0, 0}, {10, 0}, {10, 10}, {0, 10}},
		{}, // empty hole is ignored
	}

	if err := p.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(p[1]) != 0 {
		t.Errorf("empty ring should be left empty, got %d points", len(p[1]))
	}
	if !p[0].Closed() {
		t.Error("outer ring should be closed")
	}
}

func TestPolygon_Close_TooFewDistinctPoints(t *testing.T) {
	cases := []struct {
		name       string
		polygon    Polygon
		ringIndex  int
		ringLength int
	}{
		{
			name: "single point outer",
			polygon: Polygon{
				{{1, 1}},
			},
			ringIndex:  0,
			ringLength: 1,
		},
		{
			name: "two distinct points hole",
			polygon: Polygon{
				{{0, 0}, {10, 0}, {10, 10}, {0, 10}},
				{{2, 2}, {3, 2}, {2, 2}}, // 2 distinct points
			},
			ringIndex:  1,
			ringLength: 3,
		},
		{
			name: "two vertices hole",
			polygon: Polygon{
				{{0, 0}, {10, 0}, {10, 10}, {0, 10}},
				{{2, 2}, {3, 3}}, // only 2 distinct points
			},
			ringIndex:  1,
			ringLength: 2,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			backup := tc.polygon.Clone()

			err := tc.polygon.Close()
			if err == nil {
				t.Fatal("expected an error, got nil")
			}

			var ringErr *RingError
			if !errors.As(err, &ringErr) {
				t.Fatalf("expected *RingError, got %T: %v", err, err)
			}

			if !errors.Is(err, ErrRingTooFewPoints) {
				t.Errorf("error should wrap ErrRingTooFewPoints: %v", err)
			}

			if ringErr.PolygonIndex != 0 {
				t.Errorf("polygon index: %d != 0", ringErr.PolygonIndex)
			}
			if ringErr.RingIndex != tc.ringIndex {
				t.Errorf("ring index: %d != %d", ringErr.RingIndex, tc.ringIndex)
			}

			// geometry must be unmodified.
			if !tc.polygon.Equal(backup) {
				t.Errorf("polygon modified on error: got %v want %v", tc.polygon, backup)
			}
			if len(tc.polygon[tc.ringIndex]) != tc.ringLength {
				t.Errorf("invalid ring length changed: %d", len(tc.polygon[tc.ringIndex]))
			}
		})
	}
}

func TestMultiPolygon_Close(t *testing.T) {
	mp := MultiPolygon{
		{
			{{0, 0}, {10, 0}, {10, 10}, {0, 10}}, // open
		},
		{
			{{0, 0}, {5, 0}, {5, 5}, {0, 5}, {0, 0}}, // closed
		},
	}

	if err := mp.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mp[0][0]) != 5 || !mp[0][0].Closed() {
		t.Errorf("first polygon not closed: %v", mp[0][0])
	}
	if len(mp[1][0]) != 5 {
		t.Errorf("already closed ring duplicated its last point: %d", len(mp[1][0]))
	}
}

func TestMultiPolygon_Close_ErrorIndexAndAtomicity(t *testing.T) {
	mp := MultiPolygon{
		{
			{{0, 0}, {10, 0}, {10, 10}, {0, 10}}, // valid but open
		},
		{
			{{0, 0}, {5, 0}, {5, 5}, {0, 5}}, // valid open outer
			{{1, 1}, {2, 1}},                 // invalid hole: 2 distinct points
		},
	}

	backup := mp.Clone()

	err := mp.Close()
	if err == nil {
		t.Fatal("expected an error, got nil")
	}

	var ringErr *RingError
	if !errors.As(err, &ringErr) {
		t.Fatalf("expected *RingError, got %T: %v", err, err)
	}

	if ringErr.PolygonIndex != 1 {
		t.Errorf("polygon index: %d != 1", ringErr.PolygonIndex)
	}
	if ringErr.RingIndex != 1 {
		t.Errorf("ring index: %d != 1", ringErr.RingIndex)
	}

	if !mp.Equal(backup) {
		t.Errorf("multi-polygon modified on error: got %v want %v", mp, backup)
	}
}

func TestCloseRings(t *testing.T) {
	t.Run("collection", func(t *testing.T) {
		c := Collection{
			Point{1, 2}, // unaffected
			Polygon{
				{{0, 0}, {10, 0}, {10, 10}, {0, 10}},
			},
			MultiPolygon{
				{
					{{0, 0}, {5, 0}, {5, 5}, {0, 5}},
				},
			},
		}

		if err := CloseRings(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		p := c[1].(Polygon)
		if !p[0].Closed() || len(p[0]) != 5 {
			t.Errorf("polygon not closed: %v", p[0])
		}

		mp := c[2].(MultiPolygon)
		if !mp[0][0].Closed() || len(mp[0][0]) != 5 {
			t.Errorf("multi-polygon not closed: %v", mp[0][0])
		}
	})

	t.Run("non polygon geometry is a noop", func(t *testing.T) {
		ls := LineString{{0, 0}, {1, 1}, {2, 2}}
		if err := CloseRings(ls); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(ls) != 3 {
			t.Errorf("linestring modified: %v", ls)
		}
	})

	t.Run("collection error is atomic", func(t *testing.T) {
		c := Collection{
			Polygon{
				{{0, 0}, {10, 0}, {10, 10}, {0, 10}},
			},
			Polygon{
				{{0, 0}, {1, 0}}, // invalid: 2 distinct points
			},
		}
		backup := c.Clone()

		err := CloseRings(c)
		if err == nil {
			t.Fatal("expected an error, got nil")
		}

		var ringErr *RingError
		if !errors.As(err, &ringErr) {
			t.Fatalf("expected *RingError, got %T", err)
		}
		if ringErr.PolygonIndex != 0 || ringErr.RingIndex != 0 {
			t.Errorf("unexpected error location: %+v", ringErr)
		}
		if !c.Equal(backup) {
			t.Errorf("collection modified on error: %v", c)
		}
	})
}

func TestValidateRings(t *testing.T) {
	valid := Polygon{
		{{0, 0}, {10, 0}, {10, 10}, {0, 10}},
	}
	backup := valid.Clone()

	if err := ValidateRings(valid); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !valid.Equal(backup) {
		t.Error("ValidateRings modified the geometry")
	}

	invalid := Polygon{{{1, 1}, {2, 2}}}
	if err := ValidateRings(invalid); err == nil {
		t.Error("expected error for 2 distinct point ring")
	}
}
