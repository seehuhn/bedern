// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The marks of the text symbols which are made of the parts of the
// punctuation: the straight quotation marks on the baseline, the dashes of two
// thirds and of three quarters of an em, the compound word marks at the height
// of the capitals and of the ascenders, the double hyphens, the angle, double
// and quill brackets, the double vertical bar, the large circle, the open
// bullet, the low tilde and the double grave.
//
// The straight quotation marks on the baseline are the straight marks moved
// down.  The dashes are the bar of the em dash with its side spaces, in a
// fraction of its advance.  The compound word marks are empty and have no
// width.  The double hyphen is two bars of the hyphen about its middle; as a
// hyphenation character it has half the advance of the hyphen, and its ink
// hangs over the right edge of the advance.  The brackets reach from the top
// to the foot of the square brackets: an angle bracket is two straight
// strokes which meet in a point at the middle, a double bracket two stems
// between the arms of the bracket, the outer one of the vertical bar and the
// inner one as thick as the arms, a quill bracket the bracket with a third
// arm at the middle.  The double vertical bar is two
// vertical bars.  The large circle is the ring of the circled letters, the
// open bullet a ring of the size of the bullet, the low tilde the tilde of the
// accents below the baseline and the double grave the mirror image of the
// double acute.

// Shape constants of the marks.
const (
	quoteBaseTop = 0.25 // the top of the straight quotation marks on the baseline, in x-heights

	twoThirdsEm    = 2.0 / 3 // the advance of the dash of two thirds of an em, in advances of the em dash
	threeQuarterEm = 3.0 / 4 // the advance of the dash of three quarters of an em, in advances of the em dash

	dblHyphenWhite = 1.55 // the white between the bars of the double hyphen, in thicknesses of the hyphen
	dblHyphenHang  = 0.5  // the advance of the double hyphen as a hyphenation character, in advances of the hyphen

	angleLean  = 23.0 // the strokes of an angle bracket, degrees from the vertical
	anglePoint = 60.0 // the side space of an angle bracket at its point
	angleOpen  = 40.0 // the side space of an angle bracket at its open side

	dblBracketWhite = 0.86 // the white between the two stems of a double bracket, in stems of the bracket
	dblBracketReach = 0.86 // how far the arms of a double bracket reach beyond the inner stem, in reaches of the arms of the bracket

	dblBarWhite = 0.75 // the white between the bars of the double vertical bar, in bars, less dblBarGap
	dblBarGap   = 42.0 // the white between the bars of the double vertical bar beyond dblBarWhite of a bar

	lowTildeMiddle = 0.44 // the middle of the low tilde below the baseline, in descenders
)

// sideHang is the side kind of the right side of the double hyphen as a
// hyphenation character: its ink reaches beyond the edge of the advance by
// what the advance lacks of that of the hyphen.
const sideHang sideKind = 11000

// textMarksList lists the marks of this file and of textsigns.go, in glyph
// order.
var textMarksList = []entry{
	{"quotesinglbase.ts1", privateUseRune("quotesinglbase.ts1"), sideStraight, sideStraight, quoteStraightBase},
	{"quotedblbase.ts1", privateUseRune("quotedblbase.ts1"), sideStraightDouble, sideStraightDouble, quoteStraightDblBase},
	{"twelveudash", privateUseRune("twelveudash"), sideEmDash, sideEmDash, twelveudash},
	{"threequartersemdash", privateUseRune("threequartersemdash"), sideEmDash, sideEmDash, threequartersemdash},
	{"cwmcapital", privateUseRune("cwmcapital"), sideOpen, sideOpen, compoundWordMark},
	{"cwmascender", privateUseRune("cwmascender"), sideOpen, sideOpen, compoundWordMark},
	{"uni2E40", '⹀', sideHyphen, sideHyphen, doubleHyphen},
	{"hyphendbl.alt", privateUseRune("hyphendbl.alt"), sideHyphen, sideHang, doubleHyphenChar},
	{"angleleft", '〈', sideOpen, sideOpen, angleleft},
	{"angleright", '〉', sideOpen, sideOpen, angleright},
	{"uni27E6", '⟦', sideBracket, sideBracket, dblBracketLeft},
	{"uni27E7", '⟧', sideBracketRight, sideBracketRight, dblBracketRight},
	{"uni2045", '⁅', sideBracket, sideBracket, quillLeft},
	{"uni2046", '⁆', sideBracketRight, sideBracketRight, quillRight},
	{"uni2016", '‖', sideBar, sideBar, doubleBar},
	{"uni25EF", '◯', sideOpen, sideOpen, largeCircle},
	{"openbullet", '◦', sideOpen, sideOpen, openBullet},
	{"uni02F7", '˷', sideOpen, sideOpen, lowTilde},
	{"uni02F5", '˵', sideOpen, sideOpen, doubleGrave},
}

func init() {
	punctKindNames[sideHang] = "hang"
}

