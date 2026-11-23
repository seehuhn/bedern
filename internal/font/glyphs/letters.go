// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The letters þ, Þ, ſ, ĸ, ŋ, Ŋ, æ, Æ, œ, Œ, ĳ, Ĳ, ŀ, Ŀ and ŉ, made of the parts
// of the letters of the basic alphabet.

// Shape constants of these letters.
const (
	thornBowlBottom = 0.252 // the outer edge of the lower stroke of the bowl of Þ, in cap heights
	thornBowlTop    = 0.80  // the outer edge of the upper stroke of the bowl of Þ, in cap heights
	engDepth        = 0.9   // the lowest point of the hook of Ŋ, in descenders
	joinOverlap     = 6.0   // how far the left extreme of e in œ lies right of the inner edge of the o
	aeOverlap       = 5.0   // how far the left extreme of e in æ lies right of the left-hand edge of the stem of a
	capAEArm        = 0.235 // how far the top arm of Æ reaches left of the stem, in cap heights
	ijShift         = -20.0 // how much closer the j of ĳ stands to the i than in ij
	ldotHeight      = 0.79  // the middle of the dot of ŀ, in x-heights
	capLdotHeight   = 0.36  // the middle of the dot of Ŀ, in cap heights
	ldotGap         = 52.0  // the white between the stem of ŀ and Ŀ and the dot
	napGap          = 40.0  // the least distance between the apostrophe of ŉ and the n
)

// thorn draws þ: the p with its stem carried up to the ascender height, where
// it has the head serif of l.
func thorn(p design.Params) (outline.Path, float64) {
	stem := descenderStem(p, design.PStemLeft(p), p.Ascender, true)
	bowl := pBowl(p)
	var path outline.Path
	path.MoveTo(stem.start())
	stemUp(&path, p, stem, bowl.NotchLower)
	bowlOuter(dBowlShape, bowl).appendTo(&path)
	stemCap(&path, p, stem)
	stemDown(&path, p, stem)
	bowlCounter(dBowlShape, bowl).addContour(&path)
	return path, design.PAdvance(p)
}

// capThorn draws Þ: the stem of I with all four serifs and the bowl of P
// between thornBowlBottom and thornBowlTop of the cap height.
func capThorn(p design.Params) (outline.Path, float64) {
	c := p.Caps
	j := c.SerifJut
	s := capStem{Left: design.CapBowlStemLeft(p), FootLeft: j, FootRight: j, HeadLeft: j, HeadRight: j}
	b := capBowlFrom(p, s.right(p), c.PWidth, c.BowlFlat, thornBowlBottom*p.CapHeight, thornBowlTop*p.CapHeight)
	up := s.up(p, s.start(), b.Bottom)
	ch := join(up, b.outer())
	top := s.top(p, ch.end(), p.CapHeight/2)
	path := capContour(join(ch, top, s.down(p, top.end())))
	return append(path, capContour(b.counter())...), design.CapPAdvance(p)
}

// longs draws ſ: the f without its crossbar.
func longs(p design.Params) (outline.Path, float64) {
	stem := fStem(p)
	h := newFHookParts(p).chain
	var path outline.Path
	path.MoveTo(stem.start())
	stemUp(&path, p, stem, h.start().Y)
	h.appendTo(&path)
	stemDown(&path, p, stem)
	return path, design.FAdvance(p)
}

// kgreenlandic draws ĸ: the k with its stem cut at the x-height, where it has
// the head serif of n.
func kgreenlandic(p design.Params) (outline.Path, float64) {
	stem := kStemSpec(p)
	stem.Top = p.XHeight
	return kOnStem(p, stem), design.KAdvance(p)
}

