// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

import (
	"fmt"
	"math"
	"slices"
	"sync"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"

	"seehuhn.de/go/bedern/internal/fair"
	"seehuhn.de/go/bedern/internal/stroke"
)

// The integral at text and display size.  The integral is one pen stroke of
// the stroke tool, drawn upright and sheared by the slant 1/8 with the pen
// put on the sheared centre line: from the head, thin there, it turns down
// through the middle, where it is thickest and straight, to the foot, thin
// again, and ends at both ends in the ball terminal of the italic, its
// radius grown by the weight factor to the power intBallGrow; the head is the
// foot turned through half a turn (halfTurnOfFoot).  The stem and the thin
// strokes are the capital stem and arm grown by the law of the big operators
// (bigops.go), the stem times intWeight, and the reach of the hooks from the
// centre of the stem grows as the square root of the height.
//
// LuaTeX sets the superscript of an integral at its advance and the
// subscript the italics correction to the left of it; it applies no math
// kerns to a big operator.  The italics correction puts the
// subscript intSide right of the ink of the foot.
const (
	// intTextHeight and intDisplayHeight are the heights.
	intTextHeight    = 1100.0
	intDisplayHeight = 2200.0
	// intSlantNum / intSlantDen is the slant, 7.1 degrees.
	intSlantNum, intSlantDen = 1, 8
	// intWeight is the stem over the law of the big operators.
	intWeight = 0.76
	// intReach is the reach of a hook from the centre of the stem to the
	// outside of its ball at text size.
	intReach = 208.0
	// intBallX is the distance of the node at the head and at the foot from
	// the centre of the stem over that of the centre of the ball, and
	// intBallUp the depth of the centre of the ball below the tangent of the
	// outer edge there over its radius.
	intBallX  = 0.7
	intBallUp = 1.95
	// intBallGrow is the power of the weight factor by which the ball grows.
	intBallGrow = 0.5
	// intSide is the side space on each side of the ink.
	intSide = 55.0
	// intFootBand is the part of the height, from the foot, whose ink the
	// subscript clears.
	intFootBand = 0.35
)

// The ball terminal of the italic (bedern/internal/font/italic, ballEnd): the tip of the
// wedge lies the neck of the curl of the figures from the outer edge, the
// inner edge meets the circle at ballOpen degrees, and the ball is placed by
// the rule of the spiral of the stroke tool with a power in ballPower.
const ballOpen = 53.0

var ballPower = [2]float64{1, 2.5}

// An intShape is the choice of the values of the stroke which the parameter
// sets leave open: the reach of the hooks and the place of the ball at their
// ends, as shares of the values above.
type intShape struct {
	reach  float64 // the share of intReach
	ballX  float64 // intBallX
	ballUp float64 // intBallUp
	open   float64 // ballOpen
	ball   float64 // the share of the diameter of the ball
	neck   float64 // the share of the neck of the curl
}

// intShapes are the shapes in the order of preference: the first is the one
// of the regular parameters, the second that of the heavier strokes (caps-up:
// the thin parts 21 per cent heavier against a ball which grows only with the
// square root of the weight; the ball and the neck of its curl grow
// 10 and 30 per cent and the wedge opens wider) and the third that of the
// lighter strokes (caps-down); the others, a short search about them, are
// those which the integral takes where these have findings of the fairness
// analysis (integralShape).
var intShapes = func() []intShape {
	out := []intShape{
		{1, intBallX, intBallUp, ballOpen, 1, 1},
		{1, intBallX, intBallUp, 60, 1.1, 1.3},
		{0.92, intBallX, 2.1, 56, 1, 1},
	}
	for _, up := range []float64{intBallUp, 1.8, 2.1, 1.7, 2.25} {
		for _, bx := range []float64{intBallX, 0.6, 0.8, 0.5} {
			for _, reach := range []float64{1, 0.96, 1.04, 0.92, 1.08} {
				for _, open := range []float64{ballOpen, 50, 56, 60} {
					for _, ball := range []float64{1, 1.1, 1.2} {
						sh := intShape{reach, bx, up, open, ball, 1}
						if !slices.Contains(out, sh) {
							out = append(out, sh)
						}
					}
				}
			}
		}
	}
	return out
}()

