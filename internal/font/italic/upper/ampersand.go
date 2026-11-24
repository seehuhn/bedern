// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package upper

import (
	"fmt"
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
	"seehuhn.de/go/bedern/internal/stroke"
)

// Et holds the shape of the ampersand of the italic, the Et form: an
// epsilon of two bowls, the small upper one starting in a ball and the
// larger lower one, which meet in a pointed waist, the lower bowl running on
// into the stem of a small t, crossed by a hairline bar at the x-height.  All
// positions are upright, before the shear.
type Et struct {
	Left                       float64 // the outer left extreme of the lower bowl
	Waist                      float64 // the height of the waist
	Loop                       float64 // the height of the loop at the waist, between its centre lines
	WaistX                     float64 // the point of the waist
	UpperL                     float64 // the outer left extreme of the upper bowl
	UpperX                     float64 // the top of the upper bowl
	BallInk                    float64 // the rightmost ink of the ball
	BallY                      float64 // the height of the centre of the ball
	TipK                       float64 // the curvature of the inner edge at the tip of the wedge of the ball
	StalkDX, StalkDY, StalkDir float64 // the first node of the stroke, from the centre of the ball, and its direction
	BottomX                    float64 // the bottom of the lower bowl
	RiseX                      float64 // the place where the rise into the stem runs at 45 degrees
	RiseY                      float64
	StemX                      float64 // the centre line of the stem of the t
	StemLow                    float64 // the height from which the stem runs straight
	StemHigh                   float64 // the height up to which it runs straight
	TopDX, TopW                float64 // the last node of the stem, from its centre line, where it runs vertically in the italic, and its thickness there
	Top                        float64 // the highest ink of the stem, which ends round
	BarL                       float64 // the bar of the t, from BarL to BarR, centred on the x-height
	BarR                       float64
	Advance                    float64
}

// defaultEt returns the shape of the ampersand.
func defaultEt() Et {
	return Et{
		Left: 65, Waist: 390, WaistX: 295, Loop: 130,
		UpperL: 115, UpperX: 275, BallInk: 470, BallY: 565, StalkDX: -60, StalkDY: 90, StalkDir: 150, TipK: math.NaN(),
		BottomX: 295, RiseX: 475, RiseY: 95,
		StemX: 585, StemLow: 300, StemHigh: 560, TopDX: -3.5, TopW: 78, Top: 649,
		BarL: 500, BarR: 680,
		Advance: 740,
	}
}

// ampersand draws the Et with the pen of the italic at the weights of the
// capitals and shears it on the lattice; the path is that of the italic.
func ampersand(p Params, rp design.Params) (outline.Path, float64, error) {
	return etWith(p, rp, defaultEt())
}

// etWith draws the Et of the shape e.
func etWith(p Params, rp design.Params, e Et) (outline.Path, float64, error) {
	thick := 0.92 * rp.Caps.Round * p.Width // the thick of the rounds at right angles, a little lighter
	thin := rp.Caps.RoundThin
	stem := 0.9 * p.CapStem * math.Cos(math.Atan(p.slant()))
	hair := 32.0 // the bar of the t, the hairline of the italic, its edges at heights which are multiples of 4
	nib := &stroke.Nib{Major: thick, Minor: thin, Angle: p.Turn}
	auto := math.NaN()
	node := func(x, y, dir float64) stroke.Node {
		return stroke.Node{P: pt(x, y), Dir: dir, K: auto, W: auto, DW: auto}
	}
	over := rp.OvershootUpper
	r := rp.Figures.Ball / 2
	thinNode := func(x, y, dir float64) stroke.Node {
		return stroke.Node{P: pt(x, y), Dir: dir, K: auto, W: thin, DW: auto}
	}
	main := stroke.Stroke{
		Nib: nib,
		Nodes: []stroke.Node{
			{P: pt(e.BallInk-r+e.StalkDX, e.BallY+e.StalkDY), Dir: e.StalkDir, K: auto, W: rp.BallStalk + 4, DW: auto},
			node(e.UpperX, rp.CapHeight+over-thin/2, 180),
			node(e.UpperL+thick/2, (rp.CapHeight+e.Waist)/2+10, 270),
			thinNode(e.WaistX, e.Waist-e.Loop/2, 0),
			thinNode(e.WaistX+e.Loop*0.6, e.Waist, 90),
			thinNode(e.WaistX, e.Waist+e.Loop/2, 180),
			node(e.Left+thick/2, e.Waist/2-10, 270),
			node(e.BottomX, -over+thin/2, 0),
			node(e.RiseX, e.RiseY, 42),
			{P: pt(e.StemX, e.StemLow), Dir: 90, K: 0, W: stem, DW: 0},
			{P: pt(e.StemX, e.StemHigh), Dir: 90, K: 0, W: stem, DW: 0},
			{P: pt(e.StemX+e.TopDX, e.Top-e.TopW/2), Dir: 90 + 180/math.Pi*math.Atan(p.slant()), K: auto, W: e.TopW, DW: auto},
		},
		Start: stroke.End{Kind: stroke.Ball, Ball: stroke.BallSpec{
			Centre: pt(e.BallInk-r, e.BallY), R: r,
			Neck: rp.BallStalk, Open: 53, TipK: e.TipK,
			InkSide: 1, InkX: e.BallInk,
		}},
		End: stroke.End{Kind: stroke.Round},
	}
	bar := stroke.Stroke{
		Nodes: []stroke.Node{
			{P: pt(e.BarL, rp.XHeight), Dir: 0, K: 0, W: hair, DW: 0},
			{P: pt(e.BarR, rp.XHeight), Dir: 0, K: 0, W: hair, DW: 0},
		},
		Start: stroke.End{Kind: stroke.Cut},
		End:   stroke.End{Kind: stroke.Cut},
	}
	dx := -p.shift()
	var sl stroke.Letter
	for _, s := range []stroke.Stroke{main, bar} {
		s = movedStroke(s, dx).ShearedRatio(p.SlantNum, p.SlantDen)
		sl.Strokes = append(sl.Strokes, s)
	}
	res, err := sl.Build()
	if err != nil {
		return nil, 0, fmt.Errorf("ampersand: %w", err)
	}
	return atExtremes(explicit(res.Path()), p.slant(), p.SlantDen, nil), e.Advance, nil
}

// movedStroke returns the stroke moved by dx.
func movedStroke(s stroke.Stroke, dx float64) stroke.Stroke {
	out := s
	out.Nodes = append([]stroke.Node(nil), s.Nodes...)
	for i := range out.Nodes {
		out.Nodes[i].P.X += dx
	}
	mv := func(e stroke.End) stroke.End {
		e.Q.X += dx
		e.Ball.Centre.X += dx
		e.Ball.InkX += dx
		return e
	}
	out.Start, out.End = mv(s.Start), mv(s.End)
	return out
}

func init() {
	sources["ampersand"] = source{r: '&', italic: ampersand}
}
