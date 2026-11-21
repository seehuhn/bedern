// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"fmt"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// capS draws the capital S as the lowercase s is drawn (s.go): two round bowls,
// the upper one smaller, and a short heavy spine through the centre C, the
// inflection of both long edges, from which the curve to each side extreme
// leaves with no curvature and carries one radius round the extreme; the upper
// half is the lower half turned through half a turn about C and scaled by the
// ratio of the widths.  The values are those of the capitals (the CGSJQ fields
// with the names of the parameters of s), the height that of the round capitals,
// from OvershootUpper below the baseline to OvershootUpper above the cap height,
// and both terminals are the serif beaks of C and G (capSBeak).  The shape
// constants are those of s.
//
// Side kinds: the leftmost ink, the cut of the lower beak, and the rightmost ink,
// the right extreme of the lower bowl, SSideLeft and SSideRight from the edges.
func capS(p design.Params) (outline.Path, float64) {
	f := newCapSFrame(p)
	var path outline.Path
	path.MoveTo(f.C.X+f.LowerE.Inner[0].P0.X, f.C.Y+f.LowerE.Inner[0].P0.Y)
	f.circuit(&path, f.LowerE, 1)
	f.circuit(&path, f.UpperE, -1)
	return path, design.CapSAdvance(p)
}

// capSParams returns p with the values of S in place of those of s, for the
// curve code of s.
func capSParams(p design.Params) design.Params {
	g := p.Caps.CGSJQ
	q := p
	q.XHeight, q.OvershootLower = p.CapHeight, p.OvershootUpper
	q.SLowerWidth, q.SUpperWidth, q.SUpperShift = g.SLowerWidth, g.SUpperWidth, g.SUpperShift
	q.SSpineThick, q.SSpineAngle = g.SSpineThick, g.SSpineAngle
	q.SArcThick, q.SUpperArcThick, q.SSideThick = g.SArcThick, g.SUpperArcThick, g.SSideThick
	q.SUpperDrop, q.SLowerDrop = g.SUpperDrop, g.SLowerDrop
	q.SSideLeft, q.SSideRight = g.SSideLeft, g.SSideRight
	return q
}

// capSShape returns the shape constants of s for S, which are those of s.
func capSShape(p design.Params) sShapeSpec {
	return sShape
}

// capSBeak replaces the ends of the edges of a half of S, from the bottom
// (top) extremes on, by a beak of C and G (beakArc): the outer edge runs from
// the bottom extreme, with the direction and the curvature of the quarter
// before it, to a corner BeakDrop above the bottom on the cut, and the serif
// hangs from the thin arc to the end of the terminal, the drop of the half
// above the bottom; the inner edge leaves the counter at its bottom extreme with
// its direction and curvature and runs into the face of the serif.  The beak
// is drawn in the frame of the beaks of C, which is that of the half turned
// through half a turn.
func capSBeak(p design.Params, e *sEdges, h sHalf) {
	g := p.Caps.CGSJQ
	if err := sBeak(e, h, capBeakSpec{Drop: g.BeakDrop, Arc: g.BeakArc, Lift: g.BeakLift, Run: g.BeakRun}, p.Caps.ArmSerifThick, p.Caps.Bar); err != nil {
		panic(fmt.Sprintf("glyphs: a beak of S: %v; glyphs.Check reports such parameters", err))
	}
}

// newCapSFrame returns the frame of S, as newSFrame does for s, with the beaks
// of capSBeak.
func newCapSFrame(p design.Params) sFrame {
	sh := capSShape(p)
	q := capSParams(p)
	var f sFrame
	top := q.XHeight + q.OvershootLower
	bottom := -q.OvershootLower
	r := q.SUpperWidth / q.SLowerWidth
	u := q.SLowerWidth/2 - q.SUpperShift/(1+r)
	f.C = outline.Point{X: q.SSideLeft + u, Y: (top + r*bottom) / (1 + r)}
	f.Lower = sHalf{Rx: q.SLowerWidth - u, Lx: u, H: f.C.Y - bottom, Drop: q.SLowerDrop}
	f.Upper = sHalf{Rx: r * (q.SLowerWidth - u), Lx: r * u, H: top - f.C.Y, Drop: q.SUpperDrop}
	f.LowerE = sh.edges(q, f.Lower, q.SArcThick)
	f.UpperE = sh.edges(q, f.Upper, q.SUpperArcThick)
	capSBeak(p, &f.LowerE, f.Lower)
	capSBeak(p, &f.UpperE, f.Upper)
	return f
}
