// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/outline"
)

// figureNames are the glyph names of the figures 0 to 9.
var figureNames = []string{"zero", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine"}

// The flat tops (4, 5, 7) stand at one height, within half a unit, and the 1
// rises to the round tops, up to one per cent of the height above them; the round
// tops (0, 2, 3, 6, 8, 9) rise above them by 0.5 to 2 per cent of that height;
// and the round bottoms (0, 3, 5, 6, 8, 9)
// reach as far below the baseline as the bottom of O, within a unit, while the
// flat bottoms (1, 2, 4, 7) stand on it.
func TestFigureHeightsAndOvershoots(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		box := func(name string) (y0, y1 float64) {
			_, y0, _, y1 = capOutline(np, name).BBox()
			return
		}
		_, h := box("seven")
		for _, n := range []string{"four", "five"} {
			if _, top := box(n); math.Abs(top-h) > 0.5 {
				t.Errorf("%s: the top of %s is at %.1f, the top of 7 at %.1f", np.name, n, top, h)
			}
		}
		// The 1 meets the line with the point of its flag where 7 has a full
		// bar, and stands as high as the round tops or a little higher.
		_, zero := box("zero")
		if _, one := box("one"); one < zero-0.5 || one > zero+0.01*h {
			t.Errorf("%s: the top of 1 is at %.1f, the round top of 0 at %.1f", np.name, one, zero)
		}
		for _, n := range []string{"zero", "two", "three", "six", "eight", "nine"} {
			if _, top := box(n); top-h < 0.005*h || top-h > 0.02*h {
				t.Errorf("%s: the round top of %s is %.1f above the flat tops, want %.1f to %.1f", np.name, n, top-h, 0.005*h, 0.02*h)
			}
		}
		oBottom, _ := box("O")
		for _, n := range []string{"zero", "three", "five", "six", "eight", "nine"} {
			if bot, _ := box(n); math.Abs(bot-oBottom) > 1 {
				t.Errorf("%s: the bottom of %s is at %.1f, that of O at %.1f", np.name, n, bot, oBottom)
			}
		}
		for _, n := range []string{"one", "two", "four", "seven"} {
			if bot, _ := box(n); math.Abs(bot) > 0.5 {
				t.Errorf("%s: the bottom of %s is at %.1f, want the baseline", np.name, n, bot)
			}
		}
	})
}

// The figures are tabular: all ten have one advance, and the ink of each keeps
// at least 15 units inside it on both sides.
func TestFiguresAreTabular(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		_, adv0 := cachedOutline(np, "zero")
		for _, n := range figureNames {
			path, adv := cachedOutline(np, n)
			if adv != adv0 {
				t.Errorf("%s: the advance of %s is %g, that of 0 %g", np.name, n, adv, adv0)
			}
			x0, _, x1, _ := path.BBox()
			if x0 < 15 || adv-x1 < 15 {
				t.Errorf("%s: the ink of %s keeps %.1f and %.1f from the edges, want 15", np.name, n, x0, adv-x1)
			}
		}
	})
}

// The curved edges of the round figures have no kink, no bump and no straight
// stretch, by the rule of the round capitals (TestCapCGJQEdges): where two curves
// meet tangentially away from the ball, the curvature changes by at most a factor
// of 2, and every stretch of capEdgeWindow units of a run of curves turns by at
// least capEdgeTurn.  The diagonal of 2 ends straight at its foot, as a
// diagonal does, and its turning is checked only above 0.6 of the height of
// its shoulder; the edges which run into the balls have the tests of
// figwind_test.go as well.  The zero has the rule of O instead
// (TestFigureZeroIsSmooth).
func TestFigureEdges(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		f := np.params.Figures
		for _, name := range []string{"two", "three", "five", "six", "eight", "nine"} {
			path := capOutline(np, name)
			for ci, ct := range path {
				n := len(ct.Segs)
				for i := 0; i < n; i++ {
					if !ct.Segs[i].IsCurve || !ct.Segs[(i+1)%n].IsCurve {
						continue
					}
					a, b := segCubic(ct, i), segCubic(ct, i+1)
					ta, tb := a.tangent(1), b.tangent(0)
					if math.Abs(math.Atan2(ta.X*tb.Y-ta.Y*tb.X, ta.X*tb.X+ta.Y*tb.Y)) > 0.01 {
						continue // a corner: the notches, the tip of the wedge of a ball
					}
					ka, kb := a.curvature(1), b.curvature(0)
					if math.Max(math.Abs(ka), math.Abs(kb)) > 1.0/40 {
						continue // the ball, which has tests of its own
					}
					if ka*kb < 0 || math.Max(math.Abs(ka), math.Abs(kb)) > 2*math.Min(math.Abs(ka), math.Abs(kb))+1e-4 {
						t.Errorf("%s: %s contour %d: the curvature jumps from %.5f to %.5f at (%.0f, %.0f)", np.name, name, ci, ka, kb, a.P3.X, a.P3.Y)
					}
				}
				least := math.Inf(1)
				var where outline.Point
				for _, run := range capCurveRuns(ct) {
					if name == "two" {
						var upper []outline.Point
						for _, q := range run {
							if q.Y > 0.6*f.TwoShoulder {
								upper = append(upper, q)
							}
						}
						run = upper
					}
					if name == "five" {
						// The bracket at the foot of the left-hand edge
						// turns to the left into the inner edge of the
						// bowl, which turns to the right; the run is taken
						// from where the inner edge has its own turn.
						var rest []outline.Point
						for _, q := range run {
							if q.X > 180 || q.Y < 300 {
								rest = append(rest, q)
							}
						}
						run = rest
					}
					if len(run) < 2 {
						continue
					}
					s := make([]float64, len(run))
					for k := 1; k < len(run); k++ {
						s[k] = s[k-1] + math.Hypot(run[k].X-run[k-1].X, run[k].Y-run[k-1].Y)
					}
					if s[len(s)-1] < capEdgeWindow {
						continue
					}
					if l, at := slidingTurns(run, s, capEdgeWindow); l < least {
						least, where = l, at
					}
				}
				if least < capEdgeTurn {
					t.Errorf("%s: %s contour %d turns by only %.3f over %v units at (%.0f, %.0f)", np.name, name, ci, least, capEdgeWindow, where.X, where.Y)
				}
			}
		}
	})
}

