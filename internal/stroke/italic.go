// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package stroke

import (
	"fmt"
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The trial of italic forms.  These are pictures to test the tool, not a
// design of the italic: the letters are drawn upright from strokes, with entry
// and exit strokes, sheared by ItalicSlant with the pen put on the sheared
// centre lines, and joined by the union of their strokes.

// The slant of the trial: the tangent 1/4, 14.04 degrees, with a whole-unit
// point on the line of a stem every 4 units of height.
const (
	SlantNum, SlantDen = 1, 4
)

// ItalicSlant is the slant of the trial in degrees.
var ItalicSlant = math.Atan2(SlantNum, SlantDen) * 180 / math.Pi

// italicNarrow is the advance of the italic n over that of the roman n.
const italicNarrow = 0.96

// italicSide is the side space of the letters of the trial, from the ink.
const italicSide = 32.0

// An ItalicGlyph is a letter of the trial.
type ItalicGlyph struct {
	Name    string
	Letter  Letter // the strokes, sheared
	Result  *LetterResult
	Path    outline.Path // the outline, moved to the side space
	Advance float64
}

// it holds the parameters of the trial.
type it struct {
	p          design.Params
	counter    float64 // the counter of n
	top, asc   float64 // x-height and ascender of the stems
	exitTurn   float64 // height at which an exit stroke leaves the stem
	archBranch float64 // height of the node in the left stem where an arch starts
}

func newIt(p design.Params) it {
	return it{p: p, counter: 0.92 * p.NCounter, top: p.XHeight, asc: p.Ascender, exitTurn: 148, archBranch: 270}
}

// entryStroke returns the entry stroke into the top of a stem whose centre
// stands at x and whose top is at the height top: a hairline which comes in
// from the left, rising, thickens and ends inside the stem, which the union
// joins it to.
func (t it) entryStroke(x, top float64) Stroke {
	return Stroke{
		Nodes: []Node{
			{P: Point{X: x - 84, Y: top - 74}, Dir: 36, K: Auto, W: 16},
			{P: Point{X: x + 2, Y: top - 30}, Dir: 24, K: Auto, W: 30},
		},
		Start: End{Kind: Cut},
		End:   End{Kind: Cut, Hidden: true},
	}
}

// stemTopAngle is the direction of the cut at the top of a stem with an
// entry stroke: rising to the right.
const stemTopAngle = 18.0

// exit returns the nodes of the foot of a stem at x which turns into an exit
// stroke: straight down to exitTurn, round the bottom thin and up to the right.
func (t it) exit(x float64) []Node {
	return []Node{
		{P: Point{X: x, Y: t.exitTurn}, Dir: 270, W: t.p.Stem},
		{P: Point{X: x + 64, Y: 4 + 18}, Dir: 0, K: Auto, W: 36},
		{P: Point{X: x + 126, Y: 76}, Dir: 52, K: 0, W: 24},
	}
}

// foot returns the nodes of a stem at x which ends in a cut on the line y.
func (t it) foot(x, y float64) []Node {
	return []Node{{P: Point{X: x, Y: y + 32}, Dir: 270, W: t.p.Stem}, {P: Point{X: x, Y: y}, Dir: 270, W: t.p.Stem}}
}

// stem returns a stem at x from the height top, with a cut rising to the
// right for an entry stroke (entryStroke) if entry is set and along the
// horizontal otherwise, down to an exit stroke if
// exit is set and to a cut on the line bottom otherwise.
func (t it) stem(x, top float64, entry, exit bool, bottom float64) Stroke {
	var s Stroke
	if entry {
		s.Nodes = []Node{{P: Point{X: x, Y: top}, Dir: 270, W: t.p.Stem}}
		s.Start = End{Kind: Cut, Line: true, Angle: stemTopAngle}
	} else {
		s.Nodes = []Node{{P: Point{X: x, Y: top}, Dir: 270, W: t.p.Stem}}
		s.Start = End{Kind: Cut, Line: true, Angle: 0}
	}
	if exit {
		s.Nodes = append(s.Nodes, t.exit(x)...)
		s.End = End{Kind: Cut}
	} else {
		s.Nodes = append(s.Nodes, t.foot(x, bottom)...)
		s.End = End{Kind: Cut, Line: true, Angle: 0}
	}
	return s
}

// arch returns the arch from the stem at x1, which it leaves low, as italics
// branch, over the top, ThinLower thick, down the stem at x2 and into an exit
// stroke or a foot.  It starts inside the left stem; the union makes the
// crotch.
func (t it) arch(x1, x2 float64, exit bool) Stroke {
	top := t.top + t.p.ArchOvershoot
	nodes := []Node{
		{P: Point{X: x1, Y: t.archBranch}, Dir: 62, K: Auto, W: 30, DW: Auto},
		{P: Point{X: (x1+x2)/2 + 30, Y: top - t.p.ThinLower/2}, Dir: 0, K: Auto, W: t.p.ThinLower},
		{P: Point{X: x2, Y: 300}, Dir: 270, W: t.p.Stem},
	}
	s := Stroke{Start: End{Kind: Cut, Hidden: true}}
	if exit {
		s.Nodes = append(nodes, t.exit(x2)...)
		s.End = End{Kind: Cut}
	} else {
		s.Nodes = append(nodes, t.foot(x2, 0)...)
		s.End = End{Kind: Cut, Line: true, Angle: 0}
	}
	return s
}

// bowl returns the bowl of a, d, p and q by the bowl rule (BowlRule): it
// leaves the stem whose centre stands at x near its top and runs back into it
// low down, round the side of the sign side (-1 left, +1 right), and its ends
// lie inside the stem, which the union joins it to.
func (t it) bowl(x, side float64) Stroke {
	p := t.p
	edge := x + side*p.Stem/2
	w := 300.0
	return BowlRule{
		StemX: edge, Side: side, Free: edge + side*w,
		Top: t.top + p.OvershootLower, Bottom: -p.OvershootLower,
		CX: edge + side*0.5*w, ThickY: 212,
		Thin: p.ThinLower, Thick: p.CurveLower,
		NotchUpper: 404, JoinUpper: 352, NotchLower: 52, JoinLower: 104,
		JoinThick: 30, SupernessOut: p.SupernessOuter, SupernessIn: p.SupernessInner,
		Depth: 34,
	}.Stroke(false)
}

// round returns the stroke of c and e from the top of the right side over the
// top, down the left side, round the bottom and up into an exit stroke; left
// is the left side of the ink.
func (t it) round(left float64, start []Node) Stroke {
	p := t.p
	nodes := append(start,
		Node{P: Point{X: left + 190, Y: t.top + p.OvershootLower - p.ThinLower/2}, Dir: 180, K: Auto, W: p.ThinLower},
		Node{P: Point{X: left + 72, Y: 370}, Dir: 232, K: Auto, W: 66, DW: Auto},
		Node{P: Point{X: left + p.CurveLower/2, Y: 210}, Dir: 270, K: Auto, W: p.CurveLower, DW: Auto},
		Node{P: Point{X: left + 200, Y: -p.OvershootLower + p.ThinLower/2}, Dir: 0, K: Auto, W: p.ThinLower},
		Node{P: Point{X: left + 330, Y: 92}, Dir: 55, K: Auto, W: 24, DW: Auto},
	)
	return Stroke{Nodes: nodes, End: End{Kind: Cut}}
}

// cWithBall returns the c of the trial: the round of c and e which starts
// in the ball terminal of the roman, placed by its side space: its rightmost
// ink lies 372 units right of the left of the ink, below the top of the
// stalk, and the tip of its wedge lies BallStalk from the outer edge.
func (t it) cWithBall(left float64) Stroke {
	p := t.p
	s := t.round(left, []Node{
		{P: Point{X: left + 268, Y: 434}, Dir: 166, K: Auto, W: p.BallStalk + 4},
	})
	s.Start = End{Kind: Ball, Ball: BallSpec{
		Centre: Point{X: left + 320, Y: 345}, R: p.BallDiameter / 2,
		Neck: p.BallStalk, Open: ballOpen, TipK: Auto,
		InkSide: 1, InkX: left + 372,
	}}
	return s
}

// italicLetters returns the strokes of the letters of the trial, upright.
func italicLetters(p design.Params) map[string]Letter {
	t := newIt(p)
	s := p.Stem
	x1 := 80.0
	x2 := x1 + s + t.counter
	x3 := x2 + s + 0.9*t.counter
	dot := Disc(Point{X: x1 + 8, Y: p.DotCentre}, p.DotDiameter/2)
	lt := map[string]Letter{
		"n": {Strokes: []Stroke{t.stem(x1, t.top, true, false, 0), t.arch(x1, x2, true), t.entryStroke(x1, t.top)}},
		"m": {Strokes: []Stroke{t.stem(x1, t.top, true, false, 0), t.arch(x1, x2, false), t.arch(x2, x3, true), t.entryStroke(x1, t.top)}},
		"h": {Strokes: []Stroke{t.stem(x1, t.asc, true, false, 0), t.arch(x1, x2, true), t.entryStroke(x1, t.asc)}},
		"i": {Strokes: []Stroke{t.stem(x1, t.top, true, true, 0), t.entryStroke(x1, t.top)}, Extra: []Chain{dot}},
		"l": {Strokes: []Stroke{t.stem(x1, t.asc, true, true, 0), t.entryStroke(x1, t.asc)}},
		"a": {Strokes: []Stroke{t.stem(x1+250, t.top+4, false, true, 0), t.bowl(x1+250, -1)}},
		"d": {Strokes: []Stroke{t.stem(x1+250, t.asc, true, true, 0), t.bowl(x1+250, -1), t.entryStroke(x1+250, t.asc)}},
		"p": {Strokes: []Stroke{t.stem(x1, t.top, true, false, -p.Descender), t.bowl(x1, 1), t.entryStroke(x1, t.top)}},
		"e": {Strokes: []Stroke{t.round(30, []Node{
			{P: Point{X: 30 + 50, Y: 236}, Dir: 12, K: Auto, W: 30},
			{P: Point{X: 30 + 300, Y: 300}, Dir: 80, K: Auto, W: 46, DW: Auto},
			{P: Point{X: 30 + 272, Y: 412}, Dir: 128, K: Auto, W: 40, DW: Auto},
		})}},
		"c": {Strokes: []Stroke{t.cWithBall(30)}},
		"o": {Strokes: []Stroke{{
			Closed: true,
			Nib:    &Nib{Major: p.CurveLower, Minor: p.ThinLower, Angle: nibAngle},
			Nodes: []Node{
				{P: Point{X: 30 + 0.94*p.OBowlWidth - p.CurveLower/2, Y: t.top / 2}, Dir: 90, K: Auto, W: Auto, DW: Auto},
				{P: Point{X: 30 + 0.47*p.OBowlWidth, Y: t.top + p.OvershootLower - p.ThinLower/2}, Dir: 180, K: Auto, W: Auto, DW: Auto},
				{P: Point{X: 30 + p.CurveLower/2, Y: t.top / 2}, Dir: 270, K: Auto, W: Auto, DW: Auto},
				{P: Point{X: 30 + 0.47*p.OBowlWidth, Y: -p.OvershootLower + p.ThinLower/2}, Dir: 0, K: Auto, W: Auto, DW: Auto},
			},
		}}},
		"u": {Strokes: []Stroke{
			{
				Nodes: []Node{
					{P: Point{X: x1, Y: t.top}, Dir: 270, W: s},
					{P: Point{X: x1, Y: 150}, Dir: 270, W: s},
					{P: Point{X: (x1+x2)/2 - 10, Y: -p.OvershootLower + p.ThinLower/2}, Dir: 0, K: Auto, W: p.ThinLower},
					{P: Point{X: x2 - 44, Y: 60}, Dir: 45, K: Auto, W: 38, DW: Auto},
					{P: Point{X: x2 - 6, Y: 200}, Dir: 78, K: Auto, W: 42, DW: Auto},
					{P: Point{X: x2, Y: 290}, Dir: 90, K: Auto, W: 36, DW: Auto},
				},
				Start: End{Kind: Cut, Line: true, Angle: stemTopAngle}, End: End{Kind: Cut, Hidden: true},
			},
			t.stem(x2, t.top+4, false, true, 0),
			t.entryStroke(x1, t.top),
		}},
		"t": {Strokes: []Stroke{
			t.stem(x1+40, 580, false, true, 0),
			{
				Nodes: []Node{
					{P: Point{X: x1 + 40 - 95, Y: 425}, Dir: 0, W: p.Hair + 4},
					{P: Point{X: x1 + 40 + 125, Y: 425}, Dir: 0, W: p.Hair + 4},
				},
			},
		}},
	}
	// The cut at the top of t is slanted, the top of the ascender of t.
	tt := lt["t"]
	tt.Strokes[0].Start = End{Kind: Cut, Line: true, Angle: 35}
	lt["t"] = tt
	return lt
}

// nibAngle is the angle of the broad pen of the italic o: a little tilted
// against the vertical stress of the roman.
const nibAngle = -12.0

// Italic returns the letters of the trial, sheared and joined.
func Italic(p design.Params) (map[string]ItalicGlyph, error) {
	out := map[string]ItalicGlyph{}
	k := float64(SlantNum) / SlantDen
	for name, l := range italicLetters(p) {
		var sl Letter
		for _, s := range l.Strokes {
			sl.Strokes = append(sl.Strokes, s.ShearedRatio(SlantNum, SlantDen))
		}
		for _, c := range l.Extra {
			// A dot keeps its circle and moves with the shear of its centre.
			y0, y1 := math.Inf(1), math.Inf(-1)
			for _, pc := range c {
				y0, y1 = math.Min(y0, pc.C[0].Y), math.Max(y1, pc.C[0].Y)
			}
			sl.Extra = append(sl.Extra, shiftChain(c, k*(y0+y1)/2))
		}
		r, err := sl.Build()
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		path := r.Path()
		x0, _, x1, _ := path.BBox()
		adv := x1 - x0 + 2*italicSide
		if name == "n" {
			adv = italicNarrow * design.NAdvance(p)
		}
		var shifted outline.Path
		dx := italicSide - x0
		if name == "n" {
			dx = (adv-(x1-x0))/2 - x0
		}
		for _, c := range path {
			shifted = append(shifted, shift(c, dx))
		}
		out[name] = ItalicGlyph{Name: name, Letter: sl, Result: r, Path: shifted, Advance: adv}
	}
	return out, nil
}

func shift(c outline.Contour, dx float64) outline.Contour {
	out := outline.Contour{Start: Point{X: c.Start.X + dx, Y: c.Start.Y}}
	for _, s := range c.Segs {
		s.P1.X += dx
		s.P2.X += dx
		s.P3.X += dx
		out.Segs = append(out.Segs, s)
	}
	return out
}

func shiftChain(ch Chain, dx float64) Chain {
	out := make(Chain, len(ch))
	for i, pc := range ch {
		for j := range pc.C {
			pc.C[j].X += dx
		}
		out[i] = pc
	}
	return out
}
