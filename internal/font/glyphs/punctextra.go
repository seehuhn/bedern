// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The further punctuation and typographic marks: ¡ ¿ « » ‹ › ‚ „ … • · † ‡ ¶
// ′ ″ ‴ ‰ ‱ and the zero of ‰, the spaces U+00A0, U+2009, U+200A and U+202F,
// and the soft hyphen.
//
// They are made of the parts of the basic punctuation where one fits: the
// inverted marks are the exclamation and the question mark turned through
// half a turn, with the dot where the upper dot of the colon stands; the low
// quotation marks are the comma with the side spaces and the pitch of the
// quotation marks; the ellipsis is three full stops, the middle dot the full
// stop on the maths axis.  The daggers are drawn in dagger.go.  The others are
// polygons and curves of their own, with the shape constants below.

// Shape constants of the further punctuation.
const (
	guilHeight = 0.846 // the height of a guillemet, in x-heights
	guilMiddle = 0.55  // the middle of a guillemet, in x-heights
	guilWidth  = 0.409 // the width of one chevron, in x-heights
	guilPitch  = 0.415 // the distance between the two chevrons of « and », in x-heights
	guilThick  = 0.48  // the stroke of a chevron at right angles to it, in stems
	guilFlat   = 12.0  // the flat at the point of a chevron
	guilSide   = 45.0  // the side spaces of the guillemets

	bulletScale = 1.85 // the diameter of the bullet, in diameters of the dot
	bulletSide  = 95.0 // the side spaces of the bullet

	pilcrowTop    = 670.0 // the top of ¶
	pilcrowBottom = -150.0
	pilcrowBowl   = 0.45  // the height of the bowl of ¶, in its height
	pilcrowStem   = 0.54  // the stems of ¶, in stems
	pilcrowGap    = 76.0  // the white between the two stems of ¶
	pilcrowBar    = 0.5   // the bar over the white between the stems, in stems
	pilcrowWidth  = 0.592 // the ink width of ¶, in its height
	pilcrowLeft   = 45.0  // the side spaces of ¶
	pilcrowRight  = 40.0

	primeSlant = 27.5 // the lean of the prime, degrees from the vertical
	primeSide  = 56.0 // the side spaces of the primes

	perMilleGap = 50.0 // the white between the two lower zeros of ‰
)

// turned returns the path turned through half a turn about (cx, cy).
func turned(path outline.Path, cx, cy float64) outline.Path {
	var out outline.Path
	for _, c := range pathChains(path) {
		c.rotated180(outline.Point{X: cx, Y: cy}).addContour(&out)
	}
	return out
}

// invertedMark returns the glyph turned through half a turn about the middle
// of its advance, at the height which puts its lowest ink, the dot, where the
// top of the upper dot of the colon stands.
func invertedMark(p design.Params, glyph string) (outline.Path, float64) {
	path, adv, _ := Outline(glyph, p)
	_, y0, _, _ := path.BBox()
	colon, _, _ := Outline("colon", p)
	_, _, _, top := colon.BBox()
	return turned(path, adv/2, (top+y0)/2), adv
}

func exclamdown(p design.Params) (outline.Path, float64)   { return invertedMark(p, "exclam") }
func questiondown(p design.Params) (outline.Path, float64) { return invertedMark(p, "question") }

// chevron adds the chevron of a guillemet pointing to the left, its point at
// the abscissa x: two straight strokes meeting at the point, which is cut flat,
// with vertical cuts at the free ends.
func chevron(path *outline.Path, p design.Params, x float64) {
	h, w := guilHeight*p.XHeight/2, guilWidth*p.XHeight
	ay := guilMiddle * p.XHeight
	// The outer edges run from the point at (x, ay) to the corners of the box
	// at (x + w, ay ± h); the strokes have the angle beta to the horizontal.
	beta := math.Atan2(h, w)
	tan, sin := math.Tan(beta), math.Sin(beta)
	v := guilThick * p.Stem / (2 * math.Cos(beta)) // the vertical half thickness
	// The outer edges, moved by v, run through the point; the point is cut
	// where they are guilFlat apart.
	xl := x + guilFlat/(2*tan)
	xr := x + w
	xi := x + guilThick*p.Stem/sin // the crotch of the inner edges
	yo := func(x1 float64) float64 { return tan * (x1 - x) }
	addPolygon(path,
		xy(xl, ay+guilFlat/2), xy(xl, ay-guilFlat/2),
		xy(xr, ay-yo(xr)), xy(xr, ay-yo(xr)+2*v), xy(xi, ay),
		xy(xr, ay+yo(xr)-2*v), xy(xr, ay+yo(xr)))
}

