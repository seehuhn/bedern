// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/outline"
)

// The tests of B, D, P, R and U measure the outlines in every variant and
// perturbed parameter set: the rounds of the bowls and of U (each curve turns one
// way, has no flat spot and no bump, starts gently after a straight stroke and
// joins the next curve without a jump of curvature), the weight of the bowls
// between the stem and the side of O, the proportions of B, the overshoot and the
// thin bottom of U, the serifs, and the leg of R against what follows.

// capRoundSamples is the number of samples of the curvature along a curve.
const capRoundSamples = 200

// A roundSeg is one curve of a round of a capital, in the order of its contour,
// with the pieces before and after it.
type roundSeg struct {
	c            cubic
	before, next segInfo
}

// segInfo says what joins a round at one of its ends: whether it is a straight
// piece, and the curve if it is a curve.
type segInfo struct {
	line bool
	c    cubic
}

// capRounds returns the curves of the contours of path whose two ends both
// satisfy keep, with the pieces next to them.
func capRounds(path outline.Path, keep func(q outline.Point) bool) []roundSeg {
	var out []roundSeg
	for _, ct := range path {
		n := len(ct.Segs)
		for i := 0; i < n; i++ {
			s := ct.Segs[i]
			if !s.IsCurve {
				continue
			}
			c := segCubic(ct, i)
			if !keep(c.P0) || !keep(c.P3) {
				continue
			}
			info := func(j int) segInfo {
				j = (j + n) % n
				return segInfo{line: !ct.Segs[j].IsCurve, c: segCubic(ct, j)}
			}
			out = append(out, roundSeg{c: c, before: info(i - 1), next: info(i + 1)})
		}
	}
	return out
}

// curvatureProfile returns the curvature of c at capRoundSamples+1 points,
// multiplied by the sign of its mean: a curve which turns one way gives positive
// values.
func curvatureProfile(c cubic) []float64 {
	k := make([]float64, capRoundSamples+1)
	var sum float64
	for i := range k {
		k[i] = c.curvature(float64(i) / capRoundSamples)
		sum += k[i]
	}
	if sum < 0 {
		for i := range k {
			k[i] = -k[i]
		}
	}
	return k
}

// The rules for the rounds.  A curve turns one way only.  Its curvature has no
// interior minimum (no flat spot and no bump): it rises, falls, or rises to one
// maximum and falls, within capRoundNoise of its maximum.  Where it leaves a
// straight stroke it starts with at most capRoundStart times its greatest
// curvature.  Where it joins another curve with the same tangent the curvatures
// on the two sides differ by at most the ratio capRoundJump.
const (
	capRoundNoise = 0.01
	capRoundStart = 0.5
	capRoundJump  = 1.25
)

