// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The radical of text.
//
// The radical is four straight strokes in one contour: the tick, a short
// thin stroke which rises from the left end of the ink to the peak; the thick
// stroke from the peak down to the foot, which is cut flat; the long stroke
// from the foot up to the top; and the bar, whose top is the top of the sign.
// The thick stroke has the weight of the parentheses (Punct.ParenThick), the
// tick that of the bars of the capitals (Caps.Bar), and the long stroke and
// the bar the stroke of the mathematical signs (Punct.MathThick).  The sign
// reaches from radicalDepth below the baseline to the top of the parentheses.
// It has the side space of the mathematical signs on the left, and the bar
// ends at the advance.

// Shape constants of the radical.
const (
	radicalDepth   = 100.0 // the foot below the baseline
	radicalWidth   = 0.70  // the ink width, in heights of the sign
	radicalVertex  = 0.32  // the foot from the left end of the ink, in heights
	radicalTickEnd = 0.41  // the middle of the left end of the tick above the foot, in heights
	radicalPeak    = 0.47  // the peak above the foot, in heights
	radicalTick    = 80.0  // the peak from the left end of the ink, less half the thick stroke
	radicalStub    = 2.5   // the underside of the bar to the right of the long stroke, in strokes
	radicalFoot    = 30.0  // the flat at the foot
)

// radicalPath returns the radical with the left end of its ink at the
// abscissa 0, and the right end of its bar.
func radicalPath(p design.Params) (outline.Path, float64) {
	q := p.Punct
	rule, thick, tick := q.MathThick, q.ParenThick, p.Caps.Bar
	bottom, top := -radicalDepth, p.BracketTop()
	h := top - bottom
	end := math.Round(radicalWidth * h)
	v := xy(math.Round(radicalVertex*h), bottom)
	peak := xy(radicalTick+thick/2, bottom+radicalPeak*h)

	// The long stroke: its right edge from the foot to the underside of the
	// bar, and its left edge parallel to it.
	b1 := xy(end-radicalStub*rule, top-rule)
	dl := unit(v, b1)
	l0 := xy(v.X-rule*dl.Y, v.Y+rule*dl.X)
	b2 := xy(l0.X+(top-l0.Y)/dl.Y*dl.X, top)

	// The thick stroke: its left edge from the left end of the foot, its
	// right edge through the peak.
	vl := xy(v.X-radicalFoot, bottom)
	a := math.Atan2(peak.Y-vl.Y, peak.X-vl.X) + math.Asin(thick/dist(vl, peak))
	e := dirOf(a)
	crotch := meetLines(l0, dl, peak, e)

	// The tick: its upper edge from the left end of the ink to the peak, its
	// lower edge parallel to it, cut upright at the left end.
	u0 := xy(0, bottom+radicalTickEnd*h+tick/2)
	dt := unit(u0, peak)
	t0 := xy(u0.X+tick*dt.Y, u0.Y-tick*dt.X)
	l1 := xy(0, t0.Y-t0.X*dt.Y/dt.X)
	m := meetLines(t0, dt, vl, e)

	var path outline.Path
	addPolygon(&path, v, b1, xy(end, top-rule), xy(end, top), b2, crotch, peak, u0, l1, m, vl)
	return path, end
}

// radical draws the radical.
func radical(p design.Params) (outline.Path, float64) {
	path, end := radicalPath(p)
	left := math.Round(p.Punct.MathSide)
	return shiftInk(path, left, 0), left + end
}
