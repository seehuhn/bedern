// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

import (
	"fmt"
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/outline"
)

// The braces and the angle brackets.  The sizes of the braces are drawn by
// mathBrace at their height with the arm grown by the height ratio to the
// power weightPower and the ink width by its square root, as sized does.  The
// assembly of a brace has a middle piece: the construction
// of the largest weight, 2 (2 endRun - connector) lower than the largest
// size, is cut at the middle of the straight run of each arm; the bottom
// part is what lies below the lower cut, the top part what lies above the
// upper cut, the middle piece what lies between, each with endRun of straight
// stroke put in at every cut, and the extender is the arm, extender long.
// The largest size is the construction stretched at both cuts; the assembly
// with no extender is the largest size.
//
// The angle brackets are two straight strokes of the rule, at the height of
// the parentheses, whose outer edges meet in a point; their ink is angleWidth
// times the bow of the parenthesis wide.  Their width grows as the square root
// of the height and the strokes grow by weightPower with the rule, which sized
// grows.  They have the side spaces of the parenthesis and no assembly.
const (
	// angleWidth is the ink width of the angle bracket over the bow of the
	// parenthesis: the base is tan 29 degrees times half the height of the
	// parentheses wide, 265 units over a bow of 174.
	angleWidth = 265.0 / 174.0
	// withMiddle is the kind of assembly of the braces: two ends, a middle
	// piece and two extenders.
	withMiddle = withEnds + 1
)

// angleBracket draws ⟨, or ⟩ if right is true.
func angleBracket(right bool) func(r *roman) (outline.Path, float64) {
	return func(r *roman) (outline.Path, float64) {
		p := r.p
		t := p.Punct.MathThick
		top, bottom := p.BracketTop(), p.BracketBottom()
		mid := (top + bottom) / 2
		half := (top - bottom) / 2
		w := angleWidth * p.Punct.ParenSagitta
		// The outer edge of the upper stroke runs from the vertex at (0, mid)
		// to (w - a, top), where a is the run of the stroke along the cut.
		a := slantedRun(t, w, half) // the stroke across a horizontal cut
		// The inner edges meet on the middle line, a to the right of the
		// vertex.
		path := outline.Path{polygon(
			pt(0, mid), pt(w-a, bottom), pt(w, bottom), pt(a, mid), pt(w, top), pt(w-a, top))}
		sides := p.Punct.ParenLeft
		if right {
			_, _, x1, _ := path.BBox()
			path = mirroredX(path, x1/2)
			sides.Left, sides.Right = sides.Right, sides.Left
		}
		return withSides(path, sides.Left, sides.Right)
	}
}

// armCut returns the middle of the longest straight upright segment of the
// path above (or below) the line y = yc: the middle of the straight run of an
// arm of the brace.
func armCut(path outline.Path, yc float64, above bool) (float64, error) {
	best, at := 0.0, 0.0
	for _, c := range path {
		cur := c.Start
		for _, s := range c.Segs {
			a, b := cur, s.P3
			cur = b
			if s.IsCurve || math.Abs(a.X-b.X) > 1e-6 {
				continue
			}
			if above && math.Min(a.Y, b.Y) <= yc || !above && math.Max(a.Y, b.Y) >= yc {
				continue
			}
			if l := math.Abs(a.Y - b.Y); l > best {
				best, at = l, (a.Y+b.Y)/2
			}
		}
	}
	if best == 0 {
		return 0, fmt.Errorf("no straight upright run %s %v", map[bool]string{true: "above", false: "below"}[above], yc)
	}
	return math.Round(at), nil
}

// braceCuts returns the heights of the two cuts of the construction.
func braceCuts(path outline.Path, axis float64) (lo, hi float64, err error) {
	hi, err = armCut(path, axis, true)
	if err != nil {
		return 0, 0, err
	}
	lo, err = armCut(path, axis, false)
	return lo, hi, err
}

// braceConstruction returns the construction of the assembly of the brace d
// and the heights of its cuts.
func braceConstruction(p design.Params, d delimiter) (outline.Path, float64, float64, float64, error) {
	largest := delimSizes[len(delimSizes)-1]
	s := 2*endRun - connector
	q := sized(p, largest-2*s, grow(p, largest), math.Sqrt(largest/baseHeight(p)))
	path, adv := d.draw(newRomanNominal(q, largest))
	lo, hi, err := braceCuts(path, p.Punct.MathAxis)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	return path, math.Round(adv), lo, hi, nil
}

