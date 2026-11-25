// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

import (
	"fmt"
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/outline"
)

// The big unions, the big circled operators and the contour integral at text
// and display size.  Each size is drawn at its height, centred on the axis,
// with its weights grown by the law of the big operators (bigops.go); a
// display size is a drawing of its own and not the text size scaled.
//
// ⋃ is one stroke of the symbols down the left stem, round through half a
// turn and up the right stem, the stems straight and cut level at the top;
// ⋂ is ⋃ mirrored in the axis.  ⨁ and ⨂ are a ring with the bars of ⊕ and ⊗
// across its counter.  ∮ is the integral of integral.go with a ring about
// the middle of its stem on the axis.
const (
	// cupOpStem is the stem of ⋃ and ⋂ over the law of the big operators;
	// the stems have the weight of the ring of ⨁ (circOpRing), 68 in text.
	// cupOpRound is the bottom of the round over the stem.
	cupOpStem  = 0.60
	cupOpRound = 1.0
	// cupOpWidthText and cupOpWidthDisplay are the ink widths of ⋃ over its
	// height.
	cupOpWidthText    = 0.77
	cupOpWidthDisplay = 0.75
	// circOpText and circOpDisplay are the outer diameters of ⨁ and ⨂: 1.00
	// and 0.93 of the height of the summation.  circOpRing is the ring over
	// the law of the big operators and circOpBar the bars over the ring.
	circOpText    = 1000.0
	circOpDisplay = 1300.0
	circOpRing    = 0.60
	circOpBar     = 0.94
	// ointRingText and ointRingDisplay are the outer diameters of the ring
	// of ∮ over the height of the integral.  ointRingWeight is its stroke
	// over the thin stroke of the integral, which puts it at 41 and 56; the
	// stem is the dominant stroke.
	ointRingText    = 0.44
	ointRingDisplay = 0.33
	ointRingWeight  = 0.9
	// multiIntPitch is the distance between the strokes of ∬ ∭ and ∯ over the
	// ink width of the integral at the same size.  The side spaces and the
	// italics correction are those of the integral.  ∯ has one oval about the
	// stems of both strokes.
	multiIntPitch = 0.58
)

// cupOp draws ⋃ h high: a stroke down the left stem, round through half a
// turn and up the right stem, wShare of h wide, the stems cupOpStem of the
// law of the big operators and the bottom of the round cupOpRound of that,
// centred on the axis, with MathSide on each side.
func cupOp(p design.Params, h, wShare float64) (outline.Path, float64) {
	s := newOpSize(p, h)
	t := cupOpStem * s.thick
	tb := cupOpRound * t
	w := math.Round(wShare * h)
	top, bottom := s.top, s.bottom
	a := (w - t) / 2 // the radius of the centre line at the sides
	yc := bottom + tb/2 + a
	path := glyphs.Stroke([]glyphs.Station{
		{X: t / 2, Y: top, Dir: 270, W: t, Flat: true},
		{X: t / 2, Y: yc, Dir: 270, W: t},
		{X: w / 2, Y: bottom + tb/2, Dir: 0, W: tb},
		{X: w - t/2, Y: yc, Dir: 90, W: t, Flat: true},
		{X: w - t/2, Y: top, Dir: 90, W: t},
	})
	return withSides(path, p.Punct.MathSide, p.Punct.MathSide)
}

// circOp draws ⨁ (kind "plus") or ⨂ (kind "times") with the outer diameter
// d: a ring circOpRing of the law of the big operators thick and the bars
// circOpBar of the ring across its counter, their ends in the middle of the
// ring, centred on the axis, with MathSide on each side.
func circOp(p design.Params, d float64, kind string) (outline.Path, float64) {
	s := newOpSize(p, d)
	t := math.Round(circOpRing * s.thick)
	tb := math.Round(circOpBar * t)
	rad := d / 2
	o := pt(0, p.Punct.MathAxis)
	reach := rad - t/2
	bar := outline.Path{box(o.X-reach, o.Y-tb/2, o.X+reach, o.Y+tb/2)}
	parts := []outline.Path{ring(o, rad, t)}
	switch kind {
	case "plus":
		parts = append(parts, bar, rotated(bar, o, math.Pi/2))
	case "times":
		parts = append(parts, rotated(bar, o, math.Pi/4), rotated(bar, o, -math.Pi/4))
	}
	return withSides(glyphs.Union(parts...), p.Punct.MathSide, p.Punct.MathSide)
}

