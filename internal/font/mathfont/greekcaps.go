// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

import (
	"math"
	"sync"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/outline"

	"seehuhn.de/go/bedern/internal/fair"
)

// The Greek capitals of the first tier: upright, with the weights and the serifs of the roman
// capitals.  Γ, Θ, Λ, Ξ and Π are made from the parts of E, O, A and H; Δ, Σ,
// Υ, Φ, Ψ and Ω are drawn new from the same parts: the stem of the capitals
// with its bracketed serifs (glyphs.CapStemOutline), the arms with their
// vertical serifs (glyphs.CapArmOutline), the round of O
// (glyphs.SuperEllipseOutline) and the stroke of the symbols (glyphs.Stroke).
//
// Widths are ink widths over the ink width of H, weights shares of the parts
// of the roman capitals.
const (
	// gammaInk: Γ over H.
	gammaInk = 0.714
	// deltaInk: Δ over H.
	deltaInk = 0.834
	// deltaApex: the left end of the apex of Δ from the left of its ink, as
	// a share of the ink width.
	deltaApex = 0.477
	// deltaBase: the base of Δ over the bar of the arms.
	deltaBase = 1.25
	// greekBar: the bars of Ξ and of Θ over the bar of the arms of E.
	greekBar = 1.55
	// thetaBar: the bar of Θ over the width of the counter of O;
	// thetaBarHeight its middle over the cap height.
	thetaBar       = 0.657
	thetaBarHeight = 0.511
	// thetaReach: the serifs of the bar of Θ beyond its edges, over the
	// middle serif of E.
	thetaReach = 1.0
	// xiInk: Ξ over H; xiMiddle: its middle bar over its ink width.
	xiInk    = 0.797
	xiMiddle = 0.578
	// sigmaInk: Σ over H; sigmaTop and sigmaBottom its top and bottom bars
	// over the ink; sigmaVertex the tip of the vertex over the ink from its
	// left.
	sigmaInk    = 0.776
	sigmaTop    = 0.918
	sigmaBottom = 0.943
	sigmaVertex = 0.548
	// phiInk: Φ over H; phiSide: the sides of its oval
	// over the stem; phiTop: the top of the oval over the
	// cap height.
	phiInk  = 0.907
	phiSide = 1.08
	phiTop  = 0.83
	// omegaInk: Ω over H; omegaRound: the round at its widest over the ink;
	// omegaFoot: each foot over the ink; omegaGap: the white between the legs
	// at a fifth of the cap height over the ink.
	omegaInk   = 0.969
	omegaRound = 0.913
	omegaFoot  = 0.380
	omegaGap   = 0.175
	// psiInk: Ψ over H; psiCup: the lowest ink of the arms a third of the
	// ink from the middle, over the cap height.
	psiInk = 1.006
	psiCup = 0.419
	// psiTurn: the height over the cap height down to which the arms of Ψ
	// run straight before they turn into the cup; psiNeck: how far the head
	// of a capital stem on an arm reaches below its bracket.
	psiTurn = 0.61
	psiNeck = 30
	// upsilonInk: Υ over H; upsilonCrotch: the crotch over the cap height;
	// upsilonEnd: the lowest ink of the ends of the arms over the cap
	// height.
	upsilonInk    = 0.931
	upsilonCrotch = 0.623
	upsilonEnd    = 0.895
	// upsilonTip: the thickness of the blunt ends of the arms of Υ over the
	// stem;
	// upsilonCrest: the thickness at the crest of an arm over the thin of O;
	// upsilonCrestAt: the crest of the left arm from the left of the ink, as
	// a share of the ink width.
	upsilonTip     = 0.5
	upsilonCrest   = 1.3
	upsilonCrestAt = 0.22
	// upsilonStem: the top of the stem of Υ, where both arms begin with the
	// width of the stem, over the cap height.
	upsilonStem = 0.2
	// upsilonBend: the curvature of the centre line of an arm of Υ where it
	// leaves the stem, times the stem.
	upsilonBend = 0.8
)

