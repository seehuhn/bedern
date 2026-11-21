// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The letters b, p and q have the bowl of d: b and p have it on the right of
// their stem, mirrored, q on the left like d.  The tests check that the free
// side of the bowl is that of the o, that the joins and notches stand at their
// heights, that the stroke thickens from the thin parts to the thickest part,
// that the counter is bounded by the stem, and that the serifs are present or
// absent as the table of letters gives.

// bpqGlyph returns the unrounded outline of a glyph of b, p or q, which has an
// outer contour and a counter.
func bpqGlyph(t *testing.T, name string, p design.Params) (outline.Path, float64) {
	t.Helper()
	path, advance, ok := outlineOf(name, p)
	if !ok {
		t.Fatalf("glyph %s not found", name)
	}
	if len(path) != 2 {
		t.Fatalf("%s has %d contours, want an outer contour and a counter", name, len(path))
	}
	if path[0].SignedArea() <= 0 || path[1].SignedArea() >= 0 {
		t.Errorf("%s: the outer contour is not counter-clockwise or the counter not clockwise", name)
	}
	return path, advance
}

// mirrorPath returns the mirror image of the path in the vertical line x = at,
// with each contour reversed to keep its orientation.
func mirrorPath(path outline.Path, at float64) outline.Path {
	m := func(q outline.Point) outline.Point { return outline.Point{X: 2*at - q.X, Y: q.Y} }
	var out outline.Path
	for _, c := range path {
		// The contours of the letters end where they start.
		pts := []outline.Point{c.Start}
		for _, s := range c.Segs {
			pts = append(pts, s.P3)
		}
		r := outline.Contour{Start: m(lastOfContour(c))}
		for i := len(c.Segs) - 1; i >= 0; i-- {
			s := c.Segs[i]
			r.Segs = append(r.Segs, outline.Segment{IsCurve: s.IsCurve, P1: m(s.P2), P2: m(s.P1), P3: m(pts[i])})
		}
		out = append(out, r)
	}
	return out
}

// shiftPath returns the path moved up by dy.
func shiftPath(path outline.Path, dy float64) outline.Path {
	m := func(q outline.Point) outline.Point { return outline.Point{X: q.X, Y: q.Y + dy} }
	var out outline.Path
	for _, c := range path {
		r := outline.Contour{Start: m(c.Start)}
		for _, s := range c.Segs {
			r.Segs = append(r.Segs, outline.Segment{IsCurve: s.IsCurve, P1: m(s.P1), P2: m(s.P2), P3: m(s.P3)})
		}
		out = append(out, r)
	}
	return out
}

// sameSegments reports whether the run of segments of c which starts at from
// is that of want moved right by dx, to a millionth of a unit.
func sameSegments(c outline.Contour, want []outline.Segment, dx float64) bool {
	close := func(a, b outline.Point) bool { return math.Hypot(a.X-b.X-dx, a.Y-b.Y) < 1e-6 }
	for i := 0; i+len(want) <= len(c.Segs); i++ {
		ok := true
		for j, w := range want {
			s := c.Segs[i+j]
			if s.IsCurve != w.IsCurve || !close(s.P3, w.P3) || (s.IsCurve && (!close(s.P1, w.P1) || !close(s.P2, w.P2))) {
				ok = false
				break
			}
		}
		if ok {
			return true
		}
	}
	return false
}

// cornerGap returns the distance from the corner to the nearest piece of the
// contour whose end points both satisfy keep.
func cornerGap(c outline.Contour, corner outline.Point, keep func(q outline.Point) bool) float64 {
	pts := c.Flatten()
	best := math.Inf(1)
	for i, a := range pts {
		b := pts[(i+1)%len(pts)]
		if keep(a) && keep(b) {
			best = math.Min(best, segDist(corner, a, b))
		}
	}
	return best
}

