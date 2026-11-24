// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package fair

import (
	"bytes"
	"math"
	"strings"
	"testing"

	"seehuhn.de/go/bedern/internal/outline"
)

// pen draws a contour from arcs and lines, keeping the current point and
// direction.
type pen struct {
	p   *outline.Path
	x   Point
	dir float64 // radians
}

func newPen(p *outline.Path, x, y, dir float64) *pen {
	p.MoveTo(x, y)
	return &pen{p: p, x: Point{X: x, Y: y}, dir: dir * math.Pi / 180}
}

// arc turns by deg degrees (positive to the left) on a circle of radius r,
// in cubics of at most 45 degrees.
func (q *pen) arc(r, deg float64) {
	n := int(math.Ceil(math.Abs(deg) / 45))
	step := deg * math.Pi / 180 / float64(n)
	for range n {
		k := 4.0 / 3 * math.Tan(math.Abs(step)/4) * r
		t0 := Point{X: math.Cos(q.dir), Y: math.Sin(q.dir)}
		side := 1.0
		if step < 0 {
			side = -1
		}
		c := add(q.x, mul(Point{X: -t0.Y, Y: t0.X}, side*r))
		d1 := q.dir + step
		t1 := Point{X: math.Cos(d1), Y: math.Sin(d1)}
		end := sub(c, mul(Point{X: -t1.Y, Y: t1.X}, side*r))
		p1 := add(q.x, mul(t0, k))
		p2 := sub(end, mul(t1, k))
		q.p.CurveTo(p1.X, p1.Y, p2.X, p2.Y, end.X, end.Y)
		q.x, q.dir = end, d1
	}
}

// line goes straight on by l.
func (q *pen) line(l float64) {
	q.x = add(q.x, mul(Point{X: math.Cos(q.dir), Y: math.Sin(q.dir)}, l))
	q.p.LineTo(q.x.X, q.x.Y)
}

// turn changes the direction at a corner.
func (q *pen) turn(deg float64) { q.dir += deg * math.Pi / 180 }

// findingsNear returns the findings of kind k within d of (x, y).
func findingsNear(a *Analysis, k Kind, x, y, d float64) []Finding {
	var out []Finding
	for _, f := range a.Findings {
		if f.Kind == k && math.Hypot(f.X-x, f.Y-y) <= d {
			out = append(out, f)
		}
	}
	return out
}

func describe(a *Analysis) string {
	var sb strings.Builder
	for _, f := range a.Findings {
		sb.WriteString("\n  " + f.ID() + ": " + f.Detail)
	}
	return sb.String()
}

// closeWith closes a contour drawn from (0, 0) with straight lines through a
// point far below, so that the closing edges meet the curve at corners.
func closeWith(q *pen, p *outline.Path) {
	p.LineTo(q.x.X, q.x.Y-600)
	p.LineTo(0, -600)
}

func TestJumpOfTwoArcs(t *testing.T) {
	var p outline.Path
	q := newPen(&p, 0, 0, 90)
	q.arc(300, -40) // turning right: the ink is on the right, which orient fixes
	join := q.x
	q.arc(100, -50)
	closeWith(q, &p)
	a := Analyse("test", p, Default)
	if len(findingsNear(a, Jump, join.X, join.Y, 3)) != 1 {
		t.Errorf("no jump at the join (%.0f, %.0f) of radius 300 and 100:%s", join.X, join.Y, describe(a))
	}
}

func TestFlatBetweenArcs(t *testing.T) {
	var p outline.Path
	q := newPen(&p, 0, 0, 90)
	q.arc(150, -50)
	q.arc(8000, -0.4)
	mid := q.x
	q.arc(8000, -0.4)
	q.arc(150, -50)
	closeWith(q, &p)
	a := Analyse("test", p, Default)
	if len(findingsNear(a, Flat, mid.X, mid.Y, 10)) != 1 {
		t.Errorf("no flat at (%.0f, %.0f):%s", mid.X, mid.Y, describe(a))
	}
	// A short straight segment between the arcs is a flat as well.
	p = nil
	q = newPen(&p, 0, 0, 90)
	q.arc(150, -50)
	q.line(40)
	mid = sub(q.x, mul(Point{X: math.Cos(q.dir), Y: math.Sin(q.dir)}, 20))
	q.arc(150, -50)
	closeWith(q, &p)
	a = Analyse("test", p, Default)
	if len(findingsNear(a, Flat, mid.X, mid.Y, 10)) != 1 {
		t.Errorf("no flat at the straight segment (%.0f, %.0f):%s", mid.X, mid.Y, describe(a))
	}
}

func TestKink(t *testing.T) {
	var p outline.Path
	q := newPen(&p, 0, 0, 90)
	q.arc(200, -40)
	kink := q.x
	q.turn(-3)
	q.arc(200, -40)
	closeWith(q, &p)
	a := Analyse("test", p, Default)
	if fs := findingsNear(a, Kink, kink.X, kink.Y, 1); len(fs) != 1 || math.Abs(fs[0].Size-3) > 0.1 {
		t.Errorf("no kink of 3 degrees at (%.0f, %.0f):%s", kink.X, kink.Y, describe(a))
	}
	if len(a.Findings) != 1 {
		t.Errorf("want only the kink:%s", describe(a))
	}
}