// checkRound checks one curve by the rules above and returns the ratio of the
// curvature at a straight end to the greatest curvature (0 if it has no
// straight end) and the ratio of the curvatures at a join with the next curve
// (1 if there is none).
func checkRound(t *testing.T, what string, r roundSeg) (start, jump float64) {
	t.Helper()
	k := curvatureProfile(r.c)
	kmax := 0.0
	for _, v := range k {
		if v <= 0 {
			t.Errorf("%s: the curve from (%.0f, %.0f) to (%.0f, %.0f) turns both ways", what, r.c.P0.X, r.c.P0.Y, r.c.P3.X, r.c.P3.Y)
			return 0, 1
		}
		kmax = math.Max(kmax, v)
	}
	// No interior minimum: once the curvature has fallen it does not rise
	// again by more than the noise.
	low := math.Inf(1)
	falling := false
	for i := 1; i < len(k); i++ {
		if k[i] < k[i-1]-capRoundNoise*kmax/capRoundSamples {
			falling = true
		}
		if falling {
			low = math.Min(low, k[i])
			if k[i] > low+capRoundNoise*kmax {
				t.Errorf("%s: the curvature of the curve from (%.0f, %.0f) to (%.0f, %.0f) falls and rises again (%.5f after %.5f)", what, r.c.P0.X, r.c.P0.Y, r.c.P3.X, r.c.P3.Y, k[i], low)
				break
			}
		}
	}
	// A straight piece which meets the curve in a corner, as at the crotch of R,
	// does not run into it.
	smooth := func(u, v outline.Point) bool { return u.X*v.X+u.Y*v.Y > math.Cos(2*math.Pi/180) }
	for _, e := range []struct {
		info segInfo
		k    float64
		ok   bool
	}{
		{r.before, k[0], smooth(unit(r.before.c.P0, r.before.c.P3), r.c.tangent(0))},
		{r.next, k[len(k)-1], smooth(r.c.tangent(1), unit(r.next.c.P0, r.next.c.P3))},
	} {
		if e.info.line && e.ok {
			start = math.Max(start, e.k/kmax)
			if e.k > capRoundStart*kmax {
				t.Errorf("%s: the curve from (%.0f, %.0f) to (%.0f, %.0f) leaves a straight stroke with %.2f of its greatest curvature, want at most %v", what, r.c.P0.X, r.c.P0.Y, r.c.P3.X, r.c.P3.Y, e.k/kmax, capRoundStart)
			}
		}
	}
	jump = 1
	if !r.next.line {
		a, b := r.c.tangent(1), r.next.c.tangent(0)
		if smooth(a, b) {
			kn := curvatureProfile(r.next.c)[0]
			jump = math.Max(k[len(k)-1]/kn, kn/k[len(k)-1])
			if jump > capRoundJump {
				t.Errorf("%s: the curvature jumps by the ratio %.2f at (%.0f, %.0f), want at most %v", what, jump, r.c.P3.X, r.c.P3.Y, capRoundJump)
			}
		}
	}
	return start, jump
}

// stemRightAt returns the right-hand edge of the stem of a bowl letter, from the
// scan at a quarter of the cap height.
func stemRightAt(path outline.Path, cap float64) float64 { return runs(path, 0.25*cap)[0][1] }

// The rounds of B, D, P and R (the curves right of the stem and, for P and R,
// above the underside of the bowl, which leaves out the serifs and the leg) and
// the round at the bottom of U (the curves below UCurve) keep the rules for the
// rounds.
func TestCapBowlRoundsAreSmooth(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		c := np.params.Caps
		cap := capHeight(np)
		bottoms := map[string]float64{"B": 0, "D": 0, "P": c.PBar - c.Bar/2, "R": c.RBar - c.Bar/2}
		worstStart, worstJump := 0.0, 1.0
		for _, name := range []string{"B", "D", "P", "R", "U"} {
			path := capOutline(np, name)
			var keep func(q outline.Point) bool
			if name == "U" {
				keep = func(q outline.Point) bool { return q.Y < c.UCurve+1 }
			} else {
				x, y := stemRightAt(path, cap)+1, bottoms[name]-1
				keep = func(q outline.Point) bool { return q.X > x && q.Y > y }
			}
			rs := capRounds(path, keep)
			if len(rs) < 4 {
				t.Errorf("%s: only %d curves in the rounds", name, len(rs))
			}
			for _, r := range rs {
				s, j := checkRound(t, name, r)
				worstStart, worstJump = math.Max(worstStart, s), math.Max(worstJump, j)
			}
		}
		if np.name == "regular" {
			t.Logf("greatest start %.2f of the greatest curvature, greatest jump %.3f", worstStart, worstJump)
		}
	})
}

