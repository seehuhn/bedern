// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The stem of the capitals and its serifs.
//
// A capital stem stands on the baseline and reaches the cap height, with a
// bracketed serif on each side at the foot and at the head, any of which may be
// left out where an arm or a bar takes its place.  The foot serifs are the foot
// serifs of the lowercase (footUp and footDown in stem.go) with the serif values
// of the capitals (capSerifParams).  A head serif is the foot serif of the other
// side turned through half a turn: the flat top lies on the cap height, the tip
// is SerifThick thick and the bracket leaves the stem BracketHeight below the
// top.
//
// The parts return chains, which a letter joins (join) into its contour and then
// merges (mergeLines): a straight edge made of two parts, such as the baseline
// under a stem and its foot serif, becomes one line.  The parts are, for a
// stem whose left edge is at x and right edge at x + Stem:
//
//	capFootRight   from the foot of the right-hand edge, along the baseline to
//	               the tip of the right foot serif, round it and up the bracket to
//	               the edge at BracketHeight
//	capFootLeft    from the left-hand edge at BracketHeight down the bracket and
//	               round the tip of the left foot serif to the baseline
//	capHeadRight   from the right-hand edge BracketHeight below the top, up the
//	               bracket and round the tip of the right head serif to the top
//	capHeadLeft    from the top of the left-hand edge, along the top to the tip of
//	               the left head serif, round it and down the bracket to the edge
//	               BracketHeight below the top
//
// A capStem puts them together for a stem with the serifs it is given, in the
// three pieces up, top and down, as stemUp, stemCap and stemDown do for the
// lowercase: a letter with a bar or arms puts its other parts between them.

// capSerifParams returns p with the serif values of the capitals in place of
// those of the lowercase, for the foot serif code of stem.go.
func capSerifParams(p design.Params) design.Params {
	q := p
	q.SerifThick = p.Caps.SerifThick
	q.BracketHeight = p.Caps.BracketHeight
	q.BracketReach = p.Caps.BracketReach
	return q
}

// pathChain returns what draw adds to a path whose current point is from, as a
// chain.
func pathChain(from outline.Point, draw func(path *outline.Path)) chain {
	var path outline.Path
	path.MoveTo(from.X, from.Y)
	draw(&path)
	var c chain
	cur := from
	for _, s := range path[0].Segs {
		if s.IsCurve {
			c = append(c, curveSeg(cubic{cur, s.P1, s.P2, s.P3}))
		} else {
			c = append(c, lineSeg(cur, s.P3))
		}
		cur = s.P3
	}
	return c
}

// halfTurn returns the chain turned through half a turn about the point (cx, cy).
// A rotation keeps the orientation and the order of the pieces.
func (c chain) halfTurn(cx, cy float64) chain {
	return c.mapPoints(func(q outline.Point) outline.Point {
		return outline.Point{X: 2*cx - q.X, Y: 2*cy - q.Y}
	})
}

// capFootRight returns the foot serif on the right of a stem whose right-hand
// edge is at x, of the length jut, from (x, 0) to (x, BracketHeight).
func capFootRight(p design.Params, x, jut float64) chain {
	cp := capSerifParams(p)
	return pathChain(outline.Point{X: x}, func(path *outline.Path) {
		footUp(path, cp, x, 0, jut, cp.SerifThick)
	})
}

// capFootLeft returns the foot serif on the left of a stem whose left-hand edge
// is at x, of the length jut, from (x, BracketHeight) to (x - jut, 0).
func capFootLeft(p design.Params, x, jut float64) chain {
	cp := capSerifParams(p)
	return pathChain(outline.Point{X: x, Y: cp.BracketHeight}, func(path *outline.Path) {
		footDown(path, cp, x, 0, jut)
	})
}

// capHeadRight returns the head serif on the right of a stem whose right-hand
// edge is at x and whose top is at top: the left foot serif turned through half
// a turn, from (x, top - BracketHeight) to (x + jut, top).
func capHeadRight(p design.Params, x, top, jut float64) chain {
	return capFootLeft(p, 0, jut).halfTurn(x/2, top/2)
}

