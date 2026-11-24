// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"math"
	"testing"
)

// runCentre returns the centre of the run of ink of the glyph at the height y
// whose index is i, counted from the left, and the number of runs.
func runCentre(g *Glyph, y float64, i int) (float64, int) {
	rs := runs(g.Path, y)
	if i >= len(rs) {
		return math.NaN(), len(rs)
	}
	return (rs[i][0] + rs[i][1]) / 2, len(rs)
}

// TestChiIsNoCross asks of the Greek chi that the heavy stroke runs steeply down to the right and the
// light stroke leans far over to the right, so the two cross at a narrow
// angle.  Measured as the horizontal run per unit of height between 0.75 of
// the x-height and a quarter of it below the baseline: the heavy
// stroke 0.2 to 0.35 and the light stroke 0.80 to 0.90.
func TestChiIsNoCross(t *testing.T) {
	p := Default()
	g, err := Build(p, "chi")
	if err != nil {
		t.Fatal(err)
	}
	hi, lo := 0.75*p.XHeight, -0.25*p.XHeight
	// At the upper height the heavy stroke is the left run of two, at the
	// lower the right one, and the light stroke the other way round.
	hh, n1 := runCentre(g, hi, 0)
	hl, n2 := runCentre(g, lo, 1)
	lh, n3 := runCentre(g, hi, 1)
	ll, n4 := runCentre(g, lo, 0)
	if n1 != 2 || n2 != 2 || n3 != 2 || n4 != 2 {
		t.Fatalf("chi has %d and %d runs at %.0f and %.0f, want 2 and 2", n1, n2, hi, lo)
	}
	dy := hi - lo
	heavy, light := (hl-hh)/dy, (lh-ll)/dy
	if heavy > 0.45 {
		t.Errorf("the heavy stroke of chi runs %.2f to the right per unit down, want at most 0.45", heavy)
	}
	if light < 0.70 {
		t.Errorf("the light stroke of chi runs %.2f to the right per unit up, want at least 0.70", light)
	}
}

// TestUpsilonShoulder asks of the Greek upsilon a right shoulder which rises
// to the x-height and turns back to the left in a point: the ink reaches the
// x-height, and the ink just under it stays within a tenth of the width of
// the right extreme of the letter (0.84 to 0.98 of the
// right extreme there).
func TestUpsilonShoulder(t *testing.T) {
	p := Default()
	g, err := Build(p, "upsilon")
	if err != nil {
		t.Fatal(err)
	}
	x0, _, x1, y1 := bbox(g.Path)
	if y1 < p.XHeight-2 {
		t.Errorf("upsilon rises to %.0f, want the x-height %.0f", y1, p.XHeight)
	}
	_, right := greekInk(g, 0.98*p.XHeight)
	if right < x1-0.1*(x1-x0) {
		t.Errorf("the ink of upsilon at 0.98 of the x-height ends at %.0f, the right extreme is %.0f", right, x1)
	}
}

// TestDeltaIsNoPartialDifferential asks of delta and the partial differential
// what keeps them apart at 10 pt.  The neck of delta rises from the top of
// its bowl on the left and turns over to the right into the ball: above the
// x-height the ball lies to the right of the neck, and the neck stands well
// left of the right extreme of the bowl.  The back of the partial
// differential rises from the right of its bowl, which it does not join at its
// top, and turns over to the left: the terminal lies to the left of the back,
// and the back stands within 60 of the right extreme of the bowl.  On the
// outlines delta has its ball 150 right of the neck and the neck 150 left of
// the extreme of the bowl, the partial differential its terminal 160 left of
// the back and the back 35 left of the extreme.
func TestDeltaIsNoPartialDifferential(t *testing.T) {
	p := Default()
	delta, err := Build(p, "delta")
	if err != nil {
		t.Fatal(err)
	}
	partial, err := BuildSymbol(p, "u1D715")
	if err != nil {
		t.Fatal(err)
	}
	measure := func(g *Glyph) (terminal, neckToBowl float64) {
		neck, n := runCentre(g, 520, 0)
		if n != 1 {
			t.Fatalf("a neck of %d runs at 520, want 1", n)
		}
		_, bowl := greekInk(g, 0.3*p.XHeight)
		// Above the x-height there are two runs: the neck or back and the
		// terminal, the one which lies farther from the line below.
		a, n := runCentre(g, 600, 0)
		b, m := runCentre(g, 600, 1)
		if n != 2 || m != 2 {
			t.Fatalf("%d runs at 600, want 2", m)
		}
		far := a
		if math.Abs(b-neck) > math.Abs(a-neck) {
			far = b
		}
		return far - neck, bowl - neck
	}
	dt, dn := measure(delta)
	pt, pn := measure(partial)
	if dt < 60 {
		t.Errorf("the ball of delta lies %.0f right of its neck, want at least 60", dt)
	}
	if dn < 100 {
		t.Errorf("the neck of delta stands %.0f left of the right extreme of its bowl, want at least 100", dn)
	}
	if pt > -60 {
		t.Errorf("the terminal of the partial differential lies %.0f right of its back, want at least 60 left", pt)
	}
	if pn > 60 {
		t.Errorf("the back of the partial differential stands %.0f left of the right extreme of its bowl, want at most 60", pn)
	}
}
