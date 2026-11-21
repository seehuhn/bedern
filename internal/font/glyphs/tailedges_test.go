// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/outline"
)

// The tails of a and t are curled: both edges are curves from the low point of the
// foot to the tip.  Both edges turn by at
// least minTailTurn over every stretch of tailStretch units, between the low point
// and tailTipSkip units short of the tip, where the rounded corners begin.
const (
	minTailTurn = 0.08
	tailStretch = 40.0
	tailTipSkip = 15.0
	// upperSkip is the length of the cut of the tip and its second rounded corner, which
	// the upper edge starts after.
	upperSkip = 40.0
)

// tailEdges returns the outer edge of the tail, from the low point of the foot to the
// tip, and its upper edge, from the tip to the inner low point, each as a polyline
// with its arc lengths and without the tailTipSkip units next to the tip.
func tailEdges(poly []outline.Point, th float64) (outer, upper []outline.Point, so, su []float64) {
	n := len(poly)
	at := func(i int) outline.Point { return poly[(i%n+n)%n] }
	c, s := math.Cos(th), math.Sin(th)
	var xMax float64
	for _, q := range poly {
		xMax = math.Max(xMax, q.X)
	}
	ie, best := 0, math.Inf(-1)
	for i, q := range poly {
		if q.X > xMax-40 && q.Y < 250 && q.X*c+q.Y*s > best {
			ie, best = i, q.X*c+q.Y*s
		}
	}
	i0 := ie
	for at(i0-1).Y <= at(i0).Y {
		i0--
	}
	i1 := ie
	for at(i1+1).Y >= at(i1).Y {
		i1++
	}
	for at(i1+1).Y <= at(i1).Y {
		i1++
	}
	collect := func(from, to int) (pts []outline.Point, arc []float64) {
		length := 0.0
		for i := from; ; i++ {
			if len(pts) > 0 {
				length += math.Hypot(at(i).X-pts[len(pts)-1].X, at(i).Y-pts[len(pts)-1].Y)
			}
			pts, arc = append(pts, at(i)), append(arc, length)
			if i == to {
				return pts, arc
			}
		}
	}
	outer, so = collect(i0, ie)
	upper, su = collect(ie, i1)
	for total := so[len(so)-1]; len(outer) > 2 && total-so[len(so)-1] < tailTipSkip; {
		outer, so = outer[:len(outer)-1], so[:len(so)-1]
	}
	k := 0
	for k < len(upper)-2 && su[k] < tailTipSkip+upperSkip {
		k++
	}
	upper, su = upper[k:], su[k:]
	base := su[0]
	for i := range su {
		su[i] -= base
	}
	return outer, upper, so, su
}

func TestTailEdgesHaveNoStraightStretch(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			for _, name := range []string{"a", "t"} {
				path, _ := cachedOutline(np, name)
				outer, upper, so, su := tailEdges(path[0].Flatten(), tailShape(name).TipAngle*math.Pi/180)
				for _, e := range []struct {
					what string
					pts  []outline.Point
					s    []float64
				}{{"outer edge", outer, so}, {"upper edge", upper, su}} {
					least, at := slidingTurns(e.pts, e.s, tailStretch)
					if np.name == "regular" {
						t.Logf("%s %s: the least turn over %.0f units is %.3f radians at (%.0f, %.0f)", name, e.what, tailStretch, least, at.X, at.Y)
					}
					if least < minTailTurn {
						t.Errorf("parameter set %s, %s %s: a stretch %.0f long at (%.0f, %.0f) turns by only %.3f radians, want at least %v", np.name, name, e.what, tailStretch, at.X, at.Y, least, minTailTurn)
					}
				}
			}
		})
	}
}
