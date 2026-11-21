// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package outline

import (
	"math"
	"testing"
)

func square(x0, y0, x1, y1 float64, ccw bool) Contour {
	var p Path
	p.MoveTo(x0, y0)
	if ccw {
		p.LineTo(x1, y0)
		p.LineTo(x1, y1)
		p.LineTo(x0, y1)
	} else {
		p.LineTo(x0, y1)
		p.LineTo(x1, y1)
		p.LineTo(x1, y0)
	}
	return p[0]
}

func TestSignedArea(t *testing.T) {
	t.Parallel()
	if got := square(0, 0, 10, 20, true).SignedArea(); math.Abs(got-200) > 1e-9 {
		t.Errorf("ccw square: area %v, want 200", got)
	}
	if got := square(0, 0, 10, 20, false).SignedArea(); math.Abs(got+200) > 1e-9 {
		t.Errorf("cw square: area %v, want -200", got)
	}
}

// circle returns a circle of radius r around the origin, made from four cubic
// segments and starting at (r, 0).
func circle(r float64) Contour {
	const k = 0.5522847498
	var p Path
	p.MoveTo(r, 0)
	p.CurveTo(r, k*r, k*r, r, 0, r)
	p.CurveTo(-k*r, r, -r, k*r, -r, 0)
	p.CurveTo(-r, -k*r, -k*r, -r, 0, -r)
	p.CurveTo(k*r, -r, r, -k*r, r, 0)
	return p[0]
}

func TestSignedAreaCurve(t *testing.T) {
	t.Parallel()
	const r = 100.0
	got := circle(r).SignedArea()
	want := math.Pi * r * r
	if math.Abs(got-want) > 0.01*want {
		t.Errorf("circle: area %v, want about %v", got, want)
	}
}

func checkFloats(t *testing.T, got, want []float64, tol float64) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if math.Abs(got[i]-want[i]) > tol {
			t.Errorf("value %d: got %v, want %v", i, got[i], want[i])
		}
	}
}

func TestCrossingsCurve(t *testing.T) {
	t.Parallel()
	p := Path{circle(100)}
	checkFloats(t, p.CrossingsX(0), []float64{-100, 100}, 0.01)
	checkFloats(t, p.CrossingsY(0), []float64{-100, 100}, 0.01)
}

func TestBBoxCurve(t *testing.T) {
	t.Parallel()
	x0, y0, x1, y1 := Path{circle(100)}.BBox()
	for _, v := range []float64{x0 + 100, y0 + 100, x1 - 100, y1 - 100} {
		if math.Abs(v) > 0.01 {
			t.Fatalf("bbox = %v %v %v %v", x0, y0, x1, y1)
		}
	}
}

func TestExplicitClose(t *testing.T) {
	t.Parallel()
	implicit := square(10, 0, 30, 100, true)
	explicit := implicit
	explicit.Segs = append(append([]Segment(nil), implicit.Segs...),
		Segment{P3: implicit.Start})
	if a, b := explicit.SignedArea(), implicit.SignedArea(); a != b {
		t.Errorf("area with closing segment %v, without %v", a, b)
	}
	checkFloats(t, Path{explicit}.CrossingsX(50), Path{implicit}.CrossingsX(50), 0)
}

func TestCrossingsX(t *testing.T) {
	t.Parallel()
	p := Path{square(10, 0, 30, 100, true), square(50, 0, 90, 100, true)}
	got := p.CrossingsX(50)
	want := []float64{10, 30, 50, 90}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if math.Abs(got[i]-want[i]) > 1e-9 {
			t.Errorf("crossing %d: got %v, want %v", i, got[i], want[i])
		}
	}
}

func TestCrossingsY(t *testing.T) {
	t.Parallel()
	p := Path{square(10, 0, 30, 100, true), square(10, 150, 30, 200, true)}
	checkFloats(t, p.CrossingsY(20), []float64{0, 100, 150, 200}, 1e-9)
}

func TestBBox(t *testing.T) {
	t.Parallel()
	p := Path{square(10, -5, 30, 100, true)}
	x0, y0, x1, y1 := p.BBox()
	if x0 != 10 || y0 != -5 || x1 != 30 || y1 != 100 {
		t.Errorf("bbox = %v %v %v %v", x0, y0, x1, y1)
	}
}

func TestRound(t *testing.T) {
	t.Parallel()
	var p Path
	p.MoveTo(0.4, 0.6)
	p.CurveTo(1.5, 2.5, 3.49, 4.51, 5.2, 6.8)
	q := p.Round()
	c := q[0]
	if c.Start != (Point{0, 1}) {
		t.Errorf("start = %v", c.Start)
	}
	s := c.Segs[0]
	if s.P1 != (Point{2, 3}) || s.P2 != (Point{3, 5}) || s.P3 != (Point{5, 7}) {
		t.Errorf("segment = %+v", s)
	}
}

