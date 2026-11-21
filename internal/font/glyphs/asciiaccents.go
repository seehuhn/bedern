// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The ASCII tilde, circumflex and grave.
//
// These are the characters of program text and of TeX input, not combining
// accents: each stands on its own advance.

// asciitilde draws the tilde: one stroke (symstroke.go) which rises from the
// left, crests, falls through the middle, where it is thickest (TildeThick),
// passes a trough and rises to the right, the same turned through half a turn
// about its centre, which lies on the maths axis.  At the crest and the trough
// the stroke is TildeThin thick, and the ends are cut at right angles to the
// stroke.
//
// Side kinds: open, TildeSide to the ends.
func asciitilde(p design.Params) (outline.Path, float64) {
	s := p.Sym
	x0, w, ax := s.TildeSide, s.TildeWidth, p.Punct.MathAxis
	a := (s.TildeHeight - s.TildeThin) / 2 // the amplitude of the centre line
	end := 0.84 * s.TildeThick
	sts := []station{
		{X: x0, Y: ax - 0.5*a, Dir: 42, W: end},
		{X: x0 + 0.27*w, Y: ax + a, Dir: 0, W: s.TildeThin},
		{X: x0 + 0.5*w, Y: ax, Dir: -40, W: s.TildeThick},
		{X: x0 + 0.73*w, Y: ax - a, Dir: 0, W: s.TildeThin},
		{X: x0 + w, Y: ax + 0.5*a, Dir: 42, W: end},
	}
	return chainsPath(strokeChain(sts)), 2*x0 + w
}

// asciicircum draws the circumflex: two straight strokes of CircThick which
// meet in a pointed apex at CircTop, CircWidth wide and CircHeight high, the
// feet cut horizontally.  It is a flat accent raised above the x-height.
//
// Side kinds: open, CircSide to the ends.
func asciicircum(p design.Params) (outline.Path, float64) {
	s := p.Sym
	t := s.CircThick
	cx := s.CircSide + s.CircWidth/2
	// The outer edge runs from the outer corner of the foot, at the width of the
	// ink, to the apex on the axis; the inner edge lies t from it, across the
	// stroke, and is t/sin(th) to the right of it at the foot and meets the
	// axis t/cos(th) below the apex.
	th := math.Atan2(s.CircHeight, s.CircWidth/2)
	yb := s.CircTop - s.CircHeight
	dx := t / math.Sin(th)
	pt := func(x, y float64) outline.Point { return outline.Point{X: x, Y: y} }
	mir := func(q outline.Point) outline.Point { return pt(2*cx-q.X, q.Y) }
	lo, li := pt(cx-s.CircWidth/2, yb), pt(cx-s.CircWidth/2+dx, yb)
	ao, ai := pt(cx, s.CircTop), pt(cx, s.CircTop-t/math.Cos(th))
	var path outline.Path
	path.MoveTo(mir(lo).X, mir(lo).Y)
	for _, q := range []outline.Point{ao, lo, li, ai, mir(li)} {
		path.LineTo(q.X, q.Y)
	}
	return path, 2*s.CircSide + s.CircWidth
}

// grave draws the grave: a wedge which falls from a round head at the upper
// left to a small round tip at the lower right, the outline of two circles
// (GraveHead and GraveTip across) and the two straight lines which touch both,
// GraveWidth wide and GraveHeight high with its top at GraveTop.  The head is
// the heavy end.
//
// Side kinds: open, GraveSideLeft and GraveSideRight to the ink.
func grave(p design.Params) (outline.Path, float64) {
	s := p.Sym
	r0, r1 := s.GraveHead/2, s.GraveTip/2
	x0 := s.GraveSideLeft
	c0 := outline.Point{X: x0 + r0, Y: s.GraveTop - r0}
	c1 := outline.Point{X: x0 + s.GraveWidth - r1, Y: s.GraveTop - s.GraveHeight + r1}
	return chainsPath(wedgeChain(c0, r0, c1, r1)), x0 + s.GraveWidth + s.GraveSideRight
}

// wedgeChain returns the counter-clockwise outline of the two circles about c0
// and c1 of the radii r0 and r1 and of the two straight lines which touch both,
// cut at its extremes.
func wedgeChain(c0 outline.Point, r0 float64, c1 outline.Point, r1 float64) chain {
	phi := math.Atan2(c1.Y-c0.Y, c1.X-c0.X)
	beta := math.Acos((r0 - r1) / math.Hypot(c1.X-c0.X, c1.Y-c0.Y))
	on := func(c outline.Point, r, a float64) outline.Point {
		return outline.Point{X: c.X + r*math.Cos(a), Y: c.Y + r*math.Sin(a)}
	}
	arc1 := arcPieces(c1, r1, phi-beta, phi+beta)
	arc0 := arcPieces(c0, r0, phi+beta, phi-beta+2*math.Pi)
	var ch chain
	for _, part := range []chain{arc1, lineChain(on(c1, r1, phi+beta), on(c0, r0, phi+beta)), arc0, lineChain(on(c0, r0, phi-beta), on(c1, r1, phi-beta))} {
		for _, pc := range part {
			if pc.line {
				ch = append(ch, pc)
			} else {
				ch = append(ch, splitAtExtrema(pc.cubic)...)
			}
		}
	}
	return ch
}