// eng draws ŋ: the n whose right stem has no foot serifs and runs on below the
// baseline into the tail of j.
func eng(p design.Params) (outline.Path, float64) {
	left := stemSpec{
		Left: p.SideStemLeft, Width: p.Stem, Top: p.XHeight,
		FootLeft: p.SerifJut, FootRight: p.SerifJutInner,
		HeadJut: p.SerifJut, HeadDrop: p.HeadSerifDrop,
	}
	right := stemSpec{Left: left.Left + p.Stem + p.NCounter, Width: p.Stem}
	arch := archBetween(p, left, right)
	dx := right.Left - jStem(p).Left
	tail := jTail(p).mapPoints(func(q outline.Point) outline.Point {
		return outline.Point{X: q.X + dx, Y: q.Y}
	})
	var path outline.Path
	path.MoveTo(left.start())
	stemUp(&path, p, left, arch.Join)
	archUnder(nArchShape, arch).appendTo(&path)
	path.LineTo(tail.start().X, tail.start().Y)
	tail.appendTo(&path)
	path.LineTo(right.right(), arch.Shoulder)
	archOuter(nArchShape, arch).appendTo(&path)
	stemCap(&path, p, left)
	stemDown(&path, p, left)
	return path, design.NAdvance(p)
}

// capEng draws Ŋ: the N whose diagonal runs at the foot into a capital stem
// on the right, flush with the right-hand edge of the right stem, which runs
// on below the baseline and curves to the left into the tail of J, its lowest
// point engDepth of the descender below the baseline.
func capEng(p design.Params) (outline.Path, float64) {
	l, r := capNStems(p)
	xr1 := r.rightEdge()
	diag, _ := capNDiagonal(p)
	hook := capStem{Left: xr1 - p.Caps.Stem}
	low := -engDepth * p.Descender
	g := p.Caps.CGSJQ
	in := footRising(outline.Point{X: hook.Left, Y: low + g.JInnerDrop}, -math.Pi/2)
	out := footAt(outline.Point{X: xr1, Y: low + g.JOuterDrop}, math.Pi/2)
	reach := design.CapJStemLeft(p) - g.JSideLeft
	tb := must(placeAtExtreme(capJBall(p), ballPlace{Angle: capJBallAngle},
		func(x0, _ float64) float64 { return hook.Left - reach - x0 }, low, false,
		func(tb placedBall) chain { return stalkChain("Eng", in, tb, out) }))
	tail := stalkChain("Eng", in, tb, out)
	// The lower edge of the diagonal meets the left-hand edge of the hook.
	meet := diag.edgeAt(edgeLeft, diag.Y+(hook.Left-diag.edgeAt(edgeLeft, diag.Y).X)/diag.Lean)
	_, crotch := capNDiagonal(p)
	topRight := diag.edgeAt(edgeRight, p.CapHeight)
	up1 := l.up(l.start(), p.CapHeight-p.CapDiag.NJoin)
	up2 := r.top(outline.Point{X: xr1, Y: tail.end().Y}, crotch.Y)
	top1 := l.top(topRight, l.bracketHeight())
	c := join(
		up1,
		lineChain(up1.end(), meet),
		lineChain(meet, tail.start()),
		tail,
		up2,
		lineChain(up2.end(), topRight),
		top1,
		l.down(top1.end()),
	)
	return capContour(c), design.CapNAdvance(p)
}

// oe draws œ: the o and the e, the left stroke of e standing on the right
// stroke of o, its left extreme joinOverlap right of the inner edge of o.
func oe(p design.Params) (outline.Path, float64) {
	dx := math.Round(design.OBowlRight(p) - p.CurveLower + joinOverlap - design.EBowlLeft(p))
	return joinedPair(p, "o", "e", dx)
}

// ae draws æ: the a without its tail and the e, the left extreme of e
// aeOverlap right of the left-hand edge of the stem of a.  The stem of a ends
// at the bottom in a straight cut from the lower notch of its bowl up to the
// right, aeCut above the notch, which lies inside the left stroke of e, so that
// the bowl of a and the bowl of e meet below it in a crotch.
func ae(p design.Params) (outline.Path, float64) {
	dx := math.Round(design.AStemLeft(p) + aeOverlap - design.EBowlLeft(p))
	pb, adv, _ := Outline("e", p)
	parts := append(pathChains(aWithoutTail(p)), pathChains(shiftInk(pb, dx, 0))...)
	return chainsPath(union(parts...)...), dx + adv
}

