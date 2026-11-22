// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The currency signs dollar, sterling and euro.  All three stand at the figure
// height; the dollar and the sterling on the common advance of the figures
// (Figures.Advance), the euro on its own (design.EuroAdvance).

// dollarParams returns p with the values of the S of the dollar in place of
// those of the capital S, at the figure height, for the curve code of S
// (capss.go).
func dollarParams(p design.Params) design.Params {
	s := p.Sym
	q := p
	q.CapHeight, q.OvershootUpper = p.Figures.Height, p.OvershootLower
	g := &q.Caps.CGSJQ
	g.SLowerWidth, g.SUpperWidth, g.SUpperShift = s.DollarLowerWidth, s.DollarUpperWidth, s.DollarUpperShift
	g.SSpineThick, g.SSpineAngle = s.DollarSpine, s.DollarSpineAngle
	g.SArcThick, g.SUpperArcThick, g.SSideThick = s.DollarArc, s.DollarUpperArc, s.DollarSideThick
	g.SUpperDrop, g.SLowerDrop = s.DollarUpperDrop, s.DollarLowerDrop
	g.SSideLeft = math.Round((p.Figures.Advance - s.DollarLowerWidth) / 2)
	g.SSideRight = p.Figures.Advance - s.DollarLowerWidth - g.SSideLeft
	return q
}

// dollar draws the dollar: the capital S drawn at the figure height with the
// values of the dollar (narrower and lighter than S, with its beaks), and a
// vertical stroke DollarStroke thick through the middle of its ink, from
// DollarBottom to DollarTop, which runs through the S and shows in both
// counters.  The outline is the union of the two (symunion.go).
//
// Side kinds: open; the S is centred on the advance of the figures.
func dollar(p design.Params) (outline.Path, float64) {
	s := p.Sym
	q := dollarParams(p)
	f := newCapSFrame(q)
	var path outline.Path
	path.MoveTo(f.C.X+f.LowerE.Inner[0].P0.X, f.C.Y+f.LowerE.Inner[0].P0.Y)
	f.circuit(&path, f.LowerE, 1)
	f.circuit(&path, f.UpperE, -1)
	cx := math.Round(q.Caps.CGSJQ.SSideLeft + s.DollarLowerWidth/2)
	bar := rectChain(cx-s.DollarStroke/2, s.DollarBottom, cx+s.DollarStroke/2, s.DollarTop)
	return chainsPath(union(append(pathChains(path), bar)...)...), p.Figures.Advance
}

// sterling draws the pound sterling: one stroke and a bar.  The stroke starts
// in the ball terminal of the lowercase at the upper right (ball.go), not
// mirrored, as on r, runs over the top in a hook which is thin at the top
// (Thin), comes down as a straight stem SterStem thick, leaning a little to the
// right, turns at its foot to the left into a loop at the lower left which goes
// round a counter clockwise, crosses the stem above the foot and runs on to
// the right as the base.  The base dips to the baseline, where it is SterBase
// thick, and rises to a thin end cut across the stroke, below the right of the
// hook.  The bar is SterBar thick at SterBarY, SterBarLen long from the left
// of the frame of the stem.
//
// The two edges of the stroke are drawn node by node (sterlingEdges, edgeOf).
// The outline is the union of the stroke, which crosses itself where the loop
// meets the stem, and the bar (symunion.go).
//
// Side kinds: open; the ink is centred on the advance of the figures.
func sterling(p design.Params) (outline.Path, float64) {
	s := p.Sym
	w := s.SterInk
	x0 := math.Round((p.Figures.Advance - w) / 2)
	top := p.Figures.Height + p.OvershootLower
	right, left := sterlingEdges(p, x0, top)
	re, le := edgeOf(right), edgeOf(left).reversed()
	r := p.BallDiameter / 2
	x1 := x0 + w
	ball := ballStart("sterling", p, re, le, x1-0.10*w-r, span(0.70*top, 0.82*top, 0.01*top), span(-40, 60, 2), false)
	stroke := join(re, lineChain(re.end(), le.start()), le, ball)
	bx := x0 + 0.04*w
	bar := rectChain(bx, s.SterBarY-s.SterBar/2, bx+s.SterBarLen, s.SterBarY+s.SterBar/2)
	cs := union(stroke, bar)
	lo, hi := math.Inf(1), math.Inf(-1)
	for _, c := range cs {
		for _, q := range c {
			lo, hi = math.Min(lo, math.Min(q.P0.X, q.P3.X)), math.Max(hi, math.Max(q.P0.X, q.P3.X))
		}
	}
	dx := math.Round((p.Figures.Advance - lo - hi) / 2)
	for i := range cs {
		cs[i] = cs[i].shifted(dx, 0)
	}
	return chainsPath(cs...), p.Figures.Advance
}

