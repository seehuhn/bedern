// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package stroke

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
)

// A Part is a part of a roman letter described as strokes, for the comparison
// with the outline which the font builds.
type Part struct {
	Name    string // the name of the part
	Glyph   string // the glyph it is taken from
	Strokes []Stroke

	// Tolerance is the largest distance from the outline of the font which
	// the tests allow: reproTolerance, or more for a part which differs from
	// the font by design.
	Tolerance float64
}

// Build returns the outlines of the strokes of the part.
func (pt Part) Build() ([]*Result, error) {
	var rs []*Result
	for _, s := range pt.Strokes {
		r, err := s.Build()
		if err != nil {
			return nil, err
		}
		rs = append(rs, r)
	}
	return rs, nil
}

// RomanParts returns the parts of the roman letters which the package
// reproduces.
func RomanParts(p design.Params) []Part {
	return []Part{
		romanO(p),
		romanC(p),
		romanN(p),
		romanS(p),
		romanJ(p),
		romanNine(p),
		romanD(p),
	}
}

// romanO is the bowl of o: one closed stroke with nodes at the four extremes,
// CurveLower thick at the sides and ThinLower at the top and bottom, whose
// edges have at the extremes the curvatures of the super-ellipses of the
// squareness of the outer contour and of the counter.
func romanO(p design.Params) Part {
	l, r := design.OBowlLeft(p), design.OBowlRight(p)
	bot, top := -p.OvershootLower, p.XHeight+p.OvershootLower
	cx, cy := (l+r)/2, (bot+top)/2
	ao, bo := (r-l)/2, (top-bot)/2
	ai, bi := ao-p.CurveLower, bo-p.ThinLower
	oTop, oSide := SuperK(ao, bo, p.SupernessOuter)
	iTop, iSide := SuperK(ai, bi, p.SupernessInner)
	side := &EdgeK{Right: oSide, Left: iSide}
	cap := &EdgeK{Right: oTop, Left: iTop}
	xs, ys := ao-p.CurveLower/2, bo-p.ThinLower/2
	return Part{Name: "bowl of o", Glyph: "o", Strokes: []Stroke{{
		Closed: true,
		Nodes: []Node{
			{P: Point{X: cx + xs, Y: cy}, Dir: 90, W: p.CurveLower, Edge: side},
			{P: Point{X: cx, Y: cy + ys}, Dir: 180, W: p.ThinLower, Edge: cap},
			{P: Point{X: cx - xs, Y: cy}, Dir: 270, W: p.CurveLower, Edge: side},
			{P: Point{X: cx, Y: cy - ys}, Dir: 0, W: p.ThinLower, Edge: cap},
		},
	}}}
}

// romanN is the arch of n with its branch from the left stem and its run into
// the right stem.  It leaves the left stem between the notch and the join at
// the angle of the arch plus Steeper of its shape, as thick at right angles as the gap
// between notch and join allows at that angle, is ThinLower thick at the top,
// ArchOvershoot above the x-height at the fraction ArchPeak of the way between
// the centres of the stems, and runs into the right stem, Stem thick, straight
// and without curvature from ShoulderStart down.
func romanN(p design.Params) Part { return romanNWith(p, nArch) }

// An archShape holds the shape constants of the arch of n: the curvatures at
// the stem and at the top, the rate at which the arch thickens there, the angle
// by which it leaves the stem more steeply than its thickness asks, and how far
// left of the peak of ArchPeak its top node lies.
type archShape struct {
	Steeper, DWStem float64
	StemK           EdgeK
	TopShift, DWTop float64
	TopK            EdgeK
}

// nArch is the shape of the arch of n.  The values are fitted to the
// outline of the font (TestFit -fit n); the rules draw the curves.
var nArch = archShape{
	Steeper: 8.73, DWStem: 0.170, StemK: EdgeK{Right: 0.00168, Left: -0.00177},
	TopShift: 10.17, DWTop: 0.219, TopK: EdgeK{Right: -0.0156, Left: -0.00577},
}