// greekCapitals are the Greek capitals, in the order of the font: the eleven
// which differ from the Latin capitals, then the thirteen which are the Latin
// capitals of the regular; the alphabet of the upright Greek capitals is
// complete and a document can take it from the font by its name.
var greekCapitals = []derivation{
	{"uni0393", 0x0393, Ord, "F", capGamma},
	{"uni0394", 0x0394, Ord, "A", capDelta},
	{"uni0398", 0x0398, Ord, "O", capTheta},
	{"uni039B", 0x039B, Ord, "A", capLambda},
	{"uni039E", 0x039E, Ord, "E", capXi},
	{"uni03A0", 0x03A0, Ord, "H", capPi},
	{"uni03A3", 0x03A3, Ord, "E", capSigma},
	{"uni03A5", 0x03A5, Ord, "Y", capUpsilon},
	{"uni03A6", 0x03A6, Ord, "O", capPhi},
	{"uni03A8", 0x03A8, Ord, "H", capPsi},
	{"uni03A9", 0x03A9, Ord, "O", capOmega},
	{"uni0391", 0x0391, Ord, "A", copyOf("A")},
	{"uni0392", 0x0392, Ord, "B", copyOf("B")},
	{"uni0395", 0x0395, Ord, "E", copyOf("E")},
	{"uni0396", 0x0396, Ord, "Z", copyOf("Z")},
	{"uni0397", 0x0397, Ord, "H", copyOf("H")},
	{"uni0399", 0x0399, Ord, "I", copyOf("I")},
	{"uni039A", 0x039A, Ord, "K", copyOf("K")},
	{"uni039C", 0x039C, Ord, "M", copyOf("M")},
	{"uni039D", 0x039D, Ord, "N", copyOf("N")},
	{"uni039F", 0x039F, Ord, "O", copyOf("O")},
	{"uni03A1", 0x03A1, Ord, "P", copyOf("P")},
	{"uni03A4", 0x03A4, Ord, "T", copyOf("T")},
	{"uni03A7", 0x03A7, Ord, "X", copyOf("X")},
}

// GreekCapitals returns the Greek capitals of the parameters.
func GreekCapitals(p design.Params) []Glyph { return greekCaps(newRoman(p)) }

func greekCaps(r *roman) []Glyph {
	out := make([]Glyph, len(greekCapitals))
	var wg sync.WaitGroup
	for i, d := range greekCapitals {
		wg.Add(1)
		go func() {
			defer wg.Done()
			path, adv := d.draw(r)
			out[i] = Glyph{Name: d.name, Rune: d.r, Advance: math.Round(adv), Path: path, Class: d.class, Like: d.like}
		}()
	}
	wg.Wait()
	return out
}

// hInk returns the ink width of H.
func hInk(p design.Params) float64 { return design.CapHInk(p) }

// stemSides returns the side spaces of a capital whose outermost ink is the
// tip of a stem serif, those of H.
func stemSides(p design.Params) (left, right float64) {
	c := p.Caps
	return c.SideStemLeft - c.SerifJut, c.SideStemRight - c.SerifJut
}

// near reports whether two points coincide within a thousandth of a unit.
func near(a, b outline.Point) bool { return math.Hypot(a.X-b.X, a.Y-b.Y) < 1e-3 }

// startAt returns the contour started at the end of the segment which ends at
// q, the same closed curve.
func startAt(c outline.Contour, q outline.Point) outline.Contour {
	c = closed(c)
	for i, s := range c.Segs {
		if near(s.P3, q) {
			return outline.Contour{Start: s.P3, Segs: append(append([]outline.Segment(nil), c.Segs[i+1:]...), c.Segs[:i+1]...)}
		}
	}
	panic("mathfont: no node at the point")
}

// closed returns the contour with its closing line written out.
func closed(c outline.Contour) outline.Contour {
	if n := len(c.Segs); n > 0 && !near(c.Segs[n-1].P3, c.Start) {
		c.Segs = append(c.Segs[:n:n], outline.Segment{P3: c.Start})
	}
	return c
}

