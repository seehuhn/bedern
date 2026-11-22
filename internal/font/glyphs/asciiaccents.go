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
// the stroke is TildeThin thick, its edges concentric with the centre line, and
// the ends are cut at right angles to the stroke; there each edge has the
// handle of a circular arc.
//
// Side kinds: open, TildeSide to the ink.
func asciitilde(p design.Params) (outline.Path, float64) {
	s := p.Sym
	x0, w, ax := 0.0, s.TildeWidth, p.Punct.MathAxis
	a := (s.TildeHeight - s.TildeThin) / 2 // the amplitude of the centre line
	end := 0.84 * s.TildeThick
	sts := []station{
		{X: x0, Y: ax - 0.5*a, Dir: 42, W: end, ArcEnd: true},
		{X: x0 + 0.27*w, Y: ax + a, Dir: 0, W: s.TildeThin},
		{X: x0 + 0.5*w, Y: ax, Dir: -40, W: s.TildeThick},
		{X: x0 + 0.73*w, Y: ax - a, Dir: 0, W: s.TildeThin},
		{X: x0 + w, Y: ax + 0.5*a, Dir: 42, W: end, ArcEnd: true},
	}
	// At the crest and the trough the centre line has the mean curvature of
	// the two circles which touch it there and pass through the centres of the
	// stations before and after; the edges are concentric with it.
	for _, i := range []int{1, 3} {
		t := sts[i].dir()
		var k float64
		for _, j := range []int{i - 1, i + 1} {
			d := outline.Point{X: sts[j].X - sts[i].X, Y: sts[j].Y - sts[i].Y}
			k += math.Abs(t.X*d.Y-t.Y*d.X) / (d.X*d.X + d.Y*d.Y)
		}
		if i == 1 {
			k = -k
		}
		sts[i].K = k
	}
	return placeInk(chainsPath(strokeChain(sts)), design.PunctSides{Left: s.TildeSide, Right: s.TildeSide})
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
// (GraveHead and GraveTip across) and two straight lines which run into both on
// lead spirals (leadspiral.go), GraveWidth wide and GraveHeight high with its top at GraveTop.  The head is
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

// wedgeChain returns the counter-clockwise outline of the circle about c0 of
// the radius r0 and the circle about c1 of the radius r1, joined by two lines
// which run into both circles on lead spirals (leadWedge).
func wedgeChain(c0 outline.Point, r0 float64, c1 outline.Point, r1 float64) chain {
	ch, err := leadWedge(c0, r0, c1, r1).chain()
	if err != nil {
		panic("glyphs: the wedge: " + err.Error())
	}
	return ch
}