func TestSBend(t *testing.T) {
	var p outline.Path
	q := newPen(&p, 0, 0, 0)
	q.arc(150, 60)
	join := q.x
	q.arc(150, -60)
	p.LineTo(q.x.X, q.x.Y-500)
	p.LineTo(0, -500)
	a := Analyse("test", p, Default)
	if len(findingsNear(a, Inflection, join.X, join.Y, 3)) != 1 {
		t.Errorf("no inflection at (%.0f, %.0f):%s", join.X, join.Y, describe(a))
	}
}

func TestShortSegment(t *testing.T) {
	var p outline.Path
	q := newPen(&p, 0, 0, 90)
	q.arc(200, -40)
	q.turn(-30)
	q.line(6)
	mid := sub(q.x, mul(Point{X: math.Cos(q.dir), Y: math.Sin(q.dir)}, 3))
	q.turn(-30)
	q.arc(200, -40)
	closeWith(q, &p)
	a := Analyse("test", p, Default)
	if fs := findingsNear(a, Short, mid.X, mid.Y, 2); len(fs) != 1 || math.Abs(fs[0].Size-6) > 0.01 {
		t.Errorf("no short segment of 6 units at (%.0f, %.0f):%s", mid.X, mid.Y, describe(a))
	}
}

func TestWiggle(t *testing.T) {
	var p outline.Path
	q := newPen(&p, 0, 0, 90)
	q.arc(200, -30)
	q.arc(200, 3)
	mid := q.x
	q.arc(200, -3)
	q.arc(200, -30)
	closeWith(q, &p)
	a := Analyse("test", p, Default)
	if len(findingsNear(a, Wiggle, mid.X, mid.Y, 12)) == 0 {
		t.Errorf("no wiggle at (%.0f, %.0f):%s", mid.X, mid.Y, describe(a))
	}
}

func TestSpike(t *testing.T) {
	var p outline.Path
	q := newPen(&p, 0, 0, 90)
	q.arc(150, -50)
	q.arc(15, -20)
	peak := q.x
	q.arc(15, -20)
	q.arc(150, -50)
	closeWith(q, &p)
	a := Analyse("test", p, Default)
	if len(findingsNear(a, Spike, peak.X, peak.Y, 8)) != 1 {
		t.Errorf("no spike at (%.0f, %.0f):%s", peak.X, peak.Y, describe(a))
	}
}

// stalk draws a stroke of the width w along the x axis, from x = 0 to 400,
// narrowing to pw at x = 200 between x = 160 and 240, as one contour.
func stalk(w, pw float64) outline.Path {
	var p outline.Path
	p.MoveTo(0, 0)
	d := (w - pw) / 2
	p.LineTo(160, 0)
	p.CurveTo(175, 0, 185, d, 200, d)
	p.CurveTo(215, d, 225, 0, 240, 0)
	p.LineTo(400, 0)
	p.LineTo(400, w)
	p.LineTo(240, w)
	p.CurveTo(225, w, 215, w-d, 200, w-d)
	p.CurveTo(185, w-d, 175, w, 160, w)
	p.LineTo(0, w)
	return p
}

func TestPinch(t *testing.T) {
	a := Analyse("test", stalk(60, 40), Default)
	found := false
	for _, f := range append(a.Notes, a.Findings...) {
		if f.Kind == Pinch && math.Abs(f.X-200) < 10 {
			found = true
		}
	}
	if !found {
		t.Errorf("no pinch at x = 200: notes %v", a.Notes)
	}
	a = Analyse("test", stalk(60, 60), Default)
	for _, f := range append(a.Notes, a.Findings...) {
		if f.Kind == Pinch || f.Kind == Swell {
			t.Errorf("even stalk: %s at (%.0f, %.0f)", f.Kind, f.X, f.Y)
		}
	}
}

func clean(t *testing.T, name string, p outline.Path) {
	t.Helper()
	a := Analyse(name, p, Default)
	if len(a.Findings) > 0 {
		t.Errorf("%s: want no findings:%s", name, describe(a))
	}
}

