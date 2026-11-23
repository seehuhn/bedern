// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"sort"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The tests of the capitals measure the outlines, in every variant and
// perturbed parameter set, and check the optical rules which the construction
// must keep however the parameters move: the overshoot and the weight of O, the
// heights and lengths of the bars and arms, the balance of the serifs, the
// smoothness of the O and the white between the serifs of the arms.

// forCaps runs f on every parameter set as a parallel subtest.
func forCaps(t *testing.T, f func(t *testing.T, np namedParams)) {
	for _, np := range allParams() {
		if !boldApplies(t, np) {
			continue
		}
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			f(t, np)
		})
	}
}

// capOutline returns the unrounded outline of a capital in the set.
func capOutline(np namedParams, name string) outline.Path {
	path, _ := cachedOutline(np, name)
	return path
}

// runs returns the crossings of the horizontal scan at y in pairs, the runs of
// ink from left to right.
func runs(path outline.Path, y float64) [][2]float64 {
	xs := path.CrossingsX(y)
	sort.Float64s(xs)
	var r [][2]float64
	for i := 0; i+1 < len(xs); i += 2 {
		r = append(r, [2]float64{xs[i], xs[i+1]})
	}
	return r
}

// vruns returns the runs of ink of the vertical scan at x, from the bottom up.
func vruns(path outline.Path, x float64) [][2]float64 {
	ys := path.CrossingsY(x)
	sort.Float64s(ys)
	var r [][2]float64
	for i := 0; i+1 < len(ys); i += 2 {
		r = append(r, [2]float64{ys[i], ys[i+1]})
	}
	return r
}

// stemsAt returns the stems of a capital as the runs of the scan at the height
// y, which must cross the stems and nothing else.
func stemsAt(path outline.Path, y float64) [][2]float64 { return runs(path, y) }

// capHeight returns the height of the flat top of H, measured.
func capHeight(np namedParams) float64 {
	_, _, _, y1 := capOutline(np, "H").BBox()
	return y1
}

// The O overshoots the cap height and the baseline by at least 1 per cent of the
// cap height, its sides are heavier than the stem of H by at least 5 per cent and
// its thin parts are less than half the stem.
func TestCapOOvershootsAndIsHeavierThanTheStem(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		cap := capHeight(np)
		o := capOutline(np, "O")
		x0, y0, x1, y1 := o.BBox()
		if y1-cap < 0.01*cap || -y0 < 0.01*cap {
			t.Errorf("O reaches %.1f above the cap height and %.1f below the baseline, want at least %.1f", y1-cap, -y0, 0.01*cap)
		}
		h := stemsAt(capOutline(np, "H"), 0.25*cap)
		stem := h[0][1] - h[0][0]
		side := runs(o, (y0+y1)/2)[0]
		if thick := side[1] - side[0]; thick < 1.05*stem {
			t.Errorf("the side of O is %.1f thick, the stem of H %.1f: want at least 1.05 times", thick, stem)
		}
		top := vruns(o, (x0+x1)/2)
		if thin := top[len(top)-1][1] - top[len(top)-1][0]; thin > 0.5*stem {
			t.Errorf("the top of O is %.1f thick, want less than half the stem %.1f", thin, stem)
		}
	})
}

// The bar of H sits a little above the middle: its centre between 0.5 and 0.56 of
// the cap height.
func TestCapHBarSitsAboveTheMiddle(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		cap := capHeight(np)
		h := capOutline(np, "H")
		s := stemsAt(h, 0.25*cap)
		if len(s) != 2 {
			t.Fatalf("H has %d stems", len(s))
		}
		bar := vruns(h, (s[0][1]+s[1][0])/2)
		if len(bar) != 1 {
			t.Fatalf("the scan between the stems of H crosses %d runs, want the bar alone", len(bar))
		}
		if c := (bar[0][0] + bar[0][1]) / 2 / cap; c <= 0.5 || c > 0.56 {
			t.Errorf("the centre of the bar of H is at %.3f of the cap height, want above 0.5 and at most 0.56", c)
		}
	})
}

// capArms returns the arms of E or F: for each bar which leaves the right-hand
// edge of the stem, from the bottom up, the height of its centre, its edges and
// the end of the ink at the centre height.  A foot serif, which reaches less than
// twice the serif length jut beyond the stem, is not an arm.
type measuredArm struct{ lower, upper, centre, end float64 }

func capArms(path outline.Path, cap, jut float64) (stemRight float64, arms []measuredArm) {
	stem := stemsAt(path, 0.27*cap)[0]
	stemRight = stem[1]
	for _, r := range vruns(path, stemRight+3) {
		a := measuredArm{lower: r[0], upper: r[1], centre: (r[0] + r[1]) / 2}
		rs := runs(path, a.centre)
		a.end = rs[len(rs)-1][1]
		if a.end-stemRight > 2*jut {
			arms = append(arms, a)
		}
	}
	return stemRight, arms
}