func romanNWith(p design.Params, a archShape) Part {
	x1 := p.SideStemLeft + p.Stem // the right-hand edge of the left stem
	x2 := x1 + p.NCounter         // the left-hand edge of the right stem
	c1, c2 := x1-p.Stem/2, x2+p.Stem/2
	gap := p.ArchNotch - p.ArchJoin
	leave := math.Acos(math.Min(1, p.ArchThinJoin/gap))*180/math.Pi + a.Steeper
	peak := c1 + p.ArchPeak*(c2-c1) - a.TopShift
	top := p.XHeight + p.ArchOvershoot - p.ThinLower/2
	// The stroke starts inside the left stem, whose edge clips it.
	start := sub(Point{X: x1, Y: (p.ArchNotch + p.ArchJoin) / 2}, mul(dirOf(leave), 10))
	return Part{Name: "arch of n", Glyph: "n", Strokes: []Stroke{{
		Nodes: []Node{
			{P: start, Dir: leave, W: gap * math.Cos(leave*math.Pi/180), DW: a.DWStem, Edge: &a.StemK},
			{P: Point{X: peak, Y: top}, Dir: 0, W: p.ThinLower, DW: a.DWTop, Edge: &a.TopK},
			// The pen is askew so that the outer edge runs into the stem at
			// ShoulderStart and the counter at ShoulderStartInner.
			{P: Point{X: c2, Y: (p.ShoulderStart + p.ShoulderStartInner) / 2}, Dir: 270, W: p.Stem, Skew: (p.ShoulderStartInner - p.ShoulderStart) / 2},
			{P: Point{X: c2, Y: p.ShoulderStart / 2}, Dir: 270, W: p.Stem},
		},
		Start: End{Kind: Clip, Q: Point{X: x1}, Angle: 90},
		End:   End{Sever: true},
	}}}
}

// A spineShape holds the shape constants of the spine of s: the curvatures of
// the right (inner) and the left (outer) edge at the side extreme of the lower
// half, and the rate at which the stroke thickens there.
type spineShape struct {
	Side EdgeK
	DW   float64
}

// sSpine is the shape of the spine of s, fitted once to the outline of the
// font (TestFit -fit s).
var sSpine = spineShape{Side: EdgeK{Right: -0.00867, Left: -0.00703}, DW: -0.309}

func romanS(p design.Params) Part { return romanSWith(p, sSpine) }

// romanSWith is the spine of s, from the left side of the upper half to the
// right side of the lower half.  Its centre C, where it is SSpineThick thick
// and runs SSpineAngle below the horizontal without curvature, lies 1/(1+r) of
// the height of the letter above its bottom, with r = SUpperWidth /
// SLowerWidth, and the upper half is the lower half turned through half a turn
// about C and scaled by r, except the thicknesses, which lie inside the image
// of the outer edge.  The lower half has its
// right side SSideThick thick at the right of the ink, SLowerWidth from the
// left side space, at the middle between the heights of the outer and inner
// extremes of the roman (sShape.SideY and InnerSideY, 0.45 and 0.56 of the
// depth below C).
func romanSWith(p design.Params, a spineShape) Part {
	r := p.SUpperWidth / p.SLowerWidth
	bot, top := -p.OvershootLower, p.XHeight+p.OvershootLower
	cy := bot + (top-bot)/(1+r)
	cx := p.SSideLeft + (p.SLowerWidth/2-p.SUpperShift+r*p.SLowerWidth/2)/(1+r)
	c := Point{X: cx, Y: cy}
	depth := cy - bot
	n2 := Point{X: p.SSideLeft + p.SLowerWidth - p.SSideThick/2, Y: cy - (0.45+0.56)/2*depth}
	// The outer edge of the upper side is the image of that of the lower.
	half := Point{X: p.SSideThick / 2}
	n0 := add(sub(c, mul(sub(add(n2, half), c), r)), half)
	return Part{Name: "spine of s", Glyph: "s", Strokes: []Stroke{{
		Nodes: []Node{
			{P: n0, Dir: 270, W: p.SSideThick, DW: -a.DW / r, Edge: &EdgeK{Right: -a.Side.Left / r, Left: -a.Side.Right / r}},
			{P: c, Dir: -p.SSpineAngle, W: p.SSpineThick},
			{P: n2, Dir: 270, W: p.SSideThick, DW: a.DW, Edge: &a.Side},
		},
		Start: End{Sever: true},
		End:   End{Sever: true},
	}}}
}