func TestBPQAdvancesAndSides(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	for _, c := range []struct {
		name        string
		r           rune
		left, right sideKind
		want        float64
		sum         float64 // the advance as the sum of the side spaces and the body
	}{
		// 95 + 81 + 332 + 33.
		{"b", 'b', sideStemBP, sideBowlBP, 541, p.BPSideLeft + p.Stem + p.BBowl.Width + p.BPSideRight},
		{"p", 'p', sideStemBP, sideBowlBP, 541, p.BPSideLeft + p.Stem + p.PBowl.Width + p.BPSideRight},
		// 29 + 338 + 81 + 76.
		{"q", 'q', sideBowlDQ, sideStemQ, 524, p.DQSideLeft + p.QBowl.Width + p.Stem + p.QSideRight},
	} {
		_, adv := bpqGlyph(t, c.name, p)
		if adv != c.want || adv != c.sum {
			t.Errorf("advance of %s is %v, want %v, the sum of its side spaces and its body %v", c.name, adv, c.want, c.sum)
		}
		e, ok := lookup(c.name)
		if !ok || e.r != c.r || e.left != c.left || e.right != c.right {
			t.Errorf("table entry of %s: %+v, want rune %c with the sides %v and %v", c.name, e, c.r, c.left, c.right)
		}
	}
	// The advance as a fraction of the advance of n.
	n := design.NAdvance(p)
	for _, c := range []struct {
		name   string
		adv    float64
		lo, hi float64
	}{
		{"b", design.BAdvance(p), 0.906, 1.016},
		{"p", design.PAdvance(p), 0.932, 1.033},
		{"q", design.QAdvance(p), 0.879, 1.000},
	} {
		if r := c.adv / n; r < c.lo || r > c.hi {
			t.Errorf("advance of %s is %.3f of n, want %.3f to %.3f", c.name, r, c.lo, c.hi)
		}
	}
}

// The side of b and p which is the left of the stem has the head serif and, for
// p, the foot serif at the descender line; the bounding-box rule of its own
// takes the serif from BPSideLeft.
func TestBPQBoundingBoxRules(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	for _, name := range []string{"b", "p", "q"} {
		e, _ := lookup(name)
		path, adv := bpqGlyph(t, name, p)
		left, right := e.inkMargins(p)
		x0, _, x1, _ := path.BBox()
		if x0 < left-0.5 || adv-x1 < right-0.5 {
			t.Errorf("%s: ink from %v to %v in the advance %v, want margins %v and %v", name, x0, x1, adv, left, right)
		}
	}
	// The rule for the left of b and p takes the serif from BPSideLeft, and for the
	// right of q from QSideRight: the margin is the side space less SerifJut, and the
	// serif tip is there (b has it AscenderInset further in).
	e, _ := lookup("b")
	l, _ := e.inkMargins(p)
	near(t, "left margin of b", l, p.BPSideLeft-p.SerifJut)
	e, _ = lookup("q")
	_, r := e.inkMargins(p)
	near(t, "right margin of q", r, p.QSideRight-p.SerifJut)
	for _, name := range []string{"b", "p"} {
		path, _ := bpqGlyph(t, name, p)
		x0, _, _, _ := path.BBox()
		near(t, name+": the serif tip against the margin", x0, p.BPSideLeft-p.SerifJut+design.AscenderInset(p)*map[string]float64{"b": 1, "p": 0}[name])
	}
	q, adv := bpqGlyph(t, "q", p)
	_, _, x1, _ := q.BBox()
	near(t, "q: the foot serif tip against the margin", adv-x1, p.QSideRight-p.SerifJut)
}

func TestBStem(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, adv := bpqGlyph(t, "b", p)
	stemLeft, stemRight := design.BStemLeft(p), design.BStemRight(p)
	near(t, "stem width", stemRight-stemLeft, p.AscenderStem)

	x0, y0, x1, y1 := path.BBox()
	near(t, "left end of the head serif", x0, stemLeft-p.SerifJut)
	near(t, "bottom of the bowl", y0, -p.OvershootLower)
	near(t, "ascender", y1, p.Ascender)
	near(t, "right extreme of the bowl", x1, adv-p.BPSideRight)

	// Above the bowl the stem has its two upright edges.
	for y := p.XHeight + p.OvershootLower + 10; y < p.Ascender-p.HeadSerifDrop-p.SerifThick-20; y += 40 {
		xs := path.CrossingsX(y)
		if len(xs) != 2 {
			t.Fatalf("crossings above the bowl at %.0f: %v", y, xs)
		}
		near(t, "left edge above the bowl", xs[0], stemLeft)
		near(t, "right edge above the bowl", xs[1], stemRight)
	}
	// Between the joins the left edge is that of the stem and the right edge is
	// the counter's, which is the stem edge.
	for y := p.BBowl.JoinLower + 20; y < p.BBowl.JoinUpper-20; y += 40 {
		xs := path.CrossingsX(y)
		if len(xs) != 4 {
			t.Fatalf("crossings between the joins at %.0f: %v", y, xs)
		}
		near(t, "left edge between the joins", xs[0], stemLeft)
		near(t, "right edge between the joins", xs[1], stemRight)
	}
	// The head serif is that of l, h and d, on the left at the ascender height.
	checkHeadSerif(t, p, path, stemLeft, stemRight, p.Ascender)
}

