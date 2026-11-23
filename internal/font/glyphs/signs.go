// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"sync"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The signs of Latin-1 and of text: × ÷ ± ¬ µ ¹ ² ³ ¼ ½ ¾ ⁄ ª º ¦ © ® ¢ ¥ ¤.
//
// The mathematical signs are drawn with the strokes of plus and minus
// (Punct.MathThick) on the maths axis and share their advance.  The superior
// figures are the figures drawn with heavier strokes and scaled down: every
// stroke of the figures is Figures.SupWeight/supScale times as thick, and the
// figure is then scaled by supScale, so that its strokes are Figures.SupWeight
// of those of the figures where a plain scaling would leave supScale.  The fractions are
// a superior figure, the fraction slash and a superior figure on the
// baseline.  The other signs are made of the letters, scaled, with bars and
// rings of their own.

// Shape constants of the signs.
const (
	supScale = 0.61  // the size of a superior figure, in sizes of the figure: the least at which the white beside the foot of 2 keeps the clearance floor
	supBase  = 0.435 // the baseline of a superior figure, in figure heights
	supAdv   = 0.666 // the advance of a superior figure, in advances of the figures

	fracLean  = 31.8 // the fraction slash, degrees from the vertical
	fracThick = 44.0 // the fraction slash, at right angles to it
	fracGap   = 55.0 // the least white between the fraction slash and the figures
	fracSide  = 10.0 // the side spaces of the fraction slash

	timesWidth  = 0.807 // the ink width of ×, in ink widths of +
	divideDot   = 0.585 // the centres of the dots of ÷ from the axis, in half widths of +
	pmTop       = 576.0 // the top of ±
	pmGap       = 82.0  // the white between the lower bar and the upper sign of ±, that between the bars of =
	notTop      = 370.0 // the top of the bar of ¬
	notBottom   = 133.0 // the foot of the hook of ¬
	brokenGap   = 150.0 // the white in the middle of ¦
	ordScale    = 0.7   // the letters of ª and º, in sizes of a and o
	ordWeight   = 0.76  // the strokes of the letters of ª and º, in strokes of a and o
	ordBottom   = 331.0 // the underline of ª and º
	ordGap      = 55.0  // the white between the underline and the letter
	ordSide     = 30.0  // the side spaces of ª and º
	circRing    = 40.0  // the ring of © and ® where Sym.CircRing is zero
	circGap     = 21.0  // the least distance from the farthest ink of the letter of © and ® to the inside of the ring
	circSide    = 35.0  // the side spaces of © and ®
	centStroke  = 50.0  // the stroke of ¢, as that of $
	centTop     = 560.0 // the top of the stroke of ¢
	centBottom  = -30.0 // the foot of the stroke of ¢
	yenScale    = 0.975 // the Y of ¥, in sizes of Y: the height of the figures
	yenFork     = 0.59  // the crotch of ¥, in its height
	yenBarLow   = 0.33  // the middle of the lower bar of ¥, in its height
	yenBarHigh  = 0.49  // the middle of the upper bar of ¥, in its height
	yenBar      = 0.85  // the bars of ¥, in crossbars of the capitals
	yenBarWidth = 0.70  // the bars of ¥, in its ink width
	curDiam     = 0.86  // the outer diameter of the ring of ¤, in its ink width
	curWidth    = 0.62  // the ink width of ¤, in figure heights
	curMiddle   = 0.686 // the middle of ¤, in x-heights
	curSide     = 0.74  // the ring of ¤ at its sides, in stems
	curTop      = 1.05  // the ring of ¤ at its top and bottom, in rings of the degree sign
)

// scaledPath returns the path scaled by s about the origin.
func scaledPath(path outline.Path, s float64) outline.Path {
	var out outline.Path
	for _, c := range pathChains(path) {
		c.mapPoints(func(q outline.Point) outline.Point { return outline.Point{X: s * q.X, Y: s * q.Y} }).addContour(&out)
	}
	return out
}

// mathPlace returns the path centred in the advance of the mathematical signs.
func mathPlace(p design.Params, path outline.Path) (outline.Path, float64) {
	return placeCentred(path, p.MathAdvance())
}