// The middle arm of E and F stands above the middle and is at most 0.75 of the top
// arm long, measured from the stem to the end of the ink; the bottom arm of E is
// at least as long as its top arm.
func TestCapEFArms(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		cap := capHeight(np)
		for _, name := range []string{"E", "F"} {
			x, arms := capArms(capOutline(np, name), cap, np.params.Caps.SerifJut)
			n := 3
			if name == "F" {
				n = 2
			}
			if len(arms) != n {
				t.Fatalf("%s has %d arms, want %d", name, len(arms), n)
			}
			top, middle := arms[n-1], arms[n-2]
			if middle.centre <= cap/2 {
				t.Errorf("%s: the middle arm is centred at %.1f, not above the middle %.1f", name, middle.centre, cap/2)
			}
			if r := (middle.end - x) / (top.end - x); r > 0.75 {
				t.Errorf("%s: the middle arm is %.2f of the top arm, want at most 0.75", name, r)
			}
			if name == "E" && arms[0].end < top.end {
				t.Errorf("E: the bottom arm ends at %.1f, before the top arm at %.1f", arms[0].end, top.end)
			}
		}
	})
}

// The two arms of T reach equally far from the centre of the stem.
func TestCapTArmsBalanced(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		cap := capHeight(np)
		tt := capOutline(np, "T")
		s := stemsAt(tt, 0.4*cap)[0]
		c := (s[0] + s[1]) / 2
		bar := runs(tt, cap-2)[0]
		if d := (bar[1] - c) - (c - bar[0]); math.Abs(d) > 0.5 {
			t.Errorf("the right arm of T reaches %.2f further than the left", d)
		}
	})
}

// A measuredSerif is a serif of a capital stem: its length beyond the stem edge
// and the heights of the two faces at the tip (the flat is on the baseline for a
// foot and on the cap height for a head).
type measuredSerif struct {
	jut, inner, outer float64 // the length, and the height of the face towards the stem and of the flat
}

// capSerifs returns the serifs of every stem of a capital: the foot serifs at the
// height footScanY and the head serifs just below the cap height, on the left and
// right of each stem, where the ink beyond the edge is no longer than twice the
// serif; a longer run is an arm, and a side without ink beyond the edge has none.
func capSerifs(t *testing.T, path outline.Path, cap, jut, stem float64) (serifs []measuredSerif) {
	t.Helper()
	for _, s := range stemsAt(path, 0.4*cap) {
		if s[1]-s[0] < 0.8*stem {
			continue // the serif of an arm
		}
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
					if level.head {
						run = v[len(v)-1]
					}
					ms := measuredSerif{jut: j, inner: run[1], outer: run[0]}
					if level.head {
						ms.inner, ms.outer = run[0], run[1]
					}
					serifs = append(serifs, ms)
				}
			}
		}
	}
	return serifs
}

// The serifs of the capital stems are one shape: in each letter every serif has
// the same length, the same thickness at the tip and its flat on the baseline or
// on the cap height, within half a unit.  H has eight, I four, E and F three, L
// three and T two.
func TestCapSerifsAreLevelAndBalanced(t *testing.T) {
	t.Parallel()
	want := map[string]int{"H": 8, "I": 4, "E": 2, "F": 3, "L": 3, "T": 2}
	forCaps(t, func(t *testing.T, np namedParams) {
		cap := capHeight(np)
		jut := np.params.Caps.SerifJut
		var ref *measuredSerif
		for _, name := range []string{"H", "I", "E", "F", "L", "T"} {
			ss := capSerifs(t, capOutline(np, name), cap, jut, np.params.Caps.Stem)
			if len(ss) != want[name] {
				t.Errorf("%s has %d serifs, want %d", name, len(ss), want[name])
				continue
			}
			for _, s := range ss {
				thick := math.Abs(s.inner - s.outer)
				if ref == nil {
					r := s
					ref = &r
				}
				if math.Abs(s.jut-ref.jut) > 0.5 || math.Abs(thick-math.Abs(ref.inner-ref.outer)) > 0.5 {
					t.Errorf("%s: a serif is %.2f long and %.2f thick at the tip, another %.2f and %.2f", name, s.jut, thick, ref.jut, math.Abs(ref.inner-ref.outer))
				}
				if !(math.Abs(s.outer) < 0.5 || math.Abs(s.outer-cap) < 0.5) {
					t.Errorf("%s: the flat of a serif is at %.2f, not on the baseline or the cap height", name, s.outer)
				}
			}
		}
	})
}

// contourSamples returns points along the contour, sampled exactly on its curves,
// with their arc lengths.
func contourSamples(ct outline.Contour) (pts []outline.Point, s []float64) {
	length := 0.0
	for i := range ct.Segs {
		c := segCubic(ct, i)
		for k := 0; k < edgeSamples; k++ {
			q := c.at(float64(k) / edgeSamples)
			if len(pts) > 0 {
				length += math.Hypot(q.X-pts[len(pts)-1].X, q.Y-pts[len(pts)-1].Y)
			}
			pts = append(pts, q)
			s = append(s, length)
		}
	}
	return pts, s
}