// The bowls of B, D, P and R are at least as heavy as the capital stem and no
// heavier than the side of O at their right extreme.
func TestCapBowlWeight(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		cap := capHeight(np)
		h := stemsAt(capOutline(np, "H"), 0.25*cap)
		stem := h[0][1] - h[0][0]
		o := capOutline(np, "O")
		_, y0, _, y1 := o.BBox()
		os := runs(o, (y0+y1)/2)[0]
		oside := os[1] - os[0]
		// The thickness at the right extreme, the height in the range at which
		// the ink reaches furthest to the right.
		thickest := func(path outline.Path, lo, hi float64) float64 {
			right, w := math.Inf(-1), 0.0
			for y := lo; y <= hi; y++ {
				r := runs(path, y)
				if last := r[len(r)-1]; last[1] > right {
					right, w = last[1], last[1]-last[0]
				}
			}
			return w
		}
		for _, m := range []struct {
			name   string
			lo, hi float64
		}{
			{"B", 0.1, 0.4}, {"B", 0.6, 0.9}, {"D", 0.3, 0.7}, {"P", 0.6, 0.9}, {"R", 0.65, 0.9},
		} {
			w := thickest(capOutline(np, m.name), m.lo*cap, m.hi*cap)
			if w < stem-0.5 || w > oside+0.5 {
				t.Errorf("%s: the bowl is %.1f thick at its side, want between the stem %.1f and the side of O %.1f", m.name, w, stem, oside)
			}
			if np.name == "regular" {
				t.Logf("%s: the bowl is %.1f thick at its side (stem %.1f, O %.1f)", m.name, w, stem, oside)
			}
		}
	})
}

// B: the lower bowl is wider than the upper by 1.02 to 1.15 and taller; the
// waist, the thin stroke between the counters at the stem, is centred above the
// middle and at most 0.56 of the cap height; the notch in the outer contour lies
// between the two bowls and is 0.2 to 0.5 of the width of the lower bowl deep.
func TestCapBProportions(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		cap := capHeight(np)
		b := capOutline(np, "B")
		x := stemRightAt(b, cap)
		v := vruns(b, x+3)
		if len(v) != 3 {
			t.Fatalf("the scan beside the stem of B crosses %d strokes, want 3", len(v))
		}
		waist := v[1]
		centre := (waist[0] + waist[1]) / 2
		if c := centre / cap; c <= 0.5 || c > 0.56 {
			t.Errorf("the waist of B is centred at %.3f of the cap height, want above 0.5 and at most 0.56", c)
		}
		right := func(lo, hi float64) (best, at float64) {
			best = math.Inf(-1)
			for y := lo; y <= hi; y++ {
				r := runs(b, y)
				if e := r[len(r)-1][1]; e > best {
					best, at = e, y
				}
			}
			return best, at
		}
		up, _ := right(centre, cap-1)
		low, _ := right(1, centre)
		notch := math.Inf(1)
		for y := 0.38 * cap; y <= 0.62*cap; y++ {
			r := runs(b, y)
			notch = math.Min(notch, r[len(r)-1][1])
		}
		if r := (low - x) / (up - x); r < 1.02 || r > 1.15 {
			t.Errorf("the lower bowl of B is %.3f times as wide as the upper, want 1.02 to 1.15", r)
		}
		if waist[1] <= cap-waist[0] {
			t.Errorf("the lower bowl of B is %.1f high and the upper %.1f: the lower must be the taller", waist[1], cap-waist[0])
		}
		if d := (low - notch) / (low - x); d < 0.2 || d > 0.5 {
			t.Errorf("the notch of B is %.2f of the lower bowl deep, want 0.2 to 0.5", d)
		}
		if np.name == "regular" {
			t.Logf("waist at %.3f, lower over upper %.3f, notch %.1f deep", centre/cap, (low-x)/(up-x), low-notch)
		}
	})
}

