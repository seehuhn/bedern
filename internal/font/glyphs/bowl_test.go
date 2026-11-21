// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// testBowl is a bowl unlike the bowl of d: it hangs on a stem whose left edge
// is at 340, its thin parts differ, and its joins are far from those of d.
var testBowl = bowlSpec{
	X: 340, Left: 40, Half: 170,
	Bottom: -20, Top: 500,
	ThinBottom: 26, ThinTop: 30, Thick: 100,
	NotchUpper: 450, JoinUpper: 380,
	NotchLower: 70, JoinLower: 120,
	JoinThick:      36,
	SupernessOuter: 0.73, SupernessInner: 0.76,
}

// bowlLetter draws a plain stem and the bowl on its left, in the order of the
// composition convention: the stem up and down its right-hand edge, and the
// outer edge of the bowl between the notches on its left-hand edge, then the
// counter as a second contour.
func bowlLetter(p design.Params, b bowlSpec, shape bowlShape) outline.Path {
	stem := stemSpec{Left: b.X, Width: 70, Top: 600}
	var path outline.Path
	path.MoveTo(stem.start())
	stemUp(&path, p, stem, stem.Top)
	stemCap(&path, p, stem)
	path.LineTo(b.X, b.NotchUpper)
	bowlOuter(shape, b).appendTo(&path)
	stemDown(&path, p, stem)
	bowlCounter(shape, b).addContour(&path)
	return path
}

func TestBowlContours(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	b := testBowl
	path := bowlLetter(p, b, dBowlShape)
	if len(path) != 2 {
		t.Fatalf("want an outer contour and a counter, got %d contours", len(path))
	}
	if path[0].SignedArea() <= 0 {
		t.Error("the outer contour is not counter-clockwise")
	}
	if path[1].SignedArea() >= 0 {
		t.Error("the counter is not clockwise")
	}
	x0, y0, x1, y1 := path.BBox()
	near(t, "free extreme", x0, b.Left)
	near(t, "bottom with overshoot", y0, b.Bottom)
	near(t, "top with overshoot", y1, 600) // the stem is taller
	_ = x1
	if n := len(path.SelfIntersections()); n != 0 {
		t.Errorf("%d self-intersections", n)
	}
}

func TestBowlFreeSide(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	b := testBowl
	path := bowlLetter(p, b, dBowlShape)
	mid := (b.Top + b.Bottom) / 2

	// The thickest part, at mid height of the outer ellipse.
	xs := path.CrossingsX(mid)
	if len(xs) != 4 {
		t.Fatalf("crossings at mid height: %v", xs)
	}
	near(t, "thickness on the free side", xs[1]-xs[0], b.Thick)
	near(t, "free extreme", xs[0], b.Left)
	// The counter is bounded by the stem edge: its right edge is the stem's
	// left edge.
	near(t, "right edge of the counter", xs[2], b.X)

	// The thin parts, on the vertical through the extremes at top and bottom.
	xt := b.Left + b.Half
	ys := path.CrossingsY(xt)
	if len(ys) != 4 {
		t.Fatalf("crossings on the vertical through the extremes: %v", ys)
	}
	near(t, "thin part at the bottom", ys[1]-ys[0], b.ThinBottom)
	near(t, "thin part at the top", ys[3]-ys[2], b.ThinTop)
	near(t, "overshoot at the bottom", ys[0], b.Bottom)
	near(t, "overshoot at the top", ys[3], b.Top)
}

// thicknessAt returns the distance from the point q to the nearest point of
// the contour c.
func thicknessAt(q outline.Point, c outline.Contour) float64 {
	pts := c.Flatten()
	best := math.Inf(1)
	for i, a := range pts {
		bq := pts[(i+1)%len(pts)]
		best = math.Min(best, segDist(q, a, bq))
	}
	return best
}

