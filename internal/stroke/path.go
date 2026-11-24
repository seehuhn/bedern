// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package stroke

import (
	"math"

	"seehuhn.de/go/bedern/internal/outline"
)

// Path returns the contours of the result as an outline path, exact (not
// rounded).  Straight curves become lines.
func (r *Result) Path() outline.Path {
	var p outline.Path
	for _, c := range r.Contours {
		p = append(p, chainContour(c))
	}
	return p
}

// chainContour returns the closed chain as a contour.
func chainContour(c Chain) outline.Contour {
	out := outline.Contour{Start: c[0].C[0]}
	for i, pc := range c {
		end := pc.C[3]
		if i == len(c)-1 {
			end = c[0].C[0]
		}
		if pc.C.IsStraight() {
			out.Segs = append(out.Segs, outline.Segment{P3: end})
			continue
		}
		out.Segs = append(out.Segs, outline.Segment{IsCurve: true, P1: pc.C[1], P2: pc.C[2], P3: end})
	}
	return out
}

// Sheared returns the stroke under the shear x' = x + y tan(slant), slant in
// degrees: the positions, directions and curvatures of the centre line, the
// lines of the ends and the centre of a ball are sheared, and the pen is put on
// the sheared centre line, so that the thickness across the stroke stays what
// the nodes say.
func (s Stroke) Sheared(slant float64) Stroke {
	return s.shear(math.Tan(slant * math.Pi / 180))
}

// ShearedRatio is Sheared with the slant given exactly as the ratio num/den of
// its tangent, and puts the stems of the stroke on the lattice of whole units
// of that slant (see lattice).
func (s Stroke) ShearedRatio(num, den int) Stroke {
	out := s.shear(float64(num) / float64(den))
	out.Slant = [2]int{num, den}
	return out
}

// latticeDen returns the height between whole-unit points of a line of the
// slant k = num/den: den, for k the ratio of small integers.
func latticeDen(k float64) float64 {
	for d := 1.0; d <= 64; d++ {
		if n := k * d; math.Abs(n-math.Round(n)) < 1e-9 {
			return d
		}
	}
	return 1
}

func (s Stroke) shear(k float64) Stroke {
	pt := func(q Point) Point { return Point{X: q.X + k*q.Y, Y: q.Y} }
	dirDeg := func(a float64) (float64, float64) {
		d := dirOf(a)
		e := Point{X: d.X + k*d.Y, Y: d.Y}
		return angleOf(e), norm(e)
	}
	out := s
	out.Nodes = make([]Node, len(s.Nodes))
	for i, nd := range s.Nodes {
		a, g := dirDeg(nd.Dir)
		nd.P = pt(nd.P)
		nd.Dir = a
		if !isAuto(nd.K) {
			nd.K /= g * g * g
		}
		if nd.Edge != nil {
			nd.Edge = &EdgeK{Right: nd.Edge.Right / (g * g * g), Left: nd.Edge.Left / (g * g * g)}
		}
		if !isAuto(nd.DW) {
			nd.DW /= g
		}
		nd.DDW /= g * g
		out.Nodes[i] = nd
	}
	end := func(e End) End {
		if e.Kind == Clip || e.Line {
			e.Angle, _ = dirDeg(e.Angle)
		}
		e.Q = pt(e.Q)
		if e.Kind == Ball {
			// The circle stays a circle: only its centre moves.
			e.Ball.InkX += k * e.Ball.Centre.Y
			e.Ball.Centre = pt(e.Ball.Centre)
		}
		return e
	}
	out.Start, out.End = end(s.Start), end(s.End)
	return out
}
