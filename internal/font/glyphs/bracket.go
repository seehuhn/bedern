// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// Parentheses, brackets, slash and backslash.
//
// The parentheses and the brackets reach from BracketOver above the cap height
// to BracketUnder below the descender line.  A parenthesis is a crescent
// between two arcs of ellipses which are taller than they are wide: the outer
// arc reaches ParenSagitta from the line through its ends, and its tangent at
// the cut ends is ParenAngle degrees from the vertical.  The inner arc is the
// same kind of curve, flatter, so that the stroke is ParenThick thick at the
// middle and ParenEnd (measured across) at the ends; both ends are cut flat.
// The curvature of an ellipse rises towards the ends.  A bracket is a stem with
// two arms of the thickness of the thin horizontals.

// parenShape holds the two ellipses of a parenthesis, in the frame in which the
// leftmost point of the outer edge is on the abscissa 0 and the centre of the
// ellipses at the height yc: the half axes a and b of the outer one, a2 of the
// inner one (which has the vertical half axis b too), and the half angle t of
// the arc.
type parenShape struct {
	a, a2, b, t, yc, h float64
	thick              float64 // the thickness at the middle
}

// newParenShape returns the shape of the parenthesis of the parameters.
func newParenShape(p design.Params) parenShape {
	q := p.Punct
	top, bottom := p.BracketTop(), p.BracketBottom()
	s := parenShape{yc: (top + bottom) / 2, h: (top - bottom) / 2, thick: q.ParenThick}
	// With cos t = c: the sagitta is a(1-c), the half height is b sin t and the
	// tangent at the end has the slope a sin t / (b cos t), which is tan(angle).
	k := q.ParenSagitta / (s.h * math.Tan(degrees(q.ParenAngle)))
	c := k / (1 - k)
	s.t = math.Acos(c)
	s.a = q.ParenSagitta / (1 - c)
	s.b = s.h / math.Sin(s.t)
	s.a2 = (q.ParenSagitta + q.ParenEnd - q.ParenThick) / (1 - c)
	return s
}

// edge returns the cubic curves of the ellipse with the half axes a and b, whose
// leftmost point is at the abscissa x0, from the parameter angle from to the
// parameter angle to (through 0).
func (s parenShape) edge(x0, a float64, from, to float64) chain {
	at := func(f float64) outline.Point {
		return outline.Point{X: x0 + a*(1-math.Cos(f)), Y: s.yc + s.b*math.Sin(f)}
	}
	dir := func(f float64) outline.Point { return outline.Point{X: a * math.Sin(f), Y: s.b * math.Cos(f)} }
	var ch chain
	for _, seg := range [][2]float64{{from, 0}, {0, to}} {
		f0, f1 := seg[0], seg[1]
		k := 4.0 / 3 * math.Tan((f1-f0)/4)
		p0, p3 := at(f0), at(f1)
		d0, d1 := dir(f0), dir(f1)
		ch = append(ch, curveSeg(cubic{p0,
			outline.Point{X: p0.X + k*d0.X, Y: p0.Y + k*d0.Y},
			outline.Point{X: p3.X - k*d1.X, Y: p3.Y - k*d1.Y},
			p3}))
	}
	return ch
}

// parenLeft returns the left parenthesis as a closed counter-clockwise chain.
func parenLeft(p design.Params) chain {
	s := newParenShape(p)
	outer := s.edge(0, s.a, s.t, -s.t)        // from the top to the bottom
	inner := s.edge(s.thick, s.a2, -s.t, s.t) // from the bottom to the top
	return join(outer,
		chain{lineSeg(outer.end(), inner.start())},
		inner,
		chain{lineSeg(inner.end(), outer.start())},
	)
}

// parenleft draws the left parenthesis.
func parenleft(p design.Params) (outline.Path, float64) {
	var path outline.Path
	parenLeft(p).addContour(&path)
	return placeInk(path, p.Punct.ParenLeft)
}

// parenright draws the right parenthesis: the mirror image of the left one.
func parenright(p design.Params) (outline.Path, float64) {
	var path outline.Path
	parenLeft(p).addContour(&path)
	q := p.Punct.ParenLeft
	return placeInk(mirrorInk(path, 0), design.PunctSides{Left: q.Right, Right: q.Left})
}

// bracketLeft returns the left bracket: the stem and two arms.
func bracketLeft(p design.Params) outline.Path {
	q := p.Punct
	top, bottom := p.BracketTop(), p.BracketBottom()
	x := q.BracketStem + q.BracketReach
	var path outline.Path
	addPolygon(&path,
		xy(0, bottom), xy(x, bottom), xy(x, bottom+q.BracketArm), xy(q.BracketStem, bottom+q.BracketArm),
		xy(q.BracketStem, top-q.BracketArm), xy(x, top-q.BracketArm), xy(x, top), xy(0, top))
	return path
}

// bracketleft draws the left square bracket.
func bracketleft(p design.Params) (outline.Path, float64) {
	return placeInk(bracketLeft(p), p.Punct.BracketLeft)
}

// bracketright draws the right square bracket: the mirror image of the left one.
func bracketright(p design.Params) (outline.Path, float64) {
	q := p.Punct.BracketLeft
	return placeInk(mirrorInk(bracketLeft(p), 0), design.PunctSides{Left: q.Right, Right: q.Left})
}

// slashBar returns the slash: a parallelogram with flat ends from the foot of
// the slash to its top, leaning to the right by the angle SlashAngle from the
// vertical, SlashThick thick across the stroke.
func slashBar(p design.Params) outline.Path {
	q := p.Punct
	top, bottom := p.SlashTop(), p.SlashBottom()
	a := degrees(q.SlashAngle)
	w := q.SlashThick / math.Cos(a)
	run := (top - bottom) * math.Tan(a)
	var path outline.Path
	addPolygon(&path, xy(0, bottom), xy(w, bottom), xy(run+w, top), xy(run, top))
	return path
}

// slash draws the slash.
func slash(p design.Params) (outline.Path, float64) {
	return placeInk(slashBar(p), p.Punct.Slash)
}

// backslash draws the backslash: the mirror image of the slash.
func backslash(p design.Params) (outline.Path, float64) {
	return placeInk(mirrorInk(slashBar(p), 0), p.Punct.Backslash)
}
