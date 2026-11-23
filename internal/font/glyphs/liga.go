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
// of two kinds.
//
//	An f before i or j ends in a free hook whose ball takes the place of the
//	dot of i: its centre stands over the stem of the next letter where the
//	centre of the dot of i stands (fDotHook); the dot and the head serif of i
//	and j are left out.
//	An f before f or l is bridged: its hook runs level into the top of the next
//	letter.  The outer edge of its hook runs from the stem towards the crest and
//	into a straight line at the height of the crest, which it meets without
//	curvature (levelEnd); the line runs to the crest of the next f, which it
//	meets in the same way, or to the top of l, which stands at that height
//	without a head serif.
//	The underside is a bridge from the right-hand edge of its stem to the
//	left-hand edge of the next stem (bridgeUnder).
//
// The bar runs on from one stem to the next before f, i and j, and the bar of
// an f before l is the bar of f.  Where a bridged f and the next f are joined by
// the bar, the white between the stems, the bar and the bridge is a counter.
// The last f of ff keeps the hook of f.
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
			Top: hp.tail.start().Y, FootLeft: p.SerifJut, FootRight: p.SerifJut,
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
		u := bridgeUnder(hp.under, fs[k].right(), leftOf(k+1))
		path.LineTo(u.start().X, u.start().Y)
		u.appendTo(&path)
	}
	// The last letter, up to its top.
	switch last {
	case "f":
		stemDown(&path, p, tail)
		stemUp(&path, p, tail, under)
		b := fBar(p)
		b.LeftEdge, b.RightEdge = tail.Left, tail.right()
		barRight(&path, b)
		head := levelEnd(hp.head, 0).shifted(dx(n-1), 0)
		path.LineTo(head.start().X, head.start().Y)
		head.appendTo(&path)
	case "l":
		stemDown(&path, p, tail)
		stemUp(&path, p, tail, tail.Top)
	case "i":
		stemDown(&path, p, tail)
		stemUp(&path, p, tail, tail.Top)
	case "j":
		t := jTail(p).shifted(tail.Left-p.JStemLeft, 0)
		path.LineTo(t.start().X, t.start().Y)
		t.appendTo(&path)
		path.LineTo(tail.right(), tail.Top)
	}
	// Back along the top from right to left.
	if dotted >= 0 {
		path.LineTo(fs[dotted].right(), p.XHeight)
		h := fDotHook(p, hp.tb, design.LigatureCell(p, n-1)-dx(dotted)+p.Stem/2+design.DotOffset(p))
		h = h.shifted(dx(dotted), 0)
		if dotted > 0 {
			h, _ = splitAtTop(h)
			h = levelEnd(h, 0)
		}
		path.LineTo(h.start().X, h.start().Y)
		h.appendTo(&path)
	}
	if dotted != 0 {
		// The line along the crests to the crest of the first f.
		t := levelEnd(hp.tail.reversed(), 0).reversed()
		path.LineTo(t.start().X, t.start().Y)
		t.appendTo(&path)
	}
	path.LineTo(fs[0].Left, bar.Top)
	barLeft(&path, bar)
	stemDown(&path, p, fs[0])

	// The counters under the bridges between two f joined by the bar.
	for k := 0; k+1 < n; k++ {
		if k == dotted || !joined(k) {
			continue
		}
		u := bridgeUnder(hp.under, fs[k].right(), leftOf(k+1))
		path.MoveTo(fs[k].right(), bar.Top)
		path.LineTo(u.start().X, u.start().Y)
		u.appendTo(&path)
		path.LineTo(leftOf(k+1), bar.Top)
	}

	return path, design.LigatureAdvance(p, n, last)
}

