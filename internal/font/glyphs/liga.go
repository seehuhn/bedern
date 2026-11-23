// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"sync"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The ligatures of f.
//
// A ligature is one or two letters f followed by f, i, l or j, drawn as one
// glyph.  The stems stand design.LigaturePitch apart (design.LigatureCell), and
// the stems, feet, bars and terminals are those of the letters.  Each f is one
// of three kinds.
//
//	An f before i or j ends in a free hook whose ball takes the place of the
//	dot of i: its centre stands over the stem of the next letter where the
//	centre of the dot of i stands (fDotHook); the dot and the head serif of i
//	and j are left out.
//	An f before f ends in a free hook with the small ball of the ear of g
//	(smallHook), and the next f rises beside it with its own hook.
//	An f before l runs over its crest into the stem of l (archInto): the outer
//	edge of the hook of f up to the crest, and beyond it its mirror image, which
//	meets the left-hand edge of l in a corner below the top of l; the l has no
//	head serif.  The underside runs down into the stem of l (ligaArchUnder).
//
// The bar runs on from one stem to the next before f, i and j, and the bar of
// an f before l is the bar of f.  The last f of ff keeps the hook of f.
//
// The ligature takes the advance of its last letter, moved by the pitch for
// each letter before it (design.LigatureAdvance), and the side kinds of f on
// the left and of its last letter on the right.

// A ligatureEntry names a ligature and the letters it stands for.
type ligatureEntry struct {
	name  string
	r     rune
	parts []string
}

// ligatureList holds the ligatures of f.
var ligatureList = []ligatureEntry{
	{"ff", 'ﬀ', []string{"f", "f"}},
	{"fi", 'ﬁ', []string{"f", "i"}},
	{"fl", 'ﬂ', []string{"f", "l"}},
	{"ffi", 'ﬃ', []string{"f", "f", "i"}},
	{"ffl", 'ﬄ', []string{"f", "f", "l"}},
	{"f_j", 0, []string{"f", "j"}},
}

// ligatureParts holds the letters of each ligature by name.
var ligatureParts = map[string][]string{}

func init() {
	for _, l := range ligatureList {
		parts := l.parts
		ligatureParts[l.name] = parts
		last, _ := lookup(parts[len(parts)-1])
		table = append(table, entry{l.name, l.r, sideBarF, last.right, func(p design.Params) (outline.Path, float64) {
			return ligature(p, parts)
		}})
	}
}

// LigatureParts returns the letters for which the glyph with the given name
// stands, or nil if it is not a ligature.
func LigatureParts(name string) []string { return ligatureParts[name] }

// LigatureCarets returns the positions of the carets between the letters of
// the ligature with the given parts.
func LigatureCarets(p design.Params, parts []string) []float64 {
	var cs []float64
	for k := 0; k+1 < len(parts); k++ {
		cs = append(cs, design.LigatureCaret(p, k))
	}
	return cs
}