func TestBowlJoins(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	b := testBowl
	path := bowlLetter(p, b, dBowlShape)
	outer, counter := path[0], path[1]

	// The outer edge leaves the stem edge at the two notches, and the counter
	// meets it at the two joins.
	hasVertex := func(c outline.Contour, x, y float64) bool {
		if math.Hypot(c.Start.X-x, c.Start.Y-y) < 1e-9 {
			return true
		}
		for _, s := range c.Segs {
			if math.Hypot(s.P3.X-x, s.P3.Y-y) < 1e-9 {
				return true
			}
		}
		return false
	}
	for _, y := range []float64{b.NotchUpper, b.NotchLower} {
		if !hasVertex(outer, b.X, y) {
			t.Errorf("the outer contour has no vertex at (%v, %v)", b.X, y)
		}
	}
	for _, y := range []float64{b.JoinUpper, b.JoinLower} {
		if !hasVertex(counter, b.X, y) {
			t.Errorf("the counter has no vertex at (%v, %v)", b.X, y)
		}
	}

	// Between the joins the counter reaches the stem; above the upper join and
	// below the lower one the stroke has left it.
	mid := path.CrossingsX((b.JoinUpper + b.JoinLower) / 2)
	if len(mid) != 4 || math.Abs(mid[2]-b.X) > 0.01 {
		t.Errorf("between the joins the counter does not reach the stem: %v", mid)
	}
	if hi := path.CrossingsX(b.JoinUpper + 12); len(hi) != 4 || hi[2] > b.X-3 {
		t.Errorf("12 units above the upper join the stroke has not left the stem: %v", hi)
	}
	if lo := path.CrossingsX(b.JoinLower - 12); len(lo) != 4 || lo[2] > b.X-3 {
		t.Errorf("12 units below the lower join the stroke has not left the stem: %v", lo)
	}

	// The thickness at the joins: the distance from the notch to the counter's
	// edge, at right angles to the stroke, is about the join thickness.
	for _, y := range []float64{b.NotchUpper, b.NotchLower} {
		d := thicknessAt(outline.Point{X: b.X, Y: y}, counter)
		if d < 0.9*b.JoinThick || d > 1.3*b.JoinThick {
			t.Errorf("thickness at the join of the notch at %v is %.2f, want %v", y, d, b.JoinThick)
		}
	}
}

// The stroke thickens from each thin extreme to the thickest part on the free
// side, and between the join and the thin extreme it is neither thinner than
// the thinner of the two nor thicker than the thickest part.
func TestBowlStrokeGrowsFromJoinToThickest(t *testing.T) {
	t.Parallel()
	b := testBowl
	checkBowlStroke(t, bowlLetter(design.Regular(), b, dBowlShape), b)
}

// checkBowlStroke measures the stroke of the bowl b in the letter path, whose
// first contour is the outer contour and whose second is the counter.  The
// stroke is measured at right angles, as the distance of a point of the outer
// edge from the counter.
func checkBowlStroke(t *testing.T, path outline.Path, b bowlSpec) {
	t.Helper()
	checkBowlStrokeHalves(t, path, b, b.Left+b.Half)
}

