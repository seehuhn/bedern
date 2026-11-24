// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"math"

	"seehuhn.de/go/bedern/internal/stroke"
)

// z draws the cursive z in three strokes of the pen which abut at two sharp
// corners.  The top stroke starts at the left in the cut of an entry, rises,
// turns over its crest and runs down a little to the corner at the right,
// whose outer point lies on the x-height.  The diagonal, which the pen makes
// heavy, runs from that corner to the corner at the left of the baseline.
// The bottom stroke leaves that corner running down a little to the right,
// thin by the pen, passes its lowest point below the baseline and rises
// into the end of the exit.  At each corner the bar ends in a cut along the
// outer edge of the diagonal and the diagonal in a cut along the inner edge
// of the bar, so that the strokes abut there.  The corners and the
// directions of the bars there are given in the italic and taken back to
// the upright description.
func z(p Params) letter {
	h := p.ZBarW * p.Hair
	top := p.XHeight
	up := func(q Point) Point { return p.unshear(q) }
	// upDir returns the upright direction of the italic direction a.
	upDir := func(a float64) float64 {
		r := a * math.Pi / 180
		v := up(pt(math.Cos(r), math.Sin(r)+p.ShearHeight))
		o := up(pt(0, p.ShearHeight))
		return degrees(math.Atan2(v.Y-o.Y, v.X-o.X))
	}
	// The direction of the diagonal and its thickness by the pen depend on
	// each other; in the italic its upper edge passes through the corner
	// (ZTopRight, top) and its lower edge through (ZBotLeft, 0).
	dx := p.ZTopRight - p.ZBotLeft
	w := p.across(p.Stem)
	var th float64
	for it := 0; it < 6; it++ {
		if it > 0 {
			w = p.ZDiagW * p.nibWidth(upDir(180+degrees(th)))
		}
		f := func(th float64) float64 { return dx - top/math.Tan(th) - w/math.Sin(th) }
		lo, hi := 0.1, math.Pi/2
		for k := 0; k < 60; k++ {
			m := (lo + hi) / 2
			if f(m) > 0 {
				hi = m
			} else {
				lo = m
			}
		}
		th = (lo + hi) / 2
	}
	// The edges and the centre line of the diagonal, in the italic, as a
	// point and the direction (cos th, sin th) pointing up to the right.
	u := pt(math.Cos(th), math.Sin(th))
	nIn := pt(-u.Y, u.X) // the normal from the upper edge into the stroke
	upper := pt(p.ZTopRight, top)
	mid := pt(upper.X+nIn.X*w/2, upper.Y+nIn.Y*w/2)
	// meet returns the intersection of the line through a in the direction
	// da with the line through b in the direction db.
	meet := func(a, da, b, db Point) Point {
		den := da.X*db.Y - da.Y*db.X
		t := ((b.X-a.X)*db.Y - (b.Y-a.Y)*db.X) / den
		return pt(a.X+t*da.X, a.Y+t*da.Y)
	}
	unit := func(a float64) Point { return pt(math.Cos(a*math.Pi/180), math.Sin(a*math.Pi/180)) }
	dirDiag := upDir(180 + degrees(th))
	thd := degrees(th)
	// cutAt returns the end of a stroke whose cut runs through the corner c
	// in the italic direction a: the node on the centre line through m in
	// the direction d, and the cut, given upright.
	cutAt := func(m, d, c Point, a float64) (Point, stroke.End) {
		return meet(m, d, c, unit(a)), stroke.End{Kind: stroke.Cut, Line: true, Angle: upDir(a)}
	}

	// The top corner, where the outer edges of the top bar and of the
	// diagonal meet.  Each stroke ends in a cut through the corner which
	// runs inside the other, ZCornerCut off the other's outer edge, so that
	// the strokes cross there and share only the corner.
	tb := unit(p.ZTopDir)
	tn := pt(tb.Y, -tb.X) // the normal from the outer edge of the bar into it
	tEnd, tCut := cutAt(pt(upper.X+tn.X*h/2, upper.Y+tn.Y*h/2), tb, upper, 180+thd-p.ZCornerCut)
	dTop, dTopCut := cutAt(mid, u, upper, 180+p.ZTopDir+p.ZCornerCut)

	// The bottom corner (ZBotLeft, 0), likewise.
	corner := pt(p.ZBotLeft, 0)
	bb := unit(p.ZBotDir)
	bn := pt(-bb.Y, bb.X) // the normal from the outer edge of the bar into it
	bStart, bCut := cutAt(pt(corner.X+bn.X*h/2, corner.Y+bn.Y*h/2), bb, corner, thd-p.ZCornerCut)
	dBot, dBotCut := cutAt(mid, u, corner, p.ZBotDir+p.ZCornerCut)

	crest := pt(p.ZTopLeft+p.ZCrestIn, top+p.ZCrestOver-h/2)
	topStroke := stroke.Stroke{
		Nib: p.nib(),
		Nodes: []stroke.Node{
			{P: up(pt(p.ZTopLeft, p.ZStartY)), Dir: p.ZStartDir, K: auto, W: p.EndW, DW: auto},
			{P: up(crest), Dir: 0, K: auto, W: h, DW: 0},
			{P: up(tEnd), Dir: upDir(p.ZTopDir), K: 0, W: h, DW: 0},
		},
		Start: round, End: tCut,
	}
	diag := stroke.Stroke{
		Nodes: []stroke.Node{
			{P: up(dBot), Dir: dirDiag + 180, W: w},
			{P: up(dTop), Dir: dirDiag + 180, W: w},
		},
		Start: dBotCut, End: dTopCut,
	}
	low := up(pt(p.ZLowX, -p.ZDip+p.ZLowW*p.Hair/2))
	botStroke := stroke.Stroke{
		Nib: p.nib(),
		Nodes: []stroke.Node{
			{P: up(bStart), Dir: upDir(p.ZBotDir), K: 0, W: h, DW: 0},
			{P: up(pt(bStart.X+p.ZBotRun*bb.X, bStart.Y+p.ZBotRun*bb.Y)), Dir: upDir(p.ZBotDir), K: 0, W: h, DW: 0},
			{P: low, Dir: 0, K: auto, W: p.ZLowW * p.Hair, DW: 0},
			{P: pt(low.X+p.ZExitRun, p.ExitTipHeight-p.ExitEndW/2), Dir: p.ExitAngle, K: p.ExitEndK, W: p.ExitEndW, DW: 0},
		},
		Start: bCut, End: round,
	}
	return letter{strokes: []stroke.Stroke{topStroke, diag, botStroke}, advance: p.ZWidth}
}