// aeCut is the rise of the cut at the foot of the stem of a in æ.
const aeCut = 60.0

// aWithoutTail draws the a of æ: aOutline with the tail replaced by the cut
// from the lower notch of the bowl to the right-hand edge of the stem.
func aWithoutTail(p design.Params) outline.Path {
	stem := stemSpec{Left: design.AStemLeft(p), Width: p.Stem}
	bowl := aBowl(p)
	right := stem.right()
	notch := aTailChain(p, bowl, stem).start()
	var path outline.Path
	path.MoveTo(notch.X, notch.Y)
	path.LineTo(right, notch.Y+aeCut)
	path.LineTo(right, p.ShoulderStart-aOuterDrop)
	aArcChain(p, aTerminal(p)).appendTo(&path)
	path.LineTo(stem.Left, bowl.NotchUpper)
	aBowlOuter(bowl).appendTo(&path)
	aBowlCounter(bowl).addContour(&path)
	return path
}

// joinedPair returns the union of the glyph a and the glyph b moved by dx, and
// the advance of b moved by dx.
func joinedPair(p design.Params, a, b string, dx float64) (outline.Path, float64) {
	pa, _, _ := Outline(a, p)
	pb, adv, _ := Outline(b, p)
	parts := append(pathChains(pa), pathChains(shiftInk(pb, dx, 0))...)
	return chainsPath(union(parts...)...), dx + adv
}

// capOE draws Œ: the left part of O, cut inside the stem of E, and the E
// without its left serifs.
func capOE(p design.Params) (outline.Path, float64) {
	c := p.Caps
	left, bottom, top := c.SideRoundLeft, -p.OvershootUpper, p.CapHeight+p.OvershootUpper
	right := left + c.OWidth
	outer := newSuperShape(left, bottom, right, top, p.SupernessOuter)
	// The stem of E stands where the outer edge of O is half the overshoot
	// above the cap height, so that the corners of the stem lie inside the
	// round.
	o0 := outer.cubic(0, true)
	s := capStem{Left: o0.at(o0.tAtY(p.CapHeight + p.OvershootUpper/2)).X}
	xs := s.Left + c.Stem/2
	inner := newSuperShape(left+c.Round, bottom+c.RoundThin, right-c.Round, top-c.RoundThin, p.SupernessInner)
	o3 := outer.cubic(3, true)
	i0, i3 := inner.cubic(0, false), inner.cubic(3, false)
	toTop, fromBottom := o0.sub(o0.tAtX(xs), 1), o3.sub(0, o3.tAtX(xs))
	toBottom, fromTop := i0.sub(i0.tAtX(xs), 1), i3.sub(0, i3.tAtX(xs))
	round := join(
		chain{curveSeg(toTop), curveSeg(outer.cubic(1, true)), curveSeg(outer.cubic(2, true)), curveSeg(fromBottom)},
		lineChain(fromBottom.P3, toBottom.P0),
		chain{curveSeg(toBottom), curveSeg(inner.cubic(1, false)), curveSeg(inner.cubic(2, false)), curveSeg(fromTop)},
		lineChain(fromTop.P3, toTop.P0),
	)
	e := pathChains(capEOnStem(p, s))
	path := chainsPath(union(append(e, round)...)...)
	return path, math.Round(s.Left + design.CapEAdvance(p) - design.CapStemLeft(p))
}

