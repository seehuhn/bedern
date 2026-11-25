// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

import (
	"fmt"
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The radical: the base, the seven sizes of the delimiters and an assembly.  A
// radical is four straight strokes: the tick, a short thin stroke which
// rises from the left to the peak; the thick stroke from the peak down to the
// vertex at the foot; the long stroke from the vertex up to the top; and the
// bar, RadicalRuleThickness thick, whose top is the top of the glyph and
// which ends radOverlap beyond the advance, where the rule of the engine
// starts.  The bar has the rule of the signs, the tick
// the arm of the capitals and the thick stroke the crescent of the
// parenthesis, both grown by the law of the delimiters, and the long stroke
// grows from the rule by the height ratio to the power radLongPower; the
// heights of the tick and the peak are fractions of the height, and the
// horizontal distances grow as the height ratio to the power radWidthPower;
// a taller radical is steeper.
//
// The assembly is the construction radAssemblyHeight high with an upright
// long stroke, cut on a line through the long stroke above the peak: the
// bottom part holds the tick, the foot, cut flat radFoot wide, and the
// upright stroke, the top
// part the stroke turning into the bar, and the extender is the upright
// stroke, all with straight runs of endRun at the cut, as in delims.go.
const (
	// radBaseHeight, radBaseBottom: the base is 1000 units high and reaches
	// from -210 to 790.
	radBaseHeight = 1000.0
	radBaseBottom = -210.0
	// radWidth is the ink width of the base over its height.
	radWidth = 0.76
	// radLongPower is the exponent of the height ratio by which the long
	// stroke grows from the rule; the bar keeps the rule.
	radLongPower = 0.12
	// radWidthPower is the exponent of the height ratio by which the ink
	// width grows.
	radWidthPower = 0.22
	// radVertex is the distance of the vertex from the left of the ink over
	// the height of the base.
	radVertex = 0.30
	// radTickEnd is the height of the left end of the tick over the
	// height; radPeak that of the peak.
	radTickEnd = 0.41
	radPeak    = 0.47
	// radPeakX is the distance of the peak from the left of the ink over
	// that of the vertex.
	radPeakX = 0.25
	// radStub is the length of the underside of the bar to the right of the
	// long stroke, in rules.
	radStub = 1.5
	// radThickPower is the exponent of the height ratio by which the thick
	// stroke grows from the crescent of the parenthesis.
	radThickPower = 0.1
	// radOverlap is the length of the bar beyond the advance.
	radOverlap = 30.0
	// radAssemblyHeight is the height of the construction of the assembly
	// and radCut the height of its cut over its height.
	radAssemblyHeight = 2550.0
	radCut            = 0.75
	// radFoot is the width of the flat at the foot, where the thick stroke
	// meets the long one; radAssemblyHook is the height whose fractions the
	// tick and the peak of the assembly take.
	radFoot         = 30.0
	radAssemblyHook = 2000.0
	// radUpright moves the vertex of the construction of the assembly to the
	// right: the white between its thick and its upright stroke
	// is as wide as in the largest size.
	radUpright = 150.0
)

// radicalPath draws a radical h high with its foot at bottom; upright makes
// the long stroke upright.  It returns the outline, with the ink starting at
// 0, and the right end of the bar.
func radicalPath(p design.Params, h, bottom float64, upright bool) (outline.Path, float64) {
	rule := p.Punct.MathThick
	f := math.Pow(h/baseHeight(p), weightPower)
	long := rule * math.Pow(h/radBaseHeight, radLongPower)
	thick := p.Punct.ParenThick * math.Pow(h/baseHeight(p), radThickPower)
	tick := p.Caps.Bar * f
	k := math.Pow(h/radBaseHeight, radWidthPower)
	top := bottom + h
	end := math.Round(radWidth * radBaseHeight * k)
	v := pt(math.Round(radVertex*radBaseHeight*k), bottom)
	hook := h // the height whose fractions the tick and the peak take
	if upright {
		hook = radAssemblyHook
	}
	peak := pt(radPeakX*v.X, bottom+radPeak*hook)

	// The long stroke: its right edge from the vertex to the underside of
	// the bar, radStub rules left of the end, and its left edge parallel,
	// rule from it.
	b1 := pt(end-radStub*rule, top-rule)
	if upright {
		v.X += radUpright
		b1.X = v.X
		end = b1.X + radStub*rule
	}
	dl := unitVec(b1.X-v.X, b1.Y-v.Y)
	nl := pt(-dl.Y, dl.X) // to the left of the long stroke
	l0 := pt(v.X+long*nl.X, v.Y+long*nl.Y)
	b2 := atY(l0, dl, top)

	// The thick stroke: its left edge from the vertex, its right edge
	// through the peak, thick apart.
	vl := v
	vl.X -= radFoot
	e := throughWithDistance(vl, peak, thick)
	crotch := meet(l0, dl, peak, e)

	// The tick: its upper edge from the peak to the left end of the ink,
	// its lower edge tick below it, cut upright at 0.
	u0 := pt(0, bottom+radTickEnd*hook+tick/2)
	dt := unitVec(peak.X-u0.X, peak.Y-u0.Y)
	nt := pt(dt.Y, -dt.X) // below the tick
	t0 := pt(u0.X+tick*nt.X, u0.Y+tick*nt.Y)
	l1 := pt(0, t0.Y-t0.X*dt.Y/dt.X)
	m := meet(t0, dt, vl, e)

	c := polygon(v, b1, pt(end, top-rule), pt(end, top), b2, crotch, peak, u0, l1, m, vl)
	path := outline.Path{c}
	if x0, _, _, _ := path.BBox(); x0 < 0 {
		return nil, 0
	}
	return path, end
}

// unitVec returns the unit vector in the direction (x, y).
func unitVec(x, y float64) outline.Point {
	n := math.Hypot(x, y)
	return pt(x/n, y/n)
}

// throughWithDistance returns the unit direction, pointing upwards, of the
// line through a whose distance from b is d, with b to its right.
func throughWithDistance(a, b outline.Point, d float64) outline.Point {
	dx, dy := b.X-a.X, b.Y-a.Y
	r := math.Hypot(dx, dy)
	// The angle of a to b, turned to the left by asin(d / r).
	ang := math.Atan2(dy, dx) + math.Asin(d/r)
	return pt(math.Cos(ang), math.Sin(ang))
}

// radical draws a size of the radical h high with its foot at bottom, with
// the side space of the parenthesis on the left and the advance radOverlap
// left of the end of the bar.
func radical(p design.Params, h, bottom float64) (outline.Path, float64) {
	path, barEnd := radicalPath(p, h, bottom, false)
	left := math.Round(p.Punct.ParenLeft.Left)
	return shifted(path, left, 0), left + barEnd - radOverlap
}

// radicalGlyphs returns the base of the radical with its sizes and the parts
// of its assembly.
func radicalGlyphs(p design.Params) ([]Glyph, error) {
	const base = "uni221A"
	path, adv := radical(p, radBaseHeight, radBaseBottom)
	if path == nil {
		return nil, fmt.Errorf("mathfont: the radical has no room for its tick")
	}
	g := Glyph{Name: base, Rune: 0x221A, Advance: math.Round(adv), Path: path, Class: Ord, Like: "f"}
	out := []Glyph{g}
	for k, h := range delimSizes {
		path, adv := radical(p, h, p.Punct.MathAxis-h/2)
		if path == nil {
			return nil, fmt.Errorf("mathfont: the radical %v high has no room for its tick", h)
		}
		name := fmt.Sprintf("%s.v%d", base, k+1)
		out = append(out, Glyph{Name: name, Advance: math.Round(adv), Path: path, Class: Ord, Like: "f"})
		out[0].Variants = append(out[0].Variants, name)
	}

	// The assembly.
	cons, barEnd := radicalPath(p, radAssemblyHeight, 0, true)
	if cons == nil {
		return nil, fmt.Errorf("mathfont: the assembly of the radical has no room for its tick")
	}
	left := math.Round(p.Punct.ParenLeft.Left)
	cons = shifted(cons, left, 0)
	adv = math.Round(left + barEnd - radOverlap)
	cut := math.Round(radCut * radAssemblyHeight)
	c := withNodesAt(cons[0], cut)
	above, below, err := halves(c, cut)
	if err != nil {
		return nil, fmt.Errorf("mathfont: the assembly of the radical: %w", err)
	}
	top, err := end(c, cut, endRun, true)
	if err != nil {
		return nil, err
	}
	bot, err := end(c, cut, endRun, false)
	if err != nil {
		return nil, err
	}
	lo, hi := math.Min(above.start.X, below.start.X), math.Max(above.start.X, below.start.X)
	ext := outline.Path{box(lo, 0, hi, extender)}
	_, _, _, hb := bot.BBox()
	_, _, _, ht := top.BBox()
	out = append(out,
		Glyph{Name: base + ".bot", Advance: adv, Path: bot, Class: Ord, Like: "f"},
		Glyph{Name: base + ".ext", Advance: adv, Path: ext, Class: Ord, Like: "f"},
		Glyph{Name: base + ".top", Advance: adv, Path: top, Class: Ord, Like: "f"})
	out[0].Parts = []Part{
		{Name: base + ".bot", EndConn: connector, FullAdvance: hb},
		{Name: base + ".ext", StartConn: connector, EndConn: connector, FullAdvance: extender, Extender: true},
		{Name: base + ".top", StartConn: connector, FullAdvance: ht},
	}
	return out, nil
}
