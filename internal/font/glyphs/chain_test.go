// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/outline"
)

// testChain is a chain made up for the tests: a straight line, a curve which
// bulges to the right and another line, from (0, 0) up to (0, 100), as the
// right half of a stadium.  Its points are exact in binary floating point.
func testChain() chain {
	pt := func(x, y float64) outline.Point { return outline.Point{X: x, Y: y} }
	return chain{
		lineSeg(pt(0, 0), pt(40, 0)),
		curveSeg(cubic{pt(40, 0), pt(80, 0), pt(80, 100), pt(40, 100)}),
		lineSeg(pt(40, 100), pt(0, 100)),
	}
}

func samePoint(a, b outline.Point) bool { return a == b }

func sameChain(a, b chain) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].line != b[i].line {
			return false
		}
		if !samePoint(a[i].P0, b[i].P0) || !samePoint(a[i].P3, b[i].P3) {
			return false
		}
		if !a[i].line && (!samePoint(a[i].P1, b[i].P1) || !samePoint(a[i].P2, b[i].P2)) {
			return false
		}
	}
	return true
}

func pathOf(chains ...chain) outline.Path {
	var path outline.Path
	path.MoveTo(chains[0].start().X, chains[0].start().Y)
	for _, c := range chains {
		c.appendTo(&path)
	}
	return path
}

func TestChainConnected(t *testing.T) {
	t.Parallel()
	c := testChain()
	if err := c.connected(); err != nil {
		t.Fatal(err)
	}
	broken := append(chain{}, c...)
	broken[1].P0.X++
	if broken.connected() == nil {
		t.Error("a gap between two pieces was not found")
	}
}

func TestChainAppendKeepsLinesAndCurves(t *testing.T) {
	t.Parallel()
	path := pathOf(testChain())
	segs := path[0].Segs
	if len(segs) != 3 || segs[0].IsCurve || !segs[1].IsCurve || segs[2].IsCurve {
		t.Fatalf("want line, curve, line, got %+v", segs)
	}
	if segs[1].P1 != (outline.Point{X: 80, Y: 0}) || segs[1].P3 != (outline.Point{X: 40, Y: 100}) {
		t.Errorf("the curve was changed: %+v", segs[1])
	}
}

func TestChainReversed(t *testing.T) {
	t.Parallel()
	c := testChain()
	r := c.reversed()
	if err := r.connected(); err != nil {
		t.Fatal(err)
	}
	if r.start() != c.end() || r.end() != c.start() {
		t.Errorf("reversed chain runs from %v to %v, want %v to %v", r.start(), r.end(), c.end(), c.start())
	}
	if r[1].line || !r[0].line || r[2].line != true {
		t.Errorf("reversing changed which pieces are lines")
	}
	if !sameChain(r.reversed(), c) {
		t.Error("reversing twice does not give the original")
	}
}

func TestChainTransformsTwiceGiveTheOriginal(t *testing.T) {
	t.Parallel()
	c := testChain()
	for name, got := range map[string]chain{
		"mirrorX":  c.mirrorX(37).mirrorX(37),
		"flipY":    c.flipY(-12.5).flipY(-12.5),
		"reflectX": c.reflectX(37).reflectX(37),
		"reflectY": c.reflectY(3).reflectY(3),
		"shifted":  c.shifted(7, -3).shifted(-7, 3),
	} {
		if !sameChain(got, c) {
			t.Errorf("%s twice: got %+v, want %+v", name, got, c)
		}
	}
}

func TestChainEndPointsAndTangents(t *testing.T) {
	t.Parallel()
	c := testChain()
	near2 := func(what string, got, want outline.Point) {
		t.Helper()
		if math.Hypot(got.X-want.X, got.Y-want.Y) > 1e-9 {
			t.Errorf("%s: got %v, want %v", what, got, want)
		}
	}
	// Mirrored in x = 10: (x, y) -> (20 - x, y); the chain runs backwards.
	m := c.mirrorX(10)
	near2("mirror start", m.start(), outline.Point{X: 20, Y: 100})
	near2("mirror end", m.end(), outline.Point{X: 20, Y: 0})
	// c leaves (0,0) to the right and arrives at (0,100) going left.  The
	// mirror image leaves (20,0) to the left and arrives at (20,100) going
	// right; reversed, it leaves (20,100) to the left and arrives at (20,0)
	// going right.
	near2("mirror start tangent", m.startTangent(), outline.Point{X: -1, Y: 0})
	near2("mirror end tangent", m.endTangent(), outline.Point{X: 1, Y: 0})
	f := c.flipY(50) // (x, y) -> (x, 100 - y)
	near2("flip start", f.start(), outline.Point{X: 0, Y: 0})
	near2("flip end", f.end(), outline.Point{X: 0, Y: 100})
	near2("flip start tangent", f.startTangent(), outline.Point{X: 1, Y: 0})
	near2("flip end tangent", f.endTangent(), outline.Point{X: -1, Y: 0})
	// A point of the curve lands on the mirror image of that point.
	mid := c[1].at(0.3)
	got := m[1].at(0.7) // the curve is reversed, so the parameter is too
	near2("curve point", got, outline.Point{X: 20 - mid.X, Y: mid.Y})
	s := c.shifted(5, 7)
	near2("shift start", s.start(), outline.Point{X: 5, Y: 7})
	near2("shift end", s.end(), outline.Point{X: 5, Y: 107})
	// reflect keeps the order of the pieces.
	r := c.reflectX(10)
	near2("reflect start", r.start(), outline.Point{X: 20, Y: 0})
	near2("reflect end", r.end(), outline.Point{X: 20, Y: 100})
}

