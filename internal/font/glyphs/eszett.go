// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The letters ß and ẞ.
//
// Both are one construction with the sizes of the lowercase or of the
// capitals: a stem on the left with a foot serif on the left, which rises into
// a round shoulder and over the top into a smaller upper bowl on the right;
// the upper bowl comes back to the left into a short middle stroke, the
// waist, which ends in a straight cut, and the middle stroke runs on into a
// larger lower bowl, which turns at the bottom to the left and ends in the
// ball terminal near the foot of the stem.  The two bowls are drawn as the
// bowls of 3 (figthree.go): pairs of super-ellipses with the squareness of the
// o, whose outer edges cross in a notch on the right at the waist.  The upper
// bowl has the stem as its left side: its outer and its inner edge reach the
// left-hand and the right-hand edge of the stem at their left extremes, where
// they run on straight down the stem.
//
// The contour runs from the tip of the foot serif along the baseline, up the
// right-hand edge of the stem, round the inner edge of the upper bowl to the
// cut of the middle, down the cut, round the inner edge of the lower bowl to
// the ball, round it, back along the outer edge of the lower bowl to the
// notch, round the outer edge of the upper bowl over the top and down the
// left-hand edge of the stem to the foot serif.

// An eszettFrame holds the sizes of ß or ẞ.
type eszettFrame struct {
	name               string
	stem, width, jut   float64 // the left-hand edge and the width of the stem, the foot serif
	bracket            float64 // the height of the bracket of the foot serif
	serifThick, reach  float64 // the thickness and the bracket reach of the foot serif
	bottom, top, waist float64 // the lowest and the highest ink, the middle of the middle stroke
	right, upperRight  float64 // the right extremes of the lower and of the upper bowl
	thin, middle       float64 // the thin parts of the bowls and the middle stroke
	bowl, upperBowl    float64 // the thick parts of the lower and the upper bowl
	overlap, open      float64 // as ThreeOverlap and ThreeOpen of 3
	lowerLeft          float64 // the left extreme of the lower bowl
	ballLeft           float64 // the leftmost ink of the ball
	ball               ballPart
	advance            float64
	capital            bool // the foot serif is that of the capitals
}

// Shape constants of ß and ẞ.
const (
	eszettInk        = 0.956 // the ink width of ß, in ink widths of n
	eszettUpper      = 0.90  // the right extreme of the upper bowl, in ink widths from the left of the ink
	eszettWaist      = 0.87  // the middle of the middle stroke of ß, in x-heights
	capEszettInk     = 1.22  // the ink width of ẞ, in ink widths of ß
	capEszettWaist   = 0.553 // the middle of the middle stroke of ẞ, in heights of its top
	eszettOverlap    = 10.0  // ThreeOverlap of the bowls of ß
	eszettOpen       = 50.0  // ThreeOpen of the lower bowl of ß
	capEszettOpen    = 60.0  // ThreeOpen of the lower bowl of ẞ
	eszettBallGap    = 24.0  // the white between the stem and the ball of ß
	eszettLowerLeft  = 40.0  // how far the left extreme of the lower bowl lies left of the ball
	eszettCutSlope   = 40.0  // the edge of the lower counter at the cut of the middle stroke, degrees from the horizontal
	eszettCutRun     = 20.0  // the least distance of the cut of the middle stroke from the bottom of the upper counter
	eszettUpperThick = 0.9   // the thick part of the upper bowl, in thick parts of the lower
)

// newEszettFrame returns the frame of ß.
func newEszettFrame(p design.Params) eszettFrame {
	nInk := design.NAdvance(p) - p.SideStemLeft - p.SideStemRight + 2*p.SerifJut
	s := p.SideStemLeft
	x0 := s - p.SerifJut
	w := eszettInk * nInk
	f := eszettFrame{
		name: "germandbls", stem: s, width: p.Stem, jut: p.SerifJut,
		bracket: p.BracketHeight, serifThick: p.SerifThick, reach: p.BracketReach,
		bottom: -p.OvershootLower, top: p.Ascender + p.OvershootLower, waist: eszettWaist * p.XHeight,
		right: math.Round(x0 + w), upperRight: math.Round(x0 + eszettUpper*w),
		thin: p.ThinLower, middle: p.ThinLower, bowl: p.CurveLower, upperBowl: eszettUpperThick * p.CurveLower,
		overlap: eszettOverlap, open: eszettOpen,
		ball: newBallPart(p),
	}
	f.ballLeft = s + p.Stem + eszettBallGap
	f.lowerLeft = f.ballLeft - eszettLowerLeft
	f.advance = math.Round(f.right + p.SideRoundRight)
	return f
}

// newCapEszettFrame returns the frame of ẞ: that of ß with the stem, the
// serif, the heights and the weights of the capitals, and capEszettInk times
// as wide.
func newCapEszettFrame(p design.Params) eszettFrame {
	lc := newEszettFrame(p)
	c := p.Caps
	s := design.CapStemLeft(p)
	x0 := s - c.SerifJut
	w := capEszettInk * (lc.right - (lc.stem - lc.jut))
	scale := w / (lc.right - (lc.stem - lc.jut))
	top := p.CapHeight + p.OvershootUpper
	f := eszettFrame{
		name: "uni1E9E", stem: s, width: c.Stem, jut: c.SerifJut,
		bracket: c.BracketHeight, serifThick: c.SerifThick, reach: c.BracketReach,
		bottom: -p.OvershootUpper, top: top, waist: capEszettWaist * top,
		right: math.Round(x0 + w), upperRight: math.Round(x0 + eszettUpper*w),
		thin: c.RoundThin, middle: c.Bar, bowl: c.Round, upperBowl: eszettUpperThick * c.Round,
		overlap: scale * eszettOverlap, open: capEszettOpen,
		ball: newBallPart(p), capital: true,
	}
	f.ballLeft = s + c.Stem + eszettBallGap
	f.lowerLeft = f.ballLeft - eszettLowerLeft
	f.advance = math.Round(f.right + c.SideRoundRight)
	return f
}