// braceLargest returns the largest size of the brace d: the construction
// stretched at both cuts.
func braceLargest(p design.Params, d delimiter) (outline.Path, float64, error) {
	path, adv, lo, hi, err := braceConstruction(p, d)
	if err != nil {
		return nil, 0, err
	}
	s := 2*endRun - connector
	path, err = stretchPath(path, hi, s)
	if err != nil {
		return nil, 0, err
	}
	path, err = stretchPath(path, lo-s/2, s)
	if err != nil {
		return nil, 0, err
	}
	return path, adv, nil
}

// braceParts returns the parts of the assembly of the brace d from the bottom
// up, the glyphs of the bottom part, the extender, the middle piece and the
// top part.
func braceParts(p design.Params, d delimiter) ([]Glyph, []Part, error) {
	path, adv, lo, hi, err := braceConstruction(p, d)
	if err != nil {
		return nil, nil, err
	}
	return cutBrace(path, adv, lo, hi, d.base, likeOf(d.base), endRun, connector, extender)
}

// cutBrace cuts the construction of a brace at the heights lo and hi into
// the parts of its assembly from the bottom up, each with run of straight arm
// put in at every cut, the connectors conn long and the extender ext long.
func cutBrace(path outline.Path, adv, lo, hi float64, base, like string, run, conn, ext float64) ([]Glyph, []Part, error) {
	if len(path) != 1 {
		return nil, nil, fmt.Errorf("the brace has %d contours, want 1", len(path))
	}
	c := withNodesAt(withNodesAt(path[0], hi), lo)
	top, err := end(c, hi, run, true)
	if err != nil {
		return nil, nil, err
	}
	bot, err := end(c, lo, run, false)
	if err != nil {
		return nil, nil, err
	}
	// The middle piece: the part below the upper cut, then the part of that
	// above the lower cut.
	below, err := end(c, hi, run, false)
	if err != nil {
		return nil, nil, err
	}
	_, y0, _, _ := path.BBox()
	cl := withNodesAt(below[0], lo-y0)
	mid, err := end(cl, lo-y0, run, true)
	if err != nil {
		return nil, nil, err
	}
	above, under, err := halves(withNodesAt(path[0], hi), hi)
	if err != nil {
		return nil, nil, err
	}
	x0, x1 := math.Min(above.start.X, under.start.X), math.Max(above.start.X, under.start.X)
	extPath := outline.Path{box(x0, 0, x1, ext)}
	height := func(q outline.Path) float64 { _, _, _, h := q.BBox(); return h }
	gs := []Glyph{
		{Name: base + ".bot", Advance: adv, Path: bot, Class: Fence, Like: like},
		{Name: base + ".ext", Advance: adv, Path: extPath, Class: Fence, Like: like},
		{Name: base + ".mid", Advance: adv, Path: mid, Class: Fence, Like: like},
		{Name: base + ".top", Advance: adv, Path: top, Class: Fence, Like: like},
	}
	ps := []Part{
		{Name: base + ".bot", EndConn: conn, FullAdvance: height(bot)},
		{Name: base + ".ext", StartConn: conn, EndConn: conn, FullAdvance: ext, Extender: true},
		{Name: base + ".mid", StartConn: conn, EndConn: conn, FullAdvance: height(mid)},
		{Name: base + ".ext", StartConn: conn, EndConn: conn, FullAdvance: ext, Extender: true},
		{Name: base + ".top", StartConn: conn, FullAdvance: height(top)},
	}
	return gs, ps, nil
}

// The brace of the maths font: each half is one stroke through four
// stations, as the brace of the regular is, but with proportions which hold
// at every size.  The arm is BraceArm thick; the end of the hook is braceEnd
// of that; the hook turns from the end into the arm over braceHook of the
// height and the arm turns into the point over braceTurn of the height, so
// the curls keep their share of the height as the brace grows (but at least
// as long as the reach requires, see braceHookMin); at the point the stroke
// is braceTipWide times BraceTip across the horizontal cut, which makes it
// 0.8 of the arm thick across its direction, which leans BraceTipAngle
// degrees from the horizontal.  The two halves are one contour, joined along
// the cut at the middle, where the outer edges meet in the point.  A brace
// drawn for a
// construction whose weight and width are those of a larger delimiter takes
// the proportions of that one (roman.nominal).
const (
	braceHook    = 0.16
	braceTurn    = 0.14
	braceEnd     = 0.70
	braceTipWide = 2.0
	// braceHookMin is the least distance of the centre line of the hook,
	// from the middle of its end to the start of the arm, over its reach;
	// the turn into the point is at least as long as its reach.
	braceHookMin = 1.15
	// braceArmAt is the place of the centre line of the arm, over the ink
	// width from the point.
	braceArmAt = 0.52
)

