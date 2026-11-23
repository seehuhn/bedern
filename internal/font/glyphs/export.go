// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"errors"
	"fmt"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// Union returns the outline of the ink which the closed contours of the paths
// cover together, with the orientation of doc.go.  The contours must cross
// where they meet, as for union.
func Union(paths ...outline.Path) outline.Path {
	var parts []chain
	for _, p := range paths {
		parts = append(parts, pathChains(p)...)
	}
	return chainsPath(union(parts...)...)
}

// A Station is a point of the centre line of a stroke, as the symbols use it
// (symstroke.go): its point, the direction of the stroke and its thickness,
// with the optional nib, curvature and kind of the run which follows.
type Station = station

// Stroke returns the closed outline of the stroke through the stations, with
// both ends cut along the nib (symstroke.go).
func Stroke(sts []Station) outline.Path {
	return chainsPath(strokeChain(sts))
}

// CheckOutline applies the structural checks of Check to an outline which is
// not a glyph of the table, unrounded and rounded, with the bounding-box rule
// and the sharp crotches of the glyph like.  The glyph like must be in the
// table and must not be a composed letter.
func CheckOutline(p design.Params, like string, path outline.Path, adv float64) error {
	if _, ok := lookup(like); !ok || composites[like] != nil {
		return fmt.Errorf("glyphs: %s is not a glyph of the table which is drawn whole", like)
	}
	if problems := checkStages(p, like, path, adv, path.Round()); len(problems) > 0 {
		return errors.New(problems[0])
	}
	return nil
}

// The parts of the capitals below are open contours: each runs from its
// Start through its segments to its end, and an outline which takes it joins
// it into a contour of its own.  The serifs and their brackets take the serif
// values of the capitals, the serif of an arm Caps.ArmSerifThick.

// openContour returns the chain as an open contour.
func openContour(c chain) outline.Contour {
	var path outline.Path
	c.addContour(&path)
	return path[0]
}

// CapFootLeft returns the foot serif on the left of a capital stem whose
// left-hand edge is at x, jut long: from (x, Caps.BracketHeight) down the
// bracket and round the tip to (x - jut, 0).
func CapFootLeft(p design.Params, x, jut float64) outline.Contour {
	return openContour(capFootLeft(p, x, jut))
}

// CapFootRight returns the foot serif on the right of a capital stem whose
// right-hand edge is at x, jut long: from (x, 0) along the baseline, round the
// tip and up the bracket to (x, Caps.BracketHeight).
func CapFootRight(p design.Params, x, jut float64) outline.Contour {
	return openContour(capFootRight(p, x, jut))
}

// CapHeadLeft returns the head serif on the left of a capital stem whose
// left-hand edge is at x and whose top is at top, jut long: from (x, top)
// along the top, round the tip and down the bracket to (x, top -
// Caps.BracketHeight).
func CapHeadLeft(p design.Params, x, top, jut float64) outline.Contour {
	return openContour(capHeadLeft(p, x, top, jut))
}

// CapHeadRight returns the head serif on the right of a capital stem whose
// right-hand edge is at x and whose top is at top, jut long: from (x, top -
// Caps.BracketHeight) up the bracket, round the tip and along the top to
// (x + jut, top).
func CapHeadRight(p design.Params, x, top, jut float64) outline.Contour {
	return openContour(capHeadRight(p, x, top, jut))
}

// CapArmTop returns an arm of a capital whose serif hangs drop below its
// underside, as the top arm of E: from (edge, lower) along the underside, round
// the serif and up the cut to (end, upper), where end lies right of edge.
func CapArmTop(p design.Params, edge, end, lower, upper, drop float64) outline.Contour {
	return openContour(capArmTop(capArm{Edge: edge, End: end, Lower: lower, Upper: upper, Drop: drop, SerifThick: p.Caps.ArmSerifThick}))
}

// CapArmBottom returns an arm of a capital whose serif rises rise above its
// upper edge, as the bottom arm of E: from (end, lower) up the cut, round the
// serif and along the upper edge to (edge, upper), where end lies right of
// edge.
func CapArmBottom(p design.Params, edge, end, lower, upper, rise float64) outline.Contour {
	return openContour(capArmBottom(capArm{Edge: edge, End: end, Lower: lower, Upper: upper, Rise: rise, SerifThick: p.Caps.ArmSerifThick}))
}
