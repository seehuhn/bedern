// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The musical note: a quaver.
//
// The head is an ellipse which leans to the right, with its lowest ink the
// overshoot of the lowercase below the baseline.  The stem, a hairline
// (Hair) wide, rises from the head to the cap height with its right edge on
// the rightmost point of the head, so that the edge of the head runs into it
// without a corner.  The flag leaves the stem at its top, noteRoot hairlines
// high there, and sweeps down to the right, thinning to an end which is cut
// level: each of its two edges is one cubic curve, which leaves the stem
// downwards to the right and arrives at the end upright, and each of its
// handles reaches noteHandle of the way to the point where the lines through
// its two ends in their directions meet.  The sign is one contour.

// Shape constants of the musical note.
const (
	noteHeadLong  = 162.0 // the long half axis of the head
	noteHeadShort = 108.0 // the short half axis of the head
	noteHeadLean  = 23.0  // the lean of the long axis of the head, degrees above the horizontal
	noteReach     = 200.0 // the end of the flag to the right of the stem
	noteDrop      = 0.52  // the end of the flag below the top, in cap heights
	noteRoot      = 1.7   // the height of the flag at the stem, in hairlines
	noteEnd       = 0.67  // the width of the end of the flag, in hairlines
	noteUpper     = 32.0  // the angle below the horizontal at which the upper edge of the flag leaves the stem, degrees
	noteLower     = 42.0  // that of the lower edge
	noteHandle    = 0.7   // the handles of an edge of the flag, in distances from its ends to the point where the lines through the ends in their directions meet
	noteLeft      = 55.0  // the left side space
	noteRight     = 45.0  // the right side space
)

// leaningEllipse returns the ellipse about c with the half axes a and b,
// whose long axis leans by theta (radians), as a closed counter-clockwise
// chain of four pieces from its rightmost point to its highest, its leftmost,
// its lowest and back.
func leaningEllipse(c outline.Point, a, b, theta float64) chain {
	sin, cos := math.Sincos(theta)
	at := func(q outline.Point) outline.Point {
		return outline.Point{X: c.X + a*q.X*cos - b*q.Y*sin, Y: c.Y + a*q.X*sin + b*q.Y*cos}
	}
	right := math.Atan2(-b*sin, a*cos)
	top := math.Atan2(b*cos, a*sin)
	var ch chain
	for _, s := range [][2]float64{{right, top}, {top, right + math.Pi}, {right + math.Pi, top + math.Pi}, {top + math.Pi, right + 2*math.Pi}} {
		ch = append(ch, arcPieces(outline.Point{}, 1, s[0], s[1]).mapPoints(at)...)
	}
	ch[len(ch)-1].P3 = ch[0].P0
	return ch
}

// musicalnote draws the musical note.
func musicalnote(p design.Params) (outline.Path, float64) {
	theta := degrees(noteHeadLean)
	top := p.CapHeight
	half := math.Hypot(noteHeadLong*math.Sin(theta), noteHeadShort*math.Cos(theta))
	head := leaningEllipse(xy(0, half-p.OvershootLower), noteHeadLong, noteHeadShort, theta)
	e := head.start()
	xs := e.X - p.Hair

	// The flag: the lower edge from the stem to the end and the upper edge
	// back from the end to the top of the stem.
	root := noteRoot * p.Hair
	end := noteEnd * p.Hair
	tip := xy(e.X+noteReach, top-noteDrop*top)
	edge := func(from, to outline.Point, angle float64) cubic {
		// The line on which the edge leaves the stem meets the upright
		// through the end at the height y.
		d := dirOf(-degrees(angle))
		s := (to.X - from.X) / d.X
		y := from.Y + s*d.Y
		h := noteHandle * s
		return cubic{from, xy(from.X+h*d.X, from.Y+h*d.Y), xy(to.X, to.Y+noteHandle*(y-to.Y)), to}
	}
	lower := splitAtExtrema(edge(xy(e.X, top-root), xy(tip.X-end, tip.Y), noteLower))
	upper := splitAtExtrema(edge(xy(e.X, top), tip, noteUpper)).reversed()

	// The first piece of the head runs from its rightmost point to its
	// highest; the left edge of the stem cuts it.
	first := head[0].cubic
	first = first.sub(first.tAtX(xs), 1)
	c := join(
		lineChain(e, lower.start()), lower, lineChain(lower.end(), upper.start()), upper,
		lineChain(upper.end(), xy(xs, top)), lineChain(xy(xs, top), first.P0),
		chain{curveSeg(first)}, head[1:])
	return placeInk(chainsPath(c), design.PunctSides{Left: noteLeft, Right: noteRight})
}
