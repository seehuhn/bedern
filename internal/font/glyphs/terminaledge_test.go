// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The branch of the arch has no straight stretch.  Where the arch of n, h, m and
// (turned) u leaves the stem, both edges turn by at least minBranchTurn over
// every stretch of branchLength units, from the left stem to the top of the
// arch.  The shoulder on the right runs out into the stem straight
// (TestShoulderRunsIntoTheStem).
const (
	minBranchTurn = 0.07
	branchLength  = 40.0
)

// archEdges returns the ends of the two edges of each arch of the letter, in the
// order along the contour: the underside from the join to the inner shoulder, and the
// outer edge from the outer shoulder to the notch.
func archEdges(name string, p design.Params) (under, outer [][2]outline.Point) {
	pt := func(x, y float64) outline.Point { return outline.Point{X: x, Y: y} }
	add := func(a archSpec) {
		under = append(under, [2]outline.Point{pt(a.X1, a.Join), pt(a.X2, a.ShoulderInner)})
		outer = append(outer, [2]outline.Point{pt(a.X3, a.Shoulder), pt(a.X1, a.Notch)})
	}
	switch name {
	case "n":
		left := stemSpec{Left: p.SideStemLeft, Width: p.Stem}
		right := stemSpec{Left: left.right() + p.NCounter, Width: p.Stem}
		add(archBetween(p, left, right))
	case "h":
		left := stemSpec{Left: p.SideStemLeft + design.AscenderInset(p), Width: p.AscenderStem}
		inset := (p.Stem - left.Width) / 2
		right := stemSpec{Left: left.Left - inset + p.Stem + p.NCounter, Width: p.Stem}
		add(archBetween(p, left, right))
	case "m":
		first := stemSpec{Left: p.SideStemLeft, Width: p.Stem}
		middle := stemSpec{Left: first.right() + p.MCounter, Width: p.Stem}
		last := stemSpec{Left: middle.right() + p.MCounter, Width: p.Stem}
		archL := archBetween(p, first, middle)
		archL.Shoulder = archL.Notch
		add(archL)
		add(archBetween(p, middle, last))
	case "u":
		left := stemSpec{Left: p.SideStemLeft, Width: p.Stem}
		right := stemSpec{Left: left.right() + p.UCounter, Width: p.Stem}
		under = append(under, [2]outline.Point{pt(right.Left, p.UJoin), pt(left.right(), p.UShoulderInner)})
		outer = append(outer, [2]outline.Point{pt(left.Left, p.UShoulder), pt(right.Left, p.UNotch)})
	default:
		panic("archEdges: " + name)
	}
	return under, outer
}

func TestBranchOfTheArchHasNoStraightStretch(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			for _, name := range []string{"n", "h", "m", "u"} {
				path, _ := cachedOutline(np, name)
				under, outer := archEdges(name, np.params)
				for k, edge := range append(append([][2]outline.Point{}, under...), outer...) {
					what := "underside"
					if k >= len(under) {
						what = "outer edge"
					}
					pts, s := outerEdge(t, path, edge[0], edge[1])
					pts, s = branchOnly(pts, s, np.params.XHeight/2, k < len(under))
					least, at := slidingTurns(pts, s, branchLength)
					if np.name == "regular" {
						t.Logf("%s %s: the least turn over %.0f units is %.3f radians at (%.0f, %.0f)", name, what, branchLength, least, at.X, at.Y)
					}
					if least < minBranchTurn {
						t.Errorf("parameter set %s, %s %s: a stretch %.0f long at (%.0f, %.0f) turns by only %.3f radians, want at least %v", np.name, name, what, branchLength, at.X, at.Y, least, minBranchTurn)
					}
				}
			}
		})
	}
}

// branchOnly returns the part of the edge of an arch between the left stem and
// the top, the point farthest from the line at half the x-height mid: the start
// of the edge up to it if first is set, else the rest from it.
func branchOnly(pts []outline.Point, s []float64, mid float64, first bool) ([]outline.Point, []float64) {
	k := 0
	for i, q := range pts {
		if math.Abs(q.Y-mid) > math.Abs(pts[k].Y-mid) {
			k = i
		}
	}
	if first {
		return pts[:k+1], s[:k+1]
	}
	return pts[k:], s[k:]
}