// checkBowlStrokeHalves is checkBowlStroke for a bowl whose extreme at the
// bottom lies at xtLower, not under the extreme at the top (a).
func checkBowlStrokeHalves(t *testing.T, path outline.Path, b bowlSpec, xtLower float64) {
	t.Helper()
	outer, counter := path[0], path[1]
	pts := outer.Flatten()
	xtUpper := b.Left + b.Half
	mid := b.thickY()

	// The outer contour is walked counter-clockwise: the upper notch, the top
	// extreme, the free extreme, the bottom extreme, the lower notch.  Samples
	// on the stem edge, on the serifs of the stem and below the bowl are left out.
	var nearStem [2][]float64 // upper and lower: thicknesses between the stem and the extreme
	var growing []float64     // thicknesses from the top extreme to the free extreme
	for _, q := range pts {
		if q.X >= b.X-0.01 || q.Y > b.Top+0.01 || q.Y < b.Bottom-0.01 {
			continue
		}
		d := thicknessAt(q, counter)
		k, xt := 0, xtUpper
		if q.Y < mid {
			k, xt = 1, xtLower
		}
		if q.X >= xt {
			nearStem[k] = append(nearStem[k], d)
		}
		if k == 0 && q.X <= xt {
			growing = append(growing, d)
		}
	}
	for k, thin := range []float64{b.ThinTop, b.ThinBottom} {
		what := []string{"upper", "lower"}[k]
		if len(nearStem[k]) < 10 {
			t.Fatalf("%s stroke: only %d samples between the stem and the extreme", what, len(nearStem[k]))
		}
		lo, hi := math.Inf(1), math.Inf(-1)
		for _, d := range nearStem[k] {
			lo, hi = math.Min(lo, d), math.Max(hi, d)
		}
		if floor := math.Min(thin, b.JoinThick) - 2; lo < floor {
			t.Errorf("%s stroke pinches to %.1f between join and extreme, below %.1f", what, lo, floor)
		}
		if hi > b.Thick {
			t.Errorf("%s stroke swells to %.1f between join and extreme, above %v", what, hi, b.Thick)
		}
	}

	// From the thin extreme round to the free extreme the stroke does not
	// shrink, and it ends at the thickest part.
	if len(growing) < 10 {
		t.Fatalf("only %d samples between the top extreme and the free extreme", len(growing))
	}
	for i := 1; i < len(growing); i++ {
		if growing[i] < growing[i-1]-0.5 {
			t.Errorf("the stroke shrinks from %.2f to %.2f on the way to the free extreme", growing[i-1], growing[i])
		}
	}
	if end := growing[len(growing)-1]; math.Abs(end-b.Thick) > 2 {
		t.Errorf("the stroke ends at %.1f, want the thickest part %v", end, b.Thick)
	}
	if start := growing[0]; math.Abs(start-b.ThinTop) > 2 {
		t.Errorf("the stroke starts at %.1f, want the thin part %v", start, b.ThinTop)
	}

	// The lower half mirrors it: from the free extreme to the bottom extreme the
	// stroke does not grow, and it ends at the thin part.
	var shrinking []float64
	for _, q := range pts {
		if q.X < b.X-0.01 && q.X <= xtLower && q.Y <= mid {
			shrinking = append(shrinking, thicknessAt(q, counter))
		}
	}
	if len(shrinking) < 10 {
		t.Fatalf("only %d samples between the free extreme and the bottom extreme", len(shrinking))
	}
	for i := 1; i < len(shrinking); i++ {
		if shrinking[i] > shrinking[i-1]+0.5 {
			t.Errorf("the stroke grows from %.2f to %.2f on the way to the bottom extreme", shrinking[i-1], shrinking[i])
		}
	}
	if end := shrinking[len(shrinking)-1]; math.Abs(end-b.ThinBottom) > 2 {
		t.Errorf("the stroke ends at %.1f, want the thin part %v", end, b.ThinBottom)
	}
}

func TestBowlCounterBoundedByStem(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	b := testBowl
	path := bowlLetter(p, b, dBowlShape)
	c := path[1]
	if c.Start != lastOfContour(c) {
		t.Errorf("the counter starts at %v and ends at %v", c.Start, lastOfContour(c))
	}
	x0, y0, x1, y1 := (outline.Path{c}).BBox()
	near(t, "right edge of the counter", x1, b.X)
	near(t, "left edge of the counter", x0, b.Left+b.Thick)
	near(t, "bottom of the counter", y0, b.Bottom+b.ThinBottom)
	near(t, "top of the counter", y1, b.Top-b.ThinTop)
	// A straight vertical segment between the joins lies on the stem edge.
	found := false
	prev := c.Start
	for _, s := range c.Segs {
		if !s.IsCurve && prev.X == b.X && s.P3.X == b.X {
			found = true
			if prev.Y != b.JoinUpper || s.P3.Y != b.JoinLower {
				t.Errorf("the edge on the stem runs from %v to %v, want %v to %v", prev.Y, s.P3.Y, b.JoinUpper, b.JoinLower)
			}
		}
		prev = s.P3
	}
	if !found {
		t.Error("the counter has no straight segment on the stem edge")
	}
}

func TestBowlPartsEndWhereTheyShould(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	b := testBowl
	stem := stemSpec{Left: b.X, Width: 70, Top: 600}
	var path outline.Path
	path.MoveTo(stem.start())
	stemUp(&path, p, stem, stem.Top)
	stemCap(&path, p, stem)
	path.LineTo(b.X, b.NotchUpper)
	bowlOuter(dBowlShape, b).appendTo(&path)
	wantCurrent(t, &path, "bowlOuter", b.X, b.NotchLower)
	stemDown(&path, p, stem)
	wantCurrent(t, &path, "stemDown", b.X, stem.Bottom)
	if path[0].Start != lastPoint(&path) {
		t.Errorf("the outer contour starts at %v and ends at %v", path[0].Start, lastPoint(&path))
	}
	bowlCounter(dBowlShape, b).addContour(&path)
	wantCurrent(t, &path, "bowlCounter", b.X, b.JoinUpper)
}