// The stem of b ends at the baseline in a plain corner: no foot serif on either
// side and no spur.  The lower
// stroke of the bowl leaves its edge above the baseline, and the corner keeps
// the clearance floor from the stroke.
func TestBBottomLeftIsAPlainCorner(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := bpqGlyph(t, "b", p)
	stemLeft, stemRight := design.BStemLeft(p), design.BStemRight(p)

	for _, y := range []float64{footScanY, 3, 10, 15} {
		xs := path.CrossingsX(y)
		if len(xs) != 4 {
			t.Fatalf("crossings at height %v: %v", y, xs)
		}
		near(t, "left edge of the stem near the baseline", xs[0], stemLeft)
		near(t, "right edge of the stem near the baseline", xs[1], stemRight)
	}
	// The stem stands on the baseline, flat and without a bracket: the bottom
	// of the ink between its edges is at 0, a vertical line through it crosses
	// only the bottom and the top of the stem.
	for _, x := range []float64{stemLeft + 1, (stemLeft + stemRight) / 2, stemRight - 1} {
		ys := path.CrossingsY(x)
		if len(ys) != 2 || math.Abs(ys[0]) > 1e-6 {
			t.Errorf("a vertical line at %.0f crosses at %v, want the stem from the baseline", x, ys)
		}
	}
	// The lower stroke leaves the stem edge at the lower notch, above the
	// baseline, and the bowl reaches below the baseline by the overshoot.
	if p.BBowl.NotchLower <= 0 {
		t.Fatalf("the lower notch of b at %v is not above the baseline", p.BBowl.NotchLower)
	}
	corner := outline.Point{X: stemRight, Y: 0}
	gap := cornerGap(path[0], corner, func(q outline.Point) bool { return q.X > stemRight+0.5 && q.Y < 150 })
	if floor := design.ClearanceFloor(p); gap < floor {
		t.Errorf("the corner of the stem is %.1f from the lower stroke of the bowl, want at least %v", gap, floor)
	}
}

func TestPStem(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, adv := bpqGlyph(t, "p", p)
	stemLeft, stemRight := design.PStemLeft(p), design.PStemRight(p)

	x0, y0, x1, y1 := path.BBox()
	near(t, "left end of the foot serif", x0, stemLeft-p.SerifJut)
	near(t, "descender line", y0, -p.Descender)
	near(t, "top of the bowl", y1, p.XHeight+p.OvershootLower)
	near(t, "right extreme of the bowl", x1, adv-p.BPSideRight)

	// Below the bowl, between the bracket of the feet and the bowl, the stem has
	// its two upright edges and nothing else.
	for y := -p.Descender + p.BracketHeight + 20; y < -p.OvershootLower-20; y += 30 {
		xs := path.CrossingsX(y)
		if len(xs) != 2 {
			t.Fatalf("crossings below the bowl at %.0f: %v", y, xs)
		}
		near(t, "left edge of the stem below the bowl", xs[0], stemLeft)
		near(t, "right edge of the stem below the bowl", xs[1], stemRight)
	}
	// The foot serifs are those of the other letters, on both sides at the
	// descender line: the path is moved up to put the line at 0.
	low := shiftPath(path, p.Descender)
	checkFoot(t, p, low, stemRight, true, p.SerifJut, p.SerifThick)
	checkFoot(t, p, low, stemLeft, false, p.SerifJut, p.SerifThick)
	// The head serif is that of n, at the x-height; below it the upper stroke of
	// the bowl joins the stem, which has no right-hand edge there.
	checkHeadSerifBelow(t, p, path, stemLeft, stemRight, p.XHeight, false)
	// The top of the stem is flat, at the x-height, and the bowl rises above it.
	near(t, "top of the stem", lastOf(path.CrossingsY(stemLeft+p.Stem/2)), p.XHeight)
}