// eszett draws ß.
func eszett(p design.Params) (outline.Path, float64) {
	f := newEszettFrame(p)
	return eszettOutline(p, f), f.advance
}

// capEszett draws ẞ.
func capEszett(p design.Params) (outline.Path, float64) {
	f := newCapEszettFrame(p)
	return eszettOutline(p, f), f.advance
}

// eszettOutline draws ß or ẞ with the frame f, see the comment above.
func eszettOutline(p design.Params, f eszettFrame) outline.Path {
	sr := f.stem + f.width
	w, g, ov := f.waist, f.middle, f.overlap
	lo := newSuperShape(f.lowerLeft, f.bottom, f.right, w+ov, p.SupernessOuter)
	li := newSuperShape(f.lowerLeft+f.open, f.bottom+f.thin, f.right-f.bowl, w-g/2, p.SupernessInner)
	uo := newSuperShape(f.stem, w-ov, f.upperRight, f.top, p.SupernessOuter)
	ui := newSuperShape(sr, w+g/2, f.upperRight-f.upperBowl, f.top-f.thin, p.SupernessInner)
	lower := lowerLeftBall(f.name, f.ball, lo.cubic(2, true), li.cubic(1, false), f.ballLeft,
		figStroke{Rise: true, Before: li.cubic(0, false)})
	// The notch on the right, where the outer edges cross.
	loR, uoR := lo.cubic(0, true), uo.cubic(3, true)
	ta, tb := crossAt(f.name, loR, uoR)
	toNotch, fromNotch := loR.sub(0, ta), uoR.sub(tb, 1)
	fromNotch.P0 = toNotch.P3
	// The cut of the middle.
	uiL, liL := ui.cubic(1, false), li.cubic(2, false) // bottom to left, left to top
	// The middle stroke ends in a cut which leaves the edge of the lower
	// counter where that runs eszettCutSlope degrees from the horizontal.  It
	// is vertical where the edge of the upper counter reaches that abscissa at
	// least eszettCutRun left of its bottom, and otherwise it runs to the edge
	// of the upper counter there.
	xl := xAtSlope(liL, eszettCutSlope)
	xu := math.Min(xl, uiL.P0.X-eszettCutRun)
	toCut, fromCut := uiL.sub(0, uiL.tAtX(xu)), liL.sub(liL.tAtX(xl), 1)
	// The upper bowl runs into the stem without curvature: each of its edges
	// from the top to the stem is one curve which leaves the top with the
	// curvature of the super-ellipse and meets the stem straight, where such
	// a curve exists, and the quarter of the super-ellipse otherwise.
	upperLeft := uo.cubic(1, true)  // top to the left extreme, on the left-hand edge of the stem
	innerLeft := ui.cubic(2, false) // the left extreme, on the right-hand edge of the stem, to the top
	if c, err := g2Cubic(upperLeft.P0, upperLeft.tangent(0), upperLeft.curvature(0), upperLeft.P3, upperLeft.tangent(1), 0); err == nil {
		upperLeft = c
	}
	if c, err := g2Cubic(innerLeft.P0, innerLeft.tangent(0), 0, innerLeft.P3, innerLeft.tangent(1), innerLeft.curvature(1)); err == nil {
		innerLeft = c
	}
	pt := func(x, y float64) outline.Point { return outline.Point{X: x, Y: y} }
	th, bh, br := f.serifThick, f.bracket, f.reach
	foot := chain{
		curveSeg(cubic{pt(f.stem, bh), pt(f.stem, th+(bh-th)/2), pt(f.stem-br/2, th), pt(f.stem-br, th)}),
	}
	foot = join(foot, lineChain(pt(f.stem-br, th), pt(f.stem-f.jut, th)), lineChain(pt(f.stem-f.jut, th), pt(f.stem-f.jut, 0)))
	if f.capital {
		foot = capFootLeft(p, f.stem, f.jut)
	}
	c := join(
		lineChain(foot.end(), pt(sr, 0)),
		lineChain(pt(sr, 0), innerLeft.P0),
		chain{curveSeg(innerLeft), curveSeg(ui.cubic(3, false)), curveSeg(ui.cubic(0, false))},
		curveChain(toCut),
		lineChain(toCut.P3, fromCut.P0),
		curveChain(fromCut),
		chain{curveSeg(li.cubic(3, false))},
		lower,
		chain{curveSeg(lo.cubic(3, true))},
		curveChain(toNotch),
		curveChain(fromNotch),
		chain{curveSeg(uo.cubic(0, true)), curveSeg(upperLeft)},
		lineChain(upperLeft.P3, foot.start()),
		foot,
	)
	return contourPath(c)
}

// xAtSlope returns the abscissa at which the quarter c of a super-ellipse,
// which runs between a horizontal and a vertical tangent, runs at the angle
// (degrees from the horizontal).
func xAtSlope(c cubic, angle float64) float64 {
	steep := func(t float64) bool {
		d := c.tangent(t)
		return math.Abs(d.Y) > math.Tan(angle*math.Pi/180)*math.Abs(d.X)
	}
	t0, t1 := 0.0, 1.0
	rising := steep(1)
	for range 50 {
		if mid := (t0 + t1) / 2; steep(mid) == rising {
			t1 = mid
		} else {
			t0 = mid
		}
	}
	return c.at((t0 + t1) / 2).X
}
