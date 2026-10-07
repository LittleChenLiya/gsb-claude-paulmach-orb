package orb

import (
	"errors"
	"testing"
)

func TestPolygonClose(t *testing.T) {
	// outer ring unclosed, first hole already closed, second hole unclosed.
	p := Polygon{
		{{0, 0}, {4, 0}, {4, 4}, {0, 4}},         // open outer
		{{1, 1}, {2, 1}, {2, 2}, {1, 2}, {1, 1}}, // closed hole
		{{3, 3}, {3.5, 3}, {3.5, 3.5}, {3, 3.5}}, // open hole
	}
	original := p.Clone()

	closed, err := p.Close()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// lengths: first point appended to open rings only.
	if len(closed[0]) != 5 {
		t.Errorf("outer ring: expected length 5, got %d", len(closed[0]))
	}
	if len(closed[1]) != 5 {
		t.Errorf("closed hole: expected unchanged length 5, got %d", len(closed[1]))
	}
	if len(closed[2]) != 5 {
		t.Errorf("open hole: expected length 5, got %d", len(closed[2]))
	}

	// closed hole must not get its last point duplicated.
	if closed[1][len(closed[1])-2] == closed[1][len(closed[1])-1] && len(p[1]) != len(closed[1]) {
		t.Errorf("closed hole: last point was duplicated")
	}

	// every ring is closed now, order and coordinates preserved.
	for i, r := range closed {
		if !r.Closed() {
			t.Errorf("ring %d: not closed after Close", i)
		}
		for j, pt := range p[i] {
			if closed[i][j] != pt {
				t.Errorf("ring %d point %d: original coordinate changed: %v != %v", i, j, closed[i][j], pt)
			}
		}
	}
	if closed[0][4] != (Point{0, 0}) {
		t.Errorf("outer ring: appended point = %v, expected {0 0}", closed[0][4])
	}
	if closed[2][4] != (Point{3, 3}) {
		t.Errorf("open hole: appended point = %v, expected {3 3}", closed[2][4])
	}

	// dimensions preserved.
	if closed.Dimensions() != p.Dimensions() {
		t.Errorf("dimensions changed: %d != %d", closed.Dimensions(), p.Dimensions())
	}

	// the original geometry must be untouched.
	if !p.Equal(original) {
		t.Errorf("original polygon was modified")
	}
	for i, r := range p {
		if len(r) != len(original[i]) {
			t.Errorf("original ring %d length changed: %d != %d", i, len(r), len(original[i]))
		}
	}
}

func TestPolygonCloseErrorLocatesRing(t *testing.T) {
	// the second ring (a hole) has only two distinct points.
	p := Polygon{
		{{0, 0}, {4, 0}, {4, 4}, {0, 4}},
		{{1, 1}, {2, 2}, {1, 1}},
	}
	original := p.Clone()

	closed, err := p.Close()
	if err == nil {
		t.Fatal("expected an error for a hole with two distinct points")
	}
	if closed != nil {
		t.Errorf("expected nil result on error, got %v", closed)
	}

	var rce *RingClosureError
	if !errors.As(err, &rce) {
		t.Fatalf("expected *RingClosureError, got %T (%v)", err, err)
	}
	if rce.Ring != 1 {
		t.Errorf("error ring index = %d, expected 1 (the hole)", rce.Ring)
	}
	if rce.Polygon != -1 {
		t.Errorf("error polygon index = %d, expected -1", rce.Polygon)
	}
	if rce.Distinct != 2 {
		t.Errorf("error distinct count = %d, expected 2", rce.Distinct)
	}

	// original geometry must not be modified on error.
	if !p.Equal(original) {
		t.Errorf("original polygon was modified after error")
	}
}

func TestPolygonCloseEmptyRing(t *testing.T) {
	p := Polygon{
		{{0, 0}, {4, 0}, {4, 4}, {0, 4}},
		{},
	}

	closed, err := p.Close()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(closed) != 2 {
		t.Fatalf("expected 2 rings, got %d", len(closed))
	}
	if len(closed[1]) != 0 {
		t.Errorf("empty ring should stay empty, got %v", closed[1])
	}
}

func TestMultiPolygonClose(t *testing.T) {
	mp := MultiPolygon{
		{{{0, 0}, {1, 0}, {1, 1}, {0, 1}}},
		{{{2, 2}, {3, 2}, {3, 3}, {2, 3}, {2, 2}}},
	}
	original := mp.Clone()

	closed, err := mp.Close()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(closed[0][0]) != 5 || len(closed[1][0]) != 5 {
		t.Errorf("unexpected ring lengths: %d, %d", len(closed[0][0]), len(closed[1][0]))
	}
	if !mp.Equal(original) {
		t.Errorf("original multipolygon was modified")
	}

	// error should locate polygon and ring.
	bad := MultiPolygon{
		{{{0, 0}, {1, 0}, {1, 1}, {0, 1}}},
		{{{2, 2}, {3, 3}}},
	}
	_, err = bad.Close()

	var rce *RingClosureError
	if !errors.As(err, &rce) {
		t.Fatalf("expected *RingClosureError, got %T (%v)", err, err)
	}
	if rce.Polygon != 1 || rce.Ring != 0 {
		t.Errorf("error location = polygon %d ring %d, expected polygon 1 ring 0", rce.Polygon, rce.Ring)
	}
}
