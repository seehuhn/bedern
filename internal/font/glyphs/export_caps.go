// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// CapStemOutline returns the outline of a lone capital stem, one contour: the
// stem is Caps.Stem wide with its left-hand edge at left and reaches from the
// baseline to top, with bracketed serifs of the given lengths at the foot and
// at the head on either side (zero for none).
func CapStemOutline(p design.Params, left, top, footLeft, footRight, headLeft, headRight float64) outline.Path {
	q := p
	q.CapHeight = top
	s := capStem{Left: left, FootLeft: footLeft, FootRight: footRight, HeadLeft: headLeft, HeadRight: headRight}
	lo, hi := 0.0, top
	if footLeft > 0 || footRight > 0 {
		lo = p.Caps.BracketHeight
	}
	if headLeft > 0 || headRight > 0 {
		hi = top - p.Caps.BracketHeight
	}
	up := s.up(q, s.start(), (lo+hi)/2)
	tp := s.top(q, up.end(), (lo+hi)/2)
	return capContour(join(up, tp, s.down(q, tp.end())))
}

// CapArmOutline returns the outline of an arm of the capitals, one contour,
// from the vertical line x = edge to the vertical cut at end, which lies right
// of it, with its underside at lower and its upper edge at upper.  At the cut
// a vertical serif reaches drop below the underside and rise above the upper
// edge, as on the arms of E (zero for none); the contour is closed along the
// line x = edge.
func CapArmOutline(p design.Params, edge, end, lower, upper, drop, rise float64) outline.Path {
	a := capArm{Edge: edge, End: end, Lower: lower, Upper: upper, Drop: drop, Rise: rise, SerifThick: p.Caps.ArmSerifThick}
	lo, up := outline.Point{X: edge, Y: lower}, outline.Point{X: edge, Y: upper}
	var c chain
	switch {
	case drop > 0 && rise > 0:
		m := capArmMiddle(a)
		c = join(m, lineChain(m.end(), lo))
	case rise > 0:
		b := capArmBottom(a)
		c = join(b, lineChain(b.end(), lo), lineChain(lo, b.start()))
	case drop > 0:
		t := capArmTop(a)
		c = join(t, lineChain(t.end(), up), lineChain(up, lo))
	default:
		c = join(lineChain(lo, outline.Point{X: end, Y: lower}), lineChain(outline.Point{X: end, Y: lower}, outline.Point{X: end, Y: upper}),
			lineChain(outline.Point{X: end, Y: upper}, up), lineChain(up, lo))
	}
	return capContour(c)
}

// SuperEllipseOutline returns the closed curve of four cubic segments which
// touches the rectangle from (x0, y0) to (x1, y1) at the midpoints of its
// sides, with the squareness superness, counter-clockwise if ccw is set and
// clockwise otherwise.
func SuperEllipseOutline(x0, y0, x1, y1, superness float64, ccw bool) outline.Path {
	var path outline.Path
	superEllipse(&path, x0, y0, x1, y1, superness, ccw)
	return path
}