func TestQStem(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	top := p.QStemTop
	path, adv := bpqGlyph(t, "q", p)
	stemLeft := design.QStemLeft(p)
	stemRight := stemLeft + p.Stem

	x0, y0, x1, y1 := path.BBox()
	near(t, "left extreme of the bowl", x0, p.DQSideLeft)
	near(t, "descender line", y0, -p.Descender)
	near(t, "top of the bowl", y1, p.XHeight+p.OvershootLower)
	near(t, "right end of the foot serif", x1, stemRight+p.SerifJut)
	near(t, "right side space of the stem", adv-stemRight, p.QSideRight)

	for y := -p.Descender + p.BracketHeight + 20; y < -p.OvershootLower-20; y += 30 {
		xs := path.CrossingsX(y)
		if len(xs) != 2 {
			t.Fatalf("crossings below the bowl at %.0f: %v", y, xs)
		}
		near(t, "left edge of the stem below the bowl", xs[0], stemLeft)
		near(t, "right edge of the stem below the bowl", xs[1], stemRight)
	}
	low := shiftPath(path, p.Descender)
	checkFoot(t, p, low, stemRight, true, p.SerifJut, p.SerifThick)
	checkFoot(t, p, low, stemLeft, false, p.SerifJut, p.SerifThick)

	// The top of the stem is the spur, see TestQSpur; the right-hand edge is
	// upright up to it.
	for _, y := range []float64{top - 1, top - 8, top - 60} {
		xs := path.CrossingsX(y)
		if len(xs) < 2 {
			t.Fatalf("crossings at %v: %v", y, xs)
		}
		near(t, "right edge of the stem below the top", xs[len(xs)-1], stemRight)
	}
}

func TestPTopCornerKeepsClearOfTheUpperStroke(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := bpqGlyph(t, "p", p)
	x := design.PStemRight(p)
	gap := cornerGap(path[0], outline.Point{X: x, Y: p.XHeight}, func(q outline.Point) bool { return q.X > x+0.5 && q.Y > 300 })
	if floor := design.ClearanceFloor(p); gap < floor {
		t.Errorf("the top corner of the stem of p is %.1f from the upper stroke, want at least %v", gap, floor)
	}
}

// The free side of the bowl of b, p and q is that of the o, from the same curve
// code: the thickest part, the thin parts and the overshoot of the o, and its
// segments, moved by the distance of the extremes.
func TestBPQBowlIsTheBowlOfO(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	o, _, _ := outlineOf("o", p)
	for _, c := range []struct {
		name    string
		right   bool // the free side is on the right
		extreme float64
	}{
		{"b", true, design.BBowlRight(p)},
		{"p", true, design.PBowlRight(p)},
		{"q", false, design.DBowlLeft(p)},
	} {
		path, _ := bpqGlyph(t, c.name, p)
		mid := p.XHeight / 2
		xs, ox := path.CrossingsX(mid), o.CrossingsX(mid)
		if len(xs) != 4 {
			t.Fatalf("%s: crossings at mid height: %v", c.name, xs)
		}
		var thick float64
		if c.right {
			thick = xs[3] - xs[2]
			near(t, c.name+": free extreme", xs[3], c.extreme)
		} else {
			thick = xs[1] - xs[0]
			near(t, c.name+": free extreme", xs[0], c.extreme)
		}
		near(t, c.name+": thickest part", thick, ox[1]-ox[0])
		near(t, c.name+": thickest part against CurveLower", thick, p.CurveLower)

		half := p.OBowlWidth / 2
		xt := c.extreme - half
		if !c.right {
			xt = c.extreme + half
		}
		ys := path.CrossingsY(xt)
		oys := o.CrossingsY(p.SideRoundLeft + half)
		if len(ys) != 4 {
			t.Fatalf("%s: crossings through the extremes: %v", c.name, ys)
		}
		near(t, c.name+": thin part at the top", ys[3]-ys[2], oys[3]-oys[2])
		near(t, c.name+": thin part at the bottom", ys[1]-ys[0], oys[1]-oys[0])
		near(t, c.name+": overshoot at the top", ys[3], oys[3])
		near(t, c.name+": overshoot at the bottom", ys[0], oys[0])

		// Its segments are those of the half of the o.  The o runs counter-clockwise
		// from its right extreme: segments 0 to 3 are right to top, top to left, left
		// to bottom and bottom to right.  The free side of b and p is the right half,
		// segments 3 and 0, as it comes in the order of the contour; that of q is the
		// left half, segments 1 and 2.
		dx := c.extreme - oBowlExtreme(p, c.right)
		for k, run := range [][]outline.Segment{o[0].Segs[1:3], o[1].Segs[1:3]} {
			if c.right {
				run = []outline.Segment{o[k].Segs[3], o[k].Segs[0]}
			}
			if !sameSegments(path[k], run, dx) {
				t.Errorf("%s: the segments of the free side of the o are not in contour %d", c.name, k)
			}
		}
	}
}

