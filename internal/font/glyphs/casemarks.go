// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The capital forms of the accents.
//
// Each accent which stands above a letter has a capital form (marks.go),
// which the composed capitals carry.  The glyphs of this file are that form
// standing alone: as a spacing accent, with its lowest ink Acc.GapCap above
// the cap height and Acc.MarkSide on each side of its ink, and as a combining
// mark without advance, which stands over the middle of the advance of a
// preceding n, as the combining marks do.  The glyphs have no code point and
// the name of the accent with the suffix .case.

// caseSuffix is the suffix of the name of the capital form of a glyph.
const caseSuffix = ".case"

// caseSpacingList lists the accents which have a spacing capital form, by
// the name of the spacing accent.
var caseSpacingList = []struct {
	name string
	mark markKind
}{
	{"grave", markGrave},
	{"acute", markAcute},
	{"circumflex", markCircumflex},
	{"tilde", markTilde},
	{"dieresis", markDieresis},
	{"hungarumlaut", markDblAcute},
	{"ring", markRing},
	{"caron", markCaron},
	{"breve", markBreve},
	{"macron", markMacron},
	{"dotaccent", markDot},
}

// caseList lists the capital forms of the accents in glyph order: the
// spacing forms and then the combining forms, one for each combining mark
// which stands above a letter.
var caseList = func() []entry {
	var list []entry
	for _, s := range caseSpacingList {
		k := s.mark
		list = append(list, entry{s.name + caseSuffix, 0, sideOpen, sideOpen, func(p design.Params) (outline.Path, float64) {
			return capitalAccent(p, k)
		}})
	}
	for _, s := range combiningList {
		if s.mark.below() {
			continue
		}
		k := s.mark
		list = append(list, entry{s.name + caseSuffix, 0, sideMark, sideMark, func(p design.Params) (outline.Path, float64) {
			return capitalCombiningMark(p, k), 0
		}})
	}
	return list
}()

// capitalMarkHeight returns the height of the lowest ink of the capital form
// of an accent above a capital.
func capitalMarkHeight(p design.Params) float64 {
	return math.Round(p.CapHeight + p.Acc.GapCap)
}

// capitalAccent draws the capital form of an accent standing alone, with
// Acc.MarkSide on each side of its ink.
func capitalAccent(p design.Params, k markKind) (outline.Path, float64) {
	m := markOutline(p, k, formCap)
	x0, _, x1, _ := m.BBox()
	dx := math.Round(p.Acc.MarkSide - x0)
	return shiftInk(m, dx, capitalMarkHeight(p)), math.Ceil(x1 + dx + p.Acc.MarkSide - 1e-6)
}

// capitalCombiningMark draws the capital form of a combining mark, which has
// no advance.
func capitalCombiningMark(p design.Params, k markKind) outline.Path {
	return shiftInk(markOutline(p, k, formCap), combiningX(p), capitalMarkHeight(p))
}

// capitalCombiningKind returns the mark of the capital form of a combining
// mark.
func capitalCombiningKind(name string) (markKind, bool) {
	for _, s := range combiningList {
		if s.name+caseSuffix == name && !s.mark.below() {
			return s.mark, true
		}
	}
	return 0, false
}

// capitalMarkAnchor returns the anchor of the capital form of a combining
// mark: at the height of the top anchor of a capital, so that the mark keeps
// its height over a capital.
func capitalMarkAnchor(p design.Params) Anchor {
	return Anchor{Class: "top", X: combiningX(p), Y: math.Round(p.CapHeight + p.Acc.GapCap - p.Acc.Gap)}
}
