// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"sort"
	"testing"

	"seehuhn.de/go/bedern/internal/outline"
)

// The tails of a and t end in a blunt tip: a short cut at right angles to
// the stroke with rounded corners, about as thick as the thin parts of the face.
const (
	// minTipWidth is the least thickness of the tail, across the stroke, at 4 and at
	// 14 units before its end.  The thin parts of the face are 32.
	minTipWidth = 30.0
	// maxTipTurn is the largest turn of the contour, in radians, at a vertex of the
	// flattened outline within tipReach of the end of the tail: no corner is sharper
	// than the two rounded ones, which turn by 2 degrees at most per vertex.
	maxTipTurn = 12 * math.Pi / 180
	tipReach   = 45.0
)

func TestTailEndIsBlunt(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			for _, name := range []string{"a", "t"} {
				// along and across give the position of a point along the direction
				// of the stroke at the tip and to the left of it.
				th := tailShape(name).TipAngle * math.Pi / 180
				c, s := math.Cos(th), math.Sin(th)
				along := func(q outline.Point) float64 { return q.X*c + q.Y*s }
				across := func(q outline.Point) float64 { return -q.X*s + q.Y*c }
				path, _ := cachedOutline(np, name)
				poly := path[0].Flatten()
				// The end of the tail is the vertex of the lower right part
				// which lies furthest along the direction of the stroke.
				var xMax float64
				for _, q := range poly {
					xMax = math.Max(xMax, q.X)
				}
				var end outline.Point
				best := math.Inf(-1)
				for _, q := range poly {
					if q.X > xMax-40 && q.Y < 250 && along(q) > best {
						best, end = along(q), q
					}
				}
				for _, d := range []float64{4, 14} {
					var vs []float64
					for i := range poly {
						a, b := poly[i], poly[(i+1)%len(poly)]
						ua, ub := along(a)-(best-d), along(b)-(best-d)
						if (ua < 0) == (ub < 0) {
							continue
						}
						q := outline.Point{X: a.X + ua/(ua-ub)*(b.X-a.X), Y: a.Y + ua/(ua-ub)*(b.Y-a.Y)}
						if math.Abs(across(q)-across(end)) < 60 {
							vs = append(vs, across(q))
						}
					}
					sort.Float64s(vs)
					if len(vs) < 2 {
						t.Errorf("parameter set %s, %s: no ink %v before the end of the tail", np.name, name, d)
						continue
					}
					width := vs[len(vs)-1] - vs[0]
					if np.name == "regular" {
						t.Logf("%s: the tail is %.1f thick %v before its end", name, width, d)
					}
					if width < minTipWidth {
						t.Errorf("parameter set %s, %s: the tail is %.1f thick %v before its end, want at least %v", np.name, name, width, d, minTipWidth)
					}
				}
				var worst float64
				var worstAt outline.Point
				for i := range poly {
					a, b, cc := poly[(i+len(poly)-1)%len(poly)], poly[i], poly[(i+1)%len(poly)]
					if math.Hypot(b.X-end.X, b.Y-end.Y) > tipReach {
						continue
					}
					if x := math.Abs(turn(math.Atan2(b.Y-a.Y, b.X-a.X), math.Atan2(cc.Y-b.Y, cc.X-b.X))); x > worst {
						worst, worstAt = x, b
					}
				}
				if np.name == "regular" {
					t.Logf("%s: the sharpest turn at a vertex near the end of the tail is %.1f degrees", name, worst*180/math.Pi)
				}
				if worst > maxTipTurn {
					t.Errorf("parameter set %s, %s: the contour turns by %.1f degrees at (%.0f, %.0f) near the end of the tail, want at most %.0f", np.name, name, worst*180/math.Pi, worstAt.X, worstAt.Y, maxTipTurn*180/math.Pi)
				}
			}
		})
	}
}

// The white inside the curl of each tail is open: curlRise above the inner low
// point it is at least minCurlWhite wide, from the inner curve which rises to
// the stem to the upper edge of the tail.
const (
	curlRise     = 20.0
	minCurlWhite = 40.0
)

func TestTailCurlIsOpen(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			p := np.params
			for _, name := range []string{"a", "t"} {
				path, _ := cachedOutline(np, name)
				y := -p.OvershootLower + p.ThinLower + curlRise
				xs := path.CrossingsX(y)
				if len(xs) < 4 {
					t.Fatalf("parameter set %s, %s: crossings at %.0f: %v", np.name, name, y, xs)
				}
				k := len(xs)
				white := xs[k-2] - xs[k-3]
				if np.name == "regular" {
					t.Logf("%s: the white inside the curl is %.1f wide %v above the inner low point", name, white, curlRise)
				}
				if white < minCurlWhite {
					t.Errorf("parameter set %s, %s: the white inside the curl is %.1f wide at %.0f, want at least %v", np.name, name, white, y, minCurlWhite)
				}
			}
		})
	}
}

// tailShape returns the shape of the foot of the glyph with the given name,
// "a" or "t".
func tailShape(name string) tFootShape {
	if name == "t" {
		return tTail
	}
	return tFoot
}