// oBowlExtreme returns the position of the right extreme of the o if right is
// set, else of its left extreme.
func oBowlExtreme(p design.Params, right bool) float64 {
	if right {
		return design.OBowlRight(p)
	}
	return design.OBowlLeft(p)
}

func TestBPQCounter(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	o, _, _ := outlineOf("o", p)
	mid := p.XHeight / 2
	ox := o.CrossingsX(mid)
	for _, c := range []struct {
		name        string
		stemEdge    float64
		extreme     float64
		right       bool
		counterLeft float64 // the edge of the counter on the stem side
	}{
		{"b", design.BStemRight(p), design.BBowlRight(p), true, 0},
		{"p", design.PStemRight(p), design.PBowlRight(p), true, 0},
		{"q", design.QStemLeft(p), design.DBowlLeft(p), false, 0},
	} {
		path, _ := bpqGlyph(t, c.name, p)
		cx0, cy0, cx1, cy1 := (outline.Path{path[1]}).BBox()
		// The counter is bounded by the stem edge on the stem side and by the thickest
		// part on the free side; its top and bottom are the thin parts.
		if c.right {
			near(t, c.name+": edge of the counter on the stem", cx0, c.stemEdge)
			near(t, c.name+": free edge of the counter", cx1, c.extreme-p.CurveLower)
		} else {
			near(t, c.name+": edge of the counter on the stem", cx1, c.stemEdge)
			near(t, c.name+": free edge of the counter", cx0, c.extreme+p.CurveLower)
		}
		near(t, c.name+": top of the counter", cy1, p.XHeight+p.OvershootLower-p.ThinLower)
		near(t, c.name+": bottom of the counter", cy0, -p.OvershootLower+p.ThinLower)

		// The width at mid height is 0.86 to 1.08 of that of the o.
		xs := path.CrossingsX(mid)
		w := xs[2] - xs[1]
		if r := w / (ox[2] - ox[1]); r < 0.86 || r > 1.08 {
			t.Errorf("%s: the counter is %.1f wide at mid height against the o's, a ratio of %.3f outside 0.86 to 1.08", c.name, w, r)
		}
	}
}

func TestBPQJoinsAndStroke(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	for _, c := range []struct {
		name  string
		bowl  design.Bowl
		spec  bowlSpec
		right bool
	}{
		{"b", p.BBowl, bBowl(p), true},
		{"p", p.PBowl, pBowl(p), true},
		{"q", p.QBowl, qBowl(p), false},
	} {
		path, _ := bpqGlyph(t, c.name, p)
		x := c.spec.X
		if c.spec.NotchUpper != c.bowl.NotchUpper || c.spec.JoinLower != c.bowl.JoinLower || c.spec.OnRight != c.right {
			t.Fatalf("%s: the bowl does not take its heights from the parameters: %+v", c.name, c.spec)
		}
		vertex := func(ct outline.Contour, y float64) bool {
			if ct.Start.X == x && ct.Start.Y == y {
				return true
			}
			for _, s := range ct.Segs {
				if s.P3.X == x && s.P3.Y == y {
					return true
				}
			}
			return false
		}
		for _, y := range []float64{c.bowl.JoinUpper, c.bowl.JoinLower} {
			if !vertex(path[1], y) {
				t.Errorf("%s: the counter has no corner on the stem at height %v", c.name, y)
			}
		}
		for _, y := range []float64{c.bowl.NotchUpper, c.bowl.NotchLower} {
			if !vertex(path[0], y) {
				t.Errorf("%s: the outer contour has no corner on the stem at height %v", c.name, y)
			}
			// The thickness at right angles to the stroke where it meets the stem.
			if d := thicknessAt(outline.Point{X: x, Y: y}, path[1]); d < 0.9*p.BowlJoinThick || d > 1.3*p.BowlJoinThick {
				t.Errorf("%s: thickness at the join at %v is %.1f, want %v to %v", c.name, y, d, 0.9*p.BowlJoinThick, 1.3*p.BowlJoinThick)
			}
		}
		// The stroke grows from each thin part to the thickest part, measured on the
		// bowl on the left: that of b and p is mirrored in the edge of the stem.
		sp, pth := c.spec, path
		if c.right {
			sp, pth = c.spec.leftForm(), mirrorPath(path, 0)
		}
		checkBowlStroke(t, pth, sp)
	}
}