// multiply draws ×: two bars of the stroke of plus crossing at right angles on
// the maths axis, along the diagonals of a square timesWidth as wide as plus,
// with their ends cut at right angles.
func multiply(p design.Params) (outline.Path, float64) {
	q := p.Punct
	w := timesWidth * q.MathWidth
	// The ink of a bar along a diagonal reaches (L + T)/√2 across.
	l := w*math.Sqrt2 - q.MathThick
	ay := q.MathAxis
	bar := func(a float64) chain {
		d := outline.Point{X: math.Cos(a) * l / 2, Y: math.Sin(a) * l / 2}
		return taperBar(outline.Point{X: -d.X, Y: ay - d.Y}, outline.Point{X: d.X, Y: ay + d.Y}, q.MathThick)
	}
	return mathPlace(p, chainsPath(union(bar(math.Pi/4), bar(3*math.Pi/4))...))
}

// taperBar returns the straight bar from a to b, thick at right angles to it,
// with ends cut at right angles.
func taperBar(a, b outline.Point, thick float64) chain {
	d := unit(a, b)
	n := outline.Point{X: -d.Y * thick / 2, Y: d.X * thick / 2}
	return polygonChain([]outline.Point{
		{X: a.X - n.X, Y: a.Y - n.Y}, {X: b.X - n.X, Y: b.Y - n.Y}, {X: b.X + n.X, Y: b.Y + n.Y}, {X: a.X + n.X, Y: a.Y + n.Y},
	})
}

// divide draws ÷: the bar of minus with the dot of the full stop above and
// below it, their centres divideDot of half the width of plus from the axis.
func divide(p design.Params) (outline.Path, float64) {
	q := p.Punct
	h := q.MathWidth / 2
	var path outline.Path
	rect(&path, -h, q.MathAxis-q.MathThick/2, h, q.MathAxis+q.MathThick/2, true)
	addDot(&path, p, 0, q.MathAxis+divideDot*h)
	addDot(&path, p, 0, q.MathAxis-divideDot*h)
	return mathPlace(p, path)
}

// plusminus draws ±: the bar of minus on the baseline and above it a plus as
// wide, whose vertical bar reaches from pmGap above the lower bar to pmTop,
// with its horizontal bar in the middle of it.
func plusminus(p design.Params) (outline.Path, float64) {
	q := p.Punct
	h, t := q.MathWidth/2, q.MathThick
	y0 := t + pmGap
	ym := (y0 + pmTop) / 2
	var path outline.Path
	rect(&path, -h, 0, h, t, true)
	addPolygon(&path,
		xy(-t/2, y0), xy(t/2, y0), xy(t/2, ym-t/2), xy(h, ym-t/2), xy(h, ym+t/2), xy(t/2, ym+t/2),
		xy(t/2, pmTop), xy(-t/2, pmTop), xy(-t/2, ym+t/2), xy(-h, ym+t/2), xy(-h, ym-t/2), xy(-t/2, ym-t/2))
	return mathPlace(p, path)
}

// logicalnot draws ¬: the bar of minus with its top at notTop and a hook of
// the same stroke down from its right end to notBottom.
func logicalnot(p design.Params) (outline.Path, float64) {
	q := p.Punct
	h, t := q.MathWidth/2, q.MathThick
	var path outline.Path
	addPolygon(&path,
		xy(-h, notTop-t), xy(h-t, notTop-t), xy(h-t, notBottom), xy(h, notBottom), xy(h, notTop), xy(-h, notTop))
	return mathPlace(p, path)
}

// mu draws µ: the u whose left stem runs on below the baseline to the
// descender line, where it ends round (roundEnd); the outer edge of the bowl
// leaves the stem at its right-hand edge.
func mu(p design.Params) (outline.Path, float64) {
	left := stemSpec{
		Left: p.SideStemLeft, Width: p.Stem, Top: p.XHeight,
		HeadJut: p.SerifJut, HeadDrop: p.HeadSerifDrop,
	}
	right := stemSpec{
		Left: left.right() + p.UCounter, Width: p.Stem, Top: p.XHeight,
		FootRight: p.SerifJut,
		HeadJut:   p.SerifJut, HeadDrop: p.HeadSerifDrop,
	}
	arch, cx := uArch(p, left, right)
	turn := func(c chain) chain { return c.flipY(p.XHeight / 2).mirrorX(cx) }
	outer := turn(archOuter(nArchShape, arch))
	// The outer edge of the bowl from where it crosses the right-hand edge of
	// the left stem.
	var rest chain
	for i, pc := range outer {
		if pc.P0.X <= left.right() && pc.P3.X > left.right() {
			c := pc.cubic.sub(pc.cubic.tAtX(left.right()), 1)
			if pc.line {
				c = cubic{outline.Point{X: left.right(), Y: pc.P0.Y + (left.right()-pc.P0.X)*(pc.P3.Y-pc.P0.Y)/(pc.P3.X-pc.P0.X)}, outline.Point{}, outline.Point{}, pc.P3}
				rest = append(chain{lineSeg(c.P0, c.P3)}, outer[i+1:]...)
			} else {
				rest = append(chain{curveSeg(c)}, outer[i+1:]...)
			}
			break
		}
	}
	var path outline.Path
	path.MoveTo(right.start())
	stemUp(&path, p, right, right.Top)
	stemCap(&path, p, right)
	path.LineTo(right.Left, p.UJoin)
	turn(archUnder(nArchShape, arch)).appendTo(&path)
	stemCap(&path, p, left)
	end := station{X: left.Left + p.Stem/2, Y: -p.Descender + roundEndReach(p.Stem), W: p.Stem}
	path.LineTo(left.Left, end.Y)
	roundEnd(end, outline.Point{X: 0, Y: -1}).appendTo(&path)
	path.LineTo(rest.start().X, rest.start().Y)
	rest.appendTo(&path)
	path.LineTo(right.start())
	return path, design.UAdvance(p)
}