// The contours of 0 are smooth as those of O (TestCapOIsSmooth): every stretch of
// capOWindow units turns by at least capOStraight times what the ellipse in the
// same box turns at its flattest.
func TestFigureZeroIsSmooth(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		for i, ct := range capOutline(np, "zero") {
			pts, s := contourSamples(ct)
			x0, y0, x1, y1 := outline.Path{ct}.BBox()
			a, b := (x1-x0)/2, (y1-y0)/2
			ellipse := capOWindow * a / (b * b)
			if least, at := slidingTurns(pts, s, capOWindow); least < capOStraight*ellipse {
				t.Errorf("%s: contour %d of 0 turns by only %.4f over %v units at (%.0f, %.0f), want at least %.4f", np.name, i, least, capOWindow, at.X, at.Y, capOStraight*ellipse)
			}
		}
	})
}

// Every inner corner of the outer contour of a figure, where the contour turns
// against its direction (the notches of 3, 6, 8 and 9, the corners under the bars
// of 4, 5 and 7), opens by at least 30 degrees.
func TestFigureNotchesAreOpen(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		for _, name := range figureNames {
			ct := capOutline(np, name)[0]
			n := len(ct.Segs)
			for i := 0; i < n; i++ {
				ta, tb := segCubic(ct, i).tangent(1), segCubic(ct, i+1).tangent(0)
				if ct.Segs[i].IsCurve == false {
					q := segCubic(ct, i)
					ta = unit(q.P0, q.P3)
				}
				if !ct.Segs[(i+1)%n].IsCurve {
					q := segCubic(ct, i+1)
					tb = unit(q.P0, q.P3)
				}
				turnBy := math.Atan2(ta.X*tb.Y-ta.Y*tb.X, ta.X*tb.X+ta.Y*tb.Y) * 180 / math.Pi
				if turnBy < 0 && 180+turnBy < 30 {
					p := ct.Segs[i].P3
					t.Errorf("%s: %s: the notch at (%.0f, %.0f) opens by %.1f degrees", np.name, name, p.X, p.Y, 180+turnBy)
				}
			}
		}
	})
}

// The figures with two bowls stand upright: the upper bowl of 3 and of 8 is
// narrower than the lower one (0.8 to 0.95 of its width at the widest), the
// bowls of 8 are centred on one vertical within 10 units, and the waist of both
// lies above the middle of the height of the figures, at most at 0.6 of it.
func TestFigureBowlsStandUpright(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		h := np.params.Figures.Height
		for _, name := range []string{"three", "eight"} {
			path := capOutline(np, name)
			// The widest run above and below the waist, which is the
			// height at which the right extent of the ink is least.
			waist, least := 0.0, math.Inf(1)
			for y := 0.35 * h; y < 0.65*h; y++ {
				r := runs(path, y)
				if x := r[len(r)-1][1]; x < least {
					least, waist = x, y
				}
			}
			widest := func(from, to float64) (l, r float64) {
				l, r = math.Inf(1), math.Inf(-1)
				for y := from; y < to; y++ {
					rs := runs(path, y)
					l, r = math.Min(l, rs[0][0]), math.Max(r, rs[len(rs)-1][1])
				}
				return
			}
			ul, ur := widest(waist+20, h-20)
			ll, lr := widest(20, waist-20)
			ratio := (ur - ul) / (lr - ll)
			if name == "three" {
				// The open side of 3 is the ball: compare the right extents.
				ratio = (ur - ll) / (lr - ll)
			}
			if ratio < 0.8 || ratio > 0.95 {
				t.Errorf("%s: %s: the upper bowl is %.3f of the lower", np.name, name, ratio)
			}
			if name == "eight" && math.Abs((ul+ur)/2-(ll+lr)/2) > 10 {
				t.Errorf("%s: 8: the bowls are centred %.1f apart", np.name, (ul+ur)/2-(ll+lr)/2)
			}
			if waist < 0.5*h || waist > 0.6*h {
				t.Errorf("%s: %s: the waist is at %.3f of the height", np.name, name, waist/h)
			}
		}
	})
}