// b, d, p and q are one bowl seen four ways: the width, the thickness at the
// free extreme, the thin parts and the counter agree, and p has the
// upper stroke of d and q has all the values of d.
func TestBDPQShareOneBowl(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	d, b, pp, q := p.DBowl, p.BBowl, p.PBowl, p.QBowl
	if b.Width != pp.Width || q.Width != d.Width || b.Width >= d.Width {
		t.Errorf("bowl widths d %v, b %v, p %v, q %v, want b and p equal and narrower than d and q", d.Width, b.Width, pp.Width, q.Width)
	}
	if b.Run != d.Run || b.Thick != d.Thick || b.ThickPos != d.ThickPos || b.Top != d.Top {
		t.Errorf("the shape of the bowl of b is not that of d: %+v against %+v", b, d)
	}
	if q != d {
		t.Errorf("q has the bowl %+v, want that of d, %+v", q, d)
	}
	if pp.JoinUpper != d.JoinUpper || pp.NotchUpper != d.NotchUpper {
		t.Errorf("the upper stroke of p joins at %v and %v, want those of d, %v and %v", pp.JoinUpper, pp.NotchUpper, d.JoinUpper, d.NotchUpper)
	}
	if pp.JoinLower != 92 || pp.NotchLower != 40 {
		t.Errorf("the lower stroke of p joins at %v and %v, want 92 and 40", pp.JoinLower, pp.NotchLower)
	}
	if b.JoinUpper != d.JoinUpper || b.NotchUpper != d.NotchUpper {
		t.Errorf("the upper stroke of b joins at %v and %v, want those of d, %v and %v", b.JoinUpper, b.NotchUpper, d.JoinUpper, d.NotchUpper)
	}

	// The same thickness and width at mid height, in the outlines.
	var widths [4]float64
	var thick [4]float64
	mid := p.XHeight / 2
	for i, name := range []string{"b", "d", "p", "q"} {
		path, _ := bpqOrD(t, name, p)
		xs := path.CrossingsX(mid)
		if len(xs) != 4 {
			t.Fatalf("%s: crossings at mid height: %v", name, xs)
		}
		switch name {
		case "b", "p":
			widths[i], thick[i] = xs[2]-xs[1], xs[3]-xs[2]
		default:
			widths[i], thick[i] = xs[2]-xs[1], xs[1]-xs[0]
		}
	}
	near(t, "counter width of p against b", widths[2], widths[0])
	near(t, "counter width of q against d", widths[3], widths[1])
	for i := 1; i < 4; i++ {
		near(t, "thickest part against b", thick[i], thick[0])
	}
}

func bpqOrD(t *testing.T, name string, p design.Params) (outline.Path, float64) {
	if name == "d" {
		return dGlyph(t, p)
	}
	return bpqGlyph(t, name, p)
}