// TestChainMirroredHalvesKeepTheOrientation checks that a closed contour made
// from a chain and its mirrorX image, and one made from a chain and its flipY
// image, are counter-clockwise.
func TestChainMirroredHalvesKeepTheOrientation(t *testing.T) {
	t.Parallel()
	right := testChain() // from (0, 0), round the right, to (0, 100): counter-clockwise
	left := right.mirrorX(0)
	if err := left.connected(); err != nil {
		t.Fatal(err)
	}
	path := pathOf(right, left)
	if len(path[0].Segs) != 6 {
		t.Fatalf("want 6 segments, got %d", len(path[0].Segs))
	}
	if a := path[0].SignedArea(); a <= 0 {
		t.Errorf("the contour of a chain and its mirror image has area %v, want counter-clockwise", a)
	}
	x0, _, x1, _ := path.BBox()
	near(t, "width", x1-x0, 140) // the curve reaches x = 70

	// A chain from (0, 40) below y = 0 to (100, 40) continues with its
	// flip over the top, back to the start.
	pt := func(x, y float64) outline.Point { return outline.Point{X: x, Y: y} }
	bottom := chain{curveSeg(cubic{pt(0, 40), pt(0, -20), pt(100, -20), pt(100, 40)})}
	top := bottom.flipY(40)
	if err := top.connected(); err != nil {
		t.Fatal(err)
	}
	path = pathOf(bottom, top)
	if a := path[0].SignedArea(); a <= 0 {
		t.Errorf("the contour of a chain and its flip has area %v, want counter-clockwise", a)
	}
}

func TestChainJoin(t *testing.T) {
	t.Parallel()
	c := testChain()
	j := join(c, c.mirrorX(0))
	if len(j) != 6 || j.start() != c.start() || j.end() != c.start() {
		t.Errorf("join: got %d pieces from %v to %v", len(j), j.start(), j.end())
	}
	defer func() {
		if recover() == nil {
			t.Error("joining chains which do not meet did not panic")
		}
	}()
	join(c, c)
}

func TestSplitAtExtremaIsOneChain(t *testing.T) {
	t.Parallel()
	pt := func(x, y float64) outline.Point { return outline.Point{X: x, Y: y} }
	c := cubic{pt(0, 0), pt(100, 0), pt(100, 100), pt(0, 100)} // extreme in x at 0.5
	ch := splitAtExtrema(c)
	if len(ch) != 2 || ch.connected() != nil {
		t.Fatalf("want two connected pieces, got %+v", ch)
	}
	if ch.start() != c.P0 || ch.end() != c.P3 {
		t.Errorf("the pieces do not span the curve")
	}
}

// TestSplitAtExtremaSkipsExtremesNearTheEnds checks that splitAtExtrema does
// not cut at an extreme within a unit of an end of the curve.
func TestSplitAtExtremaSkipsExtremesNearTheEnds(t *testing.T) {
	t.Parallel()
	// Leaves the start slightly downwards, then rises: the lowest point lies
	// 0.2 from the start.
	c := cubic{outline.Point{X: 0, Y: 0}, outline.Point{X: 5, Y: -0.05}, outline.Point{X: 30, Y: 10}, outline.Point{X: 60, Y: 10}}
	var path outline.Path
	path.MoveTo(c.P0.X, c.P0.Y)
	splitAtExtrema(c).appendTo(&path)
	if n := len(path[0].Segs); n != 1 {
		t.Errorf("the curve is cut into %d pieces, want 1", n)
	}
	// An extreme in the middle is still a cut.
	d := cubic{outline.Point{X: 0, Y: 0}, outline.Point{X: 20, Y: -40}, outline.Point{X: 40, Y: -40}, outline.Point{X: 60, Y: 0}}
	var path2 outline.Path
	path2.MoveTo(d.P0.X, d.P0.Y)
	splitAtExtrema(d).appendTo(&path2)
	if n := len(path2[0].Segs); n != 2 {
		t.Errorf("the curve is cut into %d pieces, want 2", n)
	}
}
