// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// eFrame holds the shapes and the points from which e is drawn.
type eFrame struct {
	left, right  float64 // extremes of the bowl
	cx           float64 // the extremes at top and bottom
	thickY       float64 // height of the left extreme, where the stroke is thickest
	barLow       float64 // lower edge of the bar
	barHigh      float64 // upper edge of the bar
	outerU       superShape
	outerD       superShape
	innerU       superShape
	innerD       superShape
	eye          superShape // the upper half of the eye, from the upper edge of the bar
	innerBottomY float64
}

// eThickY returns the height of the left extreme of e, where the stroke is
// thickest.
func eThickY(p design.Params) float64 {
	bottom, top := -p.OvershootLower, p.XHeight+p.OvershootLower
	return bottom + p.EThickPos*(top-bottom)
}

// eBar returns the heights of the underside and of the upper edge of the bar of
// e, which are also those of the counter and of the outer edge of the bowl of a
// where they meet its stem (aBowl).  The bar is EBarThick thick and centred on
// EBarHeight, and its underside lies at least 4 above the left extreme of the
// eye, where the stroke is vertical.
func eBar(p design.Params) (low, high float64) {
	low = math.Max(p.EBarHeight-p.EBarThick/2, eThickY(p)+4)
	return low, low + p.EBarThick
}

func newEFrame(p design.Params) eFrame {
	return newEFrameWidth(p, p.EBowlWidth)
}

// newEFrameWidth returns the frame of a bowl of the given width, with the left
// extreme at the left side space: the bowl of e, and with another width that of
// c, which has the same curves.
func newEFrameWidth(p design.Params, width float64) eFrame {
	return newEFrameAt(p, design.EBowlLeft(p), width)
}

// newEFrameAt returns the frame of a bowl of the given width whose left extreme
// is at left: that of e at EBowlLeft, that of c at CBowlLeft.
func newEFrameAt(p design.Params, left, width float64) eFrame {
	var f eFrame
	f.left, f.right = left, left+width
	f.cx = (f.left + f.right) / 2
	bottom, top := -p.OvershootLower, p.XHeight+p.OvershootLower
	f.thickY = eThickY(p)
	f.barLow, f.barHigh = eBar(p)
	// The upper and the lower half of each contour are halves of
	// super-ellipses whose left and right extremes lie at the height of the
	// thickest part; the halves differ in height.
	y := f.thickY
	f.outerU = newSuperShape(f.left, y-(top-y), f.right, top, p.SupernessOuter)
	f.outerD = newSuperShape(f.left, bottom, f.right, y+(y-bottom), p.SupernessOuter)
	ti, bi := top-p.ThinLower, bottom+p.ThinLower
	f.innerU = newSuperShape(f.left+p.CurveLower, y-(ti-y), f.right-p.CurveLower, ti, p.SupernessInner)
	f.innerD = newSuperShape(f.left+p.CurveLower, bi, f.right-p.CurveLower, y+(y-bi), p.SupernessInner)
	// The eye has its own shape, whose vertical tangents at the left and the
	// right extreme lie at the upper edge of the bar; its sides leave the bar
	// vertically and its top is a round arch.
	f.eye = newSuperShape(f.left+p.CurveLower, f.barHigh-(ti-f.barHigh), f.right-p.CurveLower, ti, p.EEyeSuperness)
	f.innerBottomY = bi
	return f
}

// e draws the lowercase letter e as one outer contour and one counter, the
// eye above the bar.
//
// The outer contour runs counter-clockwise.  It starts where the underside of
// the bar meets the inner edge of the left stroke, and runs along the underside
// to the right end of the bar, which is the point where the outer edge of the
// bowl, on the right, is at the height of that underside.  There it turns the
// corner and goes up the outer edge of the bowl, which is the upper right
// quarter of the outer super-ellipse of the o, cut at the underside of the bar:
// the stroke is as thick there as in the o and the
// inner edge of the right side is that of the o.  Over the top and down the
// left side the contour is the outer edge of the o, with its thin part at the
// top, its thickest part at the left and its overshoot; the left extreme lies
// at EThickPos of the height, a little below the middle.  Through the lower
// left quarter it is still the outer edge of the o, and the lower right quarter
// is cut at the height of the tip of the terminal, which lies EAperture below the
// underside of the bar: the outer edge of the terminal is the outer edge of the
// o as far as that.  The tip is a straight cut at right angles to the stroke,
// ETerminalTip long.  From its upper corner the upper edge of the terminal runs
// back, in one curve (eShape), to the lowest point of the inner edge, where it
// is horizontal and the stroke is as thin as the bottom of the o, and continues
// as the inner edge of the o, round the lower left quarter and up the left side
// to the underside of the bar.  The inner edge of the left stroke between the
// underside and the upper edge of the bar is hidden in the bar.  The right side
// below the bar is open: no ink lies to the right of the left stroke between the
// terminal and the bar.
//
// The eye is a second contour, clockwise: it starts at the upper edge of the
// bar on the left, runs up the sides and over the top of its own super-ellipse
// (EEyeSuperness), whose vertical tangents lie at the upper edge of the bar, and
// down to the upper edge of the bar on the right, and back along that edge.
func e(p design.Params) (outline.Path, float64) {
	f := newEFrame(p)

	// The right end of the bar and the left end of its underside are found on
	// the curves of the o at the height of the underside of the bar; the left
	// end of the upper edge at its height.
	q := f.outerU.cubic(0, true).fromY(f.barLow)
	innerL := f.eye.cubic(2, false) // left extreme, at the upper edge of the bar, to the top
	toUnderside := outline.Point{X: innerL.P0.X, Y: f.barLow}

	// The terminal, cut at the height of its tip below the underside of the bar.
	terminal := lowerTerminal(f.outerD.cubic(3, true), // bottom to the right extreme
		f.barLow-p.EAperture, p.ETerminalTip, outline.Point{X: f.cx, Y: f.innerBottomY}, eShape)

	var path outline.Path
	path.MoveTo(toUnderside.X, toUnderside.Y)
	path.LineTo(q.P0.X, q.P0.Y)
	path.CurveTo(q.args())
	f.outerU.segment(&path, 1, true) // top to the left extreme
	f.outerD.segment(&path, 2, true) // left extreme to the bottom
	terminal.appendTo(&path)
	f.innerD.segment(&path, 1, false)         // bottom to the left extreme
	path.LineTo(toUnderside.X, toUnderside.Y) // up to the underside of the bar

	// The eye.
	path.MoveTo(innerL.P0.X, innerL.P0.Y)
	path.CurveTo(innerL.args())
	path.CurveTo(f.eye.cubic(3, false).args()) // top to the right extreme
	path.LineTo(innerL.P0.X, innerL.P0.Y)
	return path, design.EAdvance(p)
}