// U: the round overshoots the baseline by at least 1 per cent of the cap height
// and is thinnest at the bottom: the vertical thickness of the round at the middle
// between the stems is less than the right stem and at most half the left stem;
// the right stem is thinner than the left.
func TestCapURound(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		cap := capHeight(np)
		u := capOutline(np, "U")
		_, y0, _, _ := u.BBox()
		if -y0 < 0.01*cap {
			t.Errorf("U reaches %.1f below the baseline, want at least %.1f", -y0, 0.01*cap)
		}
		s := stemsAt(u, 0.5*cap)
		if len(s) != 2 {
			t.Fatalf("U has %d stems at half the cap height", len(s))
		}
		left, right := s[0][1]-s[0][0], s[1][1]-s[1][0]
		bottom := vruns(u, (s[0][1]+s[1][0])/2)[0]
		thin := bottom[1] - bottom[0]
		if right >= left || thin >= right || thin > 0.5*left {
			t.Errorf("U: the left stem is %.1f, the right %.1f and the bottom %.1f thick", left, right, thin)
		}
	})
}

// stemSerifs returns the serifs of the given stems of a capital, as capSerifs
// does for every stem: the foot serifs at footScanY and the head serifs just
// below the cap height, on both sides of each stem, no longer than twice jut.
func stemSerifs(path outline.Path, cap, jut float64, stems [][2]float64) (serifs []measuredSerif) {
	for _, s := range stems {
		for _, level := range []struct {
			y    float64
			head bool
		}{{footScanY, false}, {cap - 1, true}} {
			for _, r := range runs(path, level.y) {
				if r[0] > s[1] || r[1] < s[0] {
					continue
				}
				for _, side := range []struct{ edge, tip, dir float64 }{{s[0], r[0], -1}, {s[1], r[1], 1}} {
					j := (side.tip - side.edge) * side.dir
					if j < 1 || j > 2*jut {
						continue
					}
					v := vruns(path, side.tip-side.dir*tipInset)
					run := v[0]
					ms := measuredSerif{jut: j, inner: run[1], outer: run[0]}
					if level.head {
						run = v[len(v)-1]
						ms = measuredSerif{jut: j, inner: run[0], outer: run[1]}
					}
					serifs = append(serifs, ms)
				}
			}
		}
	}
	return serifs
}

// The serifs of B, D, P, R and U are those of H: the same length and the same
// thickness at the tip within half a unit, with their flats on the baseline or the
// cap height.  B and D have two, P and R three and U four.
func TestCapBowlSerifsAreThoseOfH(t *testing.T) {
	t.Parallel()
	want := map[string]int{"B": 2, "D": 2, "P": 3, "R": 3, "U": 4}
	forCaps(t, func(t *testing.T, np namedParams) {
		c := np.params.Caps
		cap := capHeight(np)
		ref := capSerifs(t, capOutline(np, "H"), cap, c.SerifJut, c.Stem)[0]
		refThick := math.Abs(ref.inner - ref.outer)
		for _, name := range []string{"B", "D", "P", "R", "U"} {
			path := capOutline(np, name)
			// The stems: the first run of the scan at 0.4 of the cap height, and
			// for U the last one too.
			rs := runs(path, 0.4*cap)
			stems := [][2]float64{rs[0]}
			if name == "U" {
				stems = append(stems, rs[len(rs)-1])
			}
			ss := stemSerifs(path, cap, c.SerifJut, stems)
			if len(ss) != want[name] {
				t.Errorf("%s has %d serifs, want %d", name, len(ss), want[name])
				continue
			}
			for _, s := range ss {
				if math.Abs(s.jut-ref.jut) > 0.5 || math.Abs(math.Abs(s.inner-s.outer)-refThick) > 0.5 {
					t.Errorf("%s: a serif is %.2f long and %.2f thick at the tip, those of H %.2f and %.2f", name, s.jut, math.Abs(s.inner-s.outer), ref.jut, refThick)
				}
				if !(math.Abs(s.outer) < 0.5 || math.Abs(s.outer-cap) < 0.5) {
					t.Errorf("%s: the flat of a serif is at %.2f, not on the baseline or the cap height", name, s.outer)
				}
			}
		}
	})
}

