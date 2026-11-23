// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The barred letters ł, Ł, đ, Đ, Ð, ħ, Ħ, ŧ, Ŧ, ø and Ø.
//
// Each is its base letter, unchanged, with a bar or a stroke across it: the
// letter is the union of the base and the bar (symunion.go), so that the bar
// and the strokes it crosses meet in corners.  A bar is straight, with ends cut
// vertically; the stroke of ø and Ø is straight, with ends cut horizontally.
// The bars of the lowercase are ThinLower thick, measured at right angles to
// the bar, those of the capitals Caps.Bar; the bar of ŧ is the crossbar of t.
// The places are fractions of the height of the letter and lengths in stems,
// the shape constants below.

// Shape constants of the bars of the lowercase.
const (
	lslashAngle  = 37.0  // the bar of ł, degrees above the horizontal
	lslashCentre = 0.525 // height of the middle of the bar of ł, in ascenders
	lslashBeyond = 10.0  // how far the bar of ł reaches beyond the tips of the foot serifs, at most
	dcroatCentre = 0.787 // height of the middle of the bar of đ, in ascenders
	dcroatLength = 3.82  // length of the bar of đ, in stems
	hbarCentre   = 0.768 // height of the middle of the bar of ħ, in ascenders
	hbarLength   = 3.7   // length of the bar of ħ, in stems
	tbarCentre   = 0.6   // height of the middle of the second bar of ŧ, in x-heights
	oslashAngle  = 58.0  // the stroke of ø, degrees above the horizontal
	oslashBeyond = 34.0  // how far the stroke of ø reaches above and below the o
)

// Shape constants of the bars of the capitals.
const (
	capLslashAngle  = 35.0  // the bar of Ł, degrees above the horizontal
	capLslashCentre = 0.54  // height of the middle of the bar of Ł, in cap heights
	capDcroatCentre = 0.511 // height of the middle of the bar of Đ, in cap heights
	capDcroatLength = 3.76  // length of the bar of Đ, in capital stems
	capHbarCentre   = 0.75  // height of the middle of the second bar of Ħ, in cap heights
	capTbarCentre   = 0.526 // height of the middle of the bar of Ŧ, in cap heights
	capTbarLength   = 4.06  // length of the bar of Ŧ, in capital stems
	capOslashAngle  = 53.0  // the stroke of Ø, degrees above the horizontal
	capOslashBeyond = 30.0  // how far the stroke of Ø reaches above and below the O
)

// withBars returns the union of the outline of the base glyph and the bars.
func withBars(p design.Params, base string, bars ...chain) (outline.Path, float64) {
	path, adv, _ := Outline(base, p)
	parts := append(pathChains(path), bars...)
	return chainsPath(union(parts...)...), adv
}

// slantBar returns a bar whose middle line runs through (cx, cy) at the angle
// (degrees above the horizontal), which reaches w/2 to either side of cx, is
// thick at right angles to it, and has vertical ends.
func slantBar(cx, cy, w, angle, thick float64) chain {
	t := math.Tan(angle * math.Pi / 180)
	h := thick / math.Cos(angle*math.Pi/180) / 2 // half the vertical thickness
	x0, x1 := cx-w/2, cx+w/2
	y0, y1 := cy-t*w/2, cy+t*w/2
	pts := []outline.Point{{X: x0, Y: y0 - h}, {X: x1, Y: y1 - h}, {X: x1, Y: y1 + h}, {X: x0, Y: y0 + h}}
	return polygonChain(pts)
}

// slashStroke returns a straight stroke from the height y0 to the height y1,
// whose middle line runs through (cx, cy) at the angle (degrees above the
// horizontal), thick at right angles to it, with horizontal ends.
func slashStroke(cx, cy, y0, y1, angle, thick float64) chain {
	a := angle * math.Pi / 180
	w := thick / math.Sin(a) / 2 // half the horizontal thickness
	x := func(y float64) float64 { return cx + (y-cy)/math.Tan(a) }
	pts := []outline.Point{{X: x(y0) - w, Y: y0}, {X: x(y0) + w, Y: y0}, {X: x(y1) + w, Y: y1}, {X: x(y1) - w, Y: y1}}
	return polygonChain(pts)
}

// polygonChain returns the closed chain through the points.
func polygonChain(pts []outline.Point) chain {
	var c chain
	for i, q := range pts {
		c = append(c, lineSeg(q, pts[(i+1)%len(pts)]))
	}
	return c
}

// lslash draws ł: the l with a bar rising to the right across the stem.
func lslash(p design.Params) (outline.Path, float64) {
	cx := design.NarrowStemLeft(p) + design.AscenderInset(p) + p.AscenderStem/2
	// The bar keeps 5 inside the advance where the serifs leave less room.
	beyond := math.Min(lslashBeyond, math.Min(design.NarrowSideLeft(p), design.NarrowSideRight(p))+design.AscenderInset(p)-p.SerifJut-5)
	w := p.AscenderStem + 2*p.SerifJut + 2*beyond
	return withBars(p, "l", slantBar(cx, lslashCentre*p.Ascender, w, lslashAngle, p.ThinLower))
}

