// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"math"
	"testing"
)

// TestNBranch asks that the underside of the arch meets the left stem at
// 0.53 to 0.62 of the x-height, and that the white of the notch above
// the branch, between the outer edge of the arch and the stem, opens at more
// than 23 degrees.
func TestNBranch(t *testing.T) {
	p := Default()
	n1 := p.SideStemLeft + p.Stem/2
	for _, cs := range []struct {
		name string
		c    float64
	}{{"n", n1}, {"m", n1}, {"m", n1 + p.Stem + p.MCounter}, {"h", n1 - p.HLeftLess}} {
		checkBranch(t, p, cs.name, cs.c)
	}
}

// checkBranch checks the branch and the notch of the arch which leaves the
// stem whose centre stands at c1 at ShearHeight.
func checkBranch(t *testing.T, p Params, name string, c1 float64) {
	g := glyph(t, name)
	// Going up the right edge of the left stem, the branch is the height at
	// which the run of the stem first widens into the arch.
	var branch float64
	for y := 150.5; y < p.XHeight; y++ {
		r := stemRun(t, p, g.Path, c1, y)
		if r[1]-r[0] > p.Stem+1 {
			branch = y
			break
		}
	}
	if f := branch / p.XHeight; f < 0.53 || f > 0.62 {
		t.Errorf("%s: the arch leaves the stem at %.0f, %.2f of the x-height, want 0.53 to 0.62", name, branch, f)
	}
	// The notch: above the arch, the white between the stem and the outer
	// edge of the arch opens upwards; its angle is that of the two edges
	// 10 and 25 units above its point, below the white inside the entry,
	// which the centre line of the stem enters 30 units above the point.
	// The point of the notch is the lowest height above the branch at which
	// the run of the stem is the stem alone again.
	var tip float64
	for y := branch + 1; y < p.XHeight; y++ {
		if r := stemRun(t, p, g.Path, c1, y+0.25); r[1]-r[0] < p.Stem+1 {
			tip = y
			break
		}
	}
	if tip == 0 {
		t.Fatalf("%s: no notch above the branch", name)
	}
	w := func(y float64) float64 {
		r := stemRun(t, p, g.Path, c1, y+0.25)
		for _, q := range runs(g.Path, y+0.25) {
			if q[0] > r[1] {
				return q[0] - r[1]
			}
		}
		t.Fatalf("no arch right of the stem at %.0f", y)
		return 0
	}
	// The white grows by w(tip+25)-w(tip+10) over 15 units of height
	// between two edges whose directions differ by the angle of the notch.
	angle := math.Atan((w(tip+25)-w(tip+10))/15) * 180 / math.Pi
	t.Logf("%s: branch %.0f (%.2f), notch %.0f degrees", name, branch, branch/p.XHeight, angle)
	if angle <= 23 {
		t.Errorf("%s: the notch above the branch opens at %.1f degrees, want more than 23", name, angle)
	}
}

// TestNWidths asks that at ShearHeight the sides, the stems and the counter
// of n and of u add up to their set width, and that the counter of n equals
// the white between two n within 2 units: the equal rhythm of the roman.
func TestNWidths(t *testing.T) {
	p := Default()
	for _, name := range []string{"n", "u"} {
		g := glyph(t, name)
		rs := runs(g.Path, p.ShearHeight+0.5)
		if len(rs) != 2 {
			t.Fatalf("%s: %d runs of ink at %v, want 2", name, len(rs), p.ShearHeight)
		}
		left, right := rs[0][0], g.Advance-rs[1][1]
		counter := rs[1][0] - rs[0][1]
		stems := rs[0][1] - rs[0][0] + rs[1][1] - rs[1][0]
		if sum := left + stems + counter + right; math.Abs(sum-g.Advance) > 0.5 {
			t.Errorf("%s: the parts add up to %.1f, want %v", name, sum, g.Advance)
		}
		if name == "n" {
			if gap := left + right; math.Abs(gap-counter) > 2 {
				t.Errorf("n: the white between two n is %.1f, the counter %.1f", gap, counter)
			}
		}
	}
}

// TestUTurn asks that the turn of u overshoots the baseline as the arch of n
// rises above the x-height, and that its rise meets the right stem as high
// above the baseline as the arch of n meets its stem below the x-height,
// within 4 units: the u is the n turned round.
func TestUTurn(t *testing.T) {
	p := Default()
	n, u := glyph(t, "n"), glyph(t, "u")
	_, _, _, ny1 := bbox(n.Path)
	_, uy0, _, _ := bbox(u.Path)
	_ = ny1
	// The bottom of u, away from the exit: the lowest ink left of the right
	// stem.
	c1 := p.SideStemLeft + p.Stem/2
	c2 := c1 + p.Stem + p.NCounter
	low := math.Inf(1)
	for _, pts := range polys(u.Path) {
		for _, q := range pts {
			if q.X < italicX(p, c2, q.Y)-p.Stem/2 {
				low = math.Min(low, q.Y)
			}
		}
	}
	if math.Abs(low+p.Overshoot) > 1.5 || uy0 > low {
		t.Errorf("the turn of u reaches %.1f, want %v", low, -p.Overshoot)
	}
	var join float64
	for y := p.XHeight - 100.5; y > 50; y-- {
		r := stemRun(t, p, u.Path, c2, y)
		if r[1]-r[0] > p.Stem+1 {
			join = y
			break
		}
	}
	if want := p.URise; math.Abs(join-want) > 4 {
		t.Errorf("the rise of u meets the right stem at %.0f, want %.0f", join, want)
	}
}