// ligature draws the ligature of the given letters, see the comment above.
func ligature(p design.Params, parts []string) (outline.Path, float64) {
	n := len(parts)
	last := parts[n-1]
	hp := newFHookParts(p)
	dx := func(k int) float64 { return design.LigatureCell(p, k) - design.NarrowStemLeft(p) }
	fs := make([]stemSpec, n-1)
	for k := range fs {
		fs[k] = fStem(p)
		fs[k].Left += dx(k)
	}
	bar := fBar(p)
	under := bar.Top - bar.Thick
	// dotted is the f with the free hook into the dot, or -1.
	dotted := -1
	if last == "i" || last == "j" {
		dotted = n - 2
	}
	joined := func(k int) bool { return parts[k+1] != "l" } // the bar runs on to the next stem

	var tail stemSpec
	switch last {
	case "f":
		tail = fStem(p)
	case "l":
		tail = stemSpec{
			Left: design.NarrowStemLeft(p) + design.AscenderInset(p), Width: p.AscenderStem,
			Top: p.Ascender, FootLeft: p.SerifJut, FootRight: p.SerifJut,
		}
	case "i":
		tail = stemSpec{
			Left: design.NarrowStemLeft(p), Width: p.Stem, Top: p.XHeight,
			FootLeft: p.SerifJut, FootRight: p.SerifJut,
		}
	case "j":
		tail = jStem(p)
		tail.Left = design.NarrowStemLeft(p)
	}
	tail.Left += dx(n - 1)
	// The feet which face each other between two stems are the inner feet of n.
	for k := range fs {
		if k > 0 {
			fs[k].FootLeft = p.SerifJutInner
		}
		if parts[k+1] != "j" {
			fs[k].FootRight = p.SerifJutInner
		}
	}
	if last != "j" {
		tail.FootLeft = p.SerifJutInner
	}
	leftOf := func(k int) float64 {
		if k < n-1 {
			return fs[k].Left
		}
		return tail.Left
	}
	// The f before f or l, which come before the last letter or the f before i
	// or j.
	arched := n - 1
	if dotted >= 0 {
		arched = dotted
	}

	var path outline.Path
	path.MoveTo(fs[0].start())
	// Along the bottom from left to right.
	for k := 0; k < n-1; k++ {
		if k > 0 {
			stemDown(&path, p, fs[k])
		}
		stemUp(&path, p, fs[k], under)
		if joined(k) {
			path.LineTo(leftOf(k+1), under)
			continue
		}
		b := bar
		b.LeftEdge, b.RightEdge = fs[k].Left, fs[k].right()
		barRight(&path, b)
		u := ligaArchUnder(hp, fs[k].right(), leftOf(k+1))
		path.LineTo(u.start().X, u.start().Y)
		u.appendTo(&path)
	}
	// The last letter, up to its top, and the edge along its top from right
	// to left down to the left-hand edge of its stem.
	var top chain
	switch last {
	case "f":
		stemDown(&path, p, tail)
		stemUp(&path, p, tail, under)
		b := fBar(p)
		b.LeftEdge, b.RightEdge = tail.Left, tail.right()
		barRight(&path, b)
		top = hp.chain.shifted(dx(n-1), 0)
	case "l":
		stemDown(&path, p, tail)
		stemUp(&path, p, tail, tail.Top)
		foot := hp.tail.end().Y
		top = chain{
			lineSeg(outline.Point{X: tail.right(), Y: tail.Top}, outline.Point{X: tail.Left, Y: tail.Top}),
			lineSeg(outline.Point{X: tail.Left, Y: tail.Top}, outline.Point{X: tail.Left, Y: foot}),
		}
	case "i":
		stemDown(&path, p, tail)
		stemUp(&path, p, tail, tail.Top)
	case "j":
		t := jTail(p).shifted(tail.Left-p.JStemLeft, 0)
		path.LineTo(t.start().X, t.start().Y)
		t.appendTo(&path)
		path.LineTo(tail.right(), tail.Top)
	}
	if dotted >= 0 {
		path.LineTo(fs[dotted].right(), p.XHeight)
		top = fDotHook(p, hp, design.LigatureCell(p, n-1)-dx(dotted)+p.Stem/2+design.DotOffset(p)).shifted(dx(dotted), 0)
	}
	// An f before f ends in a small free hook, and an f before l runs over
	// its crest into the top of l.
	for k := arched - 1; k >= 0; k-- {
		if parts[k+1] == "f" {
			top = freeHookBefore(p, dx(k), top, leftOf(k+1), fs[k].right(), bar.Top)
			continue
		}
		top = archInto(hp, dx(k), top)
	}
	if lastPoint(&path) != top.start() {
		path.LineTo(top.start().X, top.start().Y)
	}
	top.appendTo(&path)
	path.LineTo(fs[0].Left, bar.Top)
	barLeft(&path, bar)
	stemDown(&path, p, fs[0])

	return path, design.LigatureAdvance(p, n, last)
}

// archInto returns the top of an arched f, its stem moved by dx from the
// stem of f, followed by next, the edge along the top of the next letter
// walked from right to left.  The outer edge of the arch is the outer edge of
// the hook of f from the stem to the crest, and beyond the crest its mirror
// image in the vertical through the crest; it runs down into the next letter
// and meets its top in a corner.  The chain is next up to the corner, the
// outer edge of the arch from the corner to the crest, one curve with the
// directions and the curvatures of the mirror image at its ends (g2Cubic), and
// the edge down to the left-hand edge of the stem.
func archInto(hp fHookParts, dx float64, next chain) chain {
	left := hp.tail.shifted(dx, 0)
	right := left.mirrorX(left.start().X)
	for i, a := range next {
		for j, b := range right {
			cs := pieceCrossings(a, b)
			if len(cs) == 0 {
				continue
			}
			ta, tb := cs[0][0], cs[0][1]
			var out chain
			out = append(out, next[:i]...)
			out = append(out, cutPiece(a, 0, ta))
			// From the corner to the crest the edge is one curve with the
			// directions and the curvatures of the mirror image at its ends.
			c := right[j].cubic
			crest := left[0].cubic
			arc, err := g2Cubic(c.at(tb), c.tangent(tb), c.curvature(tb), crest.P0, crest.tangent(0), crest.curvature(0))
			if err != nil {
				panic("glyphs: a ligature of f: the outer edge of the arch: " + err.Error())
			}
			out = append(out, curveSeg(arc))
			return join(out, left)
		}
	}
	panic("glyphs: a ligature of f: the arch does not reach the next letter")
}