// sterlingEdges returns the nodes of the right-hand and the left-hand edge of
// the stroke of the sterling, both from the top of the hook to the end of the
// base, in a frame SterInk wide from x0.
//
// The stem is the band SterStem thick about the line through the points at 0.33
// and 0.30 of the frame at 0.68 of the height and at the bar; its edges are
// straight from the hook down to the bar.  The hook runs from the top, where its
// outer edge has the radius 110 and its inner edge the radius 70, to the stem.
// Below the bar the right-hand edge bends into the counter of the loop, an oval
// 68 wide and 80 high whose extremes have the radii 29 (bottom and top) and 47
// (left), turns at the top of the counter into the base and has its
// inflection inside the stem; the left-hand edge bends into the outer edge of
// the loop, 72 and 75 from the centre of the counter with the radii 69 and 78,
// and has its inflection inside the stem too.  The lower edge of the base
// runs on a circle of the radius 185 which touches the baseline less the
// overshoot, the upper edge on a circle of the radius 260 whose lowest point is
// SterBase above it, and both bend more tightly towards the end of the base,
// where the stroke is cut.
func sterlingEdges(p design.Params, x0, top float64) (right, left []edgeNode) {
	s := p.Sym
	w := s.SterInk
	c1 := outline.Point{X: x0 + 0.33*w, Y: 0.68 * top}
	c2 := outline.Point{X: x0 + 0.30*w, Y: s.SterBarY}
	sd := unit(c1, c2)
	sdeg := math.Atan2(sd.Y, sd.X) * 180 / math.Pi
	nx := s.SterStem / 2 / math.Abs(sd.Y)
	stem := func(y, side float64) edgeNode {
		return edgeNode{At: outline.Point{X: c1.X + (y-c1.Y)*sd.X/sd.Y + side*nx, Y: y}, Dir: sdeg, Fix: true}
	}
	hx := x0 + 0.60*w
	b := -p.OvershootLower
	t := b + s.SterBase
	// at returns a node at x0+x, y with the direction dir and the radius of
	// curvature r (negative where the edge turns clockwise; zero leaves the
	// curvature to edgeOf).
	at := func(x, y, dir, r float64) edgeNode {
		n := edgeNode{At: outline.Point{X: x0 + x, Y: y}, Dir: dir}
		if r != 0 {
			n.K = 1 / r
		}
		return n
	}
	// auto returns a node at x0+x, y whose direction and curvature edgeOf
	// takes from its neighbours.
	auto := func(x, y float64) edgeNode {
		return edgeNode{At: outline.Point{X: x0 + x, Y: y}, Auto: true}
	}
	// flat returns an inflection at x0+x, y with the direction dir.
	flat := func(x, y, dir float64) edgeNode {
		return edgeNode{At: outline.Point{X: x0 + x, Y: y}, Dir: dir, Fix: true}
	}
	lineStart := func(n edgeNode) edgeNode { n.Line = true; return n }
	right = []edgeNode{
		{At: outline.Point{X: hx, Y: top}, Dir: 180, K: 1.0 / 110},
		{At: outline.Point{X: hx - 89, Y: top - 41}, Dir: -135},
		lineStart(stem(420, -1)),
		stem(s.SterBarY, -1),
		auto(77.2, b+172),
		at(66.4, b+57, -116, -42),
		at(37, b+37, 180, -29),
		at(3, b+77, 90, -47),
		at(37, b+117, 0, -29),
		flat(80, b+82, -60),
		auto(105, b+52),
		at(147.5, b+24.8, -30, 185),
		at(240, b, 0, 185),
		at(345, b+32.7, 34.6, 0),
		at(407, b+112, 68, 160),
	}
	left = []edgeNode{
		{At: outline.Point{X: hx, Y: top - p.Figures.Thin}, Dir: 180, K: 1.0 / 70},
		{At: outline.Point{X: hx - 52.7, Y: top - p.Figures.Thin - 21.8}, Dir: -135},
		lineStart(stem(440, 1)),
		stem(s.SterBarY, 1),
		auto(160, b+162),
		at(88, b+22, -133.8, 0),
		at(37, b, 180, -69),
		at(-35, b+75, 90, -78),
		at(37, b+150, 0, -69),
		at(55.6, b+147.4, -15.6, -69.7),
		flat(100, b+122, -38),
		at(125, t+32, -30, 280),
		at(235, t, 0, 260),
		at(345, t+24.4, 25, 0),
		at(380, t+50, 47, 70),
	}
	return right, left
}

// An edgeNode is a node of an edge of the sterling: its point and the
// direction of the edge there in degrees.  A node with Auto set takes the
// direction of the circle through the node and its two neighbours, or, if Fix
// is set too, the direction from the node before it to the node after it.  Fix
// marks a node where the edge is straight (an inflection, or the end of a
// straight piece), Line a straight piece to the next node.
type edgeNode struct {
	At              outline.Point
	Dir, K          float64
	Auto, Fix, Line bool
}