// A hookShape holds the shape constants of the hook of j: the place of the
// lowest node, as a fraction of the way from the stem to the ball, its
// thickness and the curvatures of its edges, where it is horizontal (its
// thickness does not change there, so that the node is the extreme), and
// the height of the centre of the ball above the bottom of the hook.
type hookShape struct {
	BottomX, BottomW float64
	BottomK          EdgeK
	BallUp           float64
}

// jHook is the shape of the hook of j, fitted once to the outline of the font
// (TestFit -fit j).
var jHook = hookShape{
	BottomX: 0.717, BottomW: 38.0, BottomK: EdgeK{Right: -0.0085, Left: -0.0083},
	BallUp: 96.8,
}

// The ball terminal of the lowercase: the tip of the wedge lies jBallTip
// degrees round the circle from the direction of the stroke at the end, and
// the inner edge meets the circle there at ballOpen degrees, with the
// curvature of a circle ballInnerR radii of the ball across.
const (
	ballOpen   = 53.0
	ballInnerR = 1.506 // 2.2 radii less the stalk of 34, over the radius of 49
	jBallTip   = -70.9
	jBallOpen  = 50.0 // the j meets its ball a little more steeply than ballOpen
)

func romanJ(p design.Params) Part { return romanJWith(p, jHook) }

// romanJWith is the hook of j: the stem of j runs straight down to the height
// where its edges begin to curve, JOuterDrop above the lowest point, turns to
// the left through the lowest node, JDepth below the baseline, and ends in the
// ball terminal, whose leftmost ink lies JReach left of the stem.
func romanJWith(p design.Params, a hookShape) Part {
	left, right := p.JStemLeft, p.JStemLeft+p.Stem
	cx := (left + right) / 2
	r := p.BallDiameter / 2
	ball := Point{X: left - p.JReach + r, Y: -p.JDepth + a.BallUp}
	join := -(p.JDepth - p.JOuterDrop)
	bottom := Point{X: right + a.BottomX*(ball.X-right), Y: -p.JDepth + a.BottomW/2}
	return Part{Name: "hook of j", Glyph: "j", Tolerance: 3.5, Strokes: []Stroke{{
		Nodes: []Node{
			{P: Point{X: cx, Y: 0}, Dir: 270, W: p.Stem},
			{P: Point{X: cx, Y: join}, Dir: 270, W: p.Stem},
			{P: bottom, Dir: 180, W: a.BottomW, Edge: &a.BottomK},
		},
		Start: End{Sever: true},
		End: End{Kind: Ball, Ball: BallSpec{
			Centre: ball, R: r, Tip: jBallTip, Open: jBallOpen, TipK: -1 / (ballInnerR * r),
		}},
	}}}
}

// A tailShape holds the shape constants of the tail of 9: the thickness and
// the direction where the part begins, at the height where the tail leaves
// the bowl, with the curvatures of its edges there, the lowest node, where
// the thickness is least, its thickness and edge curvatures, and the
// curvature of the inner edge at the tip of the wedge.
type tailShape struct {
	TopW, TopDir     float64
	TopK             EdgeK
	BottomX, BottomW float64
	BottomK          EdgeK
	TipK             float64
}