// merged returns the contour with consecutive straight segments on one line
// joined, the closing segment included.
func merged(c outline.Contour) outline.Contour {
	c = closed(c)
	from := func(segs []outline.Segment, start outline.Point, i int) outline.Point {
		if i == 0 {
			return start
		}
		return segs[i-1].P3
	}
	collinear := func(a, b, q outline.Point) bool {
		u, v := pt(b.X-a.X, b.Y-a.Y), pt(q.X-b.X, q.Y-b.Y)
		lu, lv := math.Hypot(u.X, u.Y), math.Hypot(v.X, v.Y)
		return lu > 1e-9 && lv > 1e-9 && math.Abs(u.X*v.Y-u.Y*v.X)/(lu*lv) < 1e-9 && u.X*v.X+u.Y*v.Y > 0
	}
	for changed := true; changed; {
		changed = false
		n := len(c.Segs)
		for i := 0; i < n && n > 2; i++ {
			j := (i + 1) % n
			a, b := c.Segs[i], c.Segs[j]
			if a.IsCurve || b.IsCurve || !collinear(from(c.Segs, c.Start, i), a.P3, b.P3) {
				continue
			}
			if j == 0 {
				// The node between the last and the first segment is the
				// start: the first segment starts where the last did.
				c.Start = c.Segs[n-2].P3
				c.Segs = c.Segs[:n-1]
			} else {
				c.Segs[i].P3 = b.P3
				c.Segs = append(c.Segs[:j:j], c.Segs[j+1:]...)
			}
			changed = true
			break
		}
	}
	return c
}

// barOutline returns a horizontal bar of the capitals from x0 to x1, its
// underside at lower and its upper edge at upper, with the vertical serifs
// of the arms of E at both ends, reaching drop below and rise above it (zero
// for none): the arm of glyphs.CapArmOutline from the middle to the right and
// its mirror image.
func barOutline(p design.Params, x0, x1, lower, upper, drop, rise float64) outline.Contour {
	mid := (x0 + x1) / 2
	r := startAt(glyphs.CapArmOutline(p, mid, x1, lower, upper, drop, rise)[0], pt(mid, lower))
	l := startAt(mirroredX(outline.Path{r}, mid)[0], pt(mid, lower))
	n := len(r.Segs)
	c := outline.Contour{Start: r.Start, Segs: append(append([]outline.Segment(nil), r.Segs[:n-1]...), l.Segs[1:]...)}
	return merged(c)
}

// armSegs returns the segments of an arm of glyphs.CapArmOutline from the
// point from up to its closing line along x = edge.
func armSegs(p design.Params, edge, end, lower, upper, drop, rise float64, from outline.Point) []outline.Segment {
	c := startAt(glyphs.CapArmOutline(p, edge, end, lower, upper, drop, rise)[0], from)
	var out []outline.Segment
	q := c.Start
	for _, s := range c.Segs {
		if !s.IsCurve && math.Abs(q.X-edge) < 1e-6 && math.Abs(s.P3.X-edge) < 1e-6 {
			break
		}
		out = append(out, s)
		q = s.P3
	}
	return out
}

// lineTo returns a straight segment to q.
func lineTo(q outline.Point) outline.Segment { return outline.Segment{P3: q} }

// capGamma draws Γ: the stem of F with a foot serif on both sides and the head
// serif on the left, and the top arm of E with its serif, so long that the
// ink is gammaInk of that of H; the side spaces of H on the left and of E on
// the right.
func capGamma(r *roman) (outline.Path, float64) {
	p := r.p
	c := p.Caps
	cap := p.CapHeight
	j := c.SerifJut
	left := design.CapStemLeft(p)
	right := left + c.Stem
	end := left - j + math.Round(gammaInk*hInk(p))
	stem := glyphs.CapStemOutline(p, left, cap, j, j, j, 0)[0]
	arm := glyphs.CapArmOutline(p, right, end, cap-c.Bar, cap, c.TopDrop, 0)[0]
	arm = startAt(arm, pt(right, cap-c.Bar))
	armPart := arm.Segs[:len(arm.Segs)-2]
	stem = closed(stem)
	var segs []outline.Segment
	for _, s := range stem.Segs {
		if !s.IsCurve && near(s.P3, pt(right, cap)) {
			segs = append(segs, lineTo(pt(right, cap-c.Bar)))
			segs = append(segs, armPart...)
			continue
		}
		segs = append(segs, s)
	}
	path := outline.Path{merged(outline.Contour{Start: stem.Start, Segs: segs})}
	return path, end + c.SideArmE
}

