// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The basic punctuation and the signs of ASCII mathematics.
//
// Punctuation does not take part in the fit of the letters by side kinds.  Each
// sign has its own side spaces (design.PunctSides), measured from the edge of the
// advance to the outermost ink on each side, and its advance is the ink between
// them: the sign is drawn in a frame of its own, and placeInk moves it so that
// its ink keeps the side spaces and returns the advance.  The advance is a whole
// number, and the ink keeps at least the side space it is given.

// The side kinds of the punctuation.  The numbers are far from those of the
// letters.  The kind of a sign holds for both its sides, and punctSideSpace
// says the space of each.
const (
	sidePeriod sideKind = 10000 + iota
	sideComma
	sideColon
	sideSemicolon
	sideExclam
	sideQuote
	sideQuoteDouble
	sideStraight
	sideStraightDouble
	sideHyphen
	sideDash
	sideParen
	sideParenRight
	sideBracket
	sideBracketRight
	sideSlash
	sideBackslash
	sideBar
	sideUnderscore
	sideAsterisk
	sideMath
	sideEmDash
)

// punctKindNames names the side kinds of the punctuation for the table.
var punctKindNames = map[sideKind]string{
	sidePeriod: "period", sideComma: "comma", sideColon: "colon", sideSemicolon: "semicolon",
	sideExclam: "exclam", sideQuote: "quote", sideQuoteDouble: "quoteDouble",
	sideStraight: "straight", sideStraightDouble: "straightDouble", sideHyphen: "hyphen",
	sideDash: "dash", sideParen: "paren", sideParenRight: "parenRight", sideBracket: "bracket",
	sideBracketRight: "bracketRight", sideSlash: "slash", sideBackslash: "backslash",
	sideBar: "bar", sideUnderscore: "underscore", sideAsterisk: "asterisk", sideMath: "math", sideEmDash: "emdash",
}

// punctSideSpace returns the side space of the punctuation sign of the side kind k
// on its left or right side, and false if k is not the kind of a sign.
func punctSideSpace(k sideKind, p design.Params, left bool) (float64, bool) {
	q := p.Punct
	var s design.PunctSides
	switch k {
	case sidePeriod:
		s = q.Period
	case sideComma:
		s = q.Comma
	case sideColon:
		s = q.Colon
	case sideSemicolon:
		s = q.Semicolon
	case sideExclam:
		s = q.Exclam
	case sideQuote:
		s = q.Quote
	case sideQuoteDouble:
		s = q.QuoteDouble
	case sideStraight:
		s = q.Straight
	case sideStraightDouble:
		s = q.StraightDouble
	case sideHyphen:
		s = q.Hyphen
	case sideDash:
		s = q.Dash
	case sideEmDash:
		s = q.EmDash
	case sideParen:
		s = q.ParenLeft
	case sideParenRight:
		s = design.PunctSides{Left: q.ParenLeft.Right, Right: q.ParenLeft.Left}
	case sideBracket:
		s = q.BracketLeft
	case sideBracketRight:
		s = design.PunctSides{Left: q.BracketLeft.Right, Right: q.BracketLeft.Left}
	case sideSlash:
		s = q.Slash
	case sideBackslash:
		s = q.Backslash
	case sideBar:
		s = q.Bar
	case sideUnderscore:
		s = q.Underscore
	case sideAsterisk:
		s = q.Asterisk
	case sideMath:
		s = design.PunctSides{Left: q.MathSide, Right: q.MathSide}
	default:
		return 0, false
	}
	if left {
		return s.Left, true
	}
	return s.Right, true
}

// placeInk moves the path sideways so that its leftmost ink lies s.Left from the
// left edge of the advance, and returns it with the advance, which leaves s.Right
// between the rightmost ink and the right edge.  The shift and the advance are
// whole numbers, and no side space is less than it is asked to be.
func placeInk(path outline.Path, s design.PunctSides) (outline.Path, float64) {
	x0, _, x1, _ := path.BBox()
	dx := math.Ceil(s.Left - x0 - 1e-6)
	out := shiftInk(path, dx, 0)
	return out, math.Ceil(x1 + dx + s.Right - 1e-6)
}

// shiftInk returns the path moved by dx to the right and dy up.
func shiftInk(path outline.Path, dx, dy float64) outline.Path {
	var out outline.Path
	for _, ct := range path {
		c := outline.Contour{Start: outline.Point{X: ct.Start.X + dx, Y: ct.Start.Y + dy}}
		for _, s := range ct.Segs {
			c.Segs = append(c.Segs, outline.Segment{
				IsCurve: s.IsCurve,
				P1:      outline.Point{X: s.P1.X + dx, Y: s.P1.Y + dy},
				P2:      outline.Point{X: s.P2.X + dx, Y: s.P2.Y + dy},
				P3:      outline.Point{X: s.P3.X + dx, Y: s.P3.Y + dy},
			})
		}
		out = append(out, c)
	}
	return out
}

// addPolygon adds the closed polygon through the points as a contour of the path.
// The points run counter-clockwise for an outer contour and clockwise for a
// counter.
func addPolygon(path *outline.Path, pts ...outline.Point) {
	path.MoveTo(pts[0].X, pts[0].Y)
	for _, q := range pts[1:] {
		path.LineTo(q.X, q.Y)
	}
	path.LineTo(pts[0].X, pts[0].Y)
}

// xy is a short name for a point.
func xy(x, y float64) outline.Point { return outline.Point{X: x, Y: y} }

// meetLines returns the point where the line through a with the direction
// da meets the line through b with the direction db.
func meetLines(a, da, b, db outline.Point) outline.Point {
	den := da.X*db.Y - da.Y*db.X
	t := ((b.X-a.X)*db.Y - (b.Y-a.Y)*db.X) / den
	return outline.Point{X: a.X + t*da.X, Y: a.Y + t*da.Y}
}

// mirrorInk returns the path reflected in the vertical line x = at, with every
// contour reversed (outer contours stay counter-clockwise).
func mirrorInk(path outline.Path, at float64) outline.Path {
	var out outline.Path
	for _, ct := range path {
		out = append(out, mirrorInkContour(ct, at))
	}
	return out
}

// mirrorInkContour returns the contour reflected in the vertical line x = at, walked
// the other way round, which keeps its orientation.
func mirrorInkContour(ct outline.Contour, at float64) outline.Contour {
	n := len(ct.Segs)
	pts := make([]outline.Point, n+1)
	pts[0] = ct.Start
	for i, s := range ct.Segs {
		pts[i+1] = s.P3
	}
	refl := func(q outline.Point) outline.Point { return outline.Point{X: 2*at - q.X, Y: q.Y} }
	out := outline.Contour{Start: refl(pts[n])}
	for i := n - 1; i >= 0; i-- {
		s := ct.Segs[i]
		out.Segs = append(out.Segs, outline.Segment{
			IsCurve: s.IsCurve,
			P1:      refl(s.P2),
			P2:      refl(s.P1),
			P3:      refl(pts[i]),
		})
	}
	return out
}

// addDot adds the dot of i, centred at (cx, cy), as a counter-clockwise contour.
func addDot(path *outline.Path, p design.Params, cx, cy float64) {
	r := p.DotDiameter / 2
	superEllipse(path, cx-r, cy-r, cx+r, cy+r, p.DotSuperness, true)
}

// addBar adds the rectangle from x0 to x1 and y0 to y1 as a counter-clockwise
// contour.
func addBar(path *outline.Path, x0, y0, x1, y1 float64) {
	rect(path, x0, y0, x1, y1, true)
}
