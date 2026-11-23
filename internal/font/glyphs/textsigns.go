// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"sync"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The marks of the text symbols which have a form of their own: the blank
// symbol, the commercial minus, the mark ※, the pilcrow with one stem
// and the interrobangs.
//
// The blank symbol is the b with a short hairline stroke rising to the right
// across its ascender.  The commercial minus is a
// diagonal of the stroke of the percent sign from the baseline to the height
// of the figures between two dots of the full stop, in the advance of the
// figures.  The mark ※ is two strokes of the mathematical signs which
// cross at right angles on the maths axis, with a dot of the full stop in
// each of the four angles.  The pilcrow with one stem is the solid bowl of the
// pilcrow on a single stem with the foot serifs of the lowercase.  The
// interrobang is the hook of the question mark, which runs straight into the
// stem of the exclamation mark; the stem stands under the top of the hook,
// over the dot, and reaches above the hook.  The inverted interrobang is
// the interrobang turned through half a turn, as deep as the inverted
// question mark.

// Shape constants of the signs.
const (
	blankCentre = 0.33 // where the stroke of the blank symbol crosses the middle of the stem, of the way from the x-height to the ascender
	blankAngle  = 25.0 // the stroke of the blank symbol, degrees above the horizontal
	blankRight  = 1.3  // how far the stroke of the blank symbol reaches to the right of the stem, in stems

	discountInk = 0.776 // the ink width of the commercial minus, in advances of the figures
	discountDot = 14.0  // how far the dots of the commercial minus stand inside the height of the figures and above the baseline

	refMarkSize = 1.24 // the width and the height of the mark ※, in x-heights
	refMarkDot  = 0.23 // the dots of the mark ※, in widths of the sign, at most
	refMarkSide = 56.0 // the side spaces of the mark ※

	pilcrowAltStem = 0.86  // the stem of the pilcrow with one stem, in stems
	pilcrowAltBowl = 245.0 // the width of its bowl, to the left of the stem

	interroRise    = 0.045 // how far the stem of the interrobang stands above the hook, in heights of the hook
	interroStem    = 0.8   // the stem of the interrobang at its top, in widths of the top of the exclamation mark, at most
	interroClear   = 4.0   // the white between the stem of the interrobang and its ball, above the clearance floor, at least
	interroCounter = 0.62  // the white between the stem and the right side of the hook, of the distance from the middle line of the stem to that side, at least
	interroJoin    = 0.04  // how far above the foot of the stem the hook of the interrobang ends, in heights of the hook
	interroAngle   = 40.0  // the direction in which the hook runs into the stem, degrees below the horizontal
	interroEnd     = 0.5   // the hook of the interrobang where it ends, in round strokes of the lowercase
)

var textSignsList = []entry{
	{"uni2422", '␢', sideStemB, sideBowlBP, blank},
	{"uni2052", '⁒', sideFigure, sideFigure, discount},
	{"uni203B", '※', sideOpen, sideOpen, referenceMark},
	{"paragraph.alt", privateUseRune("paragraph.alt"), sideOpen, sideOpen, pilcrowAlt},
	{"uni203D", '‽', sideOpen, sideOpen, interrobang},
	{"uni2E18", '⸘', sideOpen, sideOpen, interrobangDown},
}

// blank draws the blank symbol: the b with a stroke of the hairline (Hair)
// across its ascender, which rises blankAngle degrees to the right, from under
// the tip of the head serif to blankRight stems beyond the stem.
func blank(p design.Params) (outline.Path, float64) {
	left := design.BStemLeft(p)
	x0, x1 := left-p.SerifJut, left+p.AscenderStem+blankRight*p.Stem
	cx := (x0 + x1) / 2
	cy := p.XHeight + blankCentre*(p.Ascender-p.XHeight) + math.Tan(degrees(blankAngle))*(cx-left-p.AscenderStem/2)
	return withBars(p, "b", slantBar(cx, cy, x1-x0, blankAngle, p.Hair))
}