// supParams returns the parameters of the superior figures: the strokes of
// the figures Figures.SupWeight/supScale times as thick.
func supParams(p design.Params) design.Params {
	q := p
	f := &q.Figures
	k := f.SupWeight / supScale
	for _, v := range []*float64{&f.Stem, &f.Round, &f.Thin, &f.Heavy, &f.Bowl, &f.Ball,
		&f.OneFlagThick, &f.TwoNeckThick, &f.TwoSerifThick, &f.ThreeGap, &f.ThreeUpperBowl,
		&f.FourDiag, &f.FourBar, &q.SerifThick, &q.Caps.SerifThick} {
		*v *= k
	}
	// The lower ball of 3, heavier, moves to the left to keep its white.
	f.ThreeLowerBall -= 12
	return q
}

// heavierLetters returns the parameters with the strokes of the letters k
// times as thick, for the letters of ª, º, © and ®, which are drawn so and
// scaled down, as the superior figures are.
func heavierLetters(p design.Params, k float64) design.Params {
	q := p
	c := &q.Caps
	for _, v := range []*float64{&q.Stem, &q.AscenderStem, &q.CurveLower, &q.ThinLower, &q.Hair,
		&q.SerifThick, &q.EBarThick, &q.ATailTipThick, &q.ArchThinJoin,
		&q.BowlJoinThick,
		&c.Stem, &c.Round, &c.RoundThin, &c.Bar, &c.TBar, &c.SerifThick, &c.ArmSerifThick,
		&c.BowlRound, &c.RBowlRound, &c.RLeg} {
		*v *= k
	}
	return q
}

// superiorInk returns the superior figure drawn from the glyph of the figure,
// with its baseline at the height base, centred on the abscissa 0.
func superiorInk(p design.Params, figure string, base float64) outline.Path {
	return shiftInk(supFigure(p, figure), 0, base)
}

// supFigureKey names a superior figure in one set of parameters.
type supFigureKey struct {
	p      design.Params
	figure string
}

// supFigures holds the superior figures by the parameters and the figure, so
// that the superior figures and the fractions of one set of parameters draw
// each figure once.
var supFigures sync.Map // supFigureKey -> func() outline.Path

// supFigure returns the figure drawn with supParams, scaled by supScale, with
// its baseline at 0 and centred on the abscissa 0.  The path is shared and
// must not be changed.
func supFigure(p design.Params, figure string) outline.Path {
	f, _ := supFigures.LoadOrStore(supFigureKey{p, figure}, sync.OnceValue(func() outline.Path {
		path, adv, _ := Outline(figure, supParams(p))
		return scaledPath(shiftInk(path, -adv/2, 0), supScale)
	}))
	return f.(func() outline.Path)()
}

// supAdvance returns the advance of a superior figure.
func supAdvance(p design.Params) float64 { return math.Round(supAdv * p.Figures.Advance) }

// superior draws a superior figure, centred in its advance.
func superior(p design.Params, figure string) (outline.Path, float64) {
	return placeCentred(superiorInk(p, figure, math.Round(supBase*p.Figures.Height)), supAdvance(p))
}

func onesuperior(p design.Params) (outline.Path, float64)   { return superior(p, "one") }
func twosuperior(p design.Params) (outline.Path, float64)   { return superior(p, "two") }
func threesuperior(p design.Params) (outline.Path, float64) { return superior(p, "three") }