func TestCleanShapes(t *testing.T) {
	// A circle of four cubics.
	var p outline.Path
	q := newPen(&p, 200, 0, 90)
	q.arc(200, 360)
	clean(t, "circle", p)

	// An ellipse of four cubics, the circle scaled by 1.6 in x.
	var e outline.Path
	for _, c := range p {
		e.MoveTo(c.Start.X*1.6, c.Start.Y)
		for _, s := range c.Segs {
			e.CurveTo(s.P1.X*1.6, s.P1.Y, s.P2.X*1.6, s.P2.Y, s.P3.X*1.6, s.P3.Y)
		}
	}
	clean(t, "ellipse", e)

	// A spiral whose curvature grows linearly with the arc length, from
	// radius 500 to radius 60, drawn as cubics through points of it with
	// the tangents there and handles of a third of the arc length.
	var sp outline.Path
	const n, length = 12, 600.0
	k0, k1 := 1.0/500, 1.0/60
	var pts []Point
	var dirs []float64
	x, th := Point{}, 0.0
	steps := 6000
	for i := 0; i <= steps; i++ {
		s := length * float64(i) / float64(steps)
		if i%(steps/n) == 0 {
			pts = append(pts, x)
			dirs = append(dirs, th)
		}
		ds := length / float64(steps)
		k := k0 + (k1-k0)*s/length
		x = add(x, mul(Point{X: math.Cos(th), Y: math.Sin(th)}, ds))
		th += k * ds
	}
	sp.MoveTo(pts[0].X, pts[0].Y)
	h := length / n / 3
	for i := 1; i < len(pts); i++ {
		a, b := pts[i-1], pts[i]
		ta := Point{X: math.Cos(dirs[i-1]), Y: math.Sin(dirs[i-1])}
		tb := Point{X: math.Cos(dirs[i]), Y: math.Sin(dirs[i])}
		p1, p2 := add(a, mul(ta, h)), sub(b, mul(tb, h))
		sp.CurveTo(p1.X, p1.Y, p2.X, p2.Y, b.X, b.Y)
	}
	last := pts[len(pts)-1]
	sp.LineTo(last.X, -400)
	sp.LineTo(0, -400)
	clean(t, "spiral", sp)

	// A straight stem with a foot serif, bracketed by concave quarter
	// circles of radius 20 into its two sides.
	var st outline.Path
	q = newPen(&st, 0, 0, 0)
	q.line(200)
	q.turn(90)
	q.line(20)
	q.turn(90)
	q.line(50)
	q.arc(20, -90)
	q.line(500)
	q.turn(90)
	q.line(60)
	q.turn(90)
	q.line(500)
	q.arc(20, -90)
	q.line(50)
	clean(t, "stem", st)
}

// TestIdentity checks that the identity of a finding does not depend on
// another part of the glyph: the contour index, the start of the contour or
// a change of a distant part.
func TestIdentity(t *testing.T) {
	build := func(far float64, extra bool) *Analysis {
		var p outline.Path
		if extra {
			q := newPen(&p, 1000, 0, 90)
			q.arc(100, 360)
		}
		q := newPen(&p, 0, 0, 90)
		q.arc(300, -40)
		q.arc(100, -50)
		p.LineTo(q.x.X, q.x.Y-far)
		p.LineTo(0, -far)
		return Analyse("test", p, Default)
	}
	ids := func(a *Analysis) string {
		var s []string
		for _, f := range a.Findings {
			if f.Kind == Jump {
				s = append(s, f.ID())
			}
		}
		return strings.Join(s, ",")
	}
	want := ids(build(600, false))
	if want == "" {
		t.Fatal("no jump")
	}
	for _, c := range []struct {
		far   float64
		extra bool
	}{{900, false}, {600, true}, {400, true}} {
		if got := ids(build(c.far, c.extra)); got != want {
			t.Errorf("far %g, extra contour %v: identity %q, want %q", c.far, c.extra, got, want)
		}
	}
}

func TestBaselineRoundTrip(t *testing.T) {
	var p outline.Path
	q := newPen(&p, 0, 0, 90)
	q.arc(300, -40)
	q.arc(100, -50)
	closeWith(q, &p)
	a := Analyse("test", p, Default)
	es := UpdateBaseline(a.Findings, nil)
	es[0].Intended, es[0].Reason = true, "a deliberate change of radius"
	var buf bytes.Buffer
	if err := FormatBaseline(&buf, "test", es); err != nil {
		t.Fatal(err)
	}
	back, err := ParseBaseline(&buf)
	if err != nil {
		t.Fatal(err)
	}
	newFs, gone, _ := Compare(a.Findings, back)
	if len(newFs) != 0 || len(gone) != 0 {
		t.Errorf("round trip: new %v, gone %v", newFs, gone)
	}
	if back[0].Reason != "a deliberate change of radius" || !back[0].Intended {
		t.Errorf("round trip lost the reason: %+v", back[0])
	}
	// A finding which moves by a few units still matches.
	moved := append([]Finding(nil), a.Findings...)
	moved[0].X += 4
	if newFs, gone, _ := Compare(moved, back); len(newFs) != 0 || len(gone) != 0 {
		t.Errorf("moved by 4 units: new %v, gone %v", newFs, gone)
	}
}

func TestPicture(t *testing.T) {
	var p outline.Path
	q := newPen(&p, 0, 0, 90)
	q.arc(300, -40)
	q.arc(100, -50)
	closeWith(q, &p)
	a := Analyse("test", p, Default)
	var buf bytes.Buffer
	if err := a.Picture(&buf, "test"); err != nil {
		t.Fatal(err)
	}
	s := buf.String()
	for _, want := range []string{"<svg", `class="tooth"`, `class="finding"`, "jump at", "</svg>"} {
		if !strings.Contains(s, want) {
			t.Errorf("picture lacks %q", want)
		}
	}
}
