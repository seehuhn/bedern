// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// A stemSpec describes a vertical stem.  The foot serif on a side is absent
// if its length is zero; the lengths are measured from the edge of the stem.
// The head serif is on the left of the stem, absent if its length is zero.
type stemSpec struct {
	Left, Width float64 // left edge of the stem, and its width
	Bottom, Top float64 // the baseline of the foot serifs, and the flat top
	FootLeft    float64 // length of the foot serif on the left
	FootRight   float64 // length of the foot serif on the right
	FootRise    float64 // thickness of the foot serif on the right at its tip, SerifThick if zero
	HeadJut     float64 // length of the head serif on the left
	HeadDrop    float64 // how far the tip of the head serif lies below Top
}

// right returns the position of the right-hand edge of the stem.
func (s stemSpec) right() float64 { return s.Left + s.Width }

// footRise returns the thickness of the right foot serif at its tip.
func (s stemSpec) footRise(p design.Params) float64 {
	if s.FootRise > 0 {
		return s.FootRise
	}
	return p.SerifThick
}

// start returns the point at which the contour of a letter starts: the left
// tip of the left foot serif, on the baseline of the stem.
func (s stemSpec) start() (x, y float64) { return s.Left - s.FootLeft, s.Bottom }

// stemUp adds the right foot serif of the stem and its right-hand edge up to
// the height to, which lies above the bracket.  The current point is on the
// baseline, to the left of the right foot serif.  The part ends on the
// right-hand edge at the height to.
func stemUp(path *outline.Path, p design.Params, s stemSpec, to float64) {
	if s.FootRight > 0 {
		footUp(path, p, s.right(), s.Bottom, s.FootRight, s.footRise(p))
	} else {
		path.LineTo(s.right(), s.Bottom)
	}
	path.LineTo(s.right(), to)
}

// stemCap adds the top of the stem.  The current point is on the right-hand
// edge, below the top or at it.  The part runs up to the top right corner,
// along the top and, if the stem has a head serif, along that, and ends on
// the left-hand edge: at the top left corner without a head serif, and where
// the underside of the head serif meets the stem with one.
func stemCap(path *outline.Path, p design.Params, s stemSpec) {
	if cur := lastPoint(path); cur.X != s.right() || cur.Y != s.Top {
		path.LineTo(s.right(), s.Top)
	}
	if s.HeadJut > 0 {
		headSerifLeft(path, p, s.Left, s.Top, s.HeadJut, s.HeadDrop)
	} else {
		path.LineTo(s.Left, s.Top)
	}
}

// stemDown adds the left-hand edge of the stem and its left foot serif.  The
// current point is on the left-hand edge, above the bracket.  The part ends
// on the baseline, at the left tip of the foot serif, or at the foot of the
// edge if there is none.
func stemDown(path *outline.Path, p design.Params, s stemSpec) {
	if s.FootLeft > 0 {
		path.LineTo(s.Left, s.Bottom+p.BracketHeight)
		footDown(path, p, s.Left, s.Bottom, s.FootLeft)
	} else {
		path.LineTo(s.Left, s.Bottom)
	}
}

// stemAlone adds a closed contour for a stem on its own.
func stemAlone(path *outline.Path, p design.Params, s stemSpec) {
	path.MoveTo(s.start())
	stemUp(path, p, s, s.Top)
	stemCap(path, p, s)
	stemDown(path, p, s)
}

// footUp adds the foot serif on the right of a stem whose right edge is at x
// and whose baseline is at y0.  The current point is on the baseline, at the
// left end of the serif bottom.  The part runs along the baseline to the tip,
// up the tip, back along the flat top of the serif and up the bracket, and
// ends on the stem edge at the bracket height.  The jut is the length of the
// serif beyond the stem and rise its thickness at the tip.
func footUp(path *outline.Path, p design.Params, x, y0, jut, rise float64) {
	th, bh, br := y0+rise, y0+p.BracketHeight, p.BracketReach
	path.LineTo(x+jut, y0)
	path.LineTo(x+jut, th)
	path.LineTo(x+br, th)
	path.CurveTo(x+br/2, th, x, th+(bh-th)/2, x, bh)
}

// footDown adds the foot serif on the left of a stem whose left edge is at x
// and whose baseline is at y0.  The current point is on the stem edge at the
// bracket height.  The part runs down the bracket, along the flat top of the
// serif to the tip and down the tip, and ends on the baseline at the tip.
func footDown(path *outline.Path, p design.Params, x, y0, jut float64) {
	th, bh, br := y0+p.SerifThick, y0+p.BracketHeight, p.BracketReach
	path.CurveTo(x, th+(bh-th)/2, x-br/2, th, x-br, th)
	path.LineTo(x-jut, th)
	path.LineTo(x-jut, y0)
}

// Shape constants of the underside of the head serif: the lengths of the
// handles of its Bézier curve as fractions of the jut and of the rise, and
// the distance below the lower corner of the tip at which the curve meets
// the stem.
const (
	headUnderHandleX = 0.6  // horizontal handle at the tip, as a fraction of the jut
	headUnderHandleY = 0.3  // vertical handle at the stem, as a fraction of the run
	headUnderRun     = 10.0 // the curve meets the stem this far below the lower corner of the tip
)

// headSerifLeft adds the head serif on the left of a stem whose left edge is
// at x and whose flat top is at height top.  The current point is the top
// right corner of the stem.  The part runs along the top of the stem to its
// left edge, down the sloped upper edge of the serif, which falls by drop
// over the length jut, down the tip, and along one concave curve from the
// lower corner of the tip into the stem.  The curve meets the stem, tangent
// to it, headUnderRun below the height of the lower corner of the tip, and
// the part ends there.  With drop zero the upper edge is flat.
func headSerifLeft(path *outline.Path, p design.Params, x, top, jut, drop float64) {
	th := p.SerifThick
	tip := top - drop // height of the upper edge at the tip
	path.LineTo(x, top)
	path.LineTo(x-jut, tip)
	path.LineTo(x-jut, tip-th)
	path.CurveTo(
		x-headUnderHandleX*jut, tip-th,
		x, tip-th-headUnderHandleY*headUnderRun,
		x, tip-th-headUnderRun)
}

// lastPoint returns the current point of the path.
func lastPoint(path *outline.Path) outline.Point {
	c := (*path)[len(*path)-1]
	if len(c.Segs) == 0 {
		return c.Start
	}
	return c.Segs[len(c.Segs)-1].P3
}
