# orb/planar [![Godoc Reference](https://pkg.go.dev/badge/github.com/paulmach/orb)](https://pkg.go.dev/github.com/paulmach/orb/planar)

The geometries defined in the `orb` package are generic 2d geometries.
Depending on what projection they're in, e.g. lon/lat or flat on the plane,
area and distance calculations are different. This package implements methods
that assume the planar or Euclidean context.

## Examples

Area of 3-4-5 triangle:

```go
r := orb.Ring{{0, 0}, {3, 0}, {0, 4}, {0, 0}}
a := planar.Area(r)

fmt.Println(a)
// Output:
// 6
```

Distance between two points:

```go
d := planar.Distance(orb.Point{0, 0}, orb.Point{3, 4})

fmt.Println(d)
// Output:
// 5
```

Length/circumference of a 3-4-5 triangle:

```go
r := orb.Ring{{0, 0}, {3, 0}, {0, 4}, {0, 0}}
l := planar.Length(r)

fmt.Println(l)
// Output:
// 12
```

Classify a point as interior, on the boundary or exterior of a polygon with
one outer ring and up to one hole:

```go
p := orb.Polygon{
    {{0, 0}, {0, 4}, {4, 4}, {4, 0}, {0, 0}},
    {{1, 1}, {1, 3}, {3, 3}, {3, 1}, {1, 1}},
}

planar.PolygonPointPosition(p, orb.Point{0.5, 0.5}) // planar.Interior
planar.PolygonPointPosition(p, orb.Point{2, 2})     // planar.Exterior (in the hole)
planar.PolygonPointPosition(p, orb.Point{0, 2})     // planar.Boundary
```

The result does not depend on the direction of the rings and the input
polygon is not modified.
