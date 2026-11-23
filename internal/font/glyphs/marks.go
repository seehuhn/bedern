// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"fmt"
	"math"
	"sync"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The accent marks.
//
// Each mark is drawn once, in a frame of its own whose origin is its anchor:
// a mark above a letter has its lowest ink at the height 0 and the middle of
// its ink box at the abscissa 0, unless a shift is given; a mark below a
// letter hangs from the origin.  The composed letters (compose.go) set the
// anchor of the mark on the anchor of the base, and the spacing and combining
// accents stand the mark above the x-height.  A mark has a lowercase form and
// a capital form, which is flatter and stands above the cap height and the
// ascender height; the marks whose capital form is not given use the
// lowercase form on the capitals.
//
// The acute and the grave are wedges from a round head, at the top, to a round
// tip (the outline of two circles joined by lines which run into both on lead
// spirals, as the ASCII grave).  The circumflex and the caron are two straight
// strokes of different weight which meet in a point.  The tilde is a stroke of
// the symbols (symstroke.go) thick where it falls and thin at the crest and the
// trough.  The diaeresis and the dot above are dots of the shape of the dot of
// i.  The macron is a bar.  The breve is a crescent, thick at the bottom and
// thin at the horns.  The ring is a ring with the stress of o.  The double acute
// is two wedges of the acute.  The cedilla and the ogonek are strokes which
// hang from the base and are joined to it where they meet it.  The comma below
// is the comma scaled, and the caron of d, l, t and L the right single
// quotation mark scaled.

// A markKind names a mark.
type markKind int

const (
	markGrave markKind = iota
	markAcute
	markCircumflex
	markCaron
	markTilde
	markDieresis
	markMacron
	markBreve
	markRing
	markDot
	markDblAcute
	markCedilla
	markOgonek
	markCommaBelow
	markCommaAbove
	markApostrophe
)

// below says whether the mark hangs below its base.
func (k markKind) below() bool {
	return k == markCedilla || k == markOgonek || k == markCommaBelow
}

// joinable says whether the mark may be joined to the base it hangs from.
func (k markKind) joinable() bool { return k == markCedilla || k == markOgonek }

// A markForm is the form of a mark: for the lowercase, for the capitals and
// the ascenders, or the lowercase form for the dotless i and j.
type markForm int

const (
	formLower markForm = iota
	formCap
	formNarrow
)

// markOutline returns the mark in its frame.  The marks are built once for
// each parameter set and shared: the path must not be changed.
func markOutline(p design.Params, k markKind, f markForm) outline.Path {
	key := markCacheKey{p.Acc, p.DotDiameter, p.DotSuperness, p.Punct, k, f}
	if m, ok := markCache.Load(key); ok {
		return m.(outline.Path)
	}
	m, _ := markCache.LoadOrStore(key, chainsPath(markChains(p, k, f)...))
	return m.(outline.Path)
}

// markCacheKey holds what a mark depends on: the parameters of the accents,
// of the dot of i and of the comma.
type markCacheKey struct {
	acc      design.Accents
	dot, sup float64
	punct    design.Punct
	kind     markKind
	form     markForm
}

var markCache sync.Map

// markChains returns the closed contours of the mark in its frame.
func markChains(p design.Params, k markKind, f markForm) []chain {
	a := p.Acc
	cp := f == formCap
	pick := func(lower, capital float64) float64 {
		if cp {
			return capital
		}
		return lower
	}
	switch k {
	case markAcute:
		return []chain{wedgeMark(a.AcuteShift, a.AcuteWidth, pick(a.AcuteHeight, a.AcuteHeightCap), a.AcuteHead, a.AcuteTip, true)}
	case markGrave:
		return []chain{wedgeMark(-a.AcuteShift, a.AcuteWidth, pick(a.AcuteHeight, a.AcuteHeightCap), a.AcuteHead, a.AcuteTip, false)}
	case markDblAcute:
		h := pick(a.DblHeight, a.DblHeightCap)
		return []chain{
			wedgeMark(-a.DblPitch/2, a.DblWidth, h, a.DblHead, a.DblTip, true),
			wedgeMark(a.DblPitch/2, a.DblWidth, h, a.DblHead, a.DblTip, true),
		}
	case markCircumflex:
		return []chain{circumflexChain(a.CircWidth, pick(a.CircHeight, a.CircHeightCap), a.CircThin, a.CircThick, false)}
	case markCaron:
		return []chain{circumflexChain(a.CircWidth, pick(a.CircHeight, a.CircHeightCap), a.CircThin, a.CircThick, true)}
	case markTilde:
		return []chain{tildeChain(a.TildeWidth, pick(a.TildeHeight, a.TildeHeightCap), a.TildeThick, a.TildeThin)}
	case markDieresis:
		d := a.DotScale * p.DotDiameter
		pitch := a.DiaPitch
		if f == formNarrow {
			pitch = a.DiaPitchNarrow
		}
		return []chain{markDotChain(p, -pitch/2, d/2, d), markDotChain(p, pitch/2, d/2, d)}
	case markDot:
		d := a.DotAboveScale * p.DotDiameter
		return []chain{markDotChain(p, 0, d/2, d)}
	case markMacron:
		w := pick(a.MacronWidth, a.MacronWidthCap)
		var path outline.Path
		addBar(&path, -w/2, a.MacronLift, w/2, a.MacronLift+a.MacronThick)
		return pathChains(path)
	case markBreve:
		return []chain{breveChain(a.BreveWidth, pick(a.BreveHeight, a.BreveHeightCap), a.BreveThick, a.BreveThin)}
	case markRing:
		d := pick(a.RingDiam, a.RingDiamCap)
		outer, counter := ringChains(-d/2, 0, d/2, d, a.RingThick, a.RingThin, degSuperness, degSuperness+0.01)
		return []chain{outer, counter}
	case markCedilla:
		return []chain{cedillaChain(a)}
	case markOgonek:
		return []chain{ogonekChain(a)}
	case markCommaBelow:
		return []chain{scaledComma(p, a.CommaScale, false)}
	case markCommaAbove:
		return []chain{scaledComma(p, a.CommaScale, true)}
	case markApostrophe:
		return []chain{scaledQuote(p, a.ApoScale)}
	}
	panic(fmt.Sprintf("glyphs: no mark %d", k))
}

