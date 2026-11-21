// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// u draws the lowercase letter u as one counter-clockwise contour: the arch of
// n turned through half a turn between two stems of x-height.  The bowl at the
// bottom is the underside and the outer edge of the arch of n, each turned
// round the centre of the letter, that is flipped in the horizontal line at half
// the x-height and mirrored in the vertical line through the middle of the
// stems (flipY and mirrorX).  The two reflections together are a rotation, which
// keeps a counter-clockwise contour counter-clockwise.
//
// The arch of n leaves its left stem thin (ArchThinJoin) and runs thick into
// its right stem; turned round it leaves the right stem of u thin and runs
// thick into the left stem: the stroke comes down the left stem, thickens round
// the bottom and joins the right stem thin.  The heights of the arch of n are
// replaced by heights measured from the baseline: the inner edge meets the right
// stem at UJoin and the outer edge at UNotch, and the edges leave the left stem
// at UShoulderInner and UShoulder.  The lowest point lies OvershootLower below
// the baseline and the arch is ThinLower thick there.
//
// Both stems are x-height high and have the head serif of n on the left.  The
// left stem has no foot; the right stem has a foot serif on the right only, as
// the right stem of d.  The contour runs
//
//	from the baseline at the left edge of the right stem, along the foot, up
//	its right-hand edge and over its top with the head serif (stemUp, stemCap)
//	down the left-hand edge of the right stem to UJoin
//	the inner edge of the bowl to the left stem (archUnder, turned)
//	up the right-hand edge of the left stem, over its top with the head serif
//	(stemCap), and down its left-hand edge to UShoulder
//	the outer edge of the bowl to the right stem at UNotch (archOuter, turned)
//	down the left-hand edge of the right stem to the baseline.
func u(p design.Params) (outline.Path, float64) {
	left := stemSpec{
		Left: p.SideStemLeft, Width: p.Stem, Top: p.XHeight,
		HeadJut: p.SerifJut, HeadDrop: p.HeadSerifDrop,
	}
	right := stemSpec{
		Left: left.right() + p.UCounter, Width: p.Stem, Top: p.XHeight,
		FootRight: p.SerifJut,
		HeadJut:   p.SerifJut, HeadDrop: p.HeadSerifDrop,
	}

	arch, cx := uArch(p, left, right)
	turn := func(c chain) chain { return c.flipY(p.XHeight / 2).mirrorX(cx) }

	var path outline.Path
	path.MoveTo(right.start())
	stemUp(&path, p, right, right.Top)
	stemCap(&path, p, right)
	path.LineTo(right.Left, p.UJoin)
	turn(archUnder(nArchShape, arch)).appendTo(&path)
	stemCap(&path, p, left)
	path.LineTo(left.Left, p.UShoulder)
	turn(archOuter(nArchShape, arch)).appendTo(&path)
	path.LineTo(right.start())
	return path, design.UAdvance(p)
}

// uArch returns the arch of n in the frame which the rotation of u maps onto
// the letter, and the abscissa cx of the vertical line of the rotation: the
// left stem of the arch is the image of the right stem of u, and its heights
// are measured from the top.
func uArch(p design.Params, left, right stemSpec) (archSpec, float64) {
	cx := (left.Left + right.right()) / 2
	at := func(x float64) float64 { return 2*cx - x }
	n1 := stemSpec{Left: at(right.right()), Width: p.Stem} // left stem of the turned arch
	n2 := stemSpec{Left: at(left.right()), Width: p.Stem}  // its right stem
	c1 := n1.Left + n1.Width/2
	return archSpec{
		X1: n1.right(), X2: n2.Left, X3: n2.right(),
		XPeak:    c1 + p.ArchPeak*(n2.Left+p.Stem/2-c1),
		Peak:     p.XHeight + p.OvershootLower,
		Thick:    p.ThinLower,
		Notch:    p.XHeight - p.UNotch,
		Join:     p.XHeight - p.UJoin,
		ThinJoin: p.ArchThinJoin,

		Shoulder:      p.XHeight - p.UShoulder,
		ShoulderInner: p.XHeight - p.UShoulderInner,
	}, cx
}