// nineTail is the shape of the tail of 9, fitted once to the outline of the
// font (TestFit -fit 9).
var nineTail = tailShape{
	TopW: 86.7, TopDir: 265.4, TopK: EdgeK{Right: -0.00207, Left: 0.00238},
	BottomX: 235.5, BottomW: 46.2, BottomK: EdgeK{Right: -0.0126, Left: -0.00584},
	TipK: -0.00983,
}

// The places of the tail of 9 in the roman: the height at which its inner edge
// leaves the bowl, the right-hand edges there, the centre of the ball and the
// direction of the tip of the wedge from it.
var (
	nineTop      = Point{X: 404.6, Y: 276}
	nineBall     = Point{X: 127, Y: 116}
	nineTipAngle = -83.4
)

func romanNine(p design.Params) Part { return romanNineWith(p, nineTail) }

// romanNineWith is the tail of 9: from the right side of the round, where the
// tail leaves the bowl, down round the bottom, Figures.Thin thick there, and
// up into the ball terminal of the figures, Figures.Ball across.
func romanNineWith(p design.Params, a tailShape) Part {
	bot := -p.OvershootUpper
	return Part{Name: "tail of 9", Glyph: "nine", Tolerance: 4.5, Strokes: []Stroke{{
		Nodes: []Node{
			{P: nineTop, Dir: a.TopDir, W: a.TopW, DW: Auto, Edge: &a.TopK},
			{P: Point{X: a.BottomX, Y: bot + a.BottomW/2}, Dir: 180, W: a.BottomW, Edge: &a.BottomK},
		},
		Start: End{Sever: true},
		End: End{Kind: Ball, Ball: BallSpec{
			Centre: nineBall, R: p.Figures.Ball / 2, Neck: p.Figures.CurlNeck, Open: ballOpen, TipK: a.TipK,
		}},
	}}}
}

// A BowlRule is the description of a bowl which leaves a stem and runs back
// into it: the rule for a, b, d, g, p and q.  The centre line leaves the edge
// of the stem halfway between the notch (where the outer edge meets the stem)
// and the join (where the counter meets it), at the angle whose cosine is the
// thickness at the join over the vertical gap between notch and join, so that
// the stroke is JoinThick thick at right angles there; it starts inside the
// stem and the stem's edge cuts both edges in corners, as the roman's bowls
// meet their stems.  It runs through the extremes at the top (Thin thick), at
// the free side (Thick thick) and at the bottom (Thin thick), where the edges
// have the curvatures of the super-ellipses of the outer contour and the
// counter, and back into the stem in the same way below.  The curvatures of
// the edges at the joins are JoinK (outer, counter).
type BowlRule struct {
	StemX                     float64 // the edge of the stem
	Side                      float64 // -1: the bowl stands left of the stem, +1: right
	Free                      float64 // the outer extreme of the free side
	Top, Bottom               float64 // the outer extremes at top and bottom
	CX                        float64 // the abscissa of the extremes at top and bottom
	ThickY                    float64 // the height of the free extreme
	Thin, Thick               float64
	NotchUpper, JoinUpper     float64
	NotchLower, JoinLower     float64
	JoinThick                 float64
	SupernessOut, SupernessIn float64
	JoinK                     [2]EdgeK // at the upper and the lower join, if not zero
	Depth                     float64  // how far inside the stem the centre line starts
}