// guillemets returns the left guillemet with n chevrons, or the right one, the
// mirror image, if right is set.
func guillemets(p design.Params, n int, right bool) (outline.Path, float64) {
	var path outline.Path
	for i := range n {
		chevron(&path, p, float64(i)*guilPitch*p.XHeight)
	}
	if right {
		path = mirrorInk(path, 0)
	}
	return placeInk(path, design.PunctSides{Left: guilSide, Right: guilSide})
}

func guillemotleft(p design.Params) (outline.Path, float64)  { return guillemets(p, 2, false) }
func guillemotright(p design.Params) (outline.Path, float64) { return guillemets(p, 2, true) }
func guilsinglleft(p design.Params) (outline.Path, float64)  { return guillemets(p, 1, false) }
func guilsinglright(p design.Params) (outline.Path, float64) { return guillemets(p, 1, true) }

// quotesinglbase draws the low single quotation mark: the comma, with the side
// spaces of the quotation marks.
func quotesinglbase(p design.Params) (outline.Path, float64) {
	var path outline.Path
	commaAt(p, 0, commaTop(p)).addContour(&path)
	return placeInk(path, p.Punct.Quote)
}

// quotedblbase draws the low double quotation mark: two commas, as far apart
// as the marks of the double quotation marks.
func quotedblbase(p design.Params) (outline.Path, float64) {
	var path outline.Path
	commaAt(p, 0, commaTop(p)).addContour(&path)
	commaAt(p, p.Punct.QuotePitch, commaTop(p)).addContour(&path)
	return placeInk(path, p.Punct.QuoteDouble)
}

// ellipsis draws the ellipsis: three full stops, the dots as far apart as in
// three full stops set side by side.
func ellipsis(p design.Params) (outline.Path, float64) {
	var path outline.Path
	pitch := p.DotsAdvance()
	for i := range 3 {
		addDot(&path, p, float64(i)*pitch, p.PeriodCentre())
	}
	return placeInk(path, p.Punct.Period)
}

// periodcentered draws the middle dot: the full stop with its centre on the
// maths axis.
func periodcentered(p design.Params) (outline.Path, float64) {
	var path outline.Path
	addDot(&path, p, 0, p.Punct.MathAxis)
	return placeInk(path, p.Punct.Period)
}

// bullet draws the bullet: the dot of i, bulletScale times as large, with its
// centre on the maths axis.
func bullet(p design.Params) (outline.Path, float64) {
	var path outline.Path
	r := bulletScale * p.DotDiameter / 2
	superEllipse(&path, -r, p.Punct.MathAxis-r, r, p.Punct.MathAxis+r, p.DotSuperness, true)
	return placeInk(path, design.PunctSides{Left: bulletSide, Right: bulletSide})
}

// paragraph draws the pilcrow: two stems from the top to below the baseline
// with a bar between them at the top, and on the left of the left stem a
// solid bowl, half a super-ellipse, from the top down to pilcrowBowl of the
// height.  It is one contour.
func paragraph(p design.Params) (outline.Path, float64) {
	h := pilcrowTop - pilcrowBottom
	ws := pilcrowStem * p.Stem
	w := pilcrowWidth * h
	x2 := w - ws // the left-hand edge of the right stem
	x1 := x2 - pilcrowGap - ws
	yb := pilcrowTop - pilcrowBowl*h
	bar := pilcrowTop - pilcrowBar*p.Stem
	x0 := 0.0 // the left extreme of the bowl
	q := newSuperShape(x0, yb, 2*x1-x0, pilcrowTop, p.SupernessOuter)
	q1, q2 := q.cubic(1, true), q.cubic(2, true) // top to the left extreme, and on to the bottom
	var path outline.Path
	path.MoveTo(x1, pilcrowBottom)
	path.LineTo(x1+ws, pilcrowBottom)
	path.LineTo(x1+ws, bar)
	path.LineTo(x2, bar)
	path.LineTo(x2, pilcrowBottom)
	path.LineTo(x2+ws, pilcrowBottom)
	path.LineTo(x2+ws, pilcrowTop)
	path.LineTo(q1.P0.X, q1.P0.Y)
	path.CurveTo(q1.args())
	path.CurveTo(q2.args())
	path.LineTo(x1, pilcrowBottom)
	return placeInk(path, design.PunctSides{Left: pilcrowLeft, Right: pilcrowRight})
}

