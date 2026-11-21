// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The tests of m check three equal stems, two equal counters, two arches which
// are the arch of n, six foot serifs of which the outer ones are SerifJut and
// the inner ones SerifJutInner long, with a gap of 35 to 72 between the tips of
// neighbouring ones, a head serif on the first stem only, and an advance of
// 1.5 n.

// An mFrame holds the edges of the three stems of m, read at a height of 100.
type mFrame struct {
	p       design.Params
	path    outline.Path
	advance float64
	l, r    [3]float64 // left and right edges of the stems
}

func newMFrame(t *testing.T, p design.Params) *mFrame {
	t.Helper()
	path, advance, ok := outlineOf("m", p)
	if !ok {
		t.Fatal("glyph m not found")
	}
	xs := path.CrossingsX(100)
	if len(xs) != 6 {
		t.Fatalf("m has %d crossings at height 100, want 6: %v", len(xs), xs)
	}
	f := &mFrame{p: p, path: path, advance: advance}
	for i := 0; i < 3; i++ {
		f.l[i], f.r[i] = xs[2*i], xs[2*i+1]
	}
	return f
}

// mParams returns the regular parameters with the counters of m changed, or
// the regular parameters for 0.
func mParams(counter float64) design.Params {
	p := design.Regular()
	if counter != 0 {
		p.MCounter = counter
	}
	return p
}

// archFrameAt returns the samples of the arch of m between the stems i and i+1.
func (f *mFrame) arch(i int) (top, bot []outline.Point) {
	return archSamples(f.path, f.p.XHeight, f.r[i], f.l[i+1])
}

func TestMAdvanceAndStems(t *testing.T) {
	t.Parallel()
	f := newMFrame(t, design.Regular())
	p := f.p
	if f.advance != 810 || f.advance != design.MAdvance(p) {
		t.Errorf("advance %v, want 810", f.advance)
	}
	n := design.NAdvance(p)
	if ratio := f.advance / n; ratio < 1.424 || ratio > 1.556 {
		t.Errorf("m is %.3f of n wide, want 1.424 to 1.556", ratio)
	}
	near(t, "left side", f.l[0], p.SideStemLeft)
	near(t, "right side", f.advance-f.r[2], p.SideStemRight)
	for i := 0; i < 3; i++ {
		near(t, "stem width", f.r[i]-f.l[i], p.Stem)
	}
	// The counters are equal and MCounter wide.
	c1, c2 := f.l[1]-f.r[0], f.l[2]-f.r[1]
	near(t, "first counter", c1, p.MCounter)
	near(t, "second counter", c2, p.MCounter)
	near(t, "counters equal", c1, c2)
}

// TestMCounterIsAParameter checks that the counter of m follows the parameter
// MCounter.
func TestMCounterIsAParameter(t *testing.T) {
	t.Parallel()
	for _, c := range []float64{181, 170, 210} {
		f := newMFrame(t, mParams(c))
		near(t, "counter", f.l[1]-f.r[0], c)
		near(t, "counter", f.l[2]-f.r[1], c)
		if want := design.NAdvance(f.p) + 2*f.p.Stem + 2*c - f.p.NCounter - f.p.Stem; math.Abs(f.advance-want) > 1e-9 {
			t.Errorf("advance %v for counter %v, want %v", f.advance, c, want)
		}
		// For every counter the outline keeps
		// the clearance floor and does not touch itself.
		if d, q, _ := f.path.MinDistance(design.SlitRatio); d < design.ClearanceFloor(f.p) {
			t.Errorf("counter %v: two parts of m are %.1f apart at %v", c, d, q)
		}
		if q := f.path.SelfIntersections(); len(q) > 0 {
			t.Errorf("counter %v: m touches itself at %v", c, q[0])
		}
	}
}

func TestMFeet(t *testing.T) {
	t.Parallel()
	f := newMFrame(t, design.Regular())
	p := f.p
	// Outer feet SerifJut, inner feet SerifJutInner, six in all.
	checkFoot(t, p, f.path, f.l[0], false, p.SerifJut, p.SerifThick)
	checkFoot(t, p, f.path, f.r[0], true, p.SerifJutInner, p.SerifThick)
	checkFoot(t, p, f.path, f.l[1], false, p.SerifJutInner, p.SerifThick)
	checkFoot(t, p, f.path, f.r[1], true, p.SerifJutInner, p.SerifThick)
	checkFoot(t, p, f.path, f.l[2], false, p.SerifJutInner, p.SerifThick)
	checkFoot(t, p, f.path, f.r[2], true, p.SerifJut, p.SerifThick)
	if fs := f.path.CrossingsX(footScanY); len(fs) != 6 {
		t.Errorf("%d crossings at the foot, want 6 (three runs, the feet do not touch): %v", len(fs), fs)
	}
}