// A different shape gives a different connection to the stem and the same free
// side.
func TestBowlShapeIsUsed(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	b := testBowl
	shape := dBowlShape
	shape.OuterStemHandle += 0.2
	base := bowlLetter(p, b, dBowlShape)
	other := bowlLetter(p, b, shape)
	x := b.X - 30
	c0, c1 := base.CrossingsY(x), other.CrossingsY(x)
	if len(c0) != len(c1) || math.Abs(c0[len(c0)-1]-c1[len(c1)-1]) < 0.5 {
		t.Error("the outer edge near the stem does not depend on the shape")
	}
	// The free side is not affected.
	f0, f1 := base.CrossingsY(b.Left+b.Half-20), other.CrossingsY(b.Left+b.Half-20)
	for i := range f0 {
		if f0[i] != f1[i] {
			t.Errorf("the free side changed with the shape of the connection: %v and %v", f0, f1)
			break
		}
	}
}

// A bowl whose free extreme lies below the middle, as in a and e, has its
// thickest part there, and the stroke still grows from the thin parts to it.
func TestBowlThickestPartBelowTheMiddle(t *testing.T) {
	t.Parallel()
	b := testBowl
	b.ThickPos = 0.38
	path := bowlLetter(design.Regular(), b, dBowlShape)
	y := b.Bottom + 0.38*(b.Top-b.Bottom)
	xs := path.CrossingsX(y + 0.3)
	near(t, "free extreme", xs[0], b.Left)
	near(t, "thickest part", xs[1]-xs[0], b.Thick)
	// No point of the free side is further left than the extreme at that
	// height: the leftmost point of the outer contour is at the height of the
	// thickest part.
	left, at := math.Inf(1), 0.0
	for _, q := range path[0].Flatten() {
		if q.X < left {
			left, at = q.X, q.Y
		}
	}
	near(t, "height of the free extreme", at, y)
	checkBowlStroke(t, path, b)

	// A bowl with the middle position has a free side which is the left half of
	// one shape.
	m := testBowl
	m.ThickPos = 0.5
	base := bowlLetter(design.Regular(), testBowl, dBowlShape)
	same := bowlLetter(design.Regular(), m, dBowlShape)
	if len(base[0].Segs) != len(same[0].Segs) {
		t.Fatal("different number of segments")
	}
	for i := range base[0].Segs {
		if base[0].Segs[i] != same[0].Segs[i] {
			t.Errorf("segment %d differs between ThickPos 0 and 0.5", i)
		}
	}
}

// mirroredTestBowl returns the bowl which hangs on the right-hand edge at x of
// a stem, made from the bowl b on the left-hand edge at x by the mirror image in
// that edge.
func mirroredTestBowl(b bowlSpec) bowlSpec {
	r := b
	r.OnRight = true
	r.Left = 2*b.X - b.Left
	return r
}

// bowlLetterRight draws a plain stem and the bowl on its right, as the outline
// of b is drawn: up the right-hand edge to the lower notch, the outer edge of
// the bowl up to the upper notch, on up the stem and along its top, down the
// left-hand edge, then the counter as a second contour.
func bowlLetterRight(p design.Params, b bowlSpec, shape bowlShape) outline.Path {
	stem := stemSpec{Left: b.X - 70, Width: 70, Top: 600}
	var path outline.Path
	path.MoveTo(stem.start())
	stemUp(&path, p, stem, b.NotchLower)
	bowlOuter(shape, b).appendTo(&path)
	stemCap(&path, p, stem)
	stemDown(&path, p, stem)
	bowlCounter(shape, b).addContour(&path)
	return path
}