// TestIDot asks that the dot of i is a circle of the diameter Dot whose centre
// lies on the centre line of the sheared stem.
func TestIDot(t *testing.T) {
	p := Default()
	g := glyph(t, "i")
	if len(g.Path) != 2 {
		t.Fatalf("i has %d contours, want 2", len(g.Path))
	}
	x0, y0, x1, y1 := bbox(g.Path[1:])
	if math.Abs(x1-x0-p.Dot) > 1 || math.Abs(y1-y0-p.Dot) > 1 {
		t.Errorf("the dot of i is %.0f by %.0f, want %v", x1-x0, y1-y0, p.Dot)
	}
	c := p.SideStemLeft + p.Stem/2
	if want := italicX(p, c, (y0+y1)/2); math.Abs((x0+x1)/2-want) > 1 {
		t.Errorf("the dot of i stands at %.1f, want %.1f on the centre line of the stem", (x0+x1)/2, want)
	}
}

// TestNArchStress asks that the arch of n, measured vertically, is thinnest
// at its peak or left of it, never right of it, where the pen held at
// Stress cannot make it thin, and 28 to 34 thick there.
func TestNArchStress(t *testing.T) {
	p := Default()
	g := glyph(t, "n")
	c1 := p.SideStemLeft + p.Stem/2
	c2 := c1 + p.Stem + p.NCounter
	y := p.XHeight
	x0, x1 := italicX(p, c1, y)+p.Stem/2+5, italicX(p, c2, y)-p.Stem/2-5
	least, at := math.Inf(1), 0.0
	peak, peakX := math.Inf(-1), 0.0
	for x := x0; x <= x1; x++ {
		ss := crossings(g.Path, Point{X: x + 0.25, Y: 0}, Point{Y: 1})
		if len(ss) < 2 {
			continue
		}
		top, bottom := ss[len(ss)-1], ss[len(ss)-2]
		if w := top - bottom; w < least {
			least, at = w, x
		}
		if top > peak {
			peak, peakX = top, x
		}
	}
	t.Logf("the arch is thinnest, %.1f, at %.0f, its peak at %.0f", least, at, peakX)
	if at > peakX {
		t.Errorf("the arch of n is thinnest at %.0f, right of its peak at %.0f", at, peakX)
	}
	if least < 28 || least > 34 {
		t.Errorf("the arch of n is %.1f at its thinnest, want 28 to 34", least)
	}
}

// TestMCounters asks that the two counters of m are equal within 1 unit and
// that its parts add up to its set width at ShearHeight.
func TestMCounters(t *testing.T) {
	p := Default()
	g := glyph(t, "m")
	rs := runs(g.Path, p.ShearHeight+0.5)
	if len(rs) != 3 {
		t.Fatalf("m: %d runs of ink at %v, want 3", len(rs), p.ShearHeight)
	}
	a, b := rs[1][0]-rs[0][1], rs[2][0]-rs[1][1]
	if math.Abs(a-b) > 1 {
		t.Errorf("m: the counters are %.1f and %.1f", a, b)
	}
	sum := rs[0][0] + g.Advance - rs[2][1] + a + b
	for _, r := range rs {
		sum += r[1] - r[0]
	}
	if math.Abs(sum-g.Advance) > 0.5 {
		t.Errorf("m: the parts add up to %.1f, want %v", sum, g.Advance)
	}
}

// TestTBar asks that the crossbar of t is a hairline whose top lies on the
// x-height and which reaches CrossbarLeft and CrossbarRight beyond the edges
// of the stem, within 1 unit.
func TestTBar(t *testing.T) {
	p := Default()
	g := glyph(t, "t")
	y := p.XHeight - p.Hair/2
	rs := runs(g.Path, y)
	if len(rs) != 1 {
		t.Fatalf("t: %d runs at the bar, want 1", len(rs))
	}
	stem := stemRun(t, p, g.Path, p.TSideLeft+p.Stem/2, y-p.Hair)
	shift := p.slant()
	if l := stem[0] + shift*p.Hair - rs[0][0]; math.Abs(l-p.CrossbarLeft) > 1 {
		t.Errorf("t: the bar reaches %.1f left of the stem, want %v", l, p.CrossbarLeft)
	}
	if r := rs[0][1] - stem[1] - shift*p.Hair; math.Abs(r-p.CrossbarRight) > 1 {
		t.Errorf("t: the bar reaches %.1f right of the stem, want %v", r, p.CrossbarRight)
	}
	ss := crossings(g.Path, Point{X: rs[0][0] + 10, Y: 0}, Point{Y: 1})
	top, bottom := ss[len(ss)-1], ss[len(ss)-2]
	if math.Abs(top-p.XHeight) > 1 || math.Abs(top-bottom-p.Hair) > 1 {
		t.Errorf("t: the bar runs from %.1f to %.1f, want its top at %v and %v thick", bottom, top, p.XHeight, p.Hair)
	}
}