// integralStroke returns the stroke of the integral h high, centred on the
// axis, upright, in the shape sh.
func integralStroke(p design.Params, h float64, sh intShape) stroke.Stroke {
	g := math.Pow(h/p.CapHeight, opPower)
	stem := intWeight * p.Caps.Stem * g
	thin := p.Caps.Bar * g
	reach := sh.reach * intReach * math.Sqrt(h/intTextHeight)
	r := sh.ball * p.BallDiameter * math.Pow(g, intBallGrow) / 2
	bottom, top := p.Punct.MathAxis-h/2, p.Punct.MathAxis+h/2
	cb := reach - r // the centre of a ball from the centre of the stem
	ball := stroke.End{Kind: stroke.Ball, Ball: stroke.BallSpec{
		R: r, Neck: sh.neck * p.Figures.CurlNeck * g, Open: sh.open, TipK: stroke.Auto,
		Power: ballPower, Depth: sh.ballUp * r,
	}}
	return stroke.Stroke{
		Nodes: []stroke.Node{
			{P: pt(sh.ballX*cb, top-thin/2), Dir: 180, K: stroke.Auto, W: thin},
			{P: pt(0, p.Punct.MathAxis), Dir: 270, K: 0, W: stem},
			{P: pt(-sh.ballX*cb, bottom+thin/2), Dir: 180, K: stroke.Auto, W: thin},
		},
		Start: ball,
		End:   ball,
	}
}

// integral draws ∫ h high, sheared, with intSide on each side of the ink, and
// returns its outline, its advance and its italics correction: LuaTeX sets
// the superscript of an integral at its advance and the subscript the
// italics correction to the left of it, so the correction is the distance
// from the rightmost ink of the lower intFootBand of the height, plus
// intSide, to the advance.
func integral(p design.Params, h float64) (outline.Path, float64, float64, error) {
	return integralOf(p, h, integralShape(p))
}

// integralDrawn is what integralOf returned for one list of arguments.
type integralDrawn struct {
	path     outline.Path
	adv, ic  float64
	err      error
	computed sync.Once
}

// integralCache holds the results of integralOf, which the search for the
// shape, the glyphs and the tests ask for again and again.
var integralCache sync.Map // string to *integralDrawn

// integralOf is integral in the shape sh.  It draws each integral once and
// hands out a copy of the path.
func integralOf(p design.Params, h float64, sh intShape) (outline.Path, float64, float64, error) {
	v, _ := integralCache.LoadOrStore(fmt.Sprintf("%v|%v|%v", p, h, sh), new(integralDrawn))
	d := v.(*integralDrawn)
	d.computed.Do(func() { d.path, d.adv, d.ic, d.err = drawIntegral(p, h, sh) })
	return clonePath(d.path), d.adv, d.ic, d.err
}

// clonePath returns a copy of path which shares no memory with it.
func clonePath(path outline.Path) outline.Path {
	if path == nil {
		return nil
	}
	out := make(outline.Path, len(path))
	for i, c := range path {
		out[i] = outline.Contour{Start: c.Start, Segs: append([]outline.Segment(nil), c.Segs...)}
	}
	return out
}

// drawIntegral is the drawing behind integralOf.
func drawIntegral(p design.Params, h float64, sh intShape) (outline.Path, float64, float64, error) {
	s := integralStroke(p, h, sh).ShearedRatio(intSlantNum, intSlantDen)
	r, err := s.Build()
	if err != nil {
		return nil, 0, 0, err
	}
	path, err := halfTurnOfFoot(r.Path(), p.Punct.MathAxis)
	if err != nil {
		return nil, 0, 0, err
	}
	path, adv := withSides(path, intSide, intSide)
	_, y0, _, _ := path.Round().BBox()
	right := math.Inf(-1)
	for i := 0; i <= 40; i++ {
		for _, x := range runsAt(path, y0+intFootBand*h*float64(i)/40) {
			right = math.Max(right, x)
		}
	}
	return path, adv, math.Round(adv - right - intSide), nil
}