// fractionBar returns the fraction slash, centred on the abscissa 0: a
// parallelogram fracThick across, leaning fracLean degrees, from the bottom
// to the top of the round figures, with horizontal ends.
func fractionBar(p design.Params) outline.Path {
	bottom, top := figBottom(p), figTop(p)
	return chainsPath(slashStroke(0, (bottom+top)/2, bottom, top, 90-fracLean, fracThick))
}

// fraction draws the fraction slash with fracSide on each side.
func fraction(p design.Params) (outline.Path, float64) {
	return placeInk(fractionBar(p), design.PunctSides{Left: fracSide, Right: fracSide})
}

// shiftUntilClear returns the least whole number of units by which b must be
// moved to the right of its place so that it keeps the distance gap from a.
func shiftUntilClear(a, b outline.Path, gap float64) float64 {
	lo, hi := 0.0, 1200.0
	for hi-lo > 0.5 {
		mid := (lo + hi) / 2
		if pathsDistance(a, shiftInk(b, mid, 0), gap) < gap {
			lo = mid
		} else {
			hi = mid
		}
	}
	return math.Ceil(hi)
}

// vulgarFraction draws a fraction: the numerator, a superior figure, in its
// advance; the fraction slash moved to the right until it keeps fracGap from
// the numerator; and the denominator, a superior figure with its baseline on
// the baseline, moved to the right until it keeps fracGap from the slash.
func vulgarFraction(p design.Params, num, den string) (outline.Path, float64) {
	n, adv := superior(p, num)
	bar := fractionBar(p)
	bar = shiftInk(bar, shiftUntilClear(n, bar, fracGap), 0)
	d := placeCentredPath(superiorInk(p, den, 0), adv)
	d = shiftInk(d, shiftUntilClear(bar, d, fracGap), 0)
	_, _, nx1, _ := n.BBox()
	// The right side space is that of the numerator.
	out := append(append(n, bar...), d...)
	_, _, x1, _ := out.BBox()
	return out, math.Round(x1 + (adv - nx1))
}

// placeCentredPath returns the path centred in an advance of the given width.
func placeCentredPath(path outline.Path, adv float64) outline.Path {
	out, _ := placeCentred(path, adv)
	return out
}

func onequarter(p design.Params) (outline.Path, float64) { return vulgarFraction(p, "one", "four") }
func onehalf(p design.Params) (outline.Path, float64)    { return vulgarFraction(p, "one", "two") }
func threequarters(p design.Params) (outline.Path, float64) {
	return vulgarFraction(p, "three", "four")
}

// ordinal draws ª or º: the letter drawn heavier and scaled by ordScale, so
// that its strokes are supWeight of those of the letter, over an underline of
// the thickness of the bar of e, ordGap below it, at ordBottom.
func ordinal(p design.Params, letter string) (outline.Path, float64) {
	path, _, _ := Outline(letter, heavierLetters(p, ordWeight/ordScale))
	x0, y0, x1, _ := path.BBox()
	path = scaledPath(shiftInk(path, -(x0+x1)/2, -y0), ordScale)
	base := ordBottom + p.EBarThick + ordGap
	path = shiftInk(path, 0, base)
	lx0, _, lx1, _ := path.BBox()
	rect(&path, lx0, ordBottom, lx1, ordBottom+p.EBarThick, true)
	return placeInk(path, design.PunctSides{Left: ordSide, Right: ordSide})
}

func ordfeminine(p design.Params) (outline.Path, float64)  { return ordinal(p, "a") }
func ordmasculine(p design.Params) (outline.Path, float64) { return ordinal(p, "o") }

// brokenbar draws ¦: the vertical bar with a gap of brokenGap about the maths
// axis.
func brokenbar(p design.Params) (outline.Path, float64) {
	q := p.Punct
	var path outline.Path
	w := q.BarThick / 2
	rect(&path, -w, p.BracketBottom(), w, q.MathAxis-brokenGap/2, true)
	rect(&path, -w, q.MathAxis+brokenGap/2, w, p.BracketTop(), true)
	return placeInk(path, q.Bar)
}

