// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The currency signs of TS1: the colon, the won, the naira, the peso, the
// lira, the dong, the baht, the guarani, the florin and the dollar and the
// cent of the old-style figures.
//
// The signs of the colon, the won, the naira, the peso, the baht and the
// guarani are capitals, unchanged, with strokes or bars across them: each is
// the union of the letter and its strokes (symunion.go), in the advance of
// the letter.  The bars are Caps.Bar thick and the strokes as thick as the
// stroke of the dollar.  The lira is the sterling with two bars, the dong
// the d with a bar over an underline, and the florin the f with the tail of
// j.  The dollar of the old-style figures is the dollar with two strokes and
// their cent the cent with a leaning stroke; both stand on the advance of
// the figures.

// Shape constants of the currency signs.
const (
	colonLean         = 12.0  // the strokes of the colon sign, degrees from the vertical
	colonBeyond       = 82.0  // how far the strokes of the colon sign reach above the cap height and below the baseline
	wonBar            = 0.505 // the middle between the bars of the won and of the naira, in cap heights
	wonGap            = 0.095 // the middles of the bars of the won and of the naira from wonBar, in cap heights
	wonInset          = 25.0  // how far inside the tips of the serifs of W the bars of the won end
	wonClear          = 38.0  // the least white between the bars of the won and the crotches and the notch of W
	guaraniClear      = 35.0  // the least white between the stroke of the guarani and the tip of the left serif of the stem of G
	centOldLeft       = 5.0   // how far left of the middle of the ink of c the stroke of the cent of the old-style figures crosses half the x-height
	pesoBeyond        = 0.065 // how far the bars of the peso reach beyond the bowl, in ink widths of P
	pesoSide          = 15.0  // the least right side space of the peso
	bahtMiddle        = 0.55  // the stroke of the baht from the left of the ink, in ink widths of B
	strokeReach       = 80.0  // how far the strokes of the baht and the guarani reach above the cap height and below the baseline
	liraDrop          = 0.03  // the lower bar of the lira below the bar of the sterling, in figure heights
	liraPitch         = 0.135 // the upper bar of the lira above the lower, in figure heights
	liraBar           = 0.92  // the length of the bars of the lira, in bars of the sterling
	dongGap           = 56.0  // the white between the baseline and the underline of the dong
	florinSide        = 20.0  // the right side space of the florin
	dollarDouble      = 0.8   // each stroke of the dollar of the old-style figures, in strokes of the dollar
	dollarWhite       = 45.0  // the white between the two strokes
	dollarWhiteStroke = 59.0  // the stroke of the dollar up to which the white is dollarWhite
	dollarWhiteLoss   = 0.6   // the share of a thicker stroke which the white gives up
	centOldLean       = 7.0   // the stroke of the cent of the old-style figures, degrees from the vertical
	centOldReach      = 82.0  // how far that stroke reaches above the x-height and below the baseline
)

// currencyList lists the currency signs of TS1, in glyph order.
var currencyList = []entry{
	{"colonmonetary", '\u20A1', sideCapRound, sideOpen, colonmonetary},
	{"uni20A9", '\u20A9', sideOpen, sideOpen, won},
	{"uni20A6", '\u20A6', sideOpen, sideOpen, naira},
	{"uni20B1", '\u20B1', sideOpen, sideOpen, peso},
	{"lira", '\u20A4', sideOpen, sideOpen, lira},
	{"dong", '\u20AB', sideOpen, sideStemD, dong},
	{"uni0E3F", '\u0E3F', sideCapBowlStem, sideCapBowlB, baht},
	{"uni20B2", '\u20B2', sideCapRound, sideCapG, guarani},
	{"florin", '\u0192', sideOverhang, sideOpen, florin},
	{"dollar.oldstyle", privateUseRune("dollar.oldstyle"), sideOpen, sideOpen, dollarOldstyle},
	{"cent.oldstyle", privateUseRune("cent.oldstyle"), sideOpen, sideOpen, centOldstyle},
}

func init() {
	sharpCrotches["uni20A9"] = capWCrotches
	sharpCrotches["uni20A6"] = capNCrotches
}

// struckThrough returns the union of the outline of the glyph and the
// strokes, which are given the ink box of the glyph, with the advance of the
// glyph.
func struckThrough(p design.Params, base string, strokes func(x0, y0, x1, y1 float64) []chain) (outline.Path, float64) {
	path, adv, _ := Outline(base, p)
	parts := append(pathChains(path), strokes(path.BBox())...)
	return chainsPath(union(parts...)...), adv
}

