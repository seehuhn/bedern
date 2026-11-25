// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/fair"
	"seehuhn.de/go/bedern/internal/outline"
	"seehuhn.de/go/bedern/internal/stroke"
)

// bowlLetters are the letters with the bowl of a (that of g a little
// narrower).
var bowlLetters = []string{"a", "d", "q", "g"}

// penProfile returns, for the outer edge of a bowl in the upright
// description, the thickness across the stroke (along the inward normal of
// the edge) at its thickest on the left side and the height there, the
// height of the left extreme of the edge, the thickness at its thinnest on
// the bottom and the abscissa there, and the abscissa of the lowest point.
// The left side is the edge left of cx between the heights 60 and 60 below
// the x-height, the bottom the edge between cx-120 and cx+80 below the
// middle of the x-height and above the overshoot.
func penProfile(p Params, up outline.Path, cx float64) (thick, thickY, leftY, thin, thinX, lowX float64) {
	a := fair.Analyse("bowl", up, fair.Italic)
	leftX, lowY := math.Inf(1), math.Inf(1)
	thin = math.Inf(1)
	for _, c := range a.Contours {
		for _, s := range c.Samples {
			if s.P.X < cx && s.P.Y > 60 && s.P.Y < p.XHeight-60 && s.P.X < leftX {
				leftX, leftY = s.P.X, s.P.Y
			}
			if s.P.X > cx-120 && s.P.X < cx+80 && s.P.Y < p.XHeight/2 && s.P.Y > -p.Overshoot-3 && s.P.Y < lowY {
				lowY, lowX = s.P.Y, s.P.X
			}
		}
	}
	// The places within half a unit of the thickest and the thinnest, and
	// the middle of each.
	var side, bottom []fair.Thick
	for _, c := range a.Contours {
		for _, th := range c.Thick {
			q := th.P
			switch {
			case !th.Valid:
			case q.X < cx-60 && q.Y > 60 && q.Y < p.XHeight-60:
				side = append(side, th)
				thick = math.Max(thick, th.T)
			case q.X > cx-120 && q.X < cx+80 && q.Y < 0 && q.Y > -p.Overshoot-3:
				bottom = append(bottom, th)
				thin = math.Min(thin, th.T)
			}
		}
	}
	lo, hi := math.Inf(1), math.Inf(-1)
	for _, th := range side {
		if th.T >= thick-0.5 {
			lo, hi = math.Min(lo, th.P.Y), math.Max(hi, th.P.Y)
		}
	}
	thickY = (lo + hi) / 2
	lo, hi = math.Inf(1), math.Inf(-1)
	for _, th := range bottom {
		if th.T <= thin+0.5 {
			lo, hi = math.Min(lo, th.P.X), math.Max(hi, th.P.X)
		}
	}
	thinX = (lo + hi) / 2
	return
}

// bowlEdge returns the left edge of the stem of the letter with the bowl
// of a, and the reach of its bowl.
func bowlEdge(p Params, name string) (float64, float64) {
	if name == "g" {
		r := math.Round(p.GBowl * p.BowlReach)
		return p.SideRoundLeft + r, r
	}
	return p.bowlStemEdge(), p.BowlReach
}

// bowlStressTol is how far above the left extreme of a bowl its thickest
// place may lie: the pen's thickest place lies within a few units of the
// extreme, and the test catches a stress which leans the wrong way.
const bowlStressTol = 5.0

// TestBowlStress asks that the bowl of a, d and q has the stress of the pen
// of o in the upright description, where the pen is given: the left side is
// thickest below its left extreme (within bowlStressTol) and 78 to 90 thick there, and the bottom
// is thinnest right of its lowest point and 28 to 34 thick there, the
// thickness measured across the stroke at right angles to its outer edge
// (on which o measures 79.6, where its rays give 80 to 90), each place the
// middle of the places within half a unit of the extreme.  The top of the bowl thins from the
// topmost point into the join with the stem, which is BowlJoinW thick.
func TestBowlStress(t *testing.T) {
	p := Default()
	for _, name := range bowlLetters {
		e, reach := bowlEdge(p, name)
		cx := e - reach/2
		up := unsheared(p, glyph(t, name).Path)
		thick, thickY, leftY, thin, thinX, lowX := penProfile(p, up, cx)
		if thickY > leftY+bowlStressTol {
			t.Errorf("%s: the left side of the bowl is thickest at the height %.0f, not below its left extreme at %.0f", name, thickY, leftY)
		}
		if thick < 78 || thick > 90 {
			t.Errorf("%s: the left side of the bowl is %.1f thick at its thickest, want 78 to 90", name, thick)
		}
		if thinX <= lowX {
			t.Errorf("%s: the bottom of the bowl is thinnest at %.0f, not right of its lowest point at %.0f", name, thinX, lowX)
		}
		if thin < 28 || thin > 34 {
			t.Errorf("%s: the bottom of the bowl is %.1f thick at its thinnest, want 28 to 34", name, thin)
		}
	}
}