// capLambda draws Λ: A without its crossbar, the inner edges of its diagonals
// running up to the crotch of the apex, with the side spaces of A.
func capLambda(r *roman) (outline.Path, float64) {
	p := r.p
	a, adv := r.get("A")
	outer, counter := closed(a[0]), a[1]
	crotch := counter.Start
	for _, s := range counter.Segs {
		if s.P3.Y > crotch.Y {
			crotch = s.P3
		}
	}
	lower := p.CapDiag.ABar - p.Caps.Bar/2
	var segs []outline.Segment
	from := outer.Start
	for _, s := range outer.Segs {
		if !s.IsCurve && math.Abs(from.Y-lower) < 1e-6 && math.Abs(s.P3.Y-lower) < 1e-6 && s.P3.X > from.X {
			segs[len(segs)-1].P3 = crotch
		} else {
			segs = append(segs, s)
		}
		from = s.P3
	}
	return outline.Path{merged(outline.Contour{Start: outer.Start, Segs: segs})}, adv
}

// capDelta draws Δ: a thin diagonal on the left and a thick one on the right,
// of the weights of A, which meet in the flat of the apex of A, and a base
// deltaBase times the bar of the arms; deltaInk of H wide, the left end of the
// apex deltaApex of the ink from the left; the side spaces of A.
func capDelta(r *roman) (outline.Path, float64) {
	p := r.p
	d := p.CapDiag
	w := math.Round(deltaInk * hInk(p))
	top := p.CapHeight + d.ApexRise
	base := math.Round(deltaBase * p.Caps.Bar)
	ax := deltaApex*w + d.ApexFlat/2
	tl, tr := pt(ax-d.ApexFlat/2, top), pt(ax+d.ApexFlat/2, top)
	bl, br := pt(0, 0), pt(w, 0)
	// The inner edges are the outer ones moved inwards at right angles by
	// the weights.
	inner := func(a, b outline.Point, t float64, left bool) (outline.Point, outline.Point) {
		l := math.Hypot(b.X-a.X, b.Y-a.Y)
		nx, ny := (b.Y-a.Y)/l*t, -(b.X-a.X)/l*t
		if !left {
			nx, ny = -nx, -ny
		}
		return pt(a.X+nx, a.Y+ny), pt(b.X+nx, b.Y+ny)
	}
	la, lb := inner(bl, tl, d.Thin, true)
	ra, rb := inner(br, tr, d.Thick, false)
	at := func(a, b outline.Point, y float64) outline.Point {
		return pt(a.X+(b.X-a.X)*(y-a.Y)/(b.Y-a.Y), y)
	}
	crotch := lineMeet(la, lb, ra, rb)
	path := outline.Path{
		polygon(bl, br, tr, tl),
		polygon(at(la, lb, base), crotch, at(ra, rb, base)),
	}
	return withSides(path, d.SideA.Left, d.SideA.Right)
}

// lineMeet returns the point where the line through a and b meets the line
// through c and d.
func lineMeet(a, b, c, d outline.Point) outline.Point {
	u, v := pt(b.X-a.X, b.Y-a.Y), pt(d.X-c.X, d.Y-c.Y)
	den := u.X*v.Y - u.Y*v.X
	t := ((c.X-a.X)*v.Y - (c.Y-a.Y)*v.X) / den
	return pt(a.X+t*u.X, a.Y+t*u.Y)
}

// capTheta draws Θ: O, and inside it a bar greekBar times the bar of the arms
// thick, thetaBar of the width of the counter of O long and centred at
// thetaBarHeight of the cap height, with the vertical serifs of the middle arm
// of E at both ends, thetaReach of theirs; the side spaces of O.
func capTheta(r *roman) (outline.Path, float64) {
	p := r.p
	c := p.Caps
	o, adv := r.get("O")
	cx := c.SideRoundLeft + c.OWidth/2
	half := math.Round(thetaBar*(c.OWidth-2*c.Round)) / 2
	t := math.Round(greekBar * c.Bar)
	y := math.Round(thetaBarHeight*p.CapHeight - t/2)
	reach := math.Round(thetaReach * c.MiddleReach)
	bar := barOutline(p, cx-half, cx+half, y, y+t, reach, reach)
	return append(append(outline.Path(nil), o...), bar), adv
}