func samePoints(t *testing.T, what string, got, want []Point) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s: got %v, want %v", what, got, want)
	}
	for i := range want {
		if math.Abs(got[i].X-want[i].X) > 1e-6 || math.Abs(got[i].Y-want[i].Y) > 1e-6 {
			t.Fatalf("%s: got %v, want %v", what, got, want)
		}
	}
}

func polygon(pts ...Point) Contour {
	var p Path
	p.MoveTo(pts[0].X, pts[0].Y)
	for _, q := range pts[1:] {
		p.LineTo(q.X, q.Y)
	}
	return p[0]
}

func TestSelfIntersectionsFigureOfEight(t *testing.T) {
	t.Parallel()
	eight := polygon(Point{0, 0}, Point{20, 20}, Point{20, 0}, Point{0, 20})
	samePoints(t, "figure of eight", Path{eight}.SelfIntersections(), []Point{{10, 10}})
}

func TestSelfIntersectionsCurvedEight(t *testing.T) {
	t.Parallel()
	var p Path
	p.MoveTo(0, 0)
	p.CurveTo(40, 0, 40, 100, 0, 100)
	p.CurveTo(-40, 100, -40, 0, 0, 0)
	if got := p.SelfIntersections(); len(got) != 0 {
		t.Errorf("a simple curved contour has %v", got)
	}
	// Two S-shaped curves which mirror each other cross once, in the middle.
	var q Path
	q.MoveTo(0, 0)
	q.CurveTo(30, 0, 70, 100, 100, 100)
	q.LineTo(100, 0)
	q.CurveTo(70, 0, 30, 100, 0, 100)
	if got := q.SelfIntersections(); len(got) != 1 || math.Abs(got[0].X-50) > 0.5 || math.Abs(got[0].Y-50) > 0.5 {
		t.Errorf("curved figure of eight: %v", got)
	}
}

func TestSelfIntersectionsOverlappingSquares(t *testing.T) {
	t.Parallel()
	p := Path{square(0, 0, 10, 10, true), square(5, 5, 15, 15, true)}
	samePoints(t, "squares", p.SelfIntersections(), []Point{{5, 10}, {10, 5}})
}

func TestSelfIntersectionsRing(t *testing.T) {
	t.Parallel()
	p := Path{square(0, 0, 30, 30, true), square(10, 10, 20, 20, false)}
	if got := p.SelfIntersections(); len(got) != 0 {
		t.Errorf("ring: %v", got)
	}
	if got := (Path{circle(100), circle(50)}).SelfIntersections(); len(got) != 0 {
		t.Errorf("circles: %v", got)
	}
}

func TestSelfIntersectionsCleanCircle(t *testing.T) {
	t.Parallel()
	if got := (Path{circle(100)}).SelfIntersections(); len(got) != 0 {
		t.Errorf("circle: %v", got)
	}
}

// A contour traced twice lies on top of itself.  Every edge overlaps its
// twin, and the ends of each overlap are reported: here the four corners.
func TestSelfIntersectionsTracedTwice(t *testing.T) {
	t.Parallel()
	c := square(0, 0, 10, 10, true)
	p := Path{c, c}
	samePoints(t, "twice", p.SelfIntersections(), []Point{{0, 0}, {0, 10}, {10, 0}, {10, 10}})
}

// A vertex which lies on the interior of another edge, and two contours which
// meet in a corner, count as touching.  The touching point is reported.
func TestSelfIntersectionsTouching(t *testing.T) {
	t.Parallel()
	// A hexagon-like shape whose lower right vertex returns to the middle of
	// the bottom edge.
	pinch := polygon(Point{0, 0}, Point{20, 0}, Point{20, 20}, Point{10, 0}, Point{10, 20}, Point{0, 20})
	samePoints(t, "pinch", Path{pinch}.SelfIntersections(), []Point{{10, 0}})

	corner := Path{square(0, 0, 10, 10, true), square(10, 10, 20, 20, true)}
	samePoints(t, "corner", corner.SelfIntersections(), []Point{{10, 10}})

	// Two squares which touch along part of an edge: the overlap ends are
	// reported.
	side := Path{square(0, 0, 10, 10, true), square(10, 5, 20, 15, true)}
	samePoints(t, "side", side.SelfIntersections(), []Point{{10, 5}, {10, 10}})

	// Squares which are close but apart do not touch.
	apart := Path{square(0, 0, 10, 10, true), square(10.001, 0, 20, 10, true)}
	if got := apart.SelfIntersections(); len(got) != 0 {
		t.Errorf("apart: %v", got)
	}
}

