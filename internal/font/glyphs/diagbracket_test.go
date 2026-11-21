// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/outline"
)

// The tests of this file measure the outlines of the letters, in every parameter
// set, for the brackets of the serifs on diagonals and on stems.

// A bracketRun is a run of consecutive curve segments of a contour between two
// straight segments, with the directions of the straight segments on both sides.
type bracketRun struct {
	segs       []cubic
	before     outline.Point // the direction of the line before the run
	after      outline.Point // the direction of the line after it
	start, end outline.Point
}

// bracketRuns returns the runs of curves of the contour ct.
func bracketRuns(ct outline.Contour) []bracketRun {
	n := len(ct.Segs)
	nodes := contourNodes(ct)
	isCurve := func(i int) bool { return ct.Segs[(i%n+n)%n].IsCurve }
	dir := func(i int) outline.Point {
		i = (i%n + n) % n
		return unit(nodes[i], ct.Segs[i].P3)
	}
	var runs []bracketRun
	for i := 0; i < n; i++ {
		if !isCurve(i) || isCurve(i-1) {
			continue
		}
		r := bracketRun{before: dir(i - 1), start: nodes[i]}
		j := i
		for ; isCurve(j); j++ {
			s := ct.Segs[j%n]
			r.segs = append(r.segs, cubic{nodes[j%n], s.P1, s.P2, s.P3})
		}
		r.end = r.segs[len(r.segs)-1].P3
		r.after = dir(j)
		runs = append(runs, r)
	}
	return runs
}

func (r bracketRun) startTangent() outline.Point {
	c := r.segs[0]
	if c.P1 == c.P0 {
		return unit(c.P0, c.P2)
	}
	return unit(c.P0, c.P1)
}

func (r bracketRun) endTangent() outline.Point {
	c := r.segs[len(r.segs)-1]
	if c.P2 == c.P3 {
		return unit(c.P1, c.P3)
	}
	return unit(c.P2, c.P3)
}

// concave reports whether the run turns one way only, by the sign of the turn
// from its start to its end.
func (r bracketRun) concave() bool {
	var heads []float64
	for _, c := range r.segs {
		pc := curveSeg(c)
		for k := 0; k <= 20; k++ {
			tg := pc.tangent(float64(k) / 20)
			heads = append(heads, math.Atan2(tg.Y, tg.X))
		}
	}
	total := turn(heads[0], heads[len(heads)-1])
	for k := 1; k < len(heads); k++ {
		if d := turn(heads[k-1], heads[k]); d*total < -1e-9 {
			return false
		}
	}
	return true
}

func parallel(a, b outline.Point) bool {
	return math.Abs(a.X*b.Y-a.Y*b.X) < 1e-6 && a.X*b.X+a.Y*b.Y > 0
}

// diagBracketBand is the band in which the run of a bracket on a diagonal, along
// the edge from the face of the serif, lies as a multiple of the run of the bracket
// on a stem: as generous on the obtuse side, where the corner is
// open, as about twice that of the stem, and on the acute side at least about as
// generous as that of the stem.
var diagBracketBand = [2]float64{0.9, 2.2}

// The brackets of the serifs on the diagonals are tangent to the face of the serif
// and to the edge, concave, and have a run within diagBracketBand of the bracket of
// the stem, and on every diagonal letter the run on the obtuse side is longer than
// that on the stem.
func TestDiagonalBracketsAreThoseOfTheStem(t *testing.T) {
	t.Parallel()
	letters := []string{"v", "w", "x", "y", "k"}
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			p := np.params
			for _, l := range letters {
				thick, height, top := p.SerifThick, p.BracketHeight, p.XHeight
				stemRun := height - thick
				faces := []float64{top - thick, thick}
				path, _ := cachedOutline(np, l)
				var found int
				var longest float64
				for _, ct := range path {
					for _, r := range bracketRuns(ct) {
						// A bracket on a diagonal leaves a horizontal face at one
						// end and meets a straight edge which is neither horizontal
						// nor vertical at the other.
						face, edge, edgeEnd, faceEnd := r.before, r.after, r.end, r.start
						tFace, tEdge := r.startTangent(), r.endTangent()
						if math.Abs(r.before.Y) > 1e-9 {
							face, edge, edgeEnd, faceEnd = r.after, r.before, r.start, r.end
							tFace, tEdge = r.endTangent(), r.startTangent()
						}
						onFace := false
						for _, y := range faces {
							onFace = onFace || math.Abs(faceEnd.Y-y) < 1e-6
						}
						if math.Abs(face.Y) > 1e-9 || !onFace || math.Abs(edge.X) < 1e-9 || math.Abs(edge.Y) < 1e-9 {
							continue
						}
						found++
						if !parallel(tFace, face) {
							t.Errorf("%s: the bracket from %v leaves the face of the serif in the direction %v, not along it", l, faceEnd, tFace)
						}
						if !parallel(tEdge, edge) {
							t.Errorf("%s: the bracket from %v meets the edge at %v in the direction %v, not along it (%v)", l, faceEnd, edgeEnd, tEdge, edge)
						}
						if !r.concave() {
							t.Errorf("%s: the bracket from %v to %v is not concave", l, r.start, r.end)
						}
						run := math.Abs(edgeEnd.Y - faceEnd.Y)
						if f := run / stemRun; f < diagBracketBand[0] || f > diagBracketBand[1] {
							t.Errorf("%s: the bracket from %v runs %.1f from the face, %.2f of the %.1f of the stem bracket, want %v to %v",
								l, faceEnd, run, f, stemRun, diagBracketBand[0], diagBracketBand[1])
						}
						longest = math.Max(longest, run)
					}
				}
				if found < 4 {
					t.Errorf("%s: %d brackets on diagonals found, want at least 4", l, found)
				}
				if longest <= stemRun {
					t.Errorf("%s: the longest bracket on a diagonal runs %.1f, not more than the %.1f of the stem bracket", l, longest, stemRun)
				}
			}
		})
	}
}