// levelBars returns horizontal bars of the given thickness from x0 to x1
// with their middles at the heights ys.
func levelBars(x0, x1, thick float64, ys ...float64) []chain {
	var bars []chain
	for _, y := range ys {
		y = math.Round(y)
		bars = append(bars, rectChain(x0, y-thick/2, x1, y+thick/2))
	}
	return bars
}

// colonmonetary draws the colon sign: the C with two parallel strokes
// through it, which lean colonLean degrees to the right and reach
// colonBeyond above the cap height and below the baseline.  The middle line
// of the left stroke runs through the top of the counter of C and that of
// the right stroke through its bottom.
func colonmonetary(p design.Params) (outline.Path, float64) {
	f := capRound(capCParams(p))
	path, adv := capC(p)
	parts := pathChains(path)
	for _, y := range []float64{f.innerTop, f.innerBottom} {
		parts = append(parts, slashStroke(f.cx, y, -colonBeyond, p.CapHeight+colonBeyond, 90-colonLean, p.Sym.DollarStroke))
	}
	return chainsPath(union(parts...)...), adv
}

// twoBars returns the glyph with two bars across its ink, which end inset
// inside its leftmost and its rightmost ink.  The middles of the bars lie
// wonGap of the cap height below and above wonBar, the lower no lower than
// low and the upper no higher than high.
func twoBars(p design.Params, base string, inset, low, high float64) (outline.Path, float64) {
	return struckThrough(p, base, func(x0, _, x1, _ float64) []chain {
		h := p.CapHeight
		return levelBars(x0+inset, x1-inset, p.Caps.Bar, math.Max((wonBar-wonGap)*h, low), math.Min((wonBar+wonGap)*h, high))
	})
}

// won draws the won: the W with two bars which end wonInset inside the tips
// of its outer serifs.  The bars keep wonClear from the crotches of the two
// vertices of W below them and from the notch between its inner diagonals
// above them.
func won(p design.Params) (outline.Path, float64) {
	ks := capWCrotches(p)
	notch, crotch := ks[0].Tip.Y, math.Max(ks[1].Tip.Y, ks[2].Tip.Y)
	d := wonClear + p.Caps.Bar/2
	return twoBars(p, "W", wonInset, crotch+d, notch-d)
}

// naira draws the naira: the N with two bars across the whole of its ink.
func naira(p design.Params) (outline.Path, float64) {
	return twoBars(p, "N", 0, 0, p.CapHeight)
}

// peso draws the peso: the P with two bars across its stem and its bowl,
// from the tips of the serifs on the left to pesoBeyond beyond the bowl,
// which divide the counter of the bowl into three whites of one height.  The
// advance is that of P, or wider where the bars would end less than pesoSide
// from its right edge.
func peso(p design.Params) (outline.Path, float64) {
	path, adv, _ := Outline("P", p)
	x0, _, x1, _ := path.BBox()
	_, y0, _, y1 := path[1:2].BBox()
	t := p.Caps.Bar
	white := (y1 - y0 - 2*t) / 3
	end := x1 + math.Round(pesoBeyond*(x1-x0))
	bars := levelBars(x0, end, t, y0+white+t/2, y1-white-t/2)
	return chainsPath(union(append(pathChains(path), bars...)...)...), math.Max(adv, end+pesoSide)
}

// upright returns the vertical stroke of the baht and the guarani with its
// middle at x, as thick as the stroke of the dollar, from strokeReach below
// the baseline to strokeReach above the cap height.
func upright(p design.Params, x float64) chain {
	x = math.Round(x)
	t := p.Sym.DollarStroke
	return rectChain(x-t/2, -strokeReach, x+t/2, p.CapHeight+strokeReach)
}

// baht draws the baht: the B with a vertical stroke through its bowls.
//
// The stroke stands where it is at the left of the ink, but not further right
// than bahtClear before the place where the top and bottom of the counters
// of the B turn into the bowls.
func baht(p design.Params) (outline.Path, float64) {
	b, _, _ := Outline("B", p)
	start := math.Inf(1)
	if len(b) > 1 && len(b[1].Segs) > 0 && !b[1].Segs[0].IsCurve {
		start = b[1].Segs[0].P3.X
	}
	return struckThrough(p, "B", func(x0, _, x1, _ float64) []chain {
		x := x0 + bahtMiddle*(x1-x0)
		return []chain{upright(p, math.Min(x, start-bahtClear+p.Sym.DollarStroke/2))}
	})
}

