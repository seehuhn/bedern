// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// nArchShape is the shape of the arch of n: the lengths of the handles of its
// Bézier curves as fractions of the extent of their segments, and how far the
// underside peaks to the left of the outer edge.
var nArchShape = archShape{
	UnderPeakShift:    design.ArchUnderPeakShift,
	LeftHandle:        design.ArchLeftHandle,
	PeakHandleL:       0.55,
	PeakHandleR:       0.58,
	ShoulderHandle:    0.40,
	ShoulderLow:       0.25,
	ShoulderHigh:      0.65,
	UnderLeftHandle:   design.ArchUnderLeftHandle,
	UnderPeakHandleL:  0.50,
	UnderShoulderLow:  0.25,
	UnderShoulderHigh: 0.65,
	OuterSteeper:      design.ArchOuterSteeper,
	UnderSteeper:      design.ArchUnderSteeper,
	OuterBend:         12,
	UnderBend:         8,
}

// n draws the lowercase letter n as one counter-clockwise contour: two stems
// with foot serifs, a head serif on the left of the left stem and an arch
// which springs from the left stem at an angle and runs into the right stem.
func n(p design.Params) (outline.Path, float64) {
	left := stemSpec{
		Left: p.SideStemLeft, Width: p.Stem, Top: p.XHeight,
		FootLeft: p.SerifJut, FootRight: p.SerifJutInner,
		HeadJut: p.SerifJut, HeadDrop: p.HeadSerifDrop,
	}
	return archWithStem(p, left)
}

// archWithStem draws a letter which consists of a left stem, the arch of n and a
// right stem of x-height: n if the left stem ends at the x-height, h if it
// continues to the ascender.  The left stem carries the head serif at its top.
func archWithStem(p design.Params, left stemSpec) (outline.Path, float64) {
	// The right stem stands where it does in n.  A left stem of h of width
	// AscenderStem stands in the cell of a stem of width Stem, and the counter
	// grows by its inset.
	inset := (p.Stem - left.Width) / 2
	right := stemSpec{
		Left: left.Left - inset + p.Stem + p.NCounter, Width: p.Stem,
		FootLeft: p.SerifJutInner, FootRight: p.SerifJut,
	}
	arch := archBetween(p, left, right)

	var path outline.Path
	path.MoveTo(left.start())
	stemUp(&path, p, left, arch.Join)
	archUnder(nArchShape, arch).appendTo(&path)
	stemDown(&path, p, right)
	stemUp(&path, p, right, arch.Shoulder)
	archOuter(nArchShape, arch).appendTo(&path)
	stemCap(&path, p, left)
	stemDown(&path, p, left)
	return path, design.NAdvance(p)
}

// archBetween returns the arch of n between two stems: it springs from the
// right-hand edge of the left stem and runs into the left-hand edge of the
// right one, and its outer edge leaves the right stem by its right-hand edge.
// The outer edge peaks at the fraction ArchPeak of the distance between the stem
// centres, which may be any distance: the handles of the curves are fractions of
// it.  The right stem need not be the last one of the letter.  The shoulder
// where the outer edge leaves the right stem is ShoulderStart, which m moves up
// for its first arch.
func archBetween(p design.Params, left, right stemSpec) archSpec {
	c1 := left.Left + left.Width/2
	return archSpec{
		X1: left.right(), X2: right.Left, X3: right.right(),
		XPeak:    c1 + p.ArchPeak*(right.Left+p.Stem/2-c1),
		Peak:     p.XHeight + p.ArchOvershoot,
		Thick:    p.ThinLower,
		Notch:    p.ArchNotch,
		Join:     p.ArchJoin,
		ThinJoin: p.ArchThinJoin,

		Shoulder:      p.ShoulderStart,
		ShoulderInner: p.ShoulderStartInner,
	}
}