// capXi draws Ξ: three bars greekBar times the bar of the arms thick, the top
// one at the cap height with the serifs of the top arm of E hanging at both
// ends, the bottom one on the baseline with those of the bottom arm rising,
// and the middle one, xiMiddle of the ink long, at the height of the bar of
// H with those of the middle arm; xiInk of H wide, with the left side space
// of H on both sides.
func capXi(r *roman) (outline.Path, float64) {
	p := r.p
	c := p.Caps
	cap := p.CapHeight
	w := math.Round(xiInk * hInk(p))
	t := math.Round(greekBar * c.Bar)
	m := math.Round(xiMiddle*w) / 2
	y := c.HBarHeight - t/2
	path := outline.Path{
		barOutline(p, 0, w, 0, t, 0, c.BottomRise),
		barOutline(p, w/2-m, w/2+m, y, y+t, c.MiddleReach, c.MiddleReach),
		barOutline(p, 0, w, cap-t, cap, c.TopDrop, 0),
	}
	side, _ := stemSides(p)
	side = math.Max(side, c.SideArmE)
	return withSides(path, side, side)
}

// capPi draws Π: the two stems of H with their foot serifs, and over them a
// bar of the thickness of the arms, reaching to the tips of the foot serifs
// and cut square; the side spaces of H.
func capPi(r *roman) (outline.Path, float64) {
	p := r.p
	c := p.Caps
	cap := p.CapHeight
	j := c.SerifJut
	l1 := design.CapStemLeft(p)
	l2 := l1 + c.Stem + c.HCounter
	top := cap - c.Bar/2
	bar := outline.Path{box(l1-j, cap-c.Bar, l2+c.Stem+j, cap)}
	path := glyphs.Union(bar, glyphs.CapStemOutline(p, l1, top, j, j, 0, 0), glyphs.CapStemOutline(p, l2, top, j, j, 0, 0))
	return path, design.CapHAdvance(p)
}

// capSigma draws Σ: a top bar of the thickness of the arms with the serif of
// the top arm of E hanging at its right end, a bottom bar greekBar times as
// thick with the serif of the bottom arm rising, a thick diagonal (the thick
// stroke of A) from the top left down to the vertex, sigmaVertex of the ink
// from the left at half the cap height, and a thin one (the thin stroke of A)
// from there down to the left end of the bottom bar.  It is sigmaInk of H
// wide, with the side spaces of H on the left and of E on the right.
func capSigma(r *roman) (outline.Path, float64) {
	p := r.p
	c := p.Caps
	d := p.CapDiag
	cap := p.CapHeight
	w := math.Round(sigmaInk * hInk(p))
	tb, bb := c.Bar, math.Round(greekBar*c.Bar)
	v := pt(math.Round(sigmaVertex*w), cap/2)
	topEnd, botEnd := math.Round(sigmaTop/sigmaBottom*w), w
	// The lines of the diagonals: the outer edge of the thick one through the
	// top left corner, the outer edge of the thin one through the left end of
	// the top of the bottom bar, the inner edges through the tip of the
	// vertex, each at the distance of the weight of its stroke.
	tangent := func(from outline.Point, t float64, sign float64) outline.Point {
		dx, dy := v.X-from.X, v.Y-from.Y
		l := math.Hypot(dx, dy)
		a := math.Atan2(dy, dx) + sign*math.Asin(t/l)
		return pt(math.Cos(a), math.Sin(a))
	}
	tl := pt(0, cap)
	td := tangent(tl, d.Thick, -1)
	bl := pt(0, bb)
	bd := tangent(bl, d.Thin, 1)
	crotch := lineMeet(tl, pt(tl.X+td.X, tl.Y+td.Y), bl, pt(bl.X+bd.X, bl.Y+bd.Y))
	along := func(q, dir outline.Point, y float64) outline.Point {
		return pt(q.X+dir.X*(y-q.Y)/dir.Y, y)
	}
	xa := along(v, td, cap-tb)
	xb := along(v, bd, bb)
	segs := []outline.Segment{lineTo(pt(botEnd, 0))}
	segs = append(segs, armSegs(p, xb.X, botEnd, 0, bb, 0, c.BottomRise, pt(botEnd, 0))...)
	segs = append(segs, lineTo(v), lineTo(xa))
	segs = append(segs, armSegs(p, xa.X, topEnd, cap-tb, cap, c.TopDrop, 0, xa)...)
	segs = append(segs, lineTo(tl), lineTo(crotch), lineTo(bl), lineTo(pt(0, 0)))
	path := outline.Path{merged(outline.Contour{Start: pt(0, 0), Segs: segs})}
	left, _ := stemSides(p)
	return withSides(path, left, c.SideArmE)
}