// oint draws ∮ h high: the integral with a ring about the middle of its stem
// on the axis, ointRingText or ointRingDisplay of the height across outside, its stroke
// ointRingWeight of the thin stroke of the integral; the advance is that of
// the integral, widened where the ring reaches further out than its ink, and
// the italics correction is found as that of the integral, with the ring.
func oint(p design.Params, h float64) (outline.Path, float64, float64, error) {
	path, adv, _, err := integral(p, h)
	if err != nil {
		return nil, 0, 0, err
	}
	g := math.Pow(h/p.CapHeight, opPower)
	t := math.Round(ointRingWeight * p.Caps.Bar * g)
	axis := p.Punct.MathAxis
	rs := runsOf(path, axis)
	if len(rs) != 1 {
		return nil, 0, 0, fmt.Errorf("the stem of the integral has %d runs at the axis", len(rs))
	}
	c := pt((rs[0][0]+rs[0][1])/2, axis)
	share := ointRingText
	if h > intTextHeight {
		share = ointRingDisplay
	}
	rad := math.Round(share*h) / 2
	u := glyphs.Union(path, ring(c, rad, t))
	x0, _, x1, _ := path.BBox()
	ux0, _, ux1, _ := u.BBox()
	dl, dr := math.Max(0, x0-ux0), math.Max(0, ux1-x1)
	u, adv = shifted(u, dl, 0), adv+dl+dr
	// The italics correction puts the subscript intSide right of the ink of
	// the lower intFootBand of the height, ring included, as on ∫.
	_, y0, _, _ := u.Round().BBox()
	right := math.Inf(-1)
	for i := 0; i <= 40; i++ {
		for _, x := range runsAt(u, y0+intFootBand*h*float64(i)/40) {
			right = math.Max(right, x)
		}
	}
	return u, adv, math.Round(adv - right - intSide), nil
}

// multiIntegral draws the integral h high n times in a row, ∬ for n = 2 and ∭
// for n = 3, the strokes multiIntPitch of the ink of one stroke apart, with
// the side spaces and the italics correction of the integral.  With contour
// set it adds one oval about the stems of all strokes, as ∯ has it: its
// stroke that of the ring of ∮, its height the diameter of that ring and its
// width that diameter and the distance of the stems.
func multiIntegral(p design.Params, h float64, n int, contour bool) (outline.Path, float64, float64, error) {
	single, adv, _, err := integral(p, h)
	if err != nil {
		return nil, 0, 0, err
	}
	x0, _, x1, _ := single.BBox()
	pitch := math.Round(multiIntPitch * (x1 - x0))
	var path outline.Path
	for k := range n {
		path = append(path, shifted(single, float64(k)*pitch, 0)...)
	}
	adv += float64(n-1) * pitch
	// The italics correction is that of the strokes; the oval of ∯ lies about
	// the middle of the stems, above the feet.
	_, y0, _, _ := path.Round().BBox()
	right := math.Inf(-1)
	for i := 0; i <= 40; i++ {
		for _, x := range runsAt(path, y0+intFootBand*h*float64(i)/40) {
			right = math.Max(right, x)
		}
	}
	shift := 0.0
	if contour {
		g := math.Pow(h/p.CapHeight, opPower)
		t := math.Round(ointRingWeight * p.Caps.Bar * g)
		axis := p.Punct.MathAxis
		rs := runsOf(single, axis)
		if len(rs) != 1 {
			return nil, 0, 0, fmt.Errorf("the stem of the integral has %d runs at the axis", len(rs))
		}
		share := ointRingText
		if h > intTextHeight {
			share = ointRingDisplay
		}
		rad := math.Round(share*h) / 2
		span := float64(n-1) * pitch
		c := pt((rs[0][0]+rs[0][1])/2+span/2, axis)
		a, b := rad+span/2, rad
		oval := outline.Path{
			glyphs.SuperEllipseOutline(c.X-a, c.Y-b, c.X+a, c.Y+b, ovalSuperness, true)[0],
			glyphs.SuperEllipseOutline(c.X-a+t, c.Y-b+t, c.X+a-t, c.Y+b-t, ovalSuperness, false)[0],
		}
		u := glyphs.Union(path, oval)
		ux0, _, ux1, _ := u.BBox()
		bx0, _, bx1, _ := path.BBox()
		dl, dr := math.Max(0, bx0-ux0), math.Max(0, ux1-bx1)
		path, adv, shift = shifted(u, dl, 0), adv+dl+dr, dl
	}
	return path, adv, math.Round(adv - (right + shift) - intSide), nil
}