// discount draws the commercial minus: a diagonal of the stroke of the
// percent sign from the baseline to the height of the figures, with
// horizontal ends, and two dots of the full stop in the corners of its box,
// the upper one at the left; centred in the advance of the figures.
func discount(p design.Params) (outline.Path, float64) {
	h := p.Figures.Height
	w := math.Round(discountInk * p.Figures.Advance)
	t := p.Sym.PctSlash
	// The stroke leans by the angle a from the vertical at which its ink is w
	// wide: h tan a + t / cos a = w.
	a := math.Acos(t/math.Hypot(w, h)) - math.Atan2(h, w)
	wh := t / math.Cos(a)
	r := p.DotDiameter / 2
	var path outline.Path
	addPolygon(&path, xy(0, 0), xy(wh, 0), xy(w, h), xy(w-wh, h))
	addDot(&path, p, r, h-discountDot-r)
	addDot(&path, p, w-r, discountDot+r)
	return placeCentred(path, p.Figures.Advance)
}

// referenceMark draws the mark ※: two bars of the stroke of the
// mathematical signs along the diagonals of a square refMarkSize of the
// x-height wide, which cross on the maths axis, with ends cut at right
// angles, and four dots of the shape of the full stop which touch the sides
// of the square at their middles.  A dot has the diameter of the full stop
// or, where that is less, refMarkDot of the width of the square.
func referenceMark(p design.Params) (outline.Path, float64) {
	q := p.Punct
	w := math.Round(refMarkSize * p.XHeight)
	l := w*math.Sqrt2 - q.MathThick
	ay := q.MathAxis
	bar := func(a float64) chain {
		d := outline.Point{X: math.Cos(a) * l / 2, Y: math.Sin(a) * l / 2}
		return taperBar(outline.Point{X: -d.X, Y: ay - d.Y}, outline.Point{X: d.X, Y: ay + d.Y}, q.MathThick)
	}
	path := chainsPath(union(bar(math.Pi/4), bar(3*math.Pi/4))...)
	r := math.Min(p.DotDiameter, math.Round(refMarkDot*w)) / 2
	d := w/2 - r
	for _, c := range [][2]float64{{d, 0}, {0, d}, {-d, 0}, {0, -d}} {
		superEllipse(&path, c[0]-r, ay+c[1]-r, c[0]+r, ay+c[1]+r, p.DotSuperness, true)
	}
	return placeInk(path, design.PunctSides{Left: refMarkSide, Right: refMarkSide})
}

// pilcrowAlt draws the pilcrow with one stem: a stem pilcrowAltStem stems
// wide from the top of the pilcrow to its foot, with the foot serifs of the
// lowercase, and on its left the solid bowl of the pilcrow, half a
// super-ellipse pilcrowAltBowl wide.  It is one contour.
func pilcrowAlt(p design.Params) (outline.Path, float64) {
	h := pilcrowTop - pilcrowBottom
	x1 := pilcrowAltBowl
	yb := pilcrowTop - pilcrowBowl*h
	q := newSuperShape(0, yb, 2*x1, pilcrowTop, p.SupernessOuter)
	q1, q2 := q.cubic(1, true), q.cubic(2, true) // top to the left extreme, and on to the bottom
	stem := stemSpec{
		Left: x1, Width: math.Round(pilcrowAltStem * p.Stem), Bottom: pilcrowBottom, Top: pilcrowTop,
		FootLeft: p.SerifJut, FootRight: p.SerifJut,
	}
	var path outline.Path
	path.MoveTo(stem.start())
	stemUp(&path, p, stem, pilcrowTop)
	path.LineTo(q1.P0.X, q1.P0.Y)
	path.CurveTo(q1.args())
	path.CurveTo(q2.args())
	stemDown(&path, p, stem)
	return placeInk(path, design.PunctSides{Left: pilcrowLeft, Right: pilcrowRight})
}