// halfTurnOfFoot returns the integral made of its lower half and that half
// turned through half a turn about the middle of the stem on the axis, so
// that the head is exactly the foot turned: the stroke tool places the ball
// at the start and at the end of a sheared stroke a little differently.
func halfTurnOfFoot(path outline.Path, axis float64) (outline.Path, error) {
	if len(path) != 1 {
		return nil, fmt.Errorf("the integral has %d contours, want 1", len(path))
	}
	_, below, err := halves(splitAtY(withNodesAt(path[0], axis), axis), axis)
	if err != nil {
		return nil, err
	}
	a, b := below.start, below.end()
	cx, cy := (a.X+b.X)/2, (a.Y+b.Y)/2
	top := turned(outline.Path{{Start: below.start, Segs: below.segs}}, cx, cy)[0]
	c := outline.Contour{Start: a, Segs: append([]outline.Segment(nil), below.segs...)}
	c.Segs = append(c.Segs, top.Segs...)
	return outline.Path{straightJoined(c)}, nil
}

// straightJoined returns the closed contour, started at the end of its first
// segment, with each pair of straight segments which meet in line made one.
func straightJoined(c outline.Contour) outline.Contour {
	segs := append(append([]outline.Segment(nil), c.Segs[1:]...), c.Segs[0])
	out := outline.Contour{Start: c.Segs[0].P3}
	cur, prev := out.Start, out.Start
	for _, sg := range segs {
		n := len(out.Segs)
		if n > 0 && !sg.IsCurve && !out.Segs[n-1].IsCurve {
			u, v := pt(cur.X-prev.X, cur.Y-prev.Y), pt(sg.P3.X-cur.X, sg.P3.Y-cur.Y)
			if math.Abs(u.X*v.Y-u.Y*v.X) < 1e-6*math.Hypot(u.X, u.Y)*math.Hypot(v.X, v.Y) && u.X*v.X+u.Y*v.Y > 0 {
				out.Segs[n-1].P3 = sg.P3
				cur = sg.P3
				continue
			}
		}
		out.Segs = append(out.Segs, sg)
		prev, cur = cur, sg.P3
	}
	return out
}

// splitAtY returns the contour with every curve which crosses the line
// y = yc split where it crosses it.
func splitAtY(c outline.Contour, yc float64) outline.Contour {
	out := outline.Contour{Start: c.Start}
	cur := c.Start
	for _, sg := range c.Segs {
		p0 := cur
		cur = sg.P3
		if !sg.IsCurve || (p0.Y-yc)*(sg.P3.Y-yc) >= 0 {
			out.Segs = append(out.Segs, sg)
			continue
		}
		y := func(t float64) float64 {
			u := 1 - t
			return u*u*u*p0.Y + 3*u*u*t*sg.P1.Y + 3*u*t*t*sg.P2.Y + t*t*t*sg.P3.Y
		}
		lo, hi := 0.0, 1.0
		for range 60 {
			m := (lo + hi) / 2
			if (y(lo)-yc)*(y(m)-yc) <= 0 {
				hi = m
			} else {
				lo = m
			}
		}
		t := (lo + hi) / 2
		lerp := func(a, b outline.Point) outline.Point { return pt(a.X+t*(b.X-a.X), a.Y+t*(b.Y-a.Y)) }
		a, b, d := lerp(p0, sg.P1), lerp(sg.P1, sg.P2), lerp(sg.P2, sg.P3)
		e, f := lerp(a, b), lerp(b, d)
		m := lerp(e, f)
		m.Y = yc
		out.Segs = append(out.Segs,
			outline.Segment{IsCurve: true, P1: a, P2: e, P3: m},
			outline.Segment{IsCurve: true, P1: f, P2: d, P3: sg.P3})
	}
	return out
}

// runsAt returns the right ends of the runs of ink along the line y.
func runsAt(path outline.Path, y float64) []float64 {
	var out []float64
	for _, r := range runsOf(path, y) {
		out = append(out, r[1])
	}
	return out
}

// runsOf returns the runs of ink along the horizontal line at y.
func runsOf(path outline.Path, y float64) [][2]float64 {
	xs := path.CrossingsX(y)
	slices.Sort(xs)
	var out [][2]float64
	for i := 0; i+1 < len(xs); i += 2 {
		out = append(out, [2]float64{xs[i], xs[i+1]})
	}
	return out
}