// TestBowlJoins asks where the bowl meets the stem: the white between the
// bowl and the stem, pointed at the lower right, is at least 25 wide 20
// above its point, and its point lies above the turn of the exit on a and d; at the
// top the outer edge of the bowl meets the stem below the flat cut, leaving
// a notch of white; and the bowl overshoots the x-height and the
// baseline as o does.
func TestBowlJoins(t *testing.T) {
	p := Default()
	for _, name := range bowlLetters {
		e, _ := bowlEdge(p, name)
		up := unsheared(p, glyph(t, name).Path)
		// The point of the white: going down from the middle of the bowl,
		// the height at which the run of the stem stops beginning at the
		// stem's edge.
		stemRun := func(y float64) [2]float64 {
			for _, r := range runs(up, y+0.5) {
				if r[0] <= e+p.Stem/2 && r[1] >= e+p.Stem/2 {
					return r
				}
			}
			return [2]float64{}
		}
		point := math.NaN()
		for y := p.XHeight / 2; y > 0; y-- {
			if stemRun(y)[0] < e-1 {
				point = y + 1
				break
			}
		}
		if math.IsNaN(point) {
			t.Fatalf("%s: no point of the white between the bowl and the stem", name)
		}
		if (name == "a" || name == "d") && point < p.ExitTurn {
			t.Errorf("%s: the white between the bowl and the stem ends at %.0f, below the turn of the exit at %v", name, point, p.ExitTurn)
		}
		var w float64
		for _, r := range runs(up, point+20.5) {
			if r[1] < e {
				w = e - r[1]
			}
		}
		if w < 25 {
			t.Errorf("%s: the white between the bowl and the stem is %.1f wide 20 above its point at %.0f, want at least 25", name, w, point)
		}
		// The notch at the top: just under the flat cut, the ink of the bowl
		// and that of the stem are apart.
		rs := runs(up, p.XHeight-2.5)
		var notch bool
		for i := 0; i+1 < len(rs); i++ {
			if rs[i][1] < e+1 && rs[i+1][0] >= e-1 {
				notch = true
			}
		}
		if !notch {
			t.Errorf("%s: the top of the bowl runs into the flat cut of the stem without a notch (runs %v)", name, rs)
		}
		// The bowl is the ink left of the stem, whose flat cut stands higher.
		_, y0, _, _ := bbox(up)
		y1 := y0
		for y := y0; y < p.XHeight+2*p.Overshoot; y++ {
			if rs := runs(up, y+0.5); len(rs) > 0 && rs[0][0] < e-p.Stem {
				y1 = y + 1
			}
		}
		if name == "a" && (math.Abs(y1-p.XHeight-p.Overshoot) > 1 || math.Abs(-y0-p.Overshoot) > 1) {
			t.Errorf("%s: the bowl reaches %.1f and %.1f, want the top and the bottom of o", name, y1, y0)
		}
	}
}