// Stroke returns the bowl as one stroke whose ends lie inside the stem: Clip
// ends on the edge of the stem if clip is set (to compare the bowl alone), and
// hidden cuts for the union with the stem otherwise.
func (b BowlRule) Stroke(clip bool) Stroke {
	s := b.Side // -1 for a bowl on the left
	pt := func(x, y float64) Point { return Point{X: x, Y: y} }
	angle := func(notch, join float64) float64 {
		g := math.Abs(notch - join)
		return math.Acos(math.Min(1, b.JoinThick/g)) * 180 / math.Pi
	}
	au, al := angle(b.NotchUpper, b.JoinUpper), angle(b.NotchLower, b.JoinLower)
	// Directions of travel: from the upper join away from the stem and up,
	// round the free side, and from below up into the stem.
	dirUp := 180 - au
	dirIn := al
	if s > 0 {
		dirUp, dirIn = au, 180-al
	}
	start := sub(pt(b.StemX, (b.NotchUpper+b.JoinUpper)/2), mul(dirOf(dirUp), b.Depth))
	end := add(pt(b.StemX, (b.NotchLower+b.JoinLower)/2), mul(dirOf(dirIn), b.Depth))
	ao, bo := math.Abs(b.CX-b.Free), (b.Top-b.Bottom)/2
	ai, bi := ao-b.Thick, bo-b.Thin
	oTop, oSide := SuperK(ao, bo, b.SupernessOut)
	iTop, iSide := SuperK(ai, bi, b.SupernessIn)
	sg := 1.0 // the bowl on the left runs counter-clockwise: the outer edge on the right
	if s > 0 {
		sg = -1
	}
	ek := func(outer, inner float64) *EdgeK {
		if s < 0 {
			return &EdgeK{Right: outer, Left: inner}
		}
		return &EdgeK{Right: -inner, Left: -outer}
	}
	_ = sg
	join := func(k EdgeK) (float64, *EdgeK) {
		if k == (EdgeK{}) {
			return Auto, nil
		}
		return 0, &EdgeK{Right: k.Right, Left: k.Left}
	}
	ku, eu := join(b.JoinK[0])
	kl, el := join(b.JoinK[1])
	d := func(a float64) float64 {
		if s > 0 {
			return 180 - a
		}
		return a
	}
	st := Stroke{
		Nodes: []Node{
			{P: start, Dir: dirUp, K: ku, W: b.JoinThick, DW: Auto, Edge: eu},
			{P: pt(b.CX, b.Top-b.Thin/2), Dir: d(180), W: b.Thin, Edge: ek(oTop, iTop)},
			{P: pt(b.Free-s*b.Thick/2, b.ThickY), Dir: d(270), W: b.Thick, Edge: ek(oSide, iSide)},
			{P: pt(b.CX, b.Bottom+b.Thin/2), Dir: d(360), W: b.Thin, Edge: ek(oTop, iTop)},
			{P: end, Dir: dirIn, K: kl, W: b.JoinThick, DW: Auto, Edge: el},
		},
	}
	if clip {
		st.Start = End{Kind: Clip, Q: pt(b.StemX, 0), Angle: 90}
		st.End = End{Kind: Clip, Q: pt(b.StemX, 0), Angle: 90}
	} else {
		st.Start = End{Kind: Cut, Hidden: true}
		st.End = End{Kind: Cut, Hidden: true}
	}
	return st
}

// romanD is the bowl of d by the rule of BowlRule, with the values of the
// font's bowl of d.
func romanD(p design.Params) Part { return romanDWith(p, dJoinK) }

func romanDWith(p design.Params, jk [2]EdgeK) Part {
	b := p.DBowl
	stem := design.DStemLeft(p) + design.AscenderInset(p)
	left := design.DBowlLeft(p)
	_ = b
	r := BowlRule{
		StemX: stem, Side: -1, Free: left,
		Top: p.XHeight + p.OvershootLower, Bottom: -p.OvershootLower,
		CX: left + p.OBowlWidth/2, ThickY: p.XHeight / 2,
		Thin: p.ThinLower, Thick: p.CurveLower,
		NotchUpper: 413, JoinUpper: 358, NotchLower: 50, JoinLower: 100,
		JoinThick: 32, SupernessOut: p.SupernessOuter, SupernessIn: p.SupernessInner,
		Depth: 10, JoinK: jk,
	}
	return Part{Name: "bowl of d", Glyph: "d", Tolerance: 4, Strokes: []Stroke{r.Stroke(true)}}
}

