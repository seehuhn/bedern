// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The tests of u check the stems and the counter of n, the arch of n turned round
// with the overshoot of the o below the baseline and the thin join at the right
// stem, a head serif on the left of both stems at the x-height, no foot on the
// left stem and a foot serif to the right on the right stem, the advance of n.

// A uFrame holds the edges of the two stems of u, read at the height 300, and
// the outline.
type uFrame struct {
	p              design.Params
	path           outline.Path
	advance        float64
	l0, l1, r0, r1 float64
}

func newUFrame(t *testing.T, p design.Params) *uFrame {
	t.Helper()
	path, advance, ok := outlineOf("u", p)
	if !ok {
		t.Fatal("glyph u not found")
	}
	xs := path.CrossingsX(300)
	if len(xs) != 4 {
		t.Fatalf("u has %d crossings at height 300, want 4: %v", len(xs), xs)
	}
	return &uFrame{p, path, advance, xs[0], xs[1], xs[2], xs[3]}
}

// bowl samples the bowl between the stems: the lowest crossing of the vertical
// scan is its outer edge, the next the inner edge, in the columns between the
// stems.
func (f *uFrame) bowl() (outer, inner []outline.Point) {
	for x := f.l1 + 0.5; x < f.r0-0.25; x += 0.5 {
		ys := f.path.CrossingsY(x)
		if len(ys) >= 2 && ys[1] < 0.7*f.p.XHeight {
			outer = append(outer, outline.Point{X: x, Y: ys[0]})
			inner = append(inner, outline.Point{X: x, Y: ys[1]})
		}
	}
	return outer, inner
}

// TestUAdvanceAndStems checks the advance, the stems and the counter of u.
func TestUAdvanceAndStems(t *testing.T) {
	t.Parallel()
	f := newUFrame(t, design.Regular())
	p := f.p
	if f.advance != 540 || f.advance != design.UAdvance(p) || f.advance != design.NAdvance(p) {
		t.Errorf("advance %v, want 540, that of n", f.advance)
	}
	near(t, "left side", f.l0, p.SideStemLeft)
	near(t, "left stem width", f.l1-f.l0, p.Stem)
	near(t, "counter", f.r0-f.l1, p.UCounter)
	near(t, "counter against n", f.r0-f.l1, p.NCounter)
	near(t, "right stem width", f.r1-f.r0, p.Stem)
	near(t, "right side", f.advance-f.r1, p.SideStemRight)
	if len(f.path) != 1 {
		t.Errorf("%d contours, want 1", len(f.path))
	}
	// Both stems end in the same flat top at the x-height.
	for _, x := range []float64{(f.l0 + f.l1) / 2, (f.r0 + f.r1) / 2} {
		near(t, "top of the stem", lastOf(f.path.CrossingsY(x)), p.XHeight)
	}
}

// TestUBowl checks that the bowl at the bottom is the arch of n turned round: it
// overshoots the baseline by the overshoot of the o, is ArchThinJoin thick where
// it joins the right stem and runs thick into the left stem, and its lowest point
// is the mirror image, in both axes, of the peak of the arch of n.
func TestUBowl(t *testing.T) {
	t.Parallel()
	f := newUFrame(t, design.Regular())
	p := f.p
	_, y0, _, y1 := f.path.BBox()
	near(t, "lowest point", y0, -p.OvershootLower)
	near(t, "highest point", y1, p.XHeight)

	outer, inner := f.bowl()
	if len(outer) < 20 {
		t.Fatalf("%d samples of the bowl", len(outer))
	}
	// The lowest point, as a fraction of the distance of the stem centres from
	// the right stem: the peak of the arch of n.
	var xl, yl float64 = 0, 1e9
	var sum, cnt float64
	for _, q := range outer {
		yl = math.Min(yl, q.Y)
	}
	for _, q := range outer {
		if q.Y <= yl+0.5 {
			sum += q.X
			cnt++
		}
	}
	xl = sum / cnt
	cl, cr := (f.l0+f.l1)/2, (f.r0+f.r1)/2
	if got := (cr - xl) / (cr - cl); math.Abs(got-p.ArchPeak) > 0.01 {
		t.Errorf("the lowest point is at %.3f of the distance from the right stem, want %.3f", got, p.ArchPeak)
	}
	// Thickness at the lowest point.
	ys := f.path.CrossingsY(xl)
	near(t, "bowl thickness at its lowest point", ys[1]-ys[0], p.ThinLower)

	// Thickness at right angles: ArchThinJoin 0.35 stem from the right stem, and
	// growing from there to the left stem, as the arch of n grows towards its right
	// stem.
	d := edgeThickness(outer, inner)
	at := func(x float64) float64 {
		k := 0
		for i, q := range outer {
			if math.Abs(q.X-x) < math.Abs(outer[k].X-x) {
				k = i
			}
		}
		return d[k]
	}
	right := at(f.r0 - 0.35*p.Stem)
	if math.Abs(right-p.ArchThinJoin) > 3 {
		t.Errorf("the bowl is %.1f thick 0.35 stem from the right stem, want %.1f within 3", right, p.ArchThinJoin)
	}
	left := at(f.l1 + 0.35*p.Stem)
	t.Logf("thickness at the right join %.1f, at the left join %.1f", right, left)
	if left < right+4 {
		t.Errorf("the bowl is %.1f thick near the left stem and %.1f near the right one: it should run thick into the left stem", left, right)
	}
	// From the right join to the left stem the thickness only grows (no pinch).
	var run float64
	for i := len(d) - 1; i >= 0 && outer[i].X > f.l1+10; i-- {
		if outer[i].X > f.r0-0.35*p.Stem {
			continue
		}
		run = math.Max(run, d[i])
		if d[i] < p.ArchThinJoin-1.5 || run-d[i] > 1.5 {
			t.Errorf("at x = %.0f the bowl is %.1f thick after %.1f: a pinch", outer[i].X, d[i], run)
			break
		}
	}
}