// mathBrace draws the left brace, or the right one if right is set, over the
// height of the parentheses of p, with the proportions of a delimiter nominal
// high (zero for its own height).  It returns the outline, the advance and
// whether it succeeded.
func mathBrace(p design.Params, nominal float64, right bool) (outline.Path, float64, bool) {
	s := p.Sym
	top, bottom := p.BracketTop(), p.BracketBottom()
	mid := (top + bottom) / 2
	h := top - bottom
	if nominal == 0 {
		nominal = h
	}
	t := s.BraceArm
	x0, x1 := s.BraceSide, s.BraceSide+s.BraceInk
	xa := x0 + braceArmAt*s.BraceInk
	tip := braceTipWide * s.BraceTip
	e := braceEnd * t
	half := braceHalf(top, mid, x0, x1, xa, t, e, tip, s.BraceTipAngle, math.Max(braceHook*nominal, braceHookMin*(x1-xa)+e/2), math.Max(braceTurn*nominal, x1-xa))
	if half == nil {
		return nil, 0, false
	}
	// The cut at the middle of the half: the segment which lies on y = mid.
	c := half[0]
	if n := len(c.Segs); n > 0 && c.Segs[n-1].P3 != c.Start {
		c.Segs = append(c.Segs[:n:n], outline.Segment{P3: c.Start})
	}
	cut := -1
	for i := range c.Segs {
		a, b := node(c, i), c.Segs[i].P3
		if !c.Segs[i].IsCurve && math.Abs(a.Y-mid) < 1e-6 && math.Abs(b.Y-mid) < 1e-6 {
			cut = i
		}
	}
	if cut < 0 {
		return nil, 0, false
	}
	// The run of the contour from the end of the cut round to its start, and
	// its mirror image in the middle line from the start of the cut back.
	n := len(c.Segs)
	run := outline.Contour{Start: c.Segs[cut].P3}
	for k := 1; k < n; k++ {
		run.Segs = append(run.Segs, c.Segs[(cut+k)%n])
	}
	// The ends of the cut lie on the middle line exactly.
	run.Start.Y = mid
	run.Segs[len(run.Segs)-1].P3.Y = mid
	out := outline.Contour{Start: run.Start, Segs: append([]outline.Segment(nil), run.Segs...)}
	pts := make([]outline.Point, 0, len(run.Segs)+1)
	pts = append(pts, run.Start)
	for _, sg := range run.Segs {
		pts = append(pts, sg.P3)
	}
	flip := func(q outline.Point) outline.Point { return outline.Point{X: q.X, Y: 2*mid - q.Y} }
	for i := len(run.Segs) - 1; i >= 0; i-- {
		sg := run.Segs[i]
		out.Segs = append(out.Segs, outline.Segment{IsCurve: sg.IsCurve, P1: flip(sg.P2), P2: flip(sg.P1), P3: flip(pts[i])})
	}
	out.Segs[len(out.Segs)-1].P3 = out.Start
	path := outline.Path{out}
	adv := 2*s.BraceSide + s.BraceInk
	if right {
		path = mirroredX(path, adv/2)
	}
	return path, adv, true
}

// braceHalf returns the upper half of the brace: the stroke from the end of the
// hook at the top to the point at mid.
func braceHalf(top, mid, x0, x1, xa, t, e, tip, angle, hook, turn float64) outline.Path {
	if top-hook <= mid+turn {
		return nil
	}
	return glyphs.Stroke([]glyphs.Station{
		{X: x1, Y: top - e/2, Dir: 180, W: e, Nib: -90, Ease: true},
		{X: xa, Y: top - hook, Dir: -90, W: t, Flat: true},
		{X: xa, Y: mid + turn, Dir: -90, W: t},
		{X: x0 + tip/2, Y: mid, Dir: 180 + angle, W: tip, Nib: 360, Ease: true},
	})
}