// The letters which reach below the baseline keep the structural properties in
// the tests of the whole table, and the bounding box lies at the descender line
// and not at the baseline: p and q reach it exactly, b and the rest do not go
// below the overshoot.
func TestDescenderBoundingBoxes(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		for _, name := range []string{"p", "q"} {
			path, _, _ := outlineOf(name, np.params)
			_, y0, _, _ := path.BBox()
			if y0 != -np.params.Descender {
				t.Errorf("%s, %s: the lowest ink is at %v, want the descender line %v", name, np.name, y0, -np.params.Descender)
			}
		}
		for _, name := range []string{"b", "d"} {
			path, _, _ := outlineOf(name, np.params)
			_, y0, _, _ := path.BBox()
			if math.Abs(y0+np.params.OvershootLower) > 1e-6 {
				t.Errorf("%s, %s: the lowest ink is at %v, want the overshoot %v", name, np.name, y0, -np.params.OvershootLower)
			}
		}
	}
}

// The stem of b ends at the baseline with nothing below it, p has the foot
// serifs on both sides, and q has none at the top right.
func TestBPQSerifsPresentAndAbsent(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	// b: nothing to the left of the stem at the baseline and none to the right of
	// it except the bowl, whose nearest ink at the baseline is the stroke.
	b, _ := bpqGlyph(t, "b", p)
	if xs := b.CrossingsX(footScanY); xs[0] != design.BStemLeft(p) {
		t.Errorf("b has ink at %.1f, left of its stem at %.1f", xs[0], design.BStemLeft(p))
	}
	// p and q: serifs of SerifJut on both sides at the descender line.
	for _, name := range []string{"p", "q"} {
		path, _ := bpqGlyph(t, name, p)
		xs := path.CrossingsX(-p.Descender + footScanY)
		if len(xs) != 2 {
			t.Fatalf("%s: crossings on the foot: %v", name, xs)
		}
		if w := xs[1] - xs[0]; math.Abs(w-(p.Stem+2*p.SerifJut)) > 1e-6 {
			t.Errorf("%s: the foot is %v wide, want the stem and a serif of %v on each side", name, w, p.SerifJut)
		}
	}
}

// The stem of p stands at the x-height, level with the head serifs of p and u,
// and the spur of q rises QSpur above that level at the right-hand edge only.
func TestQStemTopIsLevelWithTheHeadSerifs(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	pp, _ := bpqGlyph(t, "p", p)
	near(t, "top of the stem of p", lastOf(pp.CrossingsY(design.PStemLeft(p)+p.Stem/2)), p.XHeight)
}

// The stem of q has a small spur at the top right.  The top right is a corner
// between the upright right-hand edge and the straight cut which falls to the
// left, at between 45 and 90 degrees; the tip rises between 4 and 16 above the
// x-height, below the top of the bowl and no higher than the overshoot of the
// round letters; the cut ends at the upper notch of the bowl, which is a corner
// of the upper stroke; and nothing of the top stands to the right of the edge
// of the stem.
func TestQSpur(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			p := np.params
			path, _ := bpqGlyph(t, "q", p)
			stemLeft := design.QStemLeft(p)
			stemRight := stemLeft + p.Stem
			tip := lastOf(path.CrossingsY(stemRight - 0.5))
			if h := tip - p.XHeight; h < 4 || h > 16 {
				t.Errorf("the spur stands %.1f above the x-height, want 4 to 16", h)
			}
			if tip > p.XHeight+p.OvershootLower+0.01 {
				t.Errorf("the spur at %.1f is above the top of the bowl at %.1f", tip, p.XHeight+p.OvershootLower)
			}
			// The cut: the top at a quarter and at three quarters of the stem.
			y1 := lastOf(path.CrossingsY(stemRight - 0.25*p.Stem))
			y2 := lastOf(path.CrossingsY(stemRight - 0.75*p.Stem))
			angle := math.Atan2(y1-y2, 0.5*p.Stem) * 180 / math.Pi
			if corner := 90 - angle; corner < 45 || corner > 90 {
				t.Errorf("the top right is a corner of %.1f degrees, want 45 to 90", corner)
			}
			if y1 >= tip || y2 >= y1 {
				t.Errorf("the top does not fall to the left: %.1f, %.1f, %.1f", tip, y1, y2)
			}
			// Nothing above the foot is to the right of the edge of the stem.
			if hi := lastOf(path.CrossingsY(stemRight + 1)); hi > -p.Descender+p.BracketHeight+1 {
				t.Errorf("ink at %.1f to the right of the stem", hi)
			}
		})
	}
}