// TestMFootGap checks that the inner feet of neighbouring stems have a gap of
// 35 to 72; with the inner serifs of n it is MCounter - 2 SerifJutInner = 61.
func TestMFootGap(t *testing.T) {
	t.Parallel()
	f := newMFrame(t, design.Regular())
	fs := f.path.CrossingsX(footScanY)
	for i := 0; i < 2; i++ {
		gap := fs[2*i+2] - fs[2*i+1]
		near(t, "gap between the feet", gap, design.MFootGap(f.p))
		if gap < 35 || gap > 72 {
			t.Errorf("the gap between the feet is %.1f, want 35 to 72", gap)
		}
	}
	if g := design.MFootGap(f.p); g != 61 {
		t.Errorf("MFootGap = %v, want 61", g)
	}
	// The least distance between two parts of the outline is not smaller than
	// the clearance floor.
	if d, q, _ := f.path.MinDistance(design.SlitRatio); d < design.ClearanceFloor(f.p) {
		t.Errorf("two parts of m are %.1f apart at %v", d, q)
	}
}

func TestMHeadSerif(t *testing.T) {
	t.Parallel()
	f := newMFrame(t, design.Regular())
	checkHeadSerif(t, f.p, f.path, f.l[0], f.r[0], f.p.XHeight)
	// The other stems have none: at the height of the flat top, the stems
	// are covered by the arches, whose outer edges stand above the top of the
	// first stem by the overshoot.
	if cs := f.path.CrossingsX(f.p.XHeight + 5); len(cs) != 4 {
		t.Errorf("%d crossings above the flat top, want 4 (the two arches only): %v", len(cs), cs)
	} else if cs[0] < f.r[0] {
		t.Errorf("ink above the first stem at %.1f: a head serif above the stem", cs[0])
	}
	// Nothing sticks out to the left of the first stem except its own head
	// serif: the ink starts at the tip of the head serif and at the foot.
	x0, _, x1, _ := f.path.BBox()
	near(t, "left of the ink", x0, f.l[0]-f.p.SerifJut)
	near(t, "right of the ink", x1, f.r[2]+f.p.SerifJut)
}

// Both arches have the quantities of the arch of n: the height of the peak, its
// position between the stem centres, the thickness at the peak and where the arch
// leaves the stem, the notch and the join.
func TestMArches(t *testing.T) {
	t.Parallel()
	for _, counter := range []float64{0, 181, 210} {
		f := newMFrame(t, mParams(counter))
		p := f.p
		for i := 0; i < 2; i++ {
			top, bot := f.arch(i)
			if len(top) < 20 {
				t.Fatalf("arch %d: %d samples", i, len(top))
			}
			xt, yt := peakOf(top)
			near(t, "height of the arch", yt, p.XHeight+p.ArchOvershoot)
			c1 := (f.l[i] + f.r[i]) / 2
			d := (f.l[i+1]+f.r[i+1])/2 - c1
			if got := (xt - c1) / d; math.Abs(got-p.ArchPeak) > 0.01 {
				t.Errorf("counter %v, arch %d peaks at %.3f of the stem distance, want %.3f", counter, i, got, p.ArchPeak)
			}
			ys := f.path.CrossingsY(xt)
			if got := ys[len(ys)-1] - ys[len(ys)-2]; math.Abs(got-p.ThinLower) > 3 {
				t.Errorf("arch %d is %.1f thick at the peak, want %.1f within 3", i, got, p.ThinLower)
			}
			// The thickness where the arch leaves the stem, 0.35 stem from it, and
			// that the arch only grows thicker towards the next stem.
			d2 := edgeThickness(top, bot)
			k := 0
			for j, q := range top {
				if math.Abs(q.X-(f.r[i]+0.35*p.Stem)) < math.Abs(top[k].X-(f.r[i]+0.35*p.Stem)) {
					k = j
				}
			}
			if math.Abs(d2[k]-p.ArchThinJoin) > 3 {
				t.Errorf("counter %v, arch %d is %.1f thick where it leaves the stem, want %.1f within 3", counter, i, d2[k], p.ArchThinJoin)
			}
			var run float64
			for j := k; j < len(d2) && top[j].X < f.l[i+1]-10; j++ {
				run = math.Max(run, d2[j])
				if d2[j] < p.ArchThinJoin-1.5 || run-d2[j] > 1.5 {
					t.Errorf("counter %v, arch %d is %.1f thick at x = %.0f after %.1f: a pinch", counter, i, d2[j], top[j].X, run)
					break
				}
			}
			// The notch and the join.
			if n := notchHeight(f.path, f.r[i], p.XHeight); math.Abs(n-p.ArchNotch) > 3 {
				t.Errorf("counter %v, arch %d: notch at %.1f, want %.0f within 3", counter, i, n, p.ArchNotch)
			}
			if j := joinHeight(f.path, p.XHeight, f.r[i]); j < p.ArchJoin || j > p.ArchJoin+10 {
				t.Errorf("counter %v, arch %d: join measured at %.1f, want %.0f to %.0f", counter, i, j, p.ArchJoin, p.ArchJoin+10)
			}
		}
	}
}

