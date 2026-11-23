// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The composed letters and the marks, as data for the assembly of a font from
// a table of glyphs of its own.

// A Composition is a composed letter of the set.  Mark names its mark, as
// MarkOutline takes it, and Form the form of the mark in the letter.  Joined
// says whether the letter is the union of the base and the mark.
type Composition struct {
	Name   string
	Rune   rune
	Base   string
	Mark   string
	Form   string
	Joined bool
}

// A MarkGlyph is a spacing accent or a combining mark of the set.
type MarkGlyph struct {
	Name      string
	Rune      rune
	Mark      string
	Combining bool
}

// markNames are the names of the marks for MarkOutline.
var markNames = map[markKind]string{
	markGrave:      "grave",
	markAcute:      "acute",
	markCircumflex: "circumflex",
	markCaron:      "caron",
	markTilde:      "tilde",
	markDieresis:   "dieresis",
	markMacron:     "macron",
	markBreve:      "breve",
	markRing:       "ring",
	markDot:        "dotaccent",
	markDblAcute:   "hungarumlaut",
	markCedilla:    "cedilla",
	markOgonek:     "ogonek",
	markCommaBelow: "commabelow",
	markCommaAbove: "commaabove",
	markApostrophe: "apostrophe",
}

// formNames are the names of the forms of a mark for MarkOutline.
var formNames = map[markForm]string{
	formLower:  "lower",
	formCap:    "cap",
	formNarrow: "narrow",
}

// Compositions returns the composed letters of the set in glyph order.
func Compositions() []Composition {
	out := make([]Composition, 0, len(composedList))
	for _, c := range composedList {
		comp := c.composite
		out = append(out, Composition{
			Name: c.name, Rune: c.r, Base: c.base,
			Mark: markNames[c.mark], Form: formNames[comp.form()], Joined: comp.joined(),
		})
	}
	return out
}

// MarkGlyphs returns the spacing accents and the combining marks of the set
// in glyph order.
func MarkGlyphs() []MarkGlyph {
	var out []MarkGlyph
	for _, s := range spacingList {
		out = append(out, MarkGlyph{Name: s.name, Rune: s.r, Mark: markNames[s.mark]})
	}
	for _, s := range combiningList {
		out = append(out, MarkGlyph{Name: s.name, Rune: s.r, Mark: markNames[s.mark], Combining: true})
	}
	return out
}

// MarkOutline returns the unrounded outline of the named mark in the named
// form ("lower", "cap" or "narrow") in its frame, as marks.go describes it.
// The result is false for an unknown mark or form.
func MarkOutline(p design.Params, mark, form string) (outline.Path, bool) {
	var k markKind
	var f markForm
	found := 0
	for kk, n := range markNames {
		if n == mark {
			k, found = kk, found+1
		}
	}
	for ff, n := range formNames {
		if n == form {
			f, found = ff, found+1
		}
	}
	if found != 2 {
		return nil, false
	}
	m := markOutline(p, k, f)
	out := make(outline.Path, len(m))
	for i, c := range m {
		out[i] = outline.Contour{Start: c.Start, Segs: append([]outline.Segment(nil), c.Segs...)}
	}
	return out, true
}
