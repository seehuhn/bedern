// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/outline"

	"seehuhn.de/go/bedern/internal/stroke"
)

// The double-struck capitals U and J, by the rule of dscaps.go: a doubled stem
// which runs on into a round or a hook of one line.  The stem and the round
// are one stroke of the stroke tool whose pen is the doubled stroke along the
// straight stem and tapers to one line round the bend, and the white of the
// doubled stem is a stroke of its own which follows it into the bend as a
// lens, a line from each edge of the pen.

// dsBend describes a stem which runs on into a round: the stroke of the
// outline, the stroke of its white, and where the straight stem ends.
type dsBend struct {
	outer stroke.Result
	white stroke.Result
}

// bendOf builds the stroke of a doubled stem at x with the centre line xc
// which runs straight down from the height top to yr and then round half an
// ellipse about (xc + dir a, yr) of the half axes a and b, to the angle end
// in degrees on the ellipse (positive: counter-clockwise from the left
// extreme, for the round of U; negative: clockwise from the right extreme, for
// the hook of J), the pen a doubled stroke at the stem, wMid at the node
// between the stem and the bottom, and one line from the bottom on.
func (d dsFrame) bendOf(xc, top, yr, a, b float64, ccw bool, end float64, sh bendShape) dsBend {
	t := d.t
	// The nodes lie on the ellipse about (cx, yr): the stem node at angle 0 of
	// the walk, then every 45 degrees to the end.
	var cx float64
	phi := func(w float64) float64 { return w } // the position angle of the walk angle w
	if ccw {
		cx = xc + a
		phi = func(w float64) float64 { return 180 + w }
	} else {
		cx = xc - a
		phi = func(w float64) float64 { return -w }
	}
	at := func(w float64) (outline.Point, float64) {
		ph := phi(w) * math.Pi / 180
		q := pt(cx+a*math.Cos(ph), yr+b*math.Sin(ph))
		// The direction of travel: counter-clockwise for the round of U, clockwise
		// for the hook of J.
		var dir float64
		if ccw {
			dir = math.Atan2(b*math.Cos(ph), -a*math.Sin(ph))
		} else {
			dir = math.Atan2(-b*math.Cos(ph), a*math.Sin(ph))
		}
		return q, math.Mod(dir*180/math.Pi+360, 360)
	}
	wTaper := t + sh.taper*(d.d-t)
	walk := []struct{ w, width float64 }{{0, d.d}, {sh.angle, wTaper}, {90, t}, {135, t}}
	var nodes []stroke.Node
	nodes = append(nodes, stroke.Node{P: pt(xc, yr), Dir: 270, K: 0, W: d.d})
	for _, s := range walk[1:] {
		if s.w >= end-15 {
			break
		}
		q, dir := at(s.w)
		nodes = append(nodes, stroke.Node{P: q, Dir: dir, K: stroke.Auto, W: s.width})
	}
	q, dir := at(end)
	last := stroke.Node{P: q, Dir: dir, K: stroke.Auto, W: t}
	if ccw {
		// The round of U ends on the right stem, straight up.
		last.Dir, last.K = 90, 0
	}
	nodes = append(nodes, last)
	st := stroke.Stroke{Nodes: nodes, Start: stroke.End{Kind: stroke.Cut}, End: stroke.End{Kind: stroke.Cut}}
	r, err := st.Build()
	if err != nil {
		panic(err)
	}
	res, err := st.Resolved()
	if err != nil {
		panic(err)
	}
	// The white: from a line under the top of the stem along the straight stem
	// into the bend, to the node where the pen is dsBendWhite wide.
	wn := res[1]
	wn.W = wTaper - 2*t
	wn.DW, wn.DDW = stroke.Auto, 0
	ws := stroke.Stroke{
		Nodes: []stroke.Node{
			{P: pt(xc, top-t), Dir: 270, K: 0, W: d.g},
			{P: pt(xc, yr), Dir: 270, K: 0, W: d.g},
			wn,
		},
		Start: stroke.End{Kind: stroke.Cut}, End: stroke.End{Kind: stroke.Cut},
	}
	wr, err := ws.Build()
	if err != nil {
		panic(err)
	}
	return dsBend{*r, *wr}
}

// A bendShape holds the numbers of a bend: the share of the way from one line
// to the doubled stroke at which the pen is at the second node, the height at
// which the straight stem ends over the cap height and the angle on the
// ellipse of the second node.
type bendShape struct{ taper, top, angle float64 }

// dsU draws U: the doubled left stem with the serifs of the capitals, which
// runs on round the bottom and up the right side as a line, and the right
// stem of one line with the serifs of its head on its left, as the round of
// the roman, whose right stem the bend leaves straight up.
func dsU(d dsFrame) outline.Path {
	return fairest("U", len(dsUBends), func(i int) outline.Path { return d.settled(dsUOf(d, dsUBends[i])) })
}