// edgeOf returns the edge through the nodes, drawn as the edge of a stroke of
// no thickness whose stations are the nodes (strokeEdge): between two nodes one
// cubic curve with the direction and the curvature of the edge at its ends.
func edgeOf(ns []edgeNode) chain {
	sts := make([]station, len(ns))
	for i, n := range ns {
		d := n.Dir
		if n.Auto {
			a, b, c := ns[i-1].At, n.At, ns[i+1].At
			d = circleDir(a, b, c, n.Fix)
		}
		sts[i] = station{X: n.At.X, Y: n.At.Y, Dir: d, K: n.K, Flat: n.Line}
		if n.Fix && !n.Line && !(i > 0 && ns[i-1].Line) {
			sts[i].K = 1e-12
		}
	}
	return strokeEdge(sts, 1)
}

// circleDir returns the direction in degrees, from a towards c, of the circle
// through a, b and c at b, or of the line from a to c if chord is set or the
// points lie on a line.
func circleDir(a, b, c outline.Point, chord bool) float64 {
	ax, ay := a.X-b.X, a.Y-b.Y
	cx, cy := c.X-b.X, c.Y-b.Y
	den := 2 * (ax*cy - ay*cx)
	if chord || math.Abs(den) < 1e-9 {
		return math.Atan2(c.Y-a.Y, c.X-a.X) * 180 / math.Pi
	}
	// the centre of the circle, relative to b
	ox := (cy*(ax*ax+ay*ay) - ay*(cx*cx+cy*cy)) / den
	oy := (ax*(cx*cx+cy*cy) - cx*(ax*ax+ay*ay)) / den
	t := outline.Point{X: -oy, Y: ox}
	if t.X*(c.X-a.X)+t.Y*(c.Y-a.Y) < 0 {
		t = outline.Point{X: oy, Y: -ox}
	}
	return math.Atan2(t.Y, t.X) * 180 / math.Pi
}

// euroParams returns p with the values of C scaled for the euro
// (design.EuroScale): the round at the figure height, from OvershootUpper
// below the baseline to OverTop above the figure height, sx times the width
// of O with the round side EuroThick and the thin parts of the capitals, its
// left extreme at x; the cut of the beak and the reach of the tail scaled in
// width, the drop and the lift of the beak and the heights of the end of the
// beak and of the tail and the run of the beak (an arc length along the steep
// end of the edge) scaled in height.  The serif and the tail keep the weights
// of C.
func euroParams(p design.Params, x float64) design.Params {
	s := p.Sym
	sx, sy := design.EuroScale(p)
	ov := p.OvershootUpper
	y := func(v float64) float64 { return (v+ov)*sy - ov }
	q := capCParams(p)
	q.CapHeight = p.Figures.Height + p.Figures.OverTop - ov
	c := &q.Caps
	c.SideRoundLeft, c.OWidth, c.Round = x, sx*c.OWidth, s.EuroThick
	g := &c.CGSJQ
	g.BeakX, g.CTailReach = sx*g.BeakX, sx*g.CTailReach
	g.BeakRun *= sy
	g.BeakDrop, g.BeakLift = sy*g.BeakDrop, sy*g.BeakLift
	g.BeakEnd, g.CTailHeight = y(g.BeakEnd), y(g.CTailHeight)
	return q
}

// euro draws the euro: the capital C (capc.go) scaled to the figure height
// and to EuroRound of its proportion (euroParams), with its upper beak and its
// tail; and two horizontal bars EuroBar thick, EuroBarGap apart between their
// centres about the middle of the figure height, which start EuroReach left of
// the round with square ends and end in the counter on one line, EuroBarCut
// degrees from the horizontal, through the point EuroBarEnd right of the left
// of the ink at the middle of the figure height, so that the upper bar reaches
// further right than the lower.  The outline is the union of the round and the
// bars.
//
// Side kinds: open; the ink stands EuroSideLeft and EuroSideRight from the
// edges of its own advance (design.EuroAdvance).
func euro(p design.Params) (outline.Path, float64) {
	s := p.Sym
	x0 := s.EuroSideLeft
	q := euroParams(p, x0+s.EuroReach)
	f := capRound(q)
	round := join(capRoundLeft(f), capCTail(q, f), capRoundLeftInner(f), capUpperBeakEased(q, f))
	mid := p.Figures.Height / 2
	slope := math.Cos(s.EuroBarCut*math.Pi/180) / math.Sin(s.EuroBarCut*math.Pi/180)
	end := func(y float64) outline.Point { return outline.Point{X: x0 + s.EuroBarEnd + (y-mid)*slope, Y: y} }
	bar := func(yc float64) chain {
		y0, y1 := yc-s.EuroBar/2, yc+s.EuroBar/2
		a, b, c, d := outline.Point{X: x0, Y: y0}, end(y0), end(y1), outline.Point{X: x0, Y: y1}
		return chain{lineSeg(a, b), lineSeg(b, c), lineSeg(c, d), lineSeg(d, a)}
	}
	return chainsPath(union(round, bar(mid-s.EuroBarGap/2), bar(mid+s.EuroBarGap/2))...), design.EuroAdvance(p)
}