// cutPiece returns the part of the piece from the parameter t0 to t1.
func cutPiece(pc piece, t0, t1 float64) piece {
	if pc.line {
		return lineSeg(lerp(pc.P0, pc.P3, t0), lerp(pc.P0, pc.P3, t1))
	}
	return curveSeg(pc.sub(t0, t1))
}

// ligaArchUnder returns the underside of the arch from the right-hand edge of a
// stem at x0 to the left-hand edge of the next stem at x1: the underside of
// the hook of f from the stem up to its highest point, and from there the
// fairest curve (fairStart) down to the next stem, aimed at the height of the
// foot of the outer edge of f, which it reaches in a transition curve over the
// last stemRamp degrees of its turn (straightFoot).
func ligaArchUnder(hp fHookParts, x0, x1 float64) chain {
	left := hp.under.shifted(x0-hp.under.start().X, 0)
	last := left[len(left)-1].cubic
	foot := outline.Point{X: x1, Y: hp.tail.end().Y}
	down := outline.Point{Y: -1}
	c, err := fairStart(left.end(), left.endTangent(), last.curvature(1), foot, down)
	if err != nil {
		panic("glyphs: a ligature of f: the underside of the arch: " + err.Error())
	}
	v := tTurned(c, 1, stemRamp)
	ramp := straightFoot(c.at(v), c.tangent(v), c.curvature(v), foot, down)
	return join(left, splitAtExtrema(c.sub(0, v)), chain{curveSeg(ramp)})
}

// fDotHook returns the hook of f whose ball has its centre at the abscissa x,
// with the stem in the cell of f.  The reach is found by two steps from the
// ball of f, since the rightmost ink of the ball moves a little with the turn
// of the ball.  The underside runs to the tip of the wedge as windHook draws
// it, and leaves the stem without curvature.  The outer edge from the stem to
// the crest is that of f enlarged (dotShoulder), and from the crest it winds
// into the ball on the spiral (windBall).
func fDotHook(p design.Params, hp fHookParts, x float64) chain {
	f, _ := dotHookCache.LoadOrStore(dotHookKey{p, x}, sync.OnceValue(func() chain { return drawDotHook(p, hp, x) }))
	return f.(func() chain)()
}

// dotHookKey and dotHookCache hold the hook of fDotHook by the parameters
// and the abscissa of the ball, so that fi, ffi and f_j draw it once.
type dotHookKey struct {
	p design.Params
	x float64
}

var dotHookCache sync.Map // dotHookKey -> func() chain

// drawDotHook draws the hook of fDotHook.
func drawDotHook(p design.Params, hp fHookParts, x float64) chain {
	q := p
	tb := hp.tb
	q.FReach += x - tb.Centre.X
	tb = fTerminal(q)
	q.FReach += x - tb.Centre.X
	tb = fTerminal(q)
	st := dotStretch(hp, tb)
	in, _ := fHookFeet(q)
	in.Y = hp.under.end().Y - st*(hp.under.end().Y-in.Y)
	tip, arc := innerEnd(tb)
	a, err := edgeIn(stalkFoot{At: in, Dir: dirOf(math.Pi / 2), Kappa: math.NaN()}, tip)
	if err != nil {
		panic("glyphs: a ligature of f: the underside of the hook: " + err.Error())
	}
	u := tTurned(a, 0, dotRamp)
	rampIn := straightFoot(a.at(u), neg(a.tangent(u)), -a.curvature(u), in, outline.Point{Y: -1}).reversed()
	shoulder, kc := dotShoulder(hp, st, p.XHeight)
	wound := windBallEven(tb, shoulder.start(), kc)
	return join(chain{curveSeg(rampIn)}, splitAtExtrema(a.sub(u, 1)), arc, wound, shoulder)
}

