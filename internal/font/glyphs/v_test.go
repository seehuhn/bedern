// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The tests of v check a thick diagonal of the thickness of the stem on the left
// and a thin one of the hairline on the right, at the same angle (21.4 degrees),
// meeting in a vertex which is cut flat and dips below the baseline; head serifs
// on each side of each diagonal; an ink width of 480, serifs of 58 and 70, a
// counter of 258 and an advance of 504.

// diagReading is the measurement of one diagonal at two heights: the edges of the
// run through it.
type diagReading struct {
	left, right float64 // the edges at the lower height y0
	lean        float64 // dx/dy of the centre line
}

// width returns the horizontal width of the diagonal.
func (r diagReading) width() float64 { return r.right - r.left }

// thick returns the thickness at right angles.
func (r diagReading) thick() float64 { return r.width() / math.Hypot(1, r.lean) }

// angle returns the angle to the vertical in degrees.
func (r diagReading) angle() float64 { return math.Atan(math.Abs(r.lean)) * 180 / math.Pi }

// readDiagonals reads the diagonals from the runs of the path at the heights y0 and
// y1, where the glyph has the same number of runs, as the number of diagonals.
func readDiagonals(t *testing.T, path outline.Path, y0, y1 float64, n int) []diagReading {
	t.Helper()
	a, b := path.CrossingsX(y0), path.CrossingsX(y1)
	if len(a) != 2*n || len(b) != 2*n {
		t.Fatalf("crossings at %.0f and %.0f: %v and %v, want %d runs", y0, y1, a, b, n)
	}
	var rs []diagReading
	for i := 0; i < n; i++ {
		l0, r0, l1, r1 := a[2*i], a[2*i+1], b[2*i], b[2*i+1]
		rs = append(rs, diagReading{left: l0, right: r0, lean: ((l1+r1)/2 - (l0+r0)/2) / (y1 - y0)})
	}
	return rs
}

// vGlyph returns the outline and the advance of v after checking its contour.
func vGlyph(t *testing.T, p design.Params) (outline.Path, float64) {
	t.Helper()
	path, adv, ok := outlineOf("v", p)
	if !ok {
		t.Fatal("glyph v not found")
	}
	if len(path) != 1 {
		t.Fatalf("v has %d contours, want 1", len(path))
	}
	if path[0].SignedArea() <= 0 {
		t.Error("the contour of v is not counter-clockwise")
	}
	return path, adv
}

// TestVAdvanceAndInk checks the advance, the side spaces and the extent of the ink of v.
func TestVAdvanceAndInk(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, adv := vGlyph(t, p)
	tight(t, "advance", adv, 505)
	tight(t, "advance from side spaces and body", adv, p.DiagSideLeft+p.VInk+p.DiagSideRight)
	if r := adv / design.NAdvance(p); r < 0.788 || r > 1.0 {
		t.Errorf("v is %.3f of n wide, want 0.788 to 1.000", r)
	}
	x0, y0, x1, y1 := path.BBox()
	tight(t, "left space", x0, p.DiagSideLeft)
	tight(t, "right space", adv-x1, p.DiagSideRight)
	tight(t, "ink width", x1-x0, 480)
	if r := (x1 - x0) / p.XHeight; r < 1.02 || r > 1.15 {
		t.Errorf("the ink is %.3f x-heights wide, want 1.02 to 1.15", r)
	}
	tight(t, "top", y1, p.XHeight)
	tight(t, "bottom", y0, -p.DiagOvershoot)
}

// TestVDiagonals checks the thickness, the angle and the straightness of the two diagonals.
func TestVDiagonals(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := vGlyph(t, p)
	rs := readDiagonals(t, path, 150, 330, 2)
	thick, thin := rs[0], rs[1]
	tight(t, "thick diagonal", thick.thick(), p.Stem)
	tight(t, "thin diagonal", thin.thick(), p.Hair)
	tight(t, "thick is the stem", thick.thick(), 81)
	tight(t, "thin is the hairline", thin.thick(), 36)
	if r := thin.thick() / thick.thick(); r < 0.38 || r > 0.88 {
		t.Errorf("the thin diagonal is %.3f of the thick one, want 0.38 to 0.88", r)
	}
	// The angles: the same for both, the one the widths give, from 18.9 to 23.2.
	tight(t, "angle of the thick diagonal", thick.angle(), design.VAngle(p))
	tight(t, "angle of the thin diagonal", thin.angle(), design.VAngle(p))
	for _, r := range []diagReading{thick, thin} {
		if a := r.angle(); a < 18.9 || a > 23.2 {
			t.Errorf("angle %.2f degrees, want 18.9 to 23.2", a)
		}
	}
	if thick.lean >= 0 || thin.lean <= 0 {
		t.Errorf("the thick diagonal leans %.3f and the thin %.3f, want the thick one to the right going down", thick.lean, thin.lean)
	}
	// Both are straight: the edges are lines between the serifs and the flat.
	for _, span := range [][2]float64{{125, 355}, {-8, 110}} {
		for _, e := range edgeSamples2(path, span[0], span[1], 12) {
			if e > 1e-6 {
				t.Errorf("an edge of v deviates from a straight line by %v between %v and %v", e, span[0], span[1])
			}
		}
	}
}