// integralGlyphs returns the text and display sizes of the integral, the
// text size with the display size as its variant.
func integralGlyphs(p design.Params) ([]Glyph, error) {
	const base = "uni222B"
	var gs []Glyph
	for i, h := range []float64{intTextHeight, intDisplayHeight} {
		path, adv, ic, err := integral(p, h)
		if err != nil {
			return nil, fmt.Errorf("mathfont: the integral %v high: %w", h, err)
		}
		g := Glyph{Name: base, Rune: 0x222B, Advance: adv, Path: path, Class: Ord, Like: "f", Italic: ic}
		if i == 1 {
			g.Name, g.Rune = base+".d", 0
		} else {
			g.Variants = []string{base + ".d"}
		}
		gs = append(gs, g)
	}
	top, bot, ext, adv, err := integralParts(p)
	if err != nil {
		return nil, fmt.Errorf("mathfont: the parts of the integral: %w", err)
	}
	_, _, _, bh := bot.BBox()
	_, _, _, th := top.BBox()
	gs = append(gs,
		Glyph{Name: base + ".bot", Advance: adv, Path: bot, Class: Ord, Like: "f"},
		Glyph{Name: base + ".ext", Advance: adv, Path: ext, Class: Ord, Like: "f"},
		Glyph{Name: base + ".top", Advance: adv, Path: top, Class: Ord, Like: "f"})
	gs[0].Parts = []Part{
		{Name: base + ".bot", EndConn: connector, FullAdvance: bh},
		{Name: base + ".ext", StartConn: connector, EndConn: connector, FullAdvance: extender, Extender: true},
		{Name: base + ".top", StartConn: connector, FullAdvance: th},
	}
	return gs, nil
}

// integralShapes caches the shape of the integral by the values of the
// parameters it depends on.
var integralShapes sync.Map

// integralShape returns the first of intShapes in which the integral, at the
// text and display sizes, and the hooks of its assembly have no finding of
// the fairness analysis beyond the two inflections of the S-curve of the
// stem near its middle, or the shape with the fewest findings where none is
// free of them: the stroke tool leaves jumps and wiggles of curvature at the
// ball and the turn where the heavier or the lighter strokes of an extreme
// parameter set meet the reach and the ball of the regular, and the shape
// gives the values of the stroke for those sets.
func integralShape(p design.Params) intShape {
	key := fmt.Sprint(p.CapHeight, p.Caps.Stem, p.Caps.Bar, p.BallDiameter, p.Figures.CurlNeck, p.Punct.MathAxis)
	if v, ok := integralShapes.Load(key); ok {
		return v.(intShape)
	}
	best, bestN := intShapes[0], -1
	for _, sh := range intShapes {
		n := integralFindings(p, sh)
		if bestN < 0 || n < bestN {
			best, bestN = sh, n
		}
		if n == 0 {
			break
		}
	}
	integralShapes.Store(key, best)
	return best
}

// integralFindings returns the number of findings of the fairness analysis of
// the integral and the hooks of its assembly in the shape sh, without the
// inflections of the S-curve of the stem near its middle; a shape which
// cannot be drawn counts as many.
func integralFindings(p design.Params, sh intShape) (n int) {
	defer func() {
		if recover() != nil {
			n = 1000
		}
	}()
	for _, h := range []float64{intTextHeight, intDisplayHeight} {
		path, _, _, err := integralOf(p, h, sh)
		if err != nil {
			return 1000
		}
		path = path.Round()
		_, y0, _, y1 := path.BBox()
		for _, f := range fair.Analyse("uni222B", path, fair.Default).Findings {
			if f.Kind == fair.Inflection && math.Abs(f.Y-(y0+y1)/2) < 0.15*(y1-y0) {
				continue
			}
			n++
		}
	}
	top, bot, _, _, err := integralPartsOf(p, sh)
	if err != nil {
		return 1000
	}
	for _, path := range []outline.Path{top, bot} {
		n += len(fair.Analyse("uni222B.top", path.Round(), fair.Default).Findings)
	}
	return n
}