// capPhi draws Φ: the stem of H with all four serifs, and an oval round it
// of the squareness of O, phiInk of H wide, from the baseline mirrored up to
// phiTop of the cap height, its sides phiSide stems and its top and bottom
// the thin of O thick; the side spaces of O.
func capPhi(r *roman) (outline.Path, float64) {
	p := r.p
	c := p.Caps
	cap := p.CapHeight
	j := c.SerifJut
	w := math.Round(phiInk * hInk(p))
	side := math.Round(phiSide * c.Stem)
	y1 := math.Round(phiTop * cap)
	y0 := cap - y1
	stem := glyphs.CapStemOutline(p, (w-c.Stem)/2, cap, j, j, j, j)
	oval := append(glyphs.SuperEllipseOutline(0, y0, w, y1, p.SupernessOuter, true),
		glyphs.SuperEllipseOutline(side, y0+c.RoundThin, w-side, y1-c.RoundThin, p.SupernessInner, false)...)
	path := glyphs.Union(stem, oval)
	return withSides(path, c.SideRoundLeft, c.SideRoundRight)
}

// capOmega draws Ω: an open round of the weights of O, its sides the round of
// O and its top the thin of O, omegaRound of the ink wide at its widest,
// whose ends run inwards and down as thin legs to two feet; each foot is a
// bar greekBar times the bar of the arms thick and omegaFoot of the ink long
// with the serif of the bottom arm of E rising at its outer end, and the legs
// stand omegaGap of the ink apart.  It is omegaInk of H wide, with the side
// spaces of O.
func capOmega(r *roman) (outline.Path, float64) {
	p := r.p
	c := p.Caps
	cap := p.CapHeight
	w := math.Round(omegaInk * hInk(p))
	cx := w / 2
	round := omegaRound * w
	foot := math.Round(omegaFoot * w)
	t := math.Round(greekBar * c.Bar)
	gap := omegaGap * w
	thin := c.RoundThin
	top := cap + p.OvershootUpper
	mid := 0.56 * cap
	legX := gap/2 + thin/2
	stroke := glyphs.Stroke([]glyphs.Station{
		{X: cx - legX, Y: t / 2, Dir: 100, W: thin},
		{X: cx - round/2 + c.Round/2, Y: mid, Dir: 90, W: c.Round},
		{X: cx, Y: top - thin/2, Dir: 0, W: thin},
		{X: cx + round/2 - c.Round/2, Y: mid, Dir: -90, W: c.Round},
		{X: cx + legX, Y: t / 2, Dir: -100, W: thin},
	})
	right := glyphs.CapArmOutline(p, math.Min(w-foot, cx+legX-thin), w, 0, t, 0, c.BottomRise)
	left := mirroredX(right, cx)
	path := glyphs.Union(stroke, left, right)
	return withSides(path, c.SideRoundLeft, c.SideRoundRight)
}

// capPsi draws Ψ: the stem of H with all four serifs, and a cup of two arms
// of the stem weight whose tops have head serifs on both sides at the cap
// height, which run straight down and turn into the stem, their lowest ink
// psiCup of the cap height a third of the ink from the middle.  It is psiInk
// of H wide, with the side spaces of H.
func capPsi(r *roman) (outline.Path, float64) {
	p := r.p
	c := p.Caps
	cap := p.CapHeight
	s := c.Stem
	j := c.SerifJut
	jut := c.BracketReach
	w := math.Round(psiInk * hInk(p))
	cx := w / 2
	arm := cx - jut - s/2 // the centre of an arm from the middle
	yt := p.CapHeight - c.BracketHeight - psiNeck
	turn := psiTurn * cap
	low := psiCup * cap
	cup := glyphs.Stroke([]glyphs.Station{
		{X: cx - arm, Y: yt, Dir: -90, W: s, Flat: true},
		{X: cx - arm, Y: turn, Dir: -90, W: s},
		{X: cx, Y: low - 0.15*cap + c.RoundThin, Dir: 0, W: 1.2 * c.RoundThin},
		{X: cx + arm, Y: turn, Dir: 90, W: s, Flat: true},
		{X: cx + arm, Y: yt, Dir: 90, W: s},
	})
	cup = outline.Path{withHead(withHead(cup[0], p, cx-arm, yt, jut), p, cx+arm, yt, jut)}
	stem := glyphs.CapStemOutline(p, cx-s/2, cap, j, j, jut, jut)
	path := glyphs.Union(cup, stem)
	left, right := stemSides(p)
	return withSides(path, left, right)
}