// capAE draws Æ: the thin diagonal of A, cut at the cap height, and the E,
// whose top arm runs on to the left into the diagonal and whose middle arm
// runs on as the crossbar of A.  The stem of E stands capAEArm of the cap
// height right of the outer edge of the diagonal at the cap height, and has a
// foot serif on the left.
func capAE(p design.Params) (outline.Path, float64) {
	c := p.Caps
	l, _ := capADiagonals(p)
	topL := l.edgeAt(edgeLeft, p.CapHeight)
	s := capStem{Left: topL.X + capAEArm*p.CapHeight, FootLeft: c.SerifJut}
	r := s.right(p)
	bottom := capArmSpec(p, r, r+c.EBottomArm, c.Bar)
	bottom.Rise = c.BottomRise
	middle := capArmSpec(p, r, r+c.EMiddleArm, c.EMiddleHeight+c.Bar/2)
	middle.Rise, middle.Drop = c.MiddleReach, c.MiddleReach
	topArm := capArmSpec(p, r, r+c.ETopArm, p.CapHeight)
	topArm.Drop = c.TopDrop
	b, m, t := capArmBottom(bottom), capArmMiddle(middle), capArmTop(topArm)
	_, thin := capDiagSerifs(p)
	outL := l.footServif(edgeLeft, 0, thin)
	inL := l.footServif(edgeRight, 0, thin).reversed()
	lower, upper := c.EMiddleHeight-c.Bar/2, c.EMiddleHeight+c.Bar/2
	p1 := l.edgeAt(edgeRight, lower)
	pt := func(x, y float64) outline.Point { return outline.Point{X: x, Y: y} }
	ch := join(
		lineChain(s.start(), b.start()), b,
		lineChain(b.end(), m.start()), m,
		lineChain(m.end(), t.start()), t,
		lineChain(t.end(), topL),
		lineChain(topL, outL.start()), outL,
		lineChain(outL.end(), inL.start()), inL,
		lineChain(inL.end(), p1),
		lineChain(p1, pt(s.Left, lower)),
	)
	path := capContour(join(ch, s.down(p, ch.end())))
	q1, q2 := l.edgeAt(edgeRight, upper), l.edgeAt(edgeRight, p.CapHeight-c.Bar)
	counter := join(lineChain(q1, q2), lineChain(q2, pt(s.Left, q2.Y)), lineChain(pt(s.Left, q2.Y), pt(s.Left, upper)), lineChain(pt(s.Left, upper), q1))
	path = append(path, capContour(counter)...)
	return path, math.Round(r + design.CapEAdvance(p) - design.CapStemLeft(p) - c.Stem)
}

// ij draws ĳ: the i and the j, the j ijShift closer than in ij.
func ij(p design.Params) (outline.Path, float64) {
	return pairOf(p, "i", "j", ijShift)
}

// capIJ draws Ĳ: the I and the J, as in IJ.
func capIJ(p design.Params) (outline.Path, float64) {
	return pairOf(p, "I", "J", 0)
}

// pairOf returns the glyph a followed by the glyph b, moved by gap beyond the
// advance of a.
func pairOf(p design.Params, a, b string, gap float64) (outline.Path, float64) {
	pa, adva, _ := Outline(a, p)
	pb, advb, _ := Outline(b, p)
	return append(pa, shiftInk(pb, adva+gap, 0)...), adva + gap + advb
}

// ldot draws ŀ: the l with the dot of the middle dot at ldotHeight of the
// x-height, ldotGap right of the stem.  The advance is the one at which the
// dot stands as far from the stem of a following l, so that it stands in the
// middle between the two l of l·l.
func ldot(p design.Params) (outline.Path, float64) {
	path, _, _ := Outline("l", p)
	right := design.NarrowStemLeft(p) + design.AscenderInset(p) + p.AscenderStem
	r := p.DotDiameter / 2
	cx := right + ldotGap + r
	adv := math.Round(cx + r + ldotGap - design.NarrowStemLeft(p))
	dotChain(p, cx, ldotHeight*p.XHeight).addContour(&path)
	return path, adv
}

// capLdot draws Ŀ: the L with the dot of the middle dot at capLdotHeight of
// the cap height, ldotGap right of the stem.
func capLdot(p design.Params) (outline.Path, float64) {
	path, adv, _ := Outline("L", p)
	r := p.DotDiameter / 2
	cx := design.CapStemLeft(p) + p.Caps.Stem + ldotGap + r
	dotChain(p, cx, capLdotHeight*p.CapHeight).addContour(&path)
	return path, adv
}