// capLslash draws Ł: the L with a bar rising to the right across the stem,
// from the tip of the left foot serif, Caps.LslashWidth wide or, if that is
// zero, design.CapLslashStems stems.
func capLslash(p design.Params) (outline.Path, float64) {
	w := design.CapLslashStems * p.Caps.Stem
	if p.Caps.LslashWidth > 0 {
		w = p.Caps.LslashWidth
	}
	x0 := design.CapStemLeft(p) - p.Caps.SerifJut
	return withBars(p, "L", slantBar(x0+w/2, capLslashCentre*p.CapHeight, w, capLslashAngle, p.Caps.Bar))
}

// dcroat draws đ: the d with a horizontal bar across the ascender, which ends
// on the right where the foot serif ends.
func dcroat(p design.Params) (outline.Path, float64) {
	x1 := design.DStemLeft(p) + design.AscenderInset(p) + p.AscenderStem + p.SerifJut
	y := dcroatCentre * p.Ascender
	return withBars(p, "d", rectChain(x1-dcroatLength*p.Stem, y-p.ThinLower/2, x1, y+p.ThinLower/2))
}

// capDcroat draws Đ and Ð: the D with a horizontal bar across the stem, from
// the tip of the left foot serif into the counter.
func capDcroat(p design.Params) (outline.Path, float64) {
	x0 := capBowlStem(p, false).Left - p.Caps.SerifJut
	y := capDcroatCentre * p.CapHeight
	return withBars(p, "D", rectChain(x0, y-p.Caps.Bar/2, x0+capDcroatLength*p.Caps.Stem, y+p.Caps.Bar/2))
}

// hbar draws ħ: the h with a horizontal bar across the ascender, from the tip
// of the head serif.
func hbar(p design.Params) (outline.Path, float64) {
	x0 := p.SideStemLeft + design.AscenderInset(p) - p.SerifJut
	y := hbarCentre * p.Ascender
	return withBars(p, "h", rectChain(x0, y-p.ThinLower/2, x0+hbarLength*p.Stem, y+p.ThinLower/2))
}

// capHbar draws Ħ: the H with a second bar above the crossbar, from the tip of
// the left head serif to the tip of the right one.
func capHbar(p design.Params) (outline.Path, float64) {
	path, _, _ := Outline("H", p)
	x0, _, x1, _ := path.BBox()
	y := capHbarCentre * p.CapHeight
	return withBars(p, "H", rectChain(x0, y-p.Caps.Bar/2, x1, y+p.Caps.Bar/2))
}

// tbar draws ŧ: the t with a second bar, the crossbar of t, below the
// crossbar.
func tbar(p design.Params) (outline.Path, float64) {
	b := tBar(p)
	y := tbarCentre * p.XHeight
	return withBars(p, "t", rectChain(b.LeftEdge-b.Left, y-b.Thick/2, b.RightEdge+b.Right, y+b.Thick/2))
}

// capTbar draws Ŧ: the T with a bar across the middle of the stem.
func capTbar(p design.Params) (outline.Path, float64) {
	path, _, _ := Outline("T", p)
	lo, hi := stemAt(path, 0.5*p.CapHeight)
	cx, w := (lo+hi)/2, capTbarLength*p.Caps.Stem
	y := capTbarCentre * p.CapHeight
	return withBars(p, "T", rectChain(cx-w/2, y-p.Caps.Bar/2, cx+w/2, y+p.Caps.Bar/2))
}

// oslash draws ø: the o with a straight stroke from the lower left to the
// upper right through its middle.
func oslash(p design.Params) (outline.Path, float64) {
	cx := (design.OBowlLeft(p) + design.OBowlRight(p)) / 2
	y0, y1 := -p.OvershootLower-oslashBeyond, p.XHeight+p.OvershootLower+oslashBeyond
	return withBars(p, "o", slashStroke(cx, p.XHeight/2, y0, y1, oslashAngle, p.ThinLower))
}

// capOslash draws Ø: the O with a straight stroke from the lower left to the
// upper right through its middle.
func capOslash(p design.Params) (outline.Path, float64) {
	path, _, _ := Outline("O", p)
	x0, _, x1, _ := path.BBox()
	y0, y1 := -p.OvershootUpper-capOslashBeyond, p.CapHeight+p.OvershootUpper+capOslashBeyond
	return withBars(p, "O", slashStroke((x0+x1)/2, p.CapHeight/2, y0, y1, capOslashAngle, p.Caps.Bar))
}