// interrobangs holds the ink of the interrobang by the parameters, so that
// the interrobang and the inverted one draw it once.
var interrobangs sync.Map // design.Params -> func() outline.Path

// interrobangInk returns the interrobang in the frame of the question mark.
// The path is shared and must not be changed.
//
// The hook is the stroke of the question mark from its top round its right
// extreme, with its ball terminal.  From the extreme it runs in one curve,
// which ends without curvature, to the middle line of the stem, where it is
// interroEnd round strokes thick, runs interroAngle degrees below the
// horizontal and is cut vertically, its lower corner interroJoin of the
// height of the hook above the foot of the stem.  The stem
// is that of the exclamation mark and stands under the top of the hook: from
// the gap of the exclamation mark above the dot to interroRise of the height
// of the hook above it, interroStem of the top of the exclamation mark wide
// at its top or as wide as keeps it clear of the ball and leaves
// interroCounter of the white between its middle line and the inner edge of
// the right side of the hook.  The sign is the union of the two over the dot
// of the question mark, which stands under the stem.
func interrobangInk(p design.Params) outline.Path {
	f, _ := interrobangs.LoadOrStore(p, sync.OnceValue(func() outline.Path {
		q := p.Punct
		f := newQFrame(p)
		foot, top := p.DotDiameter+q.ExclamGap, math.Round((1+interroRise)*p.Sym.QTop)
		sts := qStations(p, f)[:3] // the third is replaced
		cx := math.Round(sts[0].X) // the middle line of the stem: under the top of the hook
		a := degrees(interroAngle)
		w := interroEnd * p.CurveLower
		end := station{X: cx, Y: foot + interroJoin*p.Sym.QTop + w/math.Cos(a)/2, Dir: interroAngle - 180, W: w / math.Cos(a), Nib: -90}
		end.K = qInflection
		sts[2] = end
		right, left := strokeOpen(sts)
		hook := join(right, lineChain(right.end(), left.start()), left, qBall(p, right, left, f.x0+p.BallDiameter/2))
		// The stem is as wide at its top as leaves the white between it and
		// the ball and between it and the right side of the hook, and at least
		// as wide as at its foot.
		room := cx - f.x0 - p.BallDiameter - design.ClearanceFloor(p) - interroClear
		room = math.Min(room, (1-interroCounter)*(sts[1].X-sts[1].W/2-cx))
		wt := math.Max(math.Min(interroStem*q.ExclamTopWidth, 2*room), q.ExclamBottomWidth)
		stem := polygonChain([]outline.Point{
			{X: cx - q.ExclamBottomWidth/2, Y: foot}, {X: cx + q.ExclamBottomWidth/2, Y: foot},
			{X: cx + wt/2, Y: top}, {X: cx - wt/2, Y: top},
		})
		cs := append(union(hook, stem), dotChain(p, cx, p.DotDiameter/2))
		return chainsPath(cs...)
	}))
	return f.(func() outline.Path)()
}

// interrobang draws the interrobang, with the advance of the question mark.
func interrobang(p design.Params) (outline.Path, float64) {
	return interrobangInk(p), newQFrame(p).x1 + p.Sym.QSide
}

// interrobangDown draws the inverted interrobang: the interrobang turned
// through half a turn about the middle of its advance, at the height at which
// its lowest ink stands where that of the inverted question mark stands.
func interrobangDown(p design.Params) (outline.Path, float64) {
	path, adv := interrobang(p)
	_, y0, _, y1 := path.BBox()
	colon, _, _ := Outline("colon", p)
	_, _, _, top := colon.BBox()
	// The inverted question mark reaches from the top of the upper dot of the
	// colon as far down as the question mark is high.
	foot := top - p.Sym.QTop
	return turned(path, adv/2, (foot+y0+y1)/2), adv
}