func dsUOf(d dsFrame, sh bendShape) outline.Path {
	cap := d.p.CapHeight
	jut := d.p.Caps.SerifJut
	x := d.stemLeft()
	xc := x + d.d/2
	rr := d.romanRun(0.3*cap, lastRun)
	xrc := d.x((rr[0]+rr[1])/2) + d.extra
	_, y0, _, _ := d.roman.BBox()
	yr := sh.top * cap
	bend := d.bendOf(xc, cap, yr, (xrc-xc)/2, yr-(y0+d.t/2), true, 180, sh)
	var c pen
	c.run(chainEdge(bend.outer.Right))
	c.run(glyphs.CapHeadRight(d.q, xrc+d.t/2, cap, jut))
	c.run(glyphs.CapHeadLeft(d.q, xrc-d.t/2, cap, jut))
	c.lineTo(pt(xrc-d.t/2, yr))
	c.run(openReversed(chainEdge(bend.outer.Left)))
	d.headRight(&c, x+d.d, cap)
	d.headLeft(&c, x, cap)
	path := outline.Path{c.close()}
	for _, w := range bend.white.Path() {
		path = append(path, reversed(pointed(w, 12)))
	}
	return path
}

// The bends of U which dsU tries in turn, and the shapes of J which dsJ tries
// in turn (the first of each tapers from the doubled stroke to the line as
// fast as the white of the lens allows, which is two lines wide where it ends;
// the bend is 121 and 115 thick at a third of the width): the bend, the angle
// on the ellipse at which the hook ends and how far its end stands right of
// the left edge of the roman.  The first of each is the shape which the
// fairness analysis accepts on the regular parameters; the others stand by
// for the places where the rounding to whole units leaves it a finding.
var (
	dsUBends  = bendShapes(0.332, 0.32, 0.35, 0.31, 0.34)
	dsJShapes = func() []struct {
		bend      bendShape
		end, hook float64
	} {
		var out []struct {
			bend      bendShape
			end, hook float64
		}
		for _, eh := range [][2]float64{{124.822, -50.128}, {124.822, -49}, {126, -50.128}, {124.822, -48}, {124.822, -52}, {122, -50.128}, {128, -50.128}} {
			for _, b := range bendShapes(0.4003, 0.39, 0.41, 0.38, 0.42) {
				out = append(out, struct {
					bend      bendShape
					end, hook float64
				}{b, eh[0], eh[1]})
			}
		}
		return out
	}()
)

// bendShapes returns the bends which dsU and dsJ try, for the heights tops of
// the start of the bend over the cap height: the share of the taper and the
// angle on the ellipse, nearest the first of each first.
func bendShapes(tops ...float64) []bendShape {
	var out []bendShape
	for _, top := range tops {
		for _, angle := range []float64{46, 48, 44, 50, 42, 40, 38} {
			for _, taper := range []float64{0.44, 0.45, 0.43, 0.46, 0.47, 0.42, 0.48} {
				out = append(out, bendShape{taper, top, angle})
			}
		}
	}
	return out
}

// dsJ draws J: the doubled stem with the serifs of the head on both sides,
// which runs on round the bottom to the left as a hook of one line, ending
// in a level cut.
func dsJ(d dsFrame) outline.Path {
	return fairest("J", len(dsJShapes), func(i int) outline.Path {
		return d.settled(dsJOf(d, dsJShapes[i].bend, dsJShapes[i].end, dsJShapes[i].hook))
	})
}

func dsJOf(d dsFrame, sh bendShape, end, hook float64) outline.Path {
	cap := d.p.CapHeight
	x := d.stemLeft()
	xc := x + d.d/2
	x0, y0, _, _ := d.roman.BBox()
	yr := sh.top * cap
	xe := d.x(x0) + d.t/2 + hook
	b := yr - (y0 + d.t/2)
	// The half axis a of the ellipse puts the end of the hook, at the angle end
	// on it, at xe.
	a := (xc - xe) / (1 - math.Cos(end*math.Pi/180))
	bend := d.bendOf(xc, cap, yr, a, b, false, end, sh)
	var c pen
	d.headLeft(&c, x, cap)
	c.run(chainEdge(bend.outer.Right))
	c.run(openReversed(chainEdge(bend.outer.Left)))
	d.headRight(&c, x+d.d, cap)
	path := outline.Path{c.close()}
	for _, w := range bend.white.Path() {
		path = append(path, reversed(pointed(w, 12)))
	}
	return path
}