// wedgeMark returns a wedge in a box w wide and h high whose bottom is at the
// height 0 and whose middle is at the abscissa cx: a round head of the
// diameter head in the upper right corner (the upper left if right is not
// set), a round tip of the diameter tip in the opposite lower corner.
func wedgeMark(cx, w, h, head, tip float64, right bool) chain {
	r0, r1 := head/2, tip/2
	x0, x1 := cx-w/2, cx+w/2
	c0 := outline.Point{X: x1 - r0, Y: h - r0}
	c1 := outline.Point{X: x0 + r1, Y: r1}
	if !right {
		c0.X = x0 + r0
		c1.X = x1 - r1
	}
	return wedgeChain(c0, r0, c1, r1)
}

// circumflexChain returns the circumflex, or with caron set the caron: two
// straight strokes w wide and h high which meet in a point, the feet cut
// horizontally.  The left stroke of the circumflex is thin thick at right
// angles to it and the right stroke thick; the caron is the circumflex turned
// upside down, so that its thick stroke is on the left.  The ink stands on the
// height 0.
func circumflexChain(w, h, thin, thick float64, caron bool) chain {
	th := math.Atan2(h, w/2) // the angle of the strokes to the horizontal
	lo, ro := xy(-w/2, 0), xy(w/2, 0)
	apex := xy(0, h)
	li := xy(-w/2+thin/math.Sin(th), 0)
	ri := xy(w/2-thick/math.Sin(th), 0)
	d := xy(math.Cos(th), math.Sin(th))
	crotch := meetLines(li, d, ri, xy(-d.X, d.Y))
	pts := []outline.Point{lo, li, crotch, ri, ro, apex}
	if caron {
		for i, q := range pts {
			pts[i] = xy(-q.X, h-q.Y) // half a turn about the middle of the box
		}
	}
	var path outline.Path
	addPolygon(&path, pts...)
	return pathChains(path)[0]
}

// tildeChain returns the tilde, w wide and h high on the height 0: one stroke
// which rises from the left, crests, falls through the middle, where it is
// thick, passes a trough and rises to the right, the same turned through half
// a turn about its centre; thin at the crest and the trough, with the ends cut
// at right angles to the stroke and the handles of a circular arc there.
func tildeChain(w, h, thick, thin float64) chain {
	a := (h - thin) / 2 // the amplitude of the centre line
	cy := h / 2
	x0 := -w / 2
	// The centre line falls through the middle at the angle of its chord from
	// the crest to the trough, made steeper by tildeSteep, and the ends rise at
	// that angle.
	fall := math.Atan(tildeSteep*2*a/(0.46*w)) * 180 / math.Pi
	end := 0.5 * (thick + thin)
	// The ends are cut where the edge of the stroke reaches the box.
	rise := fall
	sts := []station{
		{X: x0 + end/2*math.Sin(rise*math.Pi/180), Y: cy - 0.5*a, Dir: rise, W: end, ArcEnd: true},
		{X: x0 + 0.27*w, Y: cy + a, Dir: 0, W: thin},
		{X: x0 + 0.5*w, Y: cy, Dir: -fall, W: thick},
		{X: x0 + 0.73*w, Y: cy - a, Dir: 0, W: thin},
		{X: -x0 - end/2*math.Sin(rise*math.Pi/180), Y: cy + 0.5*a, Dir: rise, W: end, ArcEnd: true},
	}
	for _, i := range []int{1, 3} {
		t := sts[i].dir()
		var k float64
		for _, j := range []int{i - 1, i + 1} {
			d := outline.Point{X: sts[j].X - sts[i].X, Y: sts[j].Y - sts[i].Y}
			k += math.Abs(t.X*d.Y-t.Y*d.X) / (d.X*d.X + d.Y*d.Y)
		}
		if i == 1 {
			k = -k
		}
		sts[i].K = k
	}
	return strokeChain(sts)
}

// tildeSteep is the factor by which the middle of the tilde is steeper than
// the chord from the crest to the trough.
const tildeSteep = 2.4