// windBallEven returns the circle of the ball of tb, counter-clockwise from
// the tip of the wedge, and the spiral (windSpiral) which leaves q0 to the
// right with the curvature k0 and winds into it, walked back out to q0, as
// windBall draws them, with the spiral cut into curves which turn by at most
// spiralPieceTurn each.
func windBallEven(tb placedBall, q0 outline.Point, k0 float64) chain {
	w, err := newWindSpiral(q0, 0, k0, tb.Centre, tb.Part.R)
	if err != nil {
		panic("glyphs: a ligature of f: " + err.Error())
	}
	end := w.heading(w.L)
	n := int(math.Ceil(-end / (spiralPieceTurn * math.Pi / 180)))
	var wind chain
	for i := n; i > 0; i-- {
		a := w.lengthAt(end * float64(i) / float64(n))
		b := w.lengthAt(end * float64(i-1) / float64(n))
		if i == n {
			a = w.L
		}
		if i == 1 {
			b = 0
		}
		c, err := g2Cubic(w.point(a), dirOf(w.heading(a)+math.Pi), w.curvature(a), w.point(b), dirOf(w.heading(b)+math.Pi), w.curvature(b))
		if err != nil {
			panic("glyphs: a ligature of f: the spiral into the ball: " + err.Error())
		}
		wind = append(wind, curveSeg(c))
	}
	aW := math.Atan2(tb.Tip.Y-tb.Centre.Y, tb.Tip.X-tb.Centre.X)
	aA := end + math.Pi/2
	for aA <= aW {
		aA += 2 * math.Pi
	}
	for aA > aW+2*math.Pi {
		aA -= 2 * math.Pi
	}
	ball := circleArc(tb.Centre, tb.Part.R, aW, aA, 0)
	ball[len(ball)-1].P3 = wind.start()
	return join(ball, wind)
}

// spiralPieceTurn is the largest turn (degrees) of one curve of the spiral of
// windBallEven.
const spiralPieceTurn = 30.0

// dotShoulder returns the outer edge of the hook of an f before i or j from
// the crest to the stem, and its curvature at the crest.  The hook of f is
// enlarged by the factor s by which the rightmost ink of the ball tb lies
// further from the left-hand edge of the stem than that of the ball of f, with
// the left-hand edge of the stem and the height of the crest kept: the crest
// and the foot of the edge on the stem are those of the enlarged f, the foot
// at least halfway from the x-height (xh) to the foot of f.  The
// curvature at the crest is that of f divided by the square of s, and the edge
// is the fairest curve from the crest to the foot (fairStart), which reaches
// the stem in a transition curve over the last stemRamp degrees of its turn.
func dotShoulder(hp fHookParts, s, xh float64) (chain, float64) {
	foot := hp.tail.end()
	crest := hp.tail.start()
	crest.X = foot.X + s*(crest.X-foot.X)
	foot.Y = math.Max(crest.Y-s*(crest.Y-foot.Y), xh+(foot.Y-xh)/2)
	kc := hp.tail[0].curvature(0) / (s * s)
	down := outline.Point{Y: -1}
	b, err := fairStart(crest, outline.Point{X: -1}, kc, foot, down)
	if err != nil {
		panic("glyphs: a ligature of f: the outer edge of the hook: " + err.Error())
	}
	v := tTurned(b, 1, stemRamp)
	ramp := straightFoot(b.at(v), b.tangent(v), b.curvature(v), foot, down)
	return join(splitAtExtrema(b.sub(0, v)), chain{curveSeg(ramp)}), kc
}

// dotStretch returns the factor by which the rightmost ink of the ball tb
// lies further from the left-hand edge of the stem than that of the ball of f.
func dotStretch(hp fHookParts, tb placedBall) float64 {
	xl := hp.tail.end().X
	return (tb.Centre.X + tb.Part.R - xl) / (hp.tb.Centre.X + hp.tb.Part.R - xl)
}

// dotRamp is the turn (degrees) over which the underside of the hook of an f
// before i or j leaves the stem in a transition curve (straightFoot); the
// underside is flatter than that of f, and the turn is larger than stemRamp so
// that the transition is not short.
const dotRamp = 30.0

