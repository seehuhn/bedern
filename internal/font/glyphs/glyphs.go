// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// Glyph is one glyph of a font.
type Glyph struct {
	Name    string       // glyph name
	Rune    rune         // Unicode code point, or 0 if the glyph has none
	Advance float64      // advance width
	Path    outline.Path // the outline, with integer coordinates
}

// A sideKind says what stands on one side of a glyph, which fixes the space
// between the edge of the advance and the ink.  The advance of a letter is
// the sum of its body and the side spaces of the kinds of its two sides, and
// the bounding-box check takes the least distance from the edge of the advance
// from the same kinds.  The zero value is no kind, and an entry with it fails
// the check.
type sideKind int

const (
	sideUnknown  sideKind = iota
	sideStem              // the edge of a stem with serifs: SideStemLeft or SideStemRight to the stem
	sideRound             // the extreme of a bowl: SideRoundLeft or SideRoundRight
	sideStemD             // the right edge of the stem of d: DSideRight
	sideStemBP            // the left edge of the stem of b and p, with the serifs: BPSideLeft
	sideBowlDQ            // the left extreme of the bowl of d and q: DQSideLeft
	sideBowlBP            // the right extreme of the bowl of b and p: BPSideRight
	sideBowlC             // the left extreme of c: CSideLeft
	sideStemR             // the left edge of the stem of r, with the serifs: RSideLeft
	sideStemQ             // the right edge of the stem of q, whose foot serif is at the descender line: QSideRight
	sideNarrow            // the stem of a narrow letter, which stands in the middle: NarrowSideLeft or NarrowSideRight
	sideBowlA             // the free extreme of the bowl of a: ASideLeft
	sideTailA             // the stem of a, which ends in a tail: ATailSideRight, measured from the end of the tail
	sideOpen              // a side without a stem or bowl, or with a frame of its own: no distance is required
	sideOverhang          // a side where ink may reach beyond the edge of the advance, by at most MaxOverhang: the hook of f on the right, the tail of j on the left
	sideBarF              // the left side of f: the end of the crossbar or the left foot serif, FInkLeft
	sideBarT              // the left side of t: the end of the crossbar, TInkLeft
	sideLoopG             // the left side of g: the left extreme of the loop, GSideLeft
	sideEarG              // the right side of g: the end of the ear, or the loop, GSideRight
	sideDiag              // a side of v, w, x or y: the tip of a serif at the x-height (at the baseline for x), DiagSideLeft or DiagSideRight, with the baseline ink far inside
	sideDiagTail          // the left side of y: the tip of the serif at the x-height, DiagSideLeft, as for the diagonal letters, with the end of the tail allowed up to the edge of the advance
	sideZ                 // a side of z: the corner of the diagonal at the baseline on the left, the end of the lower bar on the right, ZSideLeft or ZSideRight
	sideK                 // the right side of k: the tip of the foot serif of the leg, KSideRight
)

// An entry describes how one glyph is made: the code which draws it and the
// kinds of its left and right side.
type entry struct {
	name        string
	r           rune // 0 for glyphs without a code point
	left, right sideKind
	draw        func(p design.Params) (path outline.Path, advance float64)
}

// space returns the distance from the edge of the advance to the stem or the
// extreme of the bowl on a side of the given kind; for the left side if left
// is true.  It is 0 for an open side.
func (k sideKind) space(p design.Params, left bool) float64 {
	switch {
	case k == sideStem && left:
		return p.SideStemLeft
	case k == sideStem:
		return p.SideStemRight
	case k == sideRound && left:
		return p.SideRoundLeft
	case k == sideRound:
		return p.SideRoundRight
	case k == sideStemD:
		return p.DSideRight
	case k == sideStemBP:
		return p.BPSideLeft
	case k == sideStemQ:
		return p.QSideRight
	case k == sideBowlDQ:
		return p.DQSideLeft
	case k == sideBowlBP:
		return p.BPSideRight
	case k == sideBowlC:
		return p.CSideLeft
	case k == sideStemR:
		return p.RSideLeft
	case k == sideNarrow && left:
		return design.NarrowSideLeft(p)
	case k == sideNarrow:
		return design.NarrowSideRight(p)
	case k == sideBowlA:
		return p.ASideLeft
	case k == sideTailA:
		return p.ATailSideRight
	case k == sideOverhang:
		return -p.MaxOverhang
	case k == sideBarF:
		return design.FInkLeft(p)
	case k == sideBarT:
		return design.TInkLeft(p)
	case k == sideLoopG:
		return p.GSideLeft
	case k == sideEarG:
		return p.GSideRight
	case k == sideZ && left:
		return p.ZSideLeft
	case k == sideZ:
		return p.ZSideRight
	case k == sideK:
		return p.KSideRight
	case k == sideDiagTail:
		return 0
	case k == sideDiag && left:
		return p.DiagSideLeft
	case k == sideDiag:
		return p.DiagSideRight
	}
	return 0
}