// circled draws © or ®: a ring from the overshoot below the baseline to the
// overshoot above the cap height, Sym.CircRing (or circRing) thick, and in it the capital drawn
// heavier and scaled as a superior figure, centred, and scaled less where it
// would come closer than circGap to the ring.
func circled(p design.Params, letter string) (outline.Path, float64) {
	bottom, top := -p.OvershootUpper, p.CapHeight+p.OvershootUpper
	r := (top - bottom) / 2
	cy := (top + bottom) / 2
	ring := circRing
	if p.Sym.CircRing > 0 {
		ring = p.Sym.CircRing
	}
	outer, counter := ringChains(-r, bottom, r, top, ring, ring, 0.7071, 0.7071)
	path := chainsPath(outer, counter)
	l, _, _ := Outline(letter, heavierLetters(p, p.Figures.SupWeight/supScale))
	lx0, _, lx1, _ := l.BBox()
	l = shiftInk(l, -(lx0+lx1)/2, -p.CapHeight/2)
	// The letter is scaled by supScale, or less where its farthest ink would
	// come closer than circGap to the inside of the ring.
	reach := 0.0
	for _, c := range l {
		for _, q := range c.Flatten() {
			reach = math.Max(reach, math.Hypot(q.X, q.Y))
		}
	}
	l = scaledPath(l, math.Min(supScale, (r-ring-circGap)/reach))
	path = append(path, shiftInk(l, 0, cy)...)
	return placeInk(path, design.PunctSides{Left: circSide, Right: circSide})
}

func copyright(p design.Params) (outline.Path, float64)  { return circled(p, "C") }
func registered(p design.Params) (outline.Path, float64) { return circled(p, "R") }

// cent draws ¢: the c, centred in the advance of the figures, with a vertical
// stroke through the middle of its bowl from centBottom to centTop.
func cent(p design.Params) (outline.Path, float64) {
	c, _, _ := Outline("c", p)
	c, adv := placeCentred(c, p.Figures.Advance)
	x0, _, x1, _ := c.BBox()
	cx := math.Round((x0 + x1) / 2)
	parts := append(pathChains(c), rectChain(cx-centStroke/2, centBottom, cx+centStroke/2, centTop))
	return chainsPath(union(parts...)...), adv
}

// yen draws ¥: the Y drawn with its ink as wide as the advance of the figures
// and its crotch at yenFork of its height, scaled by yenScale, so that it is
// as high as the figures, with two bars of the crossbar of the capitals
// across its stem below the crotch; centred in the advance of the figures.
func yen(p design.Params) (outline.Path, float64) {
	q := p
	q.CapDiag.YInk = p.Figures.Advance / yenScale
	q.CapDiag.YJunction = yenFork * p.CapHeight
	y, _ := capY(q)
	y = scaledPath(y, yenScale)
	x0, _, x1, top := y.BBox()
	cx, w := (x0+x1)/2, yenBarWidth*(x1-x0)
	parts := pathChains(y)
	for _, f := range []float64{yenBarLow, yenBarHigh} {
		yy := f * top
		parts = append(parts, rectChain(cx-w/2, yy-yenBar*p.Caps.Bar/2, cx+w/2, yy+yenBar*p.Caps.Bar/2))
	}
	path := chainsPath(union(parts...)...)
	return placeCentred(path, p.Figures.Advance)
}

// currencyWeights returns the ring of ¤ at its sides and at its top and
// bottom, and its spokes, in whole units.
func currencyWeights(p design.Params) (side, top, spoke float64) {
	return math.Round(curSide * p.Stem), math.Round(curTop * p.Sym.DegThick), p.Punct.MathThick
}

// currency draws ¤: a ring, curDiam of the ink width across, curSide stems
// thick at its sides and curTop rings of the degree sign at its top and
// bottom, with four short spokes of the stroke of the mathematical signs on
// its diagonals out to the corners of the ink, centred at curMiddle of the
// x-height in the advance of the figures.
func currency(p design.Params) (outline.Path, float64) {
	w := curWidth * p.Figures.Height
	d := curDiam * w
	cy := curMiddle * p.XHeight
	side, top, spoke := currencyWeights(p)
	outer, counter := ringChains(-d/2, cy-d/2, d/2, cy+d/2, side, top, 0.7071, 0.7071)
	parts := []chain{outer, counter}
	// A spoke runs from inside the ring out to the corner of the ink.
	reach := (w - spoke/math.Sqrt2) / 2
	for _, s := range [][2]float64{{1, 1}, {-1, 1}, {-1, -1}, {1, -1}} {
		a := outline.Point{X: s[0] * (d/2 - top/2) / math.Sqrt2, Y: cy + s[1]*(d/2-top/2)/math.Sqrt2}
		b := outline.Point{X: s[0] * reach, Y: cy + s[1]*reach}
		parts = append(parts, taperBar(a, b, spoke))
	}
	return placeCentred(chainsPath(union(parts...)...), p.Figures.Advance)
}