// inklessHeight returns the height which a glyph without ink has in the font
// metrics of TeX: the x-height for the compound word mark, the cap height and
// the height of the ascenders for the marks of those heights, and zero for
// every other glyph.
func inklessHeight(p design.Params, name string) float64 {
	switch name {
	case "uni200C":
		return p.XHeight
	case "cwmcapital":
		return p.CapHeight
	case "cwmascender":
		return p.Ascender
	}
	return 0
}

// hang returns how far the ink of the double hyphen as a hyphenation
// character reaches beyond the right edge of its advance, at most.
func hang(p design.Params) float64 {
	_, adv := hyphen(p)
	return adv - math.Round(dblHyphenHang*adv) - p.Punct.Hyphen.Right
}

// quoteBaseShift returns the distance by which the straight quotation marks
// are moved down to the baseline.
func quoteBaseShift(p design.Params) float64 {
	return p.Punct.StraightTop - math.Round(quoteBaseTop*p.XHeight)
}

// quoteStraightBase draws the straight single quotation mark on the baseline:
// the straight mark with its top at quoteBaseTop of the x-height.
func quoteStraightBase(p design.Params) (outline.Path, float64) {
	path, adv := quotesingle(p)
	return shiftInk(path, 0, -quoteBaseShift(p)), adv
}

// quoteStraightDblBase draws the straight double quotation mark on the
// baseline.
func quoteStraightDblBase(p design.Params) (outline.Path, float64) {
	path, adv := quotedbl(p)
	return shiftInk(path, 0, -quoteBaseShift(p)), adv
}

// partDash returns the dash whose advance is frac of that of the em dash: the
// bar of the em dash, as long as leaves its side spaces.
func partDash(p design.Params, frac float64) (outline.Path, float64) {
	q := p.Punct
	adv := math.Round(frac * (q.EmLength + q.EmDash.Left + q.EmDash.Right))
	return placeInk(horizontalBar(adv-q.EmDash.Left-q.EmDash.Right, q.DashThick, q.DashCentre), q.EmDash)
}

func twelveudash(p design.Params) (outline.Path, float64)         { return partDash(p, twoThirdsEm) }
func threequartersemdash(p design.Params) (outline.Path, float64) { return partDash(p, threeQuarterEm) }

// doubleHyphen draws the double hyphen: two bars of the hyphen, with
// dblHyphenWhite of their thickness of white between them, as far above the
// middle of the hyphen as below it, with the side spaces of the hyphen.
func doubleHyphen(p design.Params) (outline.Path, float64) {
	q := p.Punct
	half := math.Round((1 + dblHyphenWhite) * q.HyphenThick / 2)
	var path outline.Path
	for _, c := range []float64{q.HyphenCentre - half, q.HyphenCentre + half} {
		addBar(&path, 0, c-q.HyphenThick/2, q.HyphenLength, c+q.HyphenThick/2)
	}
	return placeInk(path, q.Hyphen)
}

// doubleHyphenChar draws the double hyphen as a hyphenation character: the
// double hyphen in dblHyphenHang of the advance of the hyphen.
func doubleHyphenChar(p design.Params) (outline.Path, float64) {
	path, adv := doubleHyphen(p)
	return path, math.Round(dblHyphenHang * adv)
}

// angleLeft returns the left angle bracket, its point at the abscissa 0: two
// strokes of the thickness of the mathematical signs from the middle of the height of the brackets to their top and
// foot, leaning angleLean degrees, with horizontal ends.
func angleLeft(p design.Params) outline.Path {
	top, bottom := p.BracketTop(), p.BracketBottom()
	mid := (top + bottom) / 2
	a := degrees(angleLean)
	run := (top - mid) * math.Tan(a)
	w := p.Punct.MathThick / math.Cos(a)
	var path outline.Path
	addPolygon(&path,
		xy(0, mid), xy(run, bottom), xy(run+w, bottom), xy(w, mid), xy(run+w, top), xy(run, top))
	return path
}

// angleleft draws the left angle bracket.
func angleleft(p design.Params) (outline.Path, float64) {
	return placeInk(angleLeft(p), design.PunctSides{Left: anglePoint, Right: angleOpen})
}

// angleright draws the right angle bracket: the mirror image of the left one.
func angleright(p design.Params) (outline.Path, float64) {
	return placeInk(mirrorInk(angleLeft(p), 0), design.PunctSides{Left: angleOpen, Right: anglePoint})
}

// dblBracket returns the left double bracket: an outer stem of the vertical
// bar and an inner stem as thick as the arms of the bracket, dblBracketWhite
// of the stem of the bracket apart, between the arms of the bracket, which
// reach dblBracketReach of their reach beyond the inner stem.  The white between the stems is a
// counter.
func dblBracket(p design.Params) outline.Path {
	q := p.Punct
	top, bottom := p.BracketTop(), p.BracketBottom()
	s := q.BarThick
	x1 := s + math.Round(dblBracketWhite*q.BracketStem) // the left edge of the inner stem
	x2 := x1 + q.BracketArm
	x := x2 + math.Round(dblBracketReach*q.BracketReach)
	var path outline.Path
	addPolygon(&path,
		xy(0, bottom), xy(x, bottom), xy(x, bottom+q.BracketArm), xy(x2, bottom+q.BracketArm),
		xy(x2, top-q.BracketArm), xy(x, top-q.BracketArm), xy(x, top), xy(0, top))
	rect(&path, s, bottom+q.BracketArm, x1, top-q.BracketArm, false)
	return path
}