// inkMargins returns the least distance of the ink from the left and from the
// right edge of the advance.  The serifs of a stem reach SerifJut beyond it, and
// the margin of such a side is that much smaller than the space.  A bowl has no
// serifs, and an open side has no requirement.
func (e entry) inkMargins(p design.Params) (left, right float64) {
	margin := func(k sideKind, isLeft bool) float64 {
		m := k.space(p, isLeft)
		switch k {
		case sideStem, sideStemD, sideStemBP, sideStemQ, sideStemR, sideNarrow:
			m -= p.SerifJut
		case sideTailA:
			// the tail is the rightmost ink: nothing to take off
		}
		if k == sideOverhang {
			return m // negative: the ink may stand beyond the edge by this much
		}
		return math.Max(m, 0)
	}
	return margin(e.left, true), margin(e.right, false)
}

// table lists all glyphs in glyph order.  The glyph .notdef comes first.
var table = []entry{
	{".notdef", 0, sideOpen, sideOpen, notdef},
	{"space", ' ', sideOpen, sideOpen, space},
	{"a", 'a', sideBowlA, sideTailA, a},
	{"b", 'b', sideStemBP, sideBowlBP, b},
	{"c", 'c', sideBowlC, sideOpen, c},
	{"d", 'd', sideBowlDQ, sideStemD, d},
	{"e", 'e', sideRound, sideOpen, e},
	{"f", 'f', sideBarF, sideOverhang, f},
	{"g", 'g', sideLoopG, sideEarG, g},
	{"h", 'h', sideStem, sideStem, h},
	{"i", 'i', sideNarrow, sideNarrow, i},
	{"j", 'j', sideOverhang, sideNarrow, j},
	{"k", 'k', sideStem, sideK, k},
	{"l", 'l', sideNarrow, sideNarrow, l},
	{"m", 'm', sideStem, sideStem, m},
	{"n", 'n', sideStem, sideStem, n},
	{"o", 'o', sideRound, sideRound, o},
	{"p", 'p', sideStemBP, sideBowlBP, p},
	{"q", 'q', sideBowlDQ, sideStemQ, q},
	{"r", 'r', sideStemR, sideOpen, r},
	{"s", 's', sideOpen, sideOpen, s},
	{"t", 't', sideBarT, sideOpen, t},
	{"u", 'u', sideStem, sideStem, u},
	{"v", 'v', sideDiag, sideDiag, v},
	{"w", 'w', sideDiag, sideDiag, w},
	{"x", 'x', sideDiag, sideDiag, x},
	{"y", 'y', sideDiagTail, sideDiag, y},
	{"z", 'z', sideZ, sideZ, z},
}

// lookup returns the entry of the glyph with the given name.
func lookup(name string) (entry, bool) {
	for _, e := range table {
		if e.name == name {
			return e, true
		}
	}
	return entry{}, false
}

// Set returns all glyphs of the font.  The glyph .notdef comes first.  The
// outlines are rounded to integer coordinates.
func Set(p design.Params) []Glyph {
	gs := make([]Glyph, len(table))
	for i, e := range table {
		path, adv := e.draw(p)
		gs[i] = Glyph{Name: e.name, Rune: e.r, Advance: adv, Path: path.Round()}
	}
	return gs
}

// Outline returns the unrounded outline and the advance width of the glyph
// with the given name.  The result is false if there is no such glyph.  The
// outline is what Set rounds.
func Outline(name string, p design.Params) (path outline.Path, advance float64, ok bool) {
	for _, e := range table {
		if e.name == name {
			path, advance = e.draw(p)
			return path, advance, true
		}
	}
	return nil, 0, false
}

// space draws the word space, which has no outline.
func space(p design.Params) (outline.Path, float64) {
	return nil, p.SpaceAdvance
}

// rect adds a rectangle to the path.
func rect(path *outline.Path, x0, y0, x1, y1 float64, ccw bool) {
	path.MoveTo(x0, y0)
	if ccw {
		path.LineTo(x1, y0)
		path.LineTo(x1, y1)
		path.LineTo(x0, y1)
	} else {
		path.LineTo(x0, y1)
		path.LineTo(x1, y1)
		path.LineTo(x1, y0)
	}
}