// withHead returns the contour of a stroke whose end is cut straight across
// a vertical stroke of the capital stem at the height y, centred at x, with
// that cut replaced by the head of a capital stem with serifs of the length
// jut on both sides at the cap height.
func withHead(c outline.Contour, p design.Params, x, y, jut float64) outline.Contour {
	s := p.Caps.Stem
	l, r := pt(x-s/2, y), pt(x+s/2, y)
	head := startAt(shifted(glyphs.CapStemOutline(p, l.X, p.CapHeight-y, 0, 0, jut, jut), 0, y)[0], r)
	head.Segs = head.Segs[:len(head.Segs)-1]
	c = closed(c)
	q := c.Start
	for i, sg := range c.Segs {
		if !sg.IsCurve && near(q, r) && near(sg.P3, l) {
			segs := append(append(append([]outline.Segment(nil), c.Segs[:i]...), head.Segs...), c.Segs[i+1:]...)
			return merged(outline.Contour{Start: c.Start, Segs: segs})
		}
		q = sg.P3
	}
	panic("mathfont: no cut across the stroke")
}

// capUpsilon draws Υ: the stem of H with its foot serifs from the baseline to
// the crotch at upsilonCrotch of the cap height, and two arms which leave the
// stem there, rise and curve outwards over the cap height and curl down into
// a ball of the figures, its lowest ink at upsilonEnd of the cap height.  It
// is upsilonInk of H wide, with the side spaces of Y.
func capUpsilon(r *roman) (outline.Path, float64) {
	p := r.p
	c := p.Caps
	cap := p.CapHeight
	j := c.SerifJut
	s := c.Stem
	w := math.Round(upsilonInk * hInk(p))
	cx := w / 2
	tip := upsilonTip * s
	thin := upsilonCrest * c.RoundThin
	y1 := upsilonStem * cap
	end := pt(tip/2, upsilonEnd*cap+tip/4)
	crest := pt(upsilonCrestAt*w, cap-thin/2)
	// The left arm and its mirror image, whose start lies one unit lower so
	// that the cuts at the starts of the arms, which overlap, do not run
	// along each other.
	stations := []glyphs.Station{
		{X: cx, Y: y1, Dir: 90, W: s, K: upsilonBend / s},
		{X: crest.X, Y: crest.Y, Dir: 180, W: thin},
		{X: end.X, Y: end.Y, Dir: 228, W: tip, ArcEnd: true},
	}
	var mirror []glyphs.Station
	for _, st := range stations {
		st.X, st.Dir, st.K = w-st.X, 180-st.Dir, -st.K
		mirror = append(mirror, st)
	}
	mirror[0].Y--
	left := glyphs.Stroke(stations)
	right := glyphs.Stroke(mirror)
	arms := glyphs.Union(left, right)
	// The arms stand on the top of the stem, whose cut they share: the
	// contour of the stem runs up its right edge, round the arms and down its
	// left edge.
	l, rt := pt(cx-s/2, y1), pt(cx+s/2, y1-1)
	a := startAt(arms[0], rt)
	var top []outline.Segment
	for _, sg := range a.Segs {
		top = append(top, sg)
		if near(sg.P3, l) {
			break
		}
	}
	st := startAt(glyphs.CapStemOutline(p, cx-s/2, y1-1, j, j, 0, 0)[0], rt)
	segs := append(top, st.Segs[1:]...)
	path := append(outline.Path{merged(outline.Contour{Start: rt, Segs: segs})}, arms[1:]...)
	return withSides(path, p.CapDiag.SideY.Left, p.CapDiag.SideY.Right)
}

// intendedCapFinding returns the reason if a finding of the fairness
// analysis on a Greek capital belongs to its form, and "" otherwise.
func intendedCapFinding(name string, p design.Params, f fair.Finding) string {
	if f.Kind != fair.Inflection {
		return ""
	}
	switch name {
	case "uni03A8":
		if f.Y > psiCup*p.CapHeight && f.Y < p.CapHeight-p.Caps.BracketHeight && math.Abs(math.Abs(f.Dir)-90) < 15 {
			return "the outer edge of an arm runs from the concave bracket of its head serif straight into the convex turn of the cup"
		}
	case "uni03A9":
		if f.Y < 0.45*p.CapHeight {
			return "the edges of the round run from its convex sides into the concave turns of the legs, which come in to the feet"
		}
	}
	return ""
}