// The two arches are one arch: the underside is the same curve and the outer edge
// is the same up to the shoulder, where the first arch is cut short by the notch
// of the second.
func TestMArchesAreEqual(t *testing.T) {
	t.Parallel()
	f := newMFrame(t, design.Regular())
	t1, b1 := f.arch(0)
	t2, b2 := f.arch(1)
	if len(t1) != len(t2) {
		t.Fatalf("the arches have %d and %d samples", len(t1), len(t2))
	}
	shift := f.r[1] - f.r[0]
	for i := range t1 {
		if math.Abs(b1[i].Y-b2[i].Y) > 0.5 || math.Abs(b1[i].X+shift-b2[i].X) > 0.5 {
			t.Errorf("undersides differ at sample %d: %v and %v", i, b1[i], b2[i])
			break
		}
		// The outer edges agree within half a unit on the first 60 per cent.
		if i < len(t1)*6/10 && math.Abs(t1[i].Y-t2[i].Y) > 0.5 {
			t.Errorf("outer edges differ by %.1f at sample %d", math.Abs(t1[i].Y-t2[i].Y), i)
			break
		}
		if math.Abs(t1[i].Y-t2[i].Y) > 6 {
			t.Errorf("outer edges differ by %.1f at sample %d", math.Abs(t1[i].Y-t2[i].Y), i)
			break
		}
	}
}

// The outer edges of the two arches meet at the notch on the right-hand edge of
// the middle stem, at the height ArchNotch, in a corner whose angle is not a slit,
// and nothing is drawn twice along the stem.
func TestMNotchBetweenArches(t *testing.T) {
	t.Parallel()
	f := newMFrame(t, design.Regular())
	p := f.p
	x := f.r[1]
	// The notch is a vertex of the path.
	found := false
	for _, c := range f.path {
		for _, s := range c.Segs {
			if math.Abs(s.P3.X-x) < 0.01 && math.Abs(s.P3.Y-p.ArchNotch) < 0.01 {
				found = true
			}
		}
	}
	if !found {
		t.Errorf("no vertex at the notch (%.0f, %.0f)", x, p.ArchNotch)
	}
	// The white of the notch: just above it the ink on the right of the stem
	// edge is arch 2 and on the left is arch 1; 20 units above the notch the gap
	// between the two outer edges is open by more than a slit.
	ys := func(dx float64) float64 { c := f.path.CrossingsY(x + dx); return c[len(c)-1] }
	up := ys(-0.5)
	if up < p.ArchNotch+3 {
		t.Errorf("the outer edge of the first arch is at %.1f half a unit left of the notch, want it rising above %.0f", up, p.ArchNotch)
	}
	if a, b := ys(15), ys(-15); a < p.ArchNotch+15 || b < p.ArchNotch+15 {
		t.Errorf("15 units either side of the notch the arches stand at %.1f and %.1f, want at least 15 above it", a, b)
	}
	// No two edges of the outline come closer than the clearance floor at the
	// notch (the slit rule).
	if d, q, _ := f.path.MinDistance(design.SlitRatio); d < design.ClearanceFloor(p) {
		t.Errorf("the notch is a slit: %.1f apart at %v", d, q)
	}
	if len(f.path) != 1 {
		t.Errorf("%d contours, want 1", len(f.path))
	}
	if q := f.path.SelfIntersections(); len(q) > 0 {
		t.Errorf("m touches itself at %v", q[0])
	}
}