// edgeSamples2 returns, for each edge of the runs of the path at the heights from y0
// to y1 in n steps, the greatest deviation of the edge from the line through its
// first and last sample.
func edgeSamples2(path outline.Path, y0, y1 float64, n int) []float64 {
	var cols [][]float64
	for i := 0; i <= n; i++ {
		cols = append(cols, path.CrossingsX(y0+(y1-y0)*float64(i)/float64(n)))
	}
	var devs []float64
	for k := range cols[0] {
		first, last := cols[0][k], cols[n][k]
		worst := 0.0
		for i := 1; i < n; i++ {
			if k >= len(cols[i]) {
				return append(devs, math.Inf(1))
			}
			want := first + (last-first)*float64(i)/float64(n)
			worst = math.Max(worst, math.Abs(cols[i][k]-want))
		}
		devs = append(devs, worst)
	}
	return devs
}

// TestVCounterAndCrotch checks the counter at the x-height and the height of the crotch.
func TestVCounterAndCrotch(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := vGlyph(t, p)
	rs := readDiagonals(t, path, 150, 330, 2)
	// The counter at the x-height, from the extended edges.
	ext := func(r diagReading, y float64, right bool) float64 {
		x := r.left
		if right {
			x = r.right
		}
		return x + r.lean*(y-150)
	}
	counter := ext(rs[1], p.XHeight, false) - ext(rs[0], p.XHeight, true)
	tight(t, "counter at the x-height", counter, design.VCounter(p))
	if counter < 252 || counter > 318 {
		t.Errorf("counter %.1f, want 252 to 318", counter)
	}
	if r := counter / p.XHeight; r < 0.551 || r > 0.679 {
		t.Errorf("counter %.3f x-heights, want 0.551 to 0.679", r)
	}
	// The crotch: where the two runs become one.
	got := math.NaN()
	for y := 300.0; y > 0; y -= 0.02 {
		if len(path.CrossingsX(y)) == 2 {
			got = y
			break
		}
	}
	if math.Abs(got-design.VCrotch(p)) > 0.1 {
		t.Errorf("the diagonals join at %.2f, want the crotch at %.2f", got, design.VCrotch(p))
	}
	// The angle of the crotch is twice that of a diagonal and wider than the slit
	// angle.
	if a := 2 * design.VAngle(p); a <= 2*math.Asin(1/design.SlitRatio)*180/math.Pi {
		t.Errorf("the crotch is %.1f degrees, not wider than the slit angle", a)
	}
}

// TestVVertex checks the depth, the flat and the width of the vertex.
func TestVVertex(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := vGlyph(t, p)
	_, y0, _, _ := path.BBox()
	tight(t, "depth of the vertex below the baseline", -y0, p.DiagOvershoot)
	if -y0 < 0 || -y0 > 20 {
		t.Errorf("the vertex dips %.1f below the baseline, want 0 to 20", -y0)
	}
	// The flat: a straight segment at the lowest height, of the flat width.
	ct := path[0]
	nodes := contourNodes(ct)
	flat := -1
	for i, s := range ct.Segs {
		if !s.IsCurve && nodes[i].Y == y0 && s.P3.Y == y0 {
			flat = i
		}
	}
	if flat < 0 {
		t.Fatal("no flat segment at the bottom")
	}
	tight(t, "width of the flat", ct.Segs[flat].P3.X-nodes[flat].X, p.DiagVertexFlat)
	tight(t, "middle of the flat", (ct.Segs[flat].P3.X+nodes[flat].X)/2, design.VCentre(p))
	// The width above the lowest point: 11 to 69 at 3 and 21 to 74 at 10 above it.
	lean := design.VLean(p)
	for _, c := range []struct {
		up, lo, hi float64
	}{{3, 11, 69}, {10, 21, 74}} {
		xs := path.CrossingsX(y0 + c.up)
		w := xs[len(xs)-1] - xs[0]
		tight(t, "width above the flat", w, p.DiagVertexFlat+2*lean*c.up)
		if w < c.lo || w > c.hi {
			t.Errorf("the vertex is %.1f wide %v above its lowest point, want %v to %v", w, c.up, c.lo, c.hi)
		}
	}
}

// TestVBrackets checks that the serifs of v have four brackets and no other curve.
func TestVBrackets(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := vGlyph(t, p)
	curves := 0
	for _, s := range path[0].Segs {
		if s.IsCurve {
			curves++
		}
	}
	if curves < 4 || curves > 8 {
		t.Errorf("%d curved segments, want one or two for each of the four brackets", curves)
	}
}