// splitAtTop splits the chain at its highest point: the part before it and
// the part after it.
func splitAtTop(c chain) (head, tail chain) {
	best, bi, bt := math.Inf(-1), 0, 0.0
	for i, pc := range c {
		ts := []float64{0, 1}
		if !pc.line {
			ts = append(ts, outline.CubicExtrema(pc.P0.Y, pc.P1.Y, pc.P2.Y, pc.P3.Y)...)
			for k := 1; k < 64; k++ {
				ts = append(ts, float64(k)/64)
			}
		}
		for _, t := range ts {
			if t < 0 || t > 1 {
				continue
			}
			if y := pc.at(t).Y; y > best {
				best, bi, bt = y, i, t
			}
		}
	}
	// The highest point to within a small fraction of a unit: the curve rises
	// to it and falls after it.
	if pc := c[bi]; !pc.line && bt > 0 && bt < 1 {
		lo, hi := math.Max(0, bt-1.0/64), math.Min(1, bt+1.0/64)
		for range 60 {
			m1, m2 := lo+(hi-lo)/3, hi-(hi-lo)/3
			if pc.at(m1).Y < pc.at(m2).Y {
				lo = m1
			} else {
				hi = m2
			}
		}
		bt = (lo + hi) / 2
	}
	head = append(head, c[:bi]...)
	switch {
	case bt <= 0:
		tail = append(tail, c[bi:]...)
	case bt >= 1:
		head = append(head, c[bi])
		tail = append(tail, c[bi+1:]...)
	default:
		head = append(head, curveSeg(c[bi].split(bt, true)))
		tail = append(tail, curveSeg(c[bi].split(bt, false)))
		tail = append(tail, c[bi+1:]...)
	}
	return head, tail
}

// An fHookParts holds the parts of the hook of f which the ligatures share: the
// ball, the hook, the hook up to its crest and from there on (head and tail),
// and the underside from the stem to its highest point.
type fHookParts struct {
	tb                       placedBall
	chain, head, tail, under chain
}

// fHookCache holds the parts of the hook of f by the parameters, so that f, ſ
// and the ligatures of one set of parameters draw the hook once.
var fHookCache sync.Map // design.Params -> func() fHookParts

// newFHookParts returns the parts of the hook of f.
func newFHookParts(p design.Params) fHookParts {
	f, _ := fHookCache.LoadOrStore(p, sync.OnceValue(func() fHookParts { return drawFHookParts(p) }))
	return f.(func() fHookParts)()
}

// drawFHookParts draws the parts of the hook of f.
func drawFHookParts(p design.Params) fHookParts {
	h := fHookParts{tb: fTerminal(p)}
	h.chain = fHookChain(p, h.tb)
	h.head, h.tail = splitAtTop(h.chain)
	end, near := 0, math.Inf(1)
	for i, pc := range h.chain {
		if d := math.Hypot(pc.P3.X-h.tb.Tip.X, pc.P3.Y-h.tb.Tip.Y); d < near {
			end, near = i, d
		}
	}
	h.under, _ = splitAtTop(h.chain[:end+1])
	return h
}

// freeHookBefore returns the top of an f before another f, its stem moved by
// dx from the stem of f, followed by next, the edge along the top of the next
// letter walked from right to left, which ends on the left-hand edge of its
// stem at xn.  The edge runs down that stem to the top of the bar, along the
// bar to the right-hand edge of the stem of the f at xr, and up into its hook
// (smallHook).
func freeHookBefore(p design.Params, dx float64, next chain, xn, xr, barTop float64) chain {
	h := smallHook(p).shifted(dx, 0)
	a := outline.Point{X: xn, Y: barTop}
	b := outline.Point{X: xr, Y: barTop}
	return join(next, chain{lineSeg(next.end(), a), lineSeg(a, b), lineSeg(b, h.start())}, h)
}

// smallHook returns the hook of an f before another f: the hook of f
// (fHookChain) with the ball and the stalk of the ear of g (GEarBall,
// GEarThick) and the reach of f less smallHookShort.
func smallHook(p design.Params) chain {
	f, _ := smallHookCache.LoadOrStore(p, sync.OnceValue(func() chain { return drawSmallHook(p) }))
	return f.(func() chain)()
}

// smallHookCache holds the hook of an f before another f by the parameters.
var smallHookCache sync.Map // design.Params -> func() chain

// drawSmallHook draws the hook of smallHook.
func drawSmallHook(p design.Params) chain {
	q := p
	q.BallDiameter = p.GEarBall
	q.BallStalk = p.GEarThick
	q.FReach -= smallHookShort
	return fHookChain(q, fTerminal(q))
}

// smallHookShort is how much shorter the reach of the hook of an f before
// another f is than that of f.
const smallHookShort = 40.0