// bahtClear is the distance of the left edge of the stroke of the baht to the
// place where the straight top of the counters of the B ends.
const bahtClear = 1.0

// guarani draws the guarani: the G with a vertical stroke through the middle
// of its round, or further left where it would come closer than
// guaraniClear to the tip of the left serif of the stem of G.
func guarani(p design.Params) (outline.Path, float64) {
	tip := capRound(p).beakX(p) - p.Caps.Stem - p.Caps.CGSJQ.GSerifLeft
	return struckThrough(p, "G", func(x0, _, _, _ float64) []chain {
		return []chain{upright(p, math.Min(x0+p.Caps.OWidth/2, tip-guaraniClear-p.Sym.DollarStroke/2))}
	})
}

// lira draws the lira: the sterling with two bars in place of its bar, each
// liraBar of that bar long: the lower liraDrop of the figure height below the place of the bar of the
// sterling and the upper liraPitch above the lower.
func lira(p design.Params) (outline.Path, float64) {
	h := p.Figures.Height
	low := p.Sym.SterBarY - math.Round(liraDrop*h)
	return sterlingBars(p, math.Round(liraBar*p.Sym.SterBarLen), low, low+math.Round(liraPitch*h))
}

// dong draws the dong: the d with its bar over an underline as thick as the
// bar and as wide as the ink, dongGap below the baseline.
func dong(p design.Params) (outline.Path, float64) {
	path, adv := dcroat(p)
	x0, _, x1, _ := path.BBox()
	rect(&path, x0, -dongGap-p.ThinLower, x1, -dongGap, true)
	return path, adv
}

// florin draws the florin: the f without its foot serifs, whose stem runs on
// below the baseline into the tail of j.  The tail overhangs the advance on
// the left as that of j does, and the hook stands florinSide from its right
// edge.
func florin(p design.Params) (outline.Path, float64) {
	stem, bar := fStem(p), fBar(p)
	js := jStem(p)
	dx := stem.Left + stem.Width/2 - (js.Left + js.Width/2)
	tail := jTail(p).shifted(dx, 0)
	h := newFHookParts(p).chain
	var path outline.Path
	path.MoveTo(tail.end().X, tail.end().Y)
	path.LineTo(stem.right(), bar.Top-bar.Thick)
	barRight(&path, bar)
	path.LineTo(h.start().X, h.start().Y)
	h.appendTo(&path)
	path.LineTo(stem.Left, bar.Top)
	barLeft(&path, bar)
	path.LineTo(tail.start().X, tail.start().Y)
	tail.appendTo(&path)
	path = shiftInk(path, math.Round(-dx), 0)
	_, _, x1, _ := path.BBox()
	return path, math.Ceil(x1 + florinSide)
}

// dollarOldstyle draws the dollar of the old-style figures: the S of the
// dollar with two vertical strokes through it in place of the one, each
// dollarDouble strokes of the dollar thick, with dollarWhite of white
// between them, from DollarBottom to DollarTop.  The white is narrower by
// dollarWhiteLoss of each unit by which the stroke of the dollar is thicker
// than dollarWhiteStroke.
func dollarOldstyle(p design.Params) (outline.Path, float64) {
	s := p.Sym
	t := math.Round(dollarDouble * s.DollarStroke)
	letter, cx := dollarS(p)
	parts := append([]chain(nil), letter...)
	white := dollarWhite - dollarWhiteLoss*math.Max(0, s.DollarStroke-dollarWhiteStroke)
	d := math.Round((t + white) / 2)
	for _, x := range []float64{cx - d, cx + d} {
		parts = append(parts, rectChain(x-t/2, s.DollarBottom, x+t/2, s.DollarTop))
	}
	return chainsPath(union(parts...)...), p.Figures.Advance
}

// centOldstyle draws the cent of the old-style figures: the c, centred in
// the advance of the figures, with a stroke as thick as that of the cent
// which crosses half the x-height centOldLeft to the left of the middle of
// the ink of c, leans centOldLean degrees to the right and reaches
// centOldReach above the x-height and below the baseline.
func centOldstyle(p design.Params) (outline.Path, float64) {
	c, _, _ := Outline("c", p)
	c, adv := placeCentred(c, p.Figures.Advance)
	x0, _, x1, _ := c.BBox()
	cx := math.Round((x0+x1)/2) - centOldLeft
	stroke := slashStroke(cx, p.XHeight/2, -centOldReach, p.XHeight+centOldReach, 90-centOldLean, centStroke)
	return chainsPath(union(append(pathChains(c), stroke)...)...), adv
}
