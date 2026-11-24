// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"math"

	"seehuhn.de/go/bedern/internal/stroke"
)

// x draws the x: two arcs back to back which meet at ShearHeight, where each
// runs along the slant of the italic (XMeetDir from the vertical of the
// italic, which is the upright vertical sheared), and overlap there by
// XOverlap.  The left arc
// is a reversed c: it starts in the entry, has the stem weight on its way down
// to the right, at the node XThickAt along the quarter of an ellipse from
// its top to the meeting, is XMeetW thick at the meeting and tapers to the hairline round its
// bottom, which ends at the lower left in a cut, as the ends of s.  The
// right arc is the left one turned through half a turn about the middle of
// the meeting, travelled the other way, so that the stroke from the upper
// left to the lower right is thick and the other thin; its lower end is the
// turn and the end of the exit instead of the entry.
func x(p Params) letter {
	mid := pt(p.XWidth/2, p.ShearHeight)
	xl := mid.X - p.XMeetW/2 + p.XOverlap/2
	top := pt(xl-p.XTopIn, p.EntryTop-p.EntryTopW/2)
	yb := -p.Overshoot + p.Hair/2
	bot := pt(xl-p.XBotIn, yb)
	cot := 1 / math.Tan((180+p.XEndDir)/2*math.Pi/180)
	end := pt(bot.X+(p.XEndY-yb)*cot, p.XEndY)
	left := []stroke.Node{
		{P: pt(top.X-p.EntryReach-p.EntryTopShift, p.EntryLow), Dir: p.EntryStartDir, K: auto, W: p.EndW, DW: auto},
		{P: top, Dir: 0, K: auto, W: p.EntryTopW, DW: auto},
		p.xThick(top, pt(xl, mid.Y)),
		{P: pt(xl, mid.Y), Dir: p.vertical() + p.XMeetDir, K: auto, W: p.XMeetW, DW: 0},
		{P: bot, Dir: 180, K: auto, W: p.Hair, DW: auto},
		{P: end, Dir: p.XEndDir, K: auto, W: p.SEndW * p.Thin, DW: 0},
	}
	var right []stroke.Node
	for k := len(left) - 1; k >= 0; k-- {
		nd := left[k]
		nd.P = pt(2*mid.X-nd.P.X, 2*mid.Y-nd.P.Y)
		// Turned through half a turn and travelled the other way, a stated
		// curvature changes its sign.
		if !math.IsNaN(nd.K) {
			nd.K = -nd.K
		}
		right = append(right, nd)
	}
	// The lower end of the right arc is the exit: its turn, a hairline
	// overshooting the baseline, and its end.
	n := len(right)
	right[n-2].P.Y, right[n-2].W, right[n-2].DW = yb, p.Hair, 0
	right[n-1] = stroke.Node{
		P:   pt(right[n-2].P.X+p.XExitRun, p.XExitTip-p.ExitEndW/2),
		Dir: p.ExitAngle, K: p.ExitEndK, W: p.ExitEndW, DW: 0,
	}
	return letter{
		strokes: []stroke.Stroke{
			{Nodes: left, Start: round, End: round},
			{Nodes: right, Start: round, End: round},
		},
		advance: p.XWidth,
	}
}

// xThick returns the node of the left arc of x where it has the stem
// weight: on the quarter of the ellipse from the top node, where the arc
// runs horizontally, to the node at the meeting, at the angle XThickAt from
// the top, travelling along the ellipse.
func (p Params) xThick(top, meet Point) stroke.Node {
	a, b := meet.X-top.X, top.Y-meet.Y
	f := p.XThickAt * math.Pi / 180
	return stroke.Node{
		P:   pt(top.X+a*math.Sin(f), meet.Y+b*math.Cos(f)),
		Dir: degrees(math.Atan2(-b*math.Sin(f), a*math.Cos(f))),
		K:   auto, W: p.across(p.Stem), DW: auto,
	}
}