// lowJoin returns the thickness of the lower join of d in stems, the least distance from the
// outer edge of the lower stroke, 0.35 stems left of the left edge of the stem at
// the height of the widest chord of the counter, to the lower edge of the
// counter, over the stem.  The outline is the glyph as it stands in the font,
// sheared.
func lowJoin(g *Glyph, p Params) float64 {
	// heights of the counter: the widest chord between the first two runs
	best, yc := 0.0, 0.0
	var ys []float64
	chord := func(y float64) (float64, float64, float64, float64, bool) {
		rs := runs(g.Path, y)
		if len(rs) < 2 {
			return 0, 0, 0, 0, false
		}
		// the gap between the first run and the second (left of the stem)
		return rs[0][1], rs[1][0], rs[1][0], rs[1][1], true
	}
	for y := 60.0; y < p.XHeight-60; y += 1 {
		l, r, _, _, ok := chord(y)
		if ok && r-l > best {
			best = r - l
		}
	}
	for y := 60.0; y < p.XHeight-60; y += 1 {
		l, r, _, _, ok := chord(y)
		if ok && r-l >= best-1 {
			ys = append(ys, y)
		}
	}
	for _, y := range ys {
		yc += y
	}
	yc /= float64(len(ys))
	h0, a2, _, b2, _ := chord(yc)
	stem := b2 - a2
	var X, LO, LOB []float64
	for x := h0 + 0.5; x < a2-0.25; x += 0.5 {
		ys := crossings(g.Path, Point{X: x, Y: 0}, Point{Y: 1})
		var rs [][2]float64
		for i := 0; i+1 < len(ys); i += 2 {
			rs = append(rs, [2]float64{ys[i], ys[i+1]})
		}
		for k := 0; k+1 < len(rs); k++ {
			if rs[k][1] < yc && yc < rs[k+1][0] {
				X = append(X, x)
				LO = append(LO, rs[k][1])
				LOB = append(LOB, rs[k][0])
				break
			}
		}
	}
	if len(X) < 20 {
		return math.NaN()
	}
	k0, bd := 0, math.Inf(1)
	for i := range X {
		if d := math.Abs(X[i] - (a2 - 0.35*stem)); d < bd {
			k0, bd = i, d
		}
	}
	// nearest point on the polyline (X, LO)
	best = math.Inf(1)
	px, py := X[k0], LOB[k0]
	for i := 0; i+1 < len(X); i++ {
		ax, ay, bx, by := X[i], LO[i], X[i+1], LO[i+1]
		dx, dy := bx-ax, by-ay
		t := ((px-ax)*dx + (py-ay)*dy) / (dx*dx + dy*dy)
		t = math.Max(0, math.Min(1, t))
		d := math.Hypot(px-(ax+t*dx), py-(ay+t*dy))
		best = math.Min(best, d)
	}
	return best / stem
}

// lowJoinMax is the thickest lower join of d, in stems, which TestBowlLowJoin
// accepts.
const lowJoinMax = 0.55

// TestBowlLowJoin asks that the lower join of the bowl of d is no thicker than
// lowJoinMax stems.
func TestBowlLowJoin(t *testing.T) {
	p := Default()
	g := glyph(t, "d")
	if w := lowJoin(g, p); !(w <= lowJoinMax) {
		t.Errorf("d: the lower join of the bowl is %.2f stems thick, want at most %.2f", w, lowJoinMax)
	}
}

// TestBowlTopFlows asks that the top stroke of the bowl, from the join with
// the stem to the top of the bowl, grows evenly in thickness: no place is
// thinner than the one before it by more than half a unit, and none swells
// beyond the pen's thickness at the top.  The thickness is the least
// distance of the left edge from the right one.
func TestBowlTopFlows(t *testing.T) {
	p := Default()
	s := p.bowl(p.bowlStemEdge())
	r, err := s.Build()
	if err != nil {
		t.Fatal(err)
	}
	sample := func(c stroke.Chain) []Point {
		var out []Point
		for _, pc := range c {
			for i := 0; i <= 20; i++ {
				out = append(out, pc.C.At(float64(i)/20))
			}
		}
		return out
	}
	left, right := sample(r.Left), sample(r.Right)
	top := p.nibWidth(0)
	prev := 0.0
	for _, q := range left {
		if q.X < s.Nodes[1].P.X {
			break
		}
		d := math.Inf(1)
		for j := 0; j+1 < len(right); j++ {
			d = math.Min(d, segDistance(q, right[j], right[j+1]))
		}
		if d < prev-0.5 || d > top+1.5 {
			t.Errorf("the top stroke of the bowl is %.1f thick at %.0f,%.0f after %.1f, the pen gives %.1f at the top", d, q.X, q.Y, prev, top)
		}
		prev = math.Max(prev, d)
	}
}

func segDistance(q, a, b Point) float64 {
	dx, dy := b.X-a.X, b.Y-a.Y
	t := 0.0
	if l := dx*dx + dy*dy; l > 0 {
		t = math.Max(0, math.Min(1, ((q.X-a.X)*dx+(q.Y-a.Y)*dy)/l))
	}
	return math.Hypot(q.X-a.X-t*dx, q.Y-a.Y-t*dy)
}