// Shape constants of ð.
const (
	ethTopX     = 0.30  // the end of the arm of ð, in widths of the bowl from its left extreme
	ethMidX     = 0.25  // the middle of the arm of ð where the bar crosses it, in widths of the bowl from its right extreme
	ethBarAngle = 25.0  // the bar of ð, degrees above the horizontal
	ethBarY     = 0.465 // the height of the middle of the bar of ð, where it crosses the arm, in the space between the top of the bowl and the ascender
	ethBarLen   = 0.42  // the horizontal extent of the bar of ð, in widths of the bowl
)

// eth draws ð: a bowl, the o, whose outer edge runs on from its right
// extreme into an arm, which rises, bends over to the left, thinning, and ends
// in a cut at the ascender height; the inner edge of the arm comes down into
// the outer edge of the bowl above its right extreme.  A bar rises to the
// right across the arm.  The arm is a stroke (symstroke.go) which leaves the
// right extreme with the thickness CurveLower, so that its outer edge leaves the
// outer edge of the bowl there.
func eth(p design.Params) (outline.Path, float64) {
	l, r := design.OBowlLeft(p), design.OBowlRight(p)
	wb := r - l
	cl := p.CurveLower
	bottom, top := -p.OvershootLower, p.XHeight+p.OvershootLower
	sts := []station{
		{X: r - cl/2, Y: (bottom + top) / 2, Dir: 90, W: cl},
		{X: r - ethMidX*wb, Y: top + ethBarY*(p.Ascender-top), Dir: 125, W: 0.8 * cl},
		{X: l + ethTopX*wb, Y: p.Ascender - p.ThinLower/2, Dir: 170, W: 1.2 * p.ThinLower},
	}
	right, left := strokeOpen(sts)
	outer := newSuperShape(l, bottom, r, top, p.SupernessOuter)
	o0 := outer.cubic(0, true) // right extreme to top
	// The inner edge of the arm meets the outer edge of the bowl.
	var down chain
	var tb float64
	found := false
	for _, pc := range left {
		if cs := pieceCrossings(pc, curveSeg(o0)); len(cs) > 0 {
			down = append(down, curveSeg(pc.cubic.sub(0, cs[0][0])))
			tb, found = cs[0][1], true
			break
		}
		down = append(down, pc)
	}
	if !found {
		panic("glyphs: glyph eth: the arm does not meet the bowl; glyphs.Check reports such parameters")
	}
	rest := o0.sub(tb, 1)
	rest.P0 = down.end()
	bowl := join(
		right, lineChain(right.end(), left.start()), down,
		chain{curveSeg(rest), curveSeg(outer.cubic(1, true)), curveSeg(outer.cubic(2, true)), curveSeg(outer.cubic(3, true))},
	)
	counter := ellipseChain(l+cl, bottom+p.ThinLower, r-cl, top-p.ThinLower, p.SupernessInner, false)
	bar := slantBar(sts[1].X, sts[1].Y, ethBarLen*wb, ethBarAngle, p.ThinLower)
	return chainsPath(union(bowl, counter, bar)...), design.OAdvance(p)
}

// napostrophe draws ŉ: the apostrophe, with its leftmost ink where that of n
// is, and the n moved to the right by the least whole number of units which
// keeps napGap between the two.
func napostrophe(p design.Params) (outline.Path, float64) {
	q, _, _ := Outline("quoteright", p)
	x0, _, _, _ := q.BBox()
	q = shiftInk(q, p.SideStemLeft-p.SerifJut-x0, 0)
	n, adv, _ := Outline("n", p)
	lo, hi := 0.0, 400.0
	for hi-lo > 0.5 {
		mid := (lo + hi) / 2
		if pathsDistance(q, shiftInk(n, mid, 0), napGap) < napGap {
			lo = mid
		} else {
			hi = mid
		}
	}
	d := math.Ceil(hi)
	return append(q, shiftInk(n, d, 0)...), adv + d
}