// TestBowlOnRightIsTheMirrorImage checks that the bowl on the right of a stem is
// the mirror image of the ordinary bowl in the edge of the stem, for the outer
// edge and for the counter, both as chains which keep the orientation.
func TestBowlOnRightIsTheMirrorImage(t *testing.T) {
	t.Parallel()
	b := testBowl
	r := mirroredTestBowl(b)
	near(t, "extremes at top and bottom", r.xExtreme(), 2*b.X-b.xExtreme())
	for name, pair := range map[string][2]chain{
		"outer edge": {bowlOuter(dBowlShape, r), bowlOuter(dBowlShape, b).mirrorX(b.X)},
		"counter":    {bowlCounter(dBowlShape, r), bowlCounter(dBowlShape, b).mirrorX(b.X)},
	} {
		got, want := pair[0], pair[1]
		if err := got.connected(); err != nil {
			t.Errorf("%s: %v", name, err)
		}
		if len(got) != len(want) {
			t.Fatalf("%s: %d pieces, want %d", name, len(got), len(want))
		}
		for i := range got {
			if got[i].line != want[i].line {
				t.Errorf("%s: piece %d differs in being a line", name, i)
			}
			for k, pair := range [][2]outline.Point{{got[i].P0, want[i].P0}, {got[i].P1, want[i].P1}, {got[i].P2, want[i].P2}, {got[i].P3, want[i].P3}} {
				if got[i].line && (k == 1 || k == 2) {
					continue
				}
				if math.Hypot(pair[0].X-pair[1].X, pair[0].Y-pair[1].Y) > 1e-9 {
					t.Errorf("%s: piece %d, point %d is %v, want %v", name, i, k, pair[0], pair[1])
				}
			}
		}
	}
}

func TestBowlOnRightContours(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	b := mirroredTestBowl(testBowl)
	path := bowlLetterRight(p, b, dBowlShape)
	if len(path) != 2 {
		t.Fatalf("want an outer contour and a counter, got %d contours", len(path))
	}
	if path[0].SignedArea() <= 0 {
		t.Error("the outer contour is not counter-clockwise")
	}
	if path[1].SignedArea() >= 0 {
		t.Error("the counter is not clockwise")
	}
	if path[0].Start != lastOfContour(path[0]) {
		t.Errorf("the outer contour starts at %v and ends at %v", path[0].Start, lastOfContour(path[0]))
	}
	if n := len(path.SelfIntersections()); n != 0 {
		t.Errorf("%d self-intersections", n)
	}
	// The outer edge runs from the lower notch to the upper notch.
	outer := bowlOuter(dBowlShape, b)
	if outer.start() != (outline.Point{X: b.X, Y: b.NotchLower}) || outer.end() != (outline.Point{X: b.X, Y: b.NotchUpper}) {
		t.Errorf("the outer edge runs from %v to %v", outer.start(), outer.end())
	}
	x0, y0, x1, y1 := path.BBox()
	near(t, "free extreme", x1, b.Left)
	near(t, "left edge of the stem", x0, b.X-70)
	near(t, "bottom with overshoot", y0, b.Bottom)
	near(t, "top", y1, 600)

	// The counter is bounded on the left by the right-hand edge of the stem.
	c := path[1]
	cx0, cy0, cx1, cy1 := (outline.Path{c}).BBox()
	near(t, "left edge of the counter", cx0, b.X)
	near(t, "right edge of the counter", cx1, b.Left-b.Thick)
	near(t, "bottom of the counter", cy0, b.Bottom+b.ThinBottom)
	near(t, "top of the counter", cy1, b.Top-b.ThinTop)
	if c.Start != (outline.Point{X: b.X, Y: b.JoinUpper}) {
		t.Errorf("the counter starts at %v, want the upper join", c.Start)
	}
	found := false
	prev := c.Start
	for _, s := range c.Segs {
		if !s.IsCurve && prev.X == b.X && s.P3.X == b.X {
			found = true
			if prev.Y != b.JoinLower || s.P3.Y != b.JoinUpper {
				t.Errorf("the edge on the stem runs from %v to %v, want %v to %v", prev.Y, s.P3.Y, b.JoinLower, b.JoinUpper)
			}
		}
		prev = s.P3
	}
	if !found {
		t.Error("the counter has no straight segment on the stem edge")
	}
}