// primes returns n primes: the tapered bar of the ASCII apostrophe, PrimeWeight
// times as wide, leaning primeSlant degrees to the right, as far apart as the
// bars of the ASCII double quote.
func primes(p design.Params, n int) (outline.Path, float64) {
	q := p.Punct
	h := q.StraightHeight
	wb, wt := q.PrimeWeight*q.StraightBottomWidth, q.PrimeWeight*q.StraightTopWidth
	top := p.Ascender
	lean := math.Tan(primeSlant*math.Pi/180) * h
	var path outline.Path
	for i := range n {
		x := float64(i) * q.StraightPitch
		addPolygon(&path,
			xy(x-wb/2, top-h), xy(x+wb/2, top-h),
			xy(x+lean+wt/2, top), xy(x+lean-wt/2, top))
	}
	return placeInk(path, design.PunctSides{Left: primeSide, Right: primeSide})
}

func minute(p design.Params) (outline.Path, float64) { return primes(p, 1) }
func second(p design.Params) (outline.Path, float64) { return primes(p, 2) }
func third(p design.Params) (outline.Path, float64)  { return primes(p, 3) }

// perthousand draws the per mille sign: the percent sign with a third zero,
// the lower one repeated perMilleGap to its right.
func perthousand(p design.Params) (outline.Path, float64) { return perZeros(p, 1) }

// permyriad draws the per ten thousand sign: the percent sign with its lower
// zero repeated twice.
func permyriad(p design.Params) (outline.Path, float64) { return perZeros(p, 2) }

// perZeros returns the percent sign with its lower zero repeated n times,
// each perMilleGap to the right of the one before.
func perZeros(p design.Params, n int) (outline.Path, float64) {
	path, adv, _ := Outline("percent", p)
	lower, dx := lowerZero(p, path)
	for i := 1; i <= n; i++ {
		path = append(path, shiftInk(lower, float64(i)*dx, 0)...)
	}
	return path, adv + float64(n)*dx
}

// lowerZero returns the lower zero of the percent sign and the distance by
// which a repetition of it is moved to the right.
func lowerZero(p design.Params, percent outline.Path) (lower outline.Path, dx float64) {
	lx0, lx1 := math.Inf(1), math.Inf(-1)
	for _, c := range percent {
		cp := outline.Path{c}
		x0, y0, x1, y1 := cp.BBox()
		if y0+y1 < p.Figures.Height && x1-x0 < p.Figures.Advance {
			lower = append(lower, c)
			lx0, lx1 = math.Min(lx0, x0), math.Max(lx1, x1)
		}
	}
	return lower, math.Round(lx1 - lx0 + perMilleGap)
}

// perthousandzero draws the zero which, set after the percent sign, makes the
// per mille sign of it, and set after that the per ten thousand sign: the
// lower zero of the percent sign, its leftmost ink at the origin, in an
// advance of the distance between two zeros.
func perthousandzero(p design.Params) (outline.Path, float64) {
	path, adv, _ := Outline("percent", p)
	lower, dx := lowerZero(p, path)
	return shiftInk(lower, dx-adv, 0), dx
}

// The spaces have no outline.  The no-break space has the advance of the word
// space; the thin space and the narrow no-break space half of it, the hair
// space an eighth.
func nbspace(p design.Params) (outline.Path, float64)    { return nil, p.SpaceAdvance }
func thinspace(p design.Params) (outline.Path, float64)  { return nil, math.Round(p.SpaceAdvance / 2) }
func hairspace(p design.Params) (outline.Path, float64)  { return nil, math.Round(p.SpaceAdvance / 8) }
func softhyphen(p design.Params) (outline.Path, float64) { return hyphen(p) }