// markDotChain returns a dot of the diameter d centred at (cx, cy), with the
// squareness of the dot of i.
func markDotChain(p design.Params, cx, cy, d float64) chain {
	r := d / 2
	q := newSuperShape(cx-r, cy-r, cx+r, cy+r, p.DotSuperness)
	var c chain
	for i := 0; i < 4; i++ {
		c = append(c, curveSeg(q.cubic(i, true)))
	}
	return c
}

// breveChain returns the breve, w wide and h high on the height 0: the lower
// halves of two ellipses, the outer one w wide and h deep, the inner one thin
// narrower on each side and thick less deep, so that the crescent is thick at
// the bottom and thin at the horns, which are cut horizontally at the top.
func breveChain(w, h, thick, thin float64) chain {
	outer := newSuperShape(-w/2, -h, w/2, h, degSuperness)
	inner := newSuperShape(-w/2+thin, -(h - thick), w/2-thin, h-thick, degSuperness)
	// Counter-clockwise: from the left horn down the outer edge to the bottom
	// and up to the right horn (the lower quarters of the outer shape), across
	// the right horn, back along the inner edge from its right to its left
	// extreme, and across the left horn.
	c := chain{curveSeg(outer.cubic(2, true)), curveSeg(outer.cubic(3, true))}
	in := chain{curveSeg(inner.cubic(0, false)), curveSeg(inner.cubic(1, false))}
	c = join(c, chain{lineSeg(c.end(), in.start())}, in, chain{lineSeg(in.end(), c.start())})
	return c.shifted(0, h)
}

// cedillaChain returns the cedilla: a straight neck CedNeck wide, which leaves
// the base at the height CedOverlap, and a hook which leaves the neck to the
// right, thin, swells to CedThick at its right, comes round below thin and ends
// in a cut CedEnd across to the lower left, as the bowl of a five hangs from
// its stem.  Neck and hook are one contour.
func cedillaChain(a design.Accents) chain {
	d := a.CedDepth
	var neck outline.Path
	addBar(&neck, -a.CedNeck/2, -0.24*d, a.CedNeck/2, a.CedOverlap)
	hook := strokeChain([]station{
		{X: -0.25 * a.CedNeck, Y: -0.24 * d, Dir: 0, W: 0.8 * a.CedNeck},
		{X: a.CedReach - a.CedThick/2, Y: -0.6 * d, Dir: -90, W: a.CedThick},
		{X: 0.15 * a.CedReach, Y: -d + 0.45*a.CedEnd, Dir: 180, W: 0.9 * a.CedEnd},
		{X: -a.CedTail + a.CedEnd/2, Y: -0.93 * d, Dir: 165, W: a.CedEnd},
	})
	cs := union(append(pathChains(neck), hook)...)
	if len(cs) != 1 {
		panic(fmt.Sprintf("glyphs: the cedilla has %d contours", len(cs)))
	}
	return cs[0]
}

// ogonekChain returns the ogonek: a stroke which leaves the base at the height
// OgoOverlap, runs down and to the left, turns at its leftmost point, comes
// round below to the right and ends in a narrow cut which points up to the
// right.
func ogonekChain(a design.Accents) chain {
	d := a.OgoDepth
	sts := []station{
		{X: 0, Y: a.OgoOverlap, Dir: -112, W: a.OgoNeck},
		{X: -a.OgoReach + a.OgoThick/2, Y: -0.55 * d, Dir: -90, W: a.OgoThick},
		{X: -0.25 * a.OgoReach, Y: -d + a.OgoThin/2, Dir: 0, W: a.OgoThin},
		{X: a.OgoTail, Y: -0.82 * d, Dir: 25, W: a.OgoTip},
	}
	return strokeChain(sts)
}

// scaledComma returns the comma scaled by s, with the centre of its head at
// the abscissa 0: below the anchor with its top at the height -CommaGap, or,
// with above set, turned through half a turn with its lowest ink at the
// height 0.
func scaledComma(p design.Params, s float64, above bool) chain {
	c := commaAt(p, 0, 0).mapPoints(func(q outline.Point) outline.Point {
		return outline.Point{X: s * q.X, Y: s * q.Y}
	})
	if above {
		var path outline.Path
		c.addContour(&path)
		x0, y0, x1, y1 := path.BBox()
		c = c.halfTurn((x0+x1)/2, (y0+y1)/2)
		var turned outline.Path
		c.addContour(&turned)
		_, ty0, _, _ := turned.BBox()
		return c.shifted(0, -ty0)
	}
	return c.shifted(0, -p.Acc.CommaGap)
}

// scaledQuote returns the right single quotation mark scaled by s, with its
// leftmost ink at the abscissa 0 and its top at the height 0.
func scaledQuote(p design.Params, s float64) chain {
	c := commaAt(p, 0, 0).mapPoints(func(q outline.Point) outline.Point {
		return outline.Point{X: s * q.X, Y: s * q.Y}
	})
	var path outline.Path
	c.addContour(&path)
	x0, _, _, _ := path.BBox()
	return c.shifted(-x0, 0)
}