// capOWindow is the arc length over which the contours of O must turn.  An
// ellipse turns least at the ends of its longer axis, by window a / b² with the
// half axes a and b; each contour of O must turn by at least capOStraight times
// that over every stretch of this length.  The regular O turns by 0.81 (outer)
// and 0.58 (inner) of the ellipse.
const (
	capOWindow   = 50.0
	capOStraight = 0.45
)

// The contours of O have no bump and no straight stretch: each turns one way
// only (the sum of the turns the other way is below edgeTolerance), and every
// stretch of capOWindow turns by at least capOStraight times what the ellipse in
// the same box turns at its flattest.
func TestCapOIsSmooth(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		o := capOutline(np, "O")
		for i, ct := range o {
			pts, s := contourSamples(ct)
			sign := 1.0
			if ct.SignedArea() < 0 {
				sign = -1
			}
			var wrong float64
			for k := 1; k+1 < len(pts); k++ {
				h0 := math.Atan2(pts[k].Y-pts[k-1].Y, pts[k].X-pts[k-1].X)
				h1 := math.Atan2(pts[k+1].Y-pts[k].Y, pts[k+1].X-pts[k].X)
				if d := sign * turn(h0, h1); d < 0 {
					wrong -= d
				}
			}
			if wrong > edgeTolerance {
				t.Errorf("contour %d of O turns the wrong way by %.4f radians", i, wrong)
			}
			x0, y0, x1, y1 := outline.Path{ct}.BBox()
			a, b := (x1-x0)/2, (y1-y0)/2
			ellipse := capOWindow * a / (b * b)
			least, at := slidingTurns(pts, s, capOWindow)
			if least < capOStraight*ellipse {
				t.Errorf("contour %d of O turns by only %.4f over %v units at (%.0f, %.0f), want at least %.4f", i, least, capOWindow, at.X, at.Y, capOStraight*ellipse)
			}
			if np.name == "regular" {
				t.Logf("contour %d: least turn %.4f over %v units, %.2f of the ellipse", i, least, capOWindow, least/ellipse)
			}
		}
	})
}

// The serifs of neighbouring arms of E and F keep clear of each other: the flat
// ends of the serif of the top arm and of the upper serif of the middle arm, and of
// the lower serif of the middle arm and the serif of the bottom arm, are at least
// three clearance floors apart.
func TestCapArmSerifsKeepApart(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		cap := capHeight(np)
		floor := 3 * design.ClearanceFloor(np.params)
		for _, name := range []string{"E", "F"} {
			path := capOutline(np, name)
			_, arms := capArms(path, cap, np.params.Caps.SerifJut)
			// The flat end of the serif of an arm, the highest or the lowest
			// ink of the vertical scan just inside the end of the arm, and its
			// width from the horizontal scan just inside it.
			end := func(a measuredArm, up bool) (y, xIn, xOut float64) {
				v := vruns(path, a.end-1)
				for _, r := range v {
					if r[0] <= a.centre && a.centre <= r[1] {
						y = r[0]
						yy := y + 1
						if up {
							y, yy = r[1], r[1]-1
						}
						rs := runs(path, yy)
						last := rs[len(rs)-1]
						for _, q := range rs {
							if q[0] <= a.end-1 && a.end-1 <= q[1] {
								last = q
							}
						}
						return y, last[0], last[1]
					}
				}
				t.Fatalf("%s: no serif at the end of the arm at %.1f", name, a.centre)
				return
			}
			gap := func(what string, lo measuredArm, hi measuredArm) {
				ya, a0, a1 := end(lo, true)
				yb, b0, b1 := end(hi, false)
				d := math.Inf(1)
				for _, q := range []outline.Point{{X: a0, Y: ya}, {X: a1, Y: ya}} {
					d = math.Min(d, segDist(q, outline.Point{X: b0, Y: yb}, outline.Point{X: b1, Y: yb}))
				}
				for _, q := range []outline.Point{{X: b0, Y: yb}, {X: b1, Y: yb}} {
					d = math.Min(d, segDist(q, outline.Point{X: a0, Y: ya}, outline.Point{X: a1, Y: ya}))
				}
				if d < floor {
					t.Errorf("%s: the serifs of the %s are %.1f apart, want at least %.1f", name, what, d, floor)
				}
				if np.name == "regular" {
					t.Logf("%s: the serifs of the %s are %.1f apart", name, what, d)
				}
			}
			n := len(arms)
			gap("middle and the top arm", arms[n-2], arms[n-1])
			if name == "E" {
				gap("bottom and the middle arm", arms[0], arms[1])
			}
		}
	})
}