// dblBracketLeft draws the left double bracket.
func dblBracketLeft(p design.Params) (outline.Path, float64) {
	return placeInk(dblBracket(p), p.Punct.BracketLeft)
}

// dblBracketRight draws the right double bracket: the mirror image of the
// left one.
func dblBracketRight(p design.Params) (outline.Path, float64) {
	q := p.Punct.BracketLeft
	return placeInk(mirrorInk(dblBracket(p), 0), design.PunctSides{Left: q.Right, Right: q.Left})
}

// quill returns the left quill bracket: the bracket with a third arm at the
// middle of its height.
func quill(p design.Params) outline.Path {
	q := p.Punct
	top, bottom := p.BracketTop(), p.BracketBottom()
	mid := (top + bottom) / 2
	st, a := q.BracketStem, q.BracketArm
	x := st + q.BracketReach
	var path outline.Path
	addPolygon(&path,
		xy(0, bottom), xy(x, bottom), xy(x, bottom+a), xy(st, bottom+a),
		xy(st, mid-a/2), xy(x, mid-a/2), xy(x, mid+a/2), xy(st, mid+a/2),
		xy(st, top-a), xy(x, top-a), xy(x, top), xy(0, top))
	return path
}

// quillLeft draws the left quill bracket.
func quillLeft(p design.Params) (outline.Path, float64) {
	return placeInk(quill(p), p.Punct.BracketLeft)
}

// quillRight draws the right quill bracket: the mirror image of the left
// one.
func quillRight(p design.Params) (outline.Path, float64) {
	q := p.Punct.BracketLeft
	return placeInk(mirrorInk(quill(p), 0), design.PunctSides{Left: q.Right, Right: q.Left})
}

// doubleBar draws the double vertical bar: two vertical bars dblBarWhite of
// their thickness and dblBarGap apart, with the side spaces of the vertical
// bar.
func doubleBar(p design.Params) (outline.Path, float64) {
	q := p.Punct
	var path outline.Path
	for _, x := range []float64{0, q.BarThick + math.Round(dblBarWhite*q.BarThick+dblBarGap)} {
		addBar(&path, x, p.BracketBottom(), x+q.BarThick, p.BracketTop())
	}
	return placeInk(path, q.Bar)
}

// circledRing returns the ring of the circled letters, centred on the
// abscissa 0: from the overshoot below the baseline to the overshoot above
// the cap height, Sym.CircRing (or circRing) thick.
func circledRing(p design.Params) (outer, counter chain) {
	ring := circRing
	if p.Sym.CircRing > 0 {
		ring = p.Sym.CircRing
	}
	return ringOfThickness(p, ring)
}

// ringOfThickness returns the ring of the circled letters, ring thick.
func ringOfThickness(p design.Params, ring float64) (outer, counter chain) {
	bottom, top := -p.OvershootUpper, p.CapHeight+p.OvershootUpper
	r := (top - bottom) / 2
	return ringChains(-r, bottom, r, top, ring, ring, 0.7071, 0.7071)
}

// largeCircle draws the large circle: the ring of the circled letters, or
// Sym.LargeRing thick where that is positive, with their side spaces.
func largeCircle(p design.Params) (outline.Path, float64) {
	if p.Sym.LargeRing > 0 {
		return placeInk(chainsPath(ringOfThickness(p, p.Sym.LargeRing)), design.PunctSides{Left: circSide, Right: circSide})
	}
	return placeInk(chainsPath(circledRing(p)), design.PunctSides{Left: circSide, Right: circSide})
}

// openBullet draws the open bullet: a ring of the size and at the place of
// the bullet, as thick as the ring of the degree sign, with the side spaces
// of the bullet.
func openBullet(p design.Params) (outline.Path, float64) {
	r := bulletScale * p.DotDiameter / 2
	y := p.Punct.MathAxis
	t := p.Sym.DegThick
	outer, counter := ringChains(-r, y-r, r, y+r, t, t, p.DotSuperness, p.DotSuperness)
	return placeInk(chainsPath(outer, counter), design.PunctSides{Left: bulletSide, Right: bulletSide})
}

// lowTilde draws the low tilde: the tilde of the accents with its middle
// lowTildeMiddle of the descender below the baseline, in the advance of the
// spacing accent.
func lowTilde(p design.Params) (outline.Path, float64) {
	path, adv := spacingAccent(p, markTilde)
	_, y0, _, y1 := path.BBox()
	return shiftInk(path, 0, -math.Round(lowTildeMiddle*p.Descender+(y0+y1)/2)), adv
}

// doubleGrave draws the double grave: the mirror image of the double acute.
func doubleGrave(p design.Params) (outline.Path, float64) {
	path, adv := spacingAccent(p, markDblAcute)
	return mirrorInk(path, adv/2), adv
}