// TestUJoinsAndShoulders checks the heights at which the bowl meets the stems:
// the inner edge meets the right stem at UJoin and the outer edge at UNotch, the
// outer edge leaves the left stem at UShoulder and the inner edge at
// UShoulderInner.
func TestUJoinsAndShoulders(t *testing.T) {
	t.Parallel()
	f := newUFrame(t, design.Regular())
	p := f.p
	// The notch: the outer edge 2 and 5 units left of the right stem, extended to
	// the stem; it lies on the bottom of the ink.
	bottom := func(dx float64) float64 { return f.path.CrossingsY(f.r0 + dx)[0] }
	y2, y5 := bottom(-2), bottom(-5)
	if notch := y2 + (y2-y5)*2/3; math.Abs(notch-p.UNotch) > 3 {
		t.Errorf("notch at the right stem %.1f, want %.0f within 3", notch, p.UNotch)
	}
	// The join: going down from 0.65 of the x-height, the height at which the
	// right end of the counter has moved more than 4 from the stem.
	join := math.NaN()
	for y := 0.65 * p.XHeight; y > 0; y -= 0.25 {
		xs := f.path.CrossingsX(y)
		if len(xs) != 4 || xs[2] < f.r0-4 {
			join = y
			break
		}
	}
	if join > p.UJoin || join < p.UJoin-10 {
		t.Errorf("join at the right stem measured at %.1f, want %.0f to %.0f", join, p.UJoin-10, p.UJoin)
	}
	// The left stem: the outer edge is straight down to UShoulder, the inner edge
	// up from UShoulderInner; they leave it with zero curvature, slowly.
	for _, c := range []struct {
		name  string
		y     float64
		index int
		edge  float64
		above bool
	}{
		{"outer edge", p.UShoulder, 0, f.l0, true},
		{"inner edge", p.UShoulderInner, 1, f.l1, false},
	} {
		xs := f.path.CrossingsX(c.y + 5)
		near(t, c.name+" above its shoulder", xs[c.index], c.edge)
		xs = f.path.CrossingsX(c.y - 50)
		if math.Abs(xs[c.index]-c.edge) < 1 {
			t.Errorf("the %s is still at the stem edge 50 units below its shoulder: %v", c.name, xs)
		}
	}
}

// TestUSerifs checks the head serifs of both stems and the foot serif of the right stem.
func TestUSerifs(t *testing.T) {
	t.Parallel()
	f := newUFrame(t, design.Regular())
	p := f.p
	// Head serifs on the left of both stems.
	checkHeadSerif(t, p, f.path, f.l0, f.l1, p.XHeight)
	checkHeadSerif(t, p, f.path, f.r0, f.r1, p.XHeight)
	// The right stem ends in a foot to the right only, as d.
	checkFoot(t, p, f.path, f.r1, true, p.SerifJut, p.SerifThick)
	fs := f.path.CrossingsX(footScanY)
	near(t, "ink at the foot ends at the tip of the foot", fs[len(fs)-1], f.r1+p.SerifJut)
	near(t, "no serif left of the right stem at the baseline", nearest(t, "left of the right foot", f.path.CrossingsX(p.BracketHeight+1), f.r0), f.r0)
	// The left stem has no foot: at the foot the ink starts at the bowl, to the
	// right of the left edge, and the bottom of the outline is the bowl.
	if fs[0] < f.l0-0.5 {
		t.Errorf("ink at %.1f left of the left stem at the baseline: a foot serif", fs[0])
	}
	x0, _, _, _ := f.path.BBox()
	near(t, "leftmost ink is the head serif", x0, f.l0-p.SerifJut)
}
