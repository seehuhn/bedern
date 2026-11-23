// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The schwa, the two letters S and the form of i of the feature locl.

// schwa draws ə: the e turned through half a turn about the middle of its
// bowl, so that the bar and the eye stand below the middle and the terminal
// ends at the upper left.  The open side is on the left, with the side space
// of the open side of e, and the round side on the right, with the side space
// of a bowl.
func schwa(p design.Params) (outline.Path, float64) {
	path, _ := e(p)
	x0, _, x1, _ := path.BBox()
	path = turned(path, (x0+x1)/2, p.XHeight/2)
	x0, _, x1, _ = path.BBox()
	dx := math.Round(p.SideOpenRight - x0)
	return shiftInk(path, dx, 0), math.Round(x1 + dx + p.SideRoundRight)
}

// Shape constants of the capital schwa.
const (
	capSchwaWidth = 0.92 // the width of the bowl, in widths of O
	capSchwaBar   = 0.47 // the height of the middle of the bar, in cap heights
)

// capSchwa draws Ə: the schwa at the height of the capitals.  It is drawn as
// the e is drawn, with the strokes, the overshoot and the side spaces of O,
// the bar of the capitals and the aperture and the tip of the terminal of e
// enlarged with the height and the thin stroke, and turned through half a
// turn.
func capSchwa(p design.Params) (outline.Path, float64) {
	path, _ := e(capSchwaParams(p))
	x0, _, x1, _ := path.BBox()
	path = turned(path, (x0+x1)/2, p.CapHeight/2)
	x0, _, x1, _ = path.BBox()
	dx := math.Round(p.Caps.SideRoundLeft - x0)
	return shiftInk(path, dx, 0), math.Round(x1 + dx + p.Caps.SideRoundRight)
}

// capSchwaParams returns the parameters with which e draws the capital schwa
// before it is turned.
func capSchwaParams(p design.Params) design.Params {
	c := p.Caps
	q := p
	q.XHeight = p.CapHeight
	q.OvershootLower = p.OvershootUpper
	q.CurveLower = c.Round
	q.ThinLower = c.RoundThin
	q.EBowlWidth = math.Round(capSchwaWidth * c.OWidth)
	q.EBarThick = c.Bar
	q.EBarHeight = math.Round((1 - capSchwaBar) * p.CapHeight)
	q.EAperture = math.Round(p.EAperture * p.CapHeight / p.XHeight)
	q.ETerminalTip = math.Round(p.ETerminalTip * c.RoundThin / p.ThinLower)
	return q
}

// capSS draws the two letters S side by side, each in the advance of S.
func capSS(p design.Params) (outline.Path, float64) {
	s, adv := capS(p)
	return append(s, shiftInk(s, adv, 0)...), 2 * adv
}