// capHeadLeft returns the head serif on the left of a stem whose left-hand edge
// is at x and whose top is at top: the right foot serif turned through half a
// turn, from (x, top) to (x, top - BracketHeight).
func capHeadLeft(p design.Params, x, top, jut float64) chain {
	return capFootRight(p, 0, jut).halfTurn(x/2, top/2)
}

// A capStem says which serifs a capital stem has: the length of each, zero for
// none.  The stem is Caps.Stem wide, from the baseline to the cap height.
type capStem struct {
	Left                float64 // the left-hand edge
	FootLeft, FootRight float64
	HeadLeft, HeadRight float64
}

func (s capStem) right(p design.Params) float64 { return s.Left + p.Caps.Stem }

// start returns the point at which the contour of a lone stem starts, the tip of
// the left foot serif on the baseline (the foot of the left-hand edge without
// one).
func (s capStem) start() outline.Point { return outline.Point{X: s.Left - s.FootLeft} }

// up returns the right-hand side of the stem from the baseline: from the point
// from on the baseline, left of the stem, along the baseline and the right foot
// serif, up the right-hand edge to the height to.
func (s capStem) up(p design.Params, from outline.Point, to float64) chain {
	x := s.right(p)
	if s.FootRight == 0 {
		return join(lineChain(from, outline.Point{X: x}), lineChain(outline.Point{X: x}, outline.Point{X: x, Y: to}))
	}
	f := capFootRight(p, x, s.FootRight)
	return join(lineChain(from, f.start()), f, lineChain(f.end(), outline.Point{X: x, Y: to}))
}

// top returns the top of the stem: from the point from on the right-hand edge,
// up to the top (with the right head serif), along it (with the left head serif)
// and down the left-hand edge to the height to.
func (s capStem) top(p design.Params, from outline.Point, to float64) chain {
	x, cap := s.right(p), p.CapHeight
	var c chain
	if s.HeadRight > 0 {
		h := capHeadRight(p, x, cap, s.HeadRight)
		c = join(lineChain(from, h.start()), h)
	} else {
		c = lineChain(from, outline.Point{X: x, Y: cap})
	}
	if s.HeadLeft > 0 {
		h := capHeadLeft(p, s.Left, cap, s.HeadLeft)
		c = join(c, lineChain(c.end(), h.start()), h)
	} else {
		c = join(c, lineChain(c.end(), outline.Point{X: s.Left, Y: cap}))
	}
	return join(c, lineChain(c.end(), outline.Point{X: s.Left, Y: to}))
}

// down returns the foot of the left-hand edge: from the point from on the edge
// down to the baseline, round the left foot serif if there is one, ending at
// start.
func (s capStem) down(p design.Params, from outline.Point) chain {
	if s.FootLeft == 0 {
		return lineChain(from, outline.Point{X: s.Left})
	}
	f := capFootLeft(p, s.Left, s.FootLeft)
	return join(lineChain(from, f.start()), f)
}

// mergeLines returns the chain with every run of straight pieces which lie on
// one line made into one piece, and without straight pieces of length zero.  The
// chain is taken to be closed: the last piece and the first are merged too, and
// the result then starts at the start of the merged piece.
func mergeLines(c chain) chain {
	collinear := func(a, b piece) bool {
		if !a.line || !b.line {
			return false
		}
		u, v := unit(a.P0, a.P3), unit(b.P0, b.P3)
		return math.Abs(u.X*v.Y-u.Y*v.X) < 1e-9 && u.X*v.X+u.Y*v.Y > 0
	}
	var out chain
	for _, q := range c {
		if q.line && math.Hypot(q.P3.X-q.P0.X, q.P3.Y-q.P0.Y) < 1e-9 {
			continue
		}
		if n := len(out); n > 0 && collinear(out[n-1], q) {
			out[n-1] = lineSeg(out[n-1].P0, q.P3)
			continue
		}
		out = append(out, q)
	}
	if n := len(out); n > 1 && collinear(out[n-1], out[0]) {
		first := lineSeg(out[n-1].P0, out[0].P3)
		out = append(chain{first}, out[1:n-1]...)
	}
	return out
}

// capContour returns the closed chain c as a path of one contour.
func capContour(c chain) outline.Path {
	if s, e := c.start(), c.end(); math.Hypot(s.X-e.X, s.Y-e.Y) > 1e-9 {
		panic("capital contour is not closed")
	}
	return contourPath(mergeLines(c))
}