// Neighbouring segments share an end point, which is not an intersection.  A
// spike, where the next segment goes back along the previous one, is reported
// at its tip.  The closing segment of a contour is a neighbour of the first
// and the last.
func TestSelfIntersectionsNeighbours(t *testing.T) {
	t.Parallel()
	if got := (Path{square(0, 0, 10, 10, true)}).SelfIntersections(); len(got) != 0 {
		t.Errorf("square: %v", got)
	}
	spike := polygon(Point{0, 0}, Point{10, 0}, Point{20, 0}, Point{10, 0}, Point{10, 10}, Point{0, 10})
	if got := (Path{spike}).SelfIntersections(); len(got) == 0 {
		t.Error("a spike is not reported")
	}
	// Repeated points make zero-length segments, which are ignored.
	dup := polygon(Point{0, 0}, Point{10, 0}, Point{10, 0}, Point{10, 10}, Point{0, 10}, Point{0, 0})
	if got := (Path{dup}).SelfIntersections(); len(got) != 0 {
		t.Errorf("square with repeated points: %v", got)
	}
}

func TestMinDistance(t *testing.T) {
	t.Parallel()
	// A U shape: arms 10 thick, base 10 thick, the slot between the arms 6 wide
	// and 40 long.
	u := polygon(Point{0, 0}, Point{26, 0}, Point{26, 50}, Point{16, 50},
		Point{16, 10}, Point{10, 10}, Point{10, 50}, Point{0, 50})
	cases := []struct {
		name string
		path Path
		want float64
		tol  float64
	}{
		{"ring of squares", Path{square(0, 0, 100, 100, true), square(20, 20, 80, 80, false)}, 20, 1e-9},
		{"ring of circles", Path{circle(100), reverse(circle(80))}, 20, 0.05},
		{"two squares 3 apart", Path{square(0, 0, 10, 10, true), square(13, 0, 23, 10, true)}, 3, 1e-9},
		{"two squares, diagonal", Path{square(0, 0, 10, 10, true), square(13, 14, 23, 24, true)}, 5, 1e-9},
		{"thin rectangle", Path{square(0, 0, 100, 4, true)}, 4, 1e-9},
		{"U shape", Path{u}, 6, 1e-9},
		// Corners, smooth curves and the far sides of a fat shape are not
		// narrow places: the arc length between them is below five times
		// their distance.
		{"big square", Path{square(0, 0, 100, 100, true)}, math.Inf(1), 0},
		{"circle", Path{circle(100)}, math.Inf(1), 0},
		// The sides of a stroke are found if it is more than four widths long.
		{"short rectangle", Path{square(0, 0, 10, 30, true)}, math.Inf(1), 0},
		{"long rectangle", Path{square(0, 0, 10, 50, true)}, 10, 1e-9},
	}
	for _, c := range cases {
		d, a, b := Path.MinDistance(c.path, 5)
		if math.IsInf(c.want, 1) {
			if !math.IsInf(d, 1) {
				t.Errorf("%s: got %v, want no pair", c.name, d)
			}
			continue
		}
		if math.Abs(d-c.want) > c.tol {
			t.Errorf("%s: got %v, want %v", c.name, d, c.want)
		}
		if got := math.Hypot(a.X-b.X, a.Y-b.Y); math.Abs(got-d) > 1e-9 {
			t.Errorf("%s: the points %v and %v are %v apart, want %v", c.name, a, b, got, d)
		}
	}
}

// reverse returns the contour with the opposite orientation.
func reverse(c Contour) Contour {
	pts := c.Flatten()
	for i, j := 0, len(pts)-1; i < j; i, j = i+1, j-1 {
		pts[i], pts[j] = pts[j], pts[i]
	}
	return polygon(pts...)
}

func TestCubicExtrema(t *testing.T) {
	t.Parallel()
	// The derivative of a cubic with the end points 0 and 1, handles 1/3 and
	// 2/3 is constant, so there is no extreme.
	if got := CubicExtrema(0, 1.0/3, 2.0/3, 1); len(got) != 0 {
		t.Errorf("straight line: %v", got)
	}
	// A bulge: y rises above both ends at t = 1/2.
	checkFloats(t, CubicExtrema(0, 50, 50, 0), []float64{0.5}, 1e-12)
	// Two extremes: the roots of 61 t^2 - 60 t + 10.
	checkFloats(t, CubicExtrema(0, 10, -10, 1), []float64{(60 - math.Sqrt(1160)) / 122, (60 + math.Sqrt(1160)) / 122}, 1e-12)
}

// With a leading coefficient of 1e-9 the roots are 0.5 and about 1e9.  The
// test checks the small root.
func TestCubicExtremaTinyLeadingCoefficient(t *testing.T) {
	t.Parallel()
	got := CubicExtrema(0, 0.5, 0.5, 1e-9)
	want := 0.5 + 0.25e-9 // the root of 1e-9 t^2 - t + 0.5 near 1/2
	if len(got) != 1 || math.Abs(got[0]-want) > 1e-12 {
		t.Errorf("got %v, want [%v]", got, want)
	}
}