// ovalSuperness is the superness of the ellipse.
const ovalSuperness = 0.7071

// multiIntegrals are the integrals with several strokes.
var multiIntegrals = []struct {
	name    string
	r       rune
	n       int
	contour bool
}{
	{"uni222C", 0x222C, 2, false},
	{"uni222D", 0x222D, 3, false},
	{"uni222F", 0x222F, 2, true},
}

// bigSetOps are the big unions and circled operators: the text size with the
// display size as its variant.
var bigSetOps = []bigOp{
	{"uni22C3", 0x22C3, func(p design.Params, d bool) (outline.Path, float64) {
		if d {
			return cupOp(p, opDisplayHeight, cupOpWidthDisplay)
		}
		return cupOp(p, opTextHeight, cupOpWidthText)
	}},
	{"uni22C2", 0x22C2, func(p design.Params, d bool) (outline.Path, float64) {
		var path outline.Path
		var adv float64
		if d {
			path, adv = cupOp(p, opDisplayHeight, cupOpWidthDisplay)
		} else {
			path, adv = cupOp(p, opTextHeight, cupOpWidthText)
		}
		return mirroredY(path, p.Punct.MathAxis), adv
	}},
	{"uni2A01", 0x2A01, func(p design.Params, d bool) (outline.Path, float64) {
		if d {
			return circOp(p, circOpDisplay, "plus")
		}
		return circOp(p, circOpText, "plus")
	}},
	{"uni2A02", 0x2A02, func(p design.Params, d bool) (outline.Path, float64) {
		if d {
			return circOp(p, circOpDisplay, "times")
		}
		return circOp(p, circOpText, "times")
	}},
}

// bigSetGlyphs returns the text and display sizes of ⋃ ⋂ ⨁ ⨂ ∮ ∬ ∭ and ∯, the text
// size with the display size as its variant.
func bigSetGlyphs(p design.Params) ([]Glyph, error) {
	var gs []Glyph
	for _, o := range bigSetOps {
		path, adv := o.draw(p, false)
		dpath, dadv := o.draw(p, true)
		gs = append(gs,
			Glyph{Name: o.base, Rune: o.r, Advance: math.Round(adv), Path: path, Class: Ord, Like: "H", Variants: []string{o.base + ".d"}},
			Glyph{Name: o.base + ".d", Advance: math.Round(dadv), Path: dpath, Class: Ord, Like: "H"})
	}
	const base = "uni222E"
	for i, h := range []float64{intTextHeight, intDisplayHeight} {
		path, adv, ic, err := oint(p, h)
		if err != nil {
			return nil, fmt.Errorf("mathfont: the contour integral %v high: %w", h, err)
		}
		g := Glyph{Name: base, Rune: 0x222E, Advance: math.Round(adv), Path: path, Class: Ord, Like: "f", Italic: ic}
		if i == 1 {
			g.Name, g.Rune = base+".d", 0
		} else {
			g.Variants = []string{base + ".d"}
		}
		gs = append(gs, g)
	}
	for _, m := range multiIntegrals {
		for i, h := range []float64{intTextHeight, intDisplayHeight} {
			path, adv, ic, err := multiIntegral(p, h, m.n, m.contour)
			if err != nil {
				return nil, fmt.Errorf("mathfont: %s %v high: %w", m.name, h, err)
			}
			g := Glyph{Name: m.name, Rune: m.r, Advance: math.Round(adv), Path: path, Class: Ord, Like: "f", Italic: ic}
			if i == 1 {
				g.Name, g.Rune = m.name+".d", 0
			} else {
				g.Variants = []string{m.name + ".d"}
			}
			gs = append(gs, g)
		}
	}
	return gs, nil
}
