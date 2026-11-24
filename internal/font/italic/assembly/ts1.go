// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package assembly

import (
	"math"
	"slices"
	"strings"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/outline"
)

// The marks and the glyphs without ink of TS1 which the roman has beyond the
// accents of T1: the tie accents, the low tilde and the double grave of the
// modifier letters, the compound word marks, and the capital forms of the
// accents.
//
// The tie accents and the two modifier signs are marks like the accents: the
// roman's glyph sheared by the slant about the lowest edge of its ink, so that
// the edge next to the letter stays and the rest leans as the mark of a
// letter does (marks.go).  The ties keep the roman's advance and place in it; the two
// modifier signs keep the margins of the accents on either side of the ink.
// The compound word marks have no ink and the advance 0.

// unfitted says whether the glyph is one of the marks of this file whose
// place in its advance is that of the roman and not a side of the fit table:
// the tie accents and the capital forms of the accents.
func unfitted(name string) bool {
	return strings.HasPrefix(name, "tieaccent") || strings.HasSuffix(name, ".case")
}

// shearedAbout returns the path sheared by the slant k about the height y0.
func shearedAbout(path outline.Path, k, y0 float64) outline.Path {
	out := make(outline.Path, len(path))
	for i, c := range path {
		out[i] = transformed(c, k, -k*y0, 0)
	}
	return out
}

// lowestEdge returns the lowest height of the outline.
func lowestEdge(path outline.Path) float64 {
	_, y0, _, _ := path.BBox()
	return y0
}

// addTS1Marks adds the tie accents, the two modifier signs and the compound
// word marks to the table after the fit; they have no sides in the fit table.
func (p Params) addTS1Marks(gs []glyphs.Glyph) []glyphs.Glyph {
	rp := design.Regular()
	k := p.slant()
	for _, t := range []struct {
		name string
		r    rune
	}{
		{"tieaccentlowercase", 0xEB65}, {"tieaccentcapital", 0xEB63},
		{"tieaccentlowercase.new", 0}, {"tieaccentcapital.new", 0},
	} {
		path, adv, ok := glyphs.Outline(t.name, rp)
		if !ok {
			panic("assembly: no glyph " + t.name + " in the roman")
		}
		path = shearedAbout(path, k, lowestEdge(path)).Round()
		gs = append(gs, glyphs.Glyph{Name: t.name, Rune: t.r, Advance: adv, Path: path})
	}
	for _, t := range []struct {
		name string
		r    rune
	}{{"uni02F7", 0x02F7}, {"uni02F5", 0x02F5}} {
		path, _, ok := glyphs.Outline(t.name, rp)
		if !ok {
			panic("assembly: no glyph " + t.name + " in the roman")
		}
		path = shearedAbout(path, k, lowestEdge(path))
		x0, _, x1, _ := path.BBox()
		dx := math.Round(rp.Acc.MarkSide - x0)
		gs = append(gs, glyphs.Glyph{Name: t.name, Rune: t.r,
			Advance: math.Ceil(x1 + dx + rp.Acc.MarkSide - 1e-6), Path: moved(path, dx, 0).Round()})
	}
	for _, t := range []struct {
		name string
		r    rune
	}{{"cwmcapital", 0xEB12}, {"cwmascender", 0xEB11}} {
		_, adv, ok := glyphs.Outline(t.name, rp)
		if !ok {
			panic("assembly: no glyph " + t.name + " in the roman")
		}
		gs = append(gs, glyphs.Glyph{Name: t.name, Rune: t.r, Advance: adv})
	}
	return gs
}

// addCaseMarks adds the capital forms of the accents: the spacing accents,
// which stand with their lowest ink Acc.GapCap above the cap height and
// Acc.MarkSide on either side of the ink, and the combining marks above a
// letter, without advance, whose anchor is at the height of the top anchor of
// a capital.  The mark is the capital form of the sloped mark (the form "cap"
// of the roman's marks, as the composed capitals use it), placed as the lower
// case form is in compose.
func (p Params) addCaseMarks(gs []glyphs.Glyph) []glyphs.Glyph {
	rp := design.Regular()
	acc := rp.Acc
	k := p.slant()
	frame := math.Round(rp.CapHeight + acc.GapCap)
	anchorY := math.Round(rp.CapHeight + acc.GapCap - acc.Gap)
	n := slices.IndexFunc(gs, func(g glyphs.Glyph) bool { return g.Name == "n" })
	cx := -math.Round(gs[n].Advance / 2)
	// The spacing accents are named by their marks.
	for _, name := range []string{"grave", "acute", "circumflex", "tilde", "dieresis", "hungarumlaut", "ring", "caron", "breve", "macron", "dotaccent"} {
		mk := p.sloped(name, "cap")
		x0, _, x1, _ := mk.BBox()
		dx := math.Round(acc.MarkSide - x0)
		gs = append(gs, glyphs.Glyph{Name: name + ".case",
			Advance: math.Ceil(x1 + dx + acc.MarkSide - 1e-6), Path: moved(mk, dx, frame).Round()})
	}
	for _, m := range glyphs.MarkGlyphs() {
		switch m.Mark {
		case "cedilla", "ogonek", "commabelow", "dotbelow", "commaabove", "apostrophe":
			continue
		}
		if !m.Combining {
			continue
		}
		mk := p.sloped(m.Mark, "cap")
		path := moved(mk, cx+math.Round(k*acc.Gap), frame)
		gs = append(gs, glyphs.Glyph{Name: m.Name + ".case", Path: path.Round(),
			Anchors: []glyphs.Anchor{{Class: "top", X: cx, Y: anchorY}}})
	}
	return gs
}