// fDotHook returns the hook of f whose ball has its centre at the abscissa x,
// with the stem in the cell of f.  The edges run to the ball as stalkChain
// draws them and leave the stem without curvature.  The reach is found by
// two steps from the ball tb of f, since the rightmost ink of the ball moves a
// little with the turn of the ball.
func fDotHook(p design.Params, tb placedBall, x float64) chain {
	q := p
	q.FReach += x - tb.Centre.X
	tb = fTerminal(q)
	q.FReach += x - tb.Centre.X
	tb = fTerminal(q)
	in, out := fHookFeet(q)
	inFoot := stalkFoot{At: in, Dir: dirOf(math.Pi / 2), Kappa: math.NaN()}
	outFoot := stalkFoot{At: out, Dir: dirOf(-math.Pi / 2), Kappa: math.NaN()}
	a, err := edgeIn(inFoot, tb.In)
	if err != nil {
		panic("glyphs: a ligature of f: the underside of the hook: " + err.Error())
	}
	b, err := edgeOut(tb.Out, outFoot)
	if err != nil {
		panic("glyphs: a ligature of f: the outer edge of the hook: " + err.Error())
	}
	u := tTurned(a, 0, dotRamp)
	rampIn := straightFoot(a.at(u), neg(a.tangent(u)), -a.curvature(u), in, neg(inFoot.Dir)).reversed()
	v := tTurned(b, 1, stemRamp)
	rampOut := straightFoot(b.at(v), b.tangent(v), b.curvature(v), out, outFoot.Dir)
	return join(chain{curveSeg(rampIn)}, splitAtExtrema(a.sub(u, 1)), tb.Chain,
		splitAtExtrema(b.sub(0, v)), chain{curveSeg(rampOut)})
}

// dotRamp is the turn (degrees) over which the underside of the hook of an f
// before i or j leaves the stem in a transition curve (straightFoot); the
// underside is flatter than that of f, and the turn is larger than stemRamp so
// that the transition is not short.
const dotRamp = 30.0

// levelEnd returns the chain, which ends at its highest point, with the end of
// its last piece replaced by a transition curve (straightFoot) which arrives on
// the level line through the highest point without curvature, as the edges of
// the hook arrive on the stem.  The transition covers the last turn degrees of
// the piece, or the whole piece if turn is zero.
func levelEnd(c chain, turn float64) chain {
	last := c[len(c)-1].cubic
	v := 0.0
	if turn > 0 {
		v = tTurned(last, 1, turn)
	}
	level := outline.Point{X: math.Copysign(1, last.tangent(1).X)}
	ramp := straightFoot(last.at(v), last.tangent(v), last.curvature(v), last.P3, level)
	out := append(chain{}, c[:len(c)-1]...)
	if v > 0 {
		out = append(out, curveSeg(last.sub(0, v)))
	}
	return append(out, curveSeg(ramp))
}

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

// bridgeUnder returns the underside of the bridge from the right-hand edge of
// a stem at x0 to the left-hand edge of the next stem at x1: the underside of
// the hook of f from the stem towards its highest point (under), which runs
// into the level of that point without curvature over the last stemRamp
// degrees of its turn (levelEnd), a level line, and the same
// curve mirrored down into the next stem.  Where the two curves would overlap,
// each is cut where they meet midway between the stems.
func bridgeUnder(under chain, x0, x1 float64) chain {
	left := levelEnd(under, stemRamp).shifted(x0-under.start().X, 0)
	xm := (x0 + x1) / 2
	if left.end().X > xm {
		var cut chain
		for _, pc := range left {
			if pc.P3.X <= xm {
				cut = append(cut, pc)
				continue
			}
			if pc.line {
				q := lerp(pc.P0, pc.P3, (xm-pc.P0.X)/(pc.P3.X-pc.P0.X))
				cut = append(cut, lineSeg(pc.P0, q))
			} else {
				cut = append(cut, curveSeg(pc.split(pc.tAtX(xm), true)))
			}
			break
		}
		left = cut
	}
	right := left.mirrorX(xm)
	if right.start().X-left.end().X < 1e-6 {
		return join(left, right)
	}
	return join(left, chain{lineSeg(left.end(), right.start())}, right)
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