// dJoinK are the curvatures of the edges of the bowl of d at its joins, upper
// and lower (outer edge, counter), fitted once to the font (TestFit -fit d).
var dJoinK = [2]EdgeK{{Right: -0.00039, Left: 0.00443}, {Right: 0.00518, Left: 0.00854}}

// superQuarter returns the quarter of the super-ellipse of the bowls about
// (cx, cy) with the half axes a and b from the extreme on the vertical axis,
// where it runs horizontally in the direction of the sign sx, to the extreme
// on the horizontal axis; sy is the sign of the height of the first above the
// centre.
func superQuarter(cx, cy, a, b, superness, sx, sy float64) Cubic {
	k := (8*superness - 4) / 3
	p0 := Point{X: cx, Y: cy + sy*b}
	p3 := Point{X: cx + sx*a, Y: cy}
	return Cubic{p0, {X: cx + sx*k*a, Y: p0.Y}, {X: p3.X, Y: cy + sy*k*b}, p3}
}

// tAtY returns the parameter at which the curve, rising or falling, has the
// height y.
func (c Cubic) tAtY(y float64) float64 {
	lo, hi := 0.0, 1.0
	up := c[3].Y > c[0].Y
	for i := 0; i < 80; i++ {
		m := (lo + hi) / 2
		if (c.At(m).Y < y) == up {
			lo = m
		} else {
			hi = m
		}
	}
	return (lo + hi) / 2
}

// cTipInnerK is the curvature of the inner edge of c at the cut of the lower
// terminal: a radius of 167 units, about the least for which the edge also
// has the curvature of the counter at the bottom.
const cTipInnerK = 0.006

// romanC is the lower stroke of c with its terminal: from the left extreme,
// where the stroke is CurveLower thick, round the bottom, ThinLower thick, to
// the cut at right angles, ETerminalTip long, whose outer corner lies
// CTipHeight above the baseline and CBowlWidth right of the left extreme.  The
// edges have the curvatures of the super-ellipses of the o at the left and at
// the bottom, and the outer edge that of its own at the cut; the inner edge
// leaves the cut with the curvature cTipInnerK.
func romanC(p design.Params) Part {
	l := design.CBowlLeft(p)
	bot, top := -p.OvershootLower, p.XHeight+p.OvershootLower
	y := bot + p.EThickPos*(top-bot)
	// The width of the outer super-ellipse which puts the corner of the cut at
	// CBowlWidth.
	w := p.CBowlWidth
	var outer Cubic
	for i := 0; i < 50; i++ {
		outer = superQuarter(l+w/2, y, w/2, y-bot, p.SupernessOuter, 1, -1)
		w += l + p.CBowlWidth - outer.At(outer.tAtY(p.CTipHeight)).X
	}
	cx := l + w/2
	ao, bo := w/2, y-bot
	ai, bi := ao-p.CurveLower, bo-p.ThinLower
	oBot, oSide := SuperK(ao, bo, p.SupernessOuter)
	iBot, iSide := SuperK(ai, bi, p.SupernessInner)
	tc := outer.tAtY(p.CTipHeight)
	tip, dir := outer.At(tc), outer.Tangent(tc)
	return Part{Name: "lower stroke of c", Glyph: "c", Strokes: []Stroke{{
		Nodes: []Node{
			{P: Point{X: l + p.CurveLower/2, Y: y}, Dir: 270, W: p.CurveLower, Edge: &EdgeK{Right: oSide, Left: iSide}},
			{P: Point{X: cx, Y: bot + p.ThinLower/2}, Dir: 0, W: p.ThinLower, Edge: &EdgeK{Right: oBot, Left: iBot}},
			{P: add(tip, mul(rot90(dir), p.ETerminalTip/2)), Dir: angleOf(dir), W: p.ETerminalTip,
				Edge: &EdgeK{Right: outer.Curvature(tc), Left: cTipInnerK}},
		},
		Start: End{Sever: true},
	}}}
}