// The foot of the leg of R keeps at least pairClearance from every letter which follows it,
// the capitals and the lowercase.
func TestCapRLegKeepsAwayFromWhatFollows(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		r, adv := cachedOutline(np, "R")
		for _, e := range table {
			if e.r == 0 || e.r == ' ' {
				continue
			}
			next, _ := cachedOutline(np, e.name)
			d, a, _ := minDistanceBetween(r, shiftedPath(next, adv))
			if d < pairClearance {
				t.Errorf("R followed by %s: nearest approach %.1f at (%.0f, %.0f), want at least %v", e.name, d, a.X, a.Y, pairClearance)
			}
		}
	})
}

// inkArea returns the area of the ink of a capital, the sum of the runs of the
// scans at the middles of the unit rows.
func inkArea(path outline.Path) float64 {
	_, y0, _, y1 := path.BBox()
	area := 0.0
	for y := math.Floor(y0) + 0.5; y < y1; y++ {
		for _, r := range runs(path, y) {
			area += r[1] - r[0]
		}
	}
	return area
}

// The ink of B is 0.95 to 0.98 of that of H and its advance 0.91 to 0.94 of that
// of H; the ink of B is 0.97 to 1.045 of that of D, and the bowls of B are
// thinner than those of D.
func TestCapBColour(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		if np.name != "regular" {
			t.Skip("the band is that of the regular font")
		}
		h, hAdv := cachedOutline(np, "H")
		b, bAdv := cachedOutline(np, "B")
		d := capOutline(np, "D")
		if r := inkArea(b) / inkArea(h); r < 0.95 || r > 0.98 {
			t.Errorf("the ink of B is %.3f of that of H, want 0.95 to 0.98", r)
		}
		if r := bAdv / hAdv; r < 0.91 || r > 0.94 {
			t.Errorf("the advance of B is %.3f of that of H, want 0.91 to 0.94", r)
		}
		if r := inkArea(b) / inkArea(d); r < 0.97 || r > 1.045 {
			t.Errorf("the ink of B is %.3f of that of D, want 0.97 to 1.045", r)
		}
		c := np.params.Caps
		if c.BUpperRound > c.BLowerRound || c.BLowerRound >= c.BowlRound {
			t.Errorf("the bowls of B are %.0f (upper) and %.0f (lower) thick, want the upper no thicker than the lower and the lower thinner than the %.0f of D", c.BUpperRound, c.BLowerRound, c.BowlRound)
		}
	})
}

// The counters of B are open: each is at least 0.95 of the counter of the bowl
// of R at the middle of the bowl, and the waist between them is a thin clean
// stroke, at most Bar thick and at least 0.8 of it.
func TestCapBCountersAreOpen(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		cap := capHeight(np)
		b, r := capOutline(np, "B"), capOutline(np, "R")
		counter := func(path outline.Path, y float64) float64 {
			rs := runs(path, y)
			return rs[1][0] - rs[0][1]
		}
		x := stemRightAt(b, cap)
		v := vruns(b, x+3)
		if len(v) != 3 {
			t.Fatalf("the scan beside the stem of B crosses %d strokes, want 3", len(v))
		}
		waist := v[1][1] - v[1][0]
		c := np.params.Caps
		if waist > c.Bar+0.5 || waist < 0.8*c.Bar {
			t.Errorf("the waist of B is %.1f thick, want %.1f down to %.1f", waist, c.Bar, 0.8*c.Bar)
		}
		rc := counter(r, 0.74*cap)
		lower := counter(b, (v[0][1]+v[1][0])/2)
		upper := counter(b, (v[1][1]+v[2][0])/2)
		if lower < 0.95*rc || upper < 0.95*rc {
			t.Errorf("the counters of B are %.1f (lower) and %.1f (upper) wide, the counter of R %.1f: want at least 0.95 of it", lower, upper, rc)
		}
		if np.name == "regular" {
			t.Logf("counters of B %.1f and %.1f, of R %.1f, waist %.1f", lower, upper, rc, waist)
		}
	})
}
