// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package build

import (
	"slices"

	"golang.org/x/text/language"

	"seehuhn.de/go/postscript/funit"

	"seehuhn.de/go/sfnt/glyph"
	"seehuhn.de/go/sfnt/opentype/anchor"
	"seehuhn.de/go/sfnt/opentype/classdef"
	"seehuhn.de/go/sfnt/opentype/coverage"
	"seehuhn.de/go/sfnt/opentype/gdef"
	"seehuhn.de/go/sfnt/opentype/gtab"
	"seehuhn.de/go/sfnt/opentype/markarray"

	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/kern"
)

// kernInfo returns the GPOS table with the feature kern for the scripts latn
// and DFLT, or nil if there are no kerns.  The glyph index is the glyph ID.
//
// The lookup has two subtables, tried in this order: the pairs whose kern
// differs from that of their classes (a pair with the kern 0 is a match
// which leaves the pair unkerned), and then the pairs of classes.
func kernInfo(t *kern.Table) *gtab.Info {
	if len(t.Pairs) == 0 && len(t.Single) == 0 {
		return nil
	}
	adj := func(k int) *gtab.PairAdjust {
		return &gtab.PairAdjust{First: &gtab.GposValueRecord{XAdvance: funit.Int16(k)}}
	}
	var subtables []gtab.Subtable
	if len(t.Single) > 0 {
		pairs := make(map[glyph.Pair]*gtab.PairAdjust, len(t.Single))
		for key, k := range t.Single {
			pairs[glyph.Pair{Left: glyph.ID(key[0]), Right: glyph.ID(key[1])}] = adj(k)
		}
		subtables = append(subtables, &gtab.Gpos2_1{Adjust: pairs})
	}
	if len(t.Values) > 1 {
		cov := coverage.Set{}
		c1 := classdef.Table{}
		for g, c := range t.Class1 {
			cov[glyph.ID(g)] = true
			c1[glyph.ID(g)] = uint16(c)
		}
		c2 := classdef.Table{}
		for g, c := range t.Class2 {
			c2[glyph.ID(g)] = uint16(c)
		}
		rows := make([][]*gtab.PairAdjust, len(t.Values))
		for i, row := range t.Values {
			rows[i] = make([]*gtab.PairAdjust, len(row))
			for j, k := range row {
				// A cell without value records is the zero adjustment.
				rows[i][j] = &gtab.PairAdjust{}
				if k != 0 {
					rows[i][j] = adj(k)
				}
			}
		}
		subtables = append(subtables, &gtab.Gpos2_2{Cov: cov, Class1: c1, Class2: c2, Adjust: rows})
	}
	features := &gtab.Features{Required: 0xFFFF, Optional: []gtab.FeatureIndex{0}}
	return &gtab.Info{
		ScriptList: gtab.ScriptListInfo{
			language.MustParse("und-Latn"): features,
			language.MustParse("und-Zzzz"): features,
		},
		FeatureList: gtab.FeatureListInfo{{Tag: "kern", Lookups: []gtab.LookupIndex{0}}},
		LookupList: gtab.LookupList{{
			Meta:      &gtab.LookupMetaInfo{LookupType: 2},
			Subtables: subtables,
		}},
	}
}

// markClasses are the anchor classes of the combining marks, in the order of
// their mark class numbers.
var markClasses = []string{"top", "bottom", "ogonek"}

// gposInfo returns the GPOS table with the feature kern of the kerns, if there
// are any, and the feature mark, which sets the combining marks on the anchors
// of the letters (lookup type 4), or nil if there is neither.
func gposInfo(t *kern.Table, gs []glyphs.Glyph) *gtab.Info {
	info := kernInfo(t)
	mark := markLookup(gs)
	if mark == nil {
		return info
	}
	if info == nil {
		info = &gtab.Info{}
	}
	info.LookupList = append(info.LookupList, mark)
	info.FeatureList = append(info.FeatureList, &gtab.Feature{
		Tag: "mark", Lookups: []gtab.LookupIndex{gtab.LookupIndex(len(info.LookupList) - 1)},
	})
	var optional []gtab.FeatureIndex
	for i := range info.FeatureList {
		optional = append(optional, gtab.FeatureIndex(i))
	}
	features := &gtab.Features{Required: 0xFFFF, Optional: optional}
	info.ScriptList = gtab.ScriptListInfo{
		language.MustParse("und-Latn"): features,
		language.MustParse("und-Zzzz"): features,
	}
	return info
}

// markLookup returns the mark-to-base lookup of the combining marks, or nil
// if the font has none.
func markLookup(gs []glyphs.Glyph) *gtab.LookupTable {
	sub := &gtab.Gpos4_1{MarkCov: coverage.Table{}, BaseCov: coverage.Table{}}
	for i, g := range gs {
		if !isMark(g) {
			continue
		}
		a := g.Anchors[0]
		sub.MarkCov[glyph.ID(i)] = len(sub.MarkArray)
		sub.MarkArray = append(sub.MarkArray, markarray.Record{
			Class: uint16(slices.Index(markClasses, a.Class)),
			Table: anchor.Table{X: funit.Int16(a.X), Y: funit.Int16(a.Y)},
		})
	}
	if len(sub.MarkArray) == 0 {
		return nil
	}
	for i, g := range gs {
		if isMark(g) || len(g.Anchors) == 0 {
			continue
		}
		row := make([]*anchor.Table, len(markClasses))
		for _, a := range g.Anchors {
			if c := slices.Index(markClasses, a.Class); c >= 0 {
				row[c] = &anchor.Table{X: funit.Int16(a.X), Y: funit.Int16(a.Y)}
			}
		}
		sub.BaseCov[glyph.ID(i)] = len(sub.BaseArray)
		sub.BaseArray = append(sub.BaseArray, row)
	}
	return &gtab.LookupTable{
		Meta:      &gtab.LookupMetaInfo{LookupType: 4},
		Subtables: []gtab.Subtable{sub},
	}
}

// isMark says whether the glyph is a combining mark: a glyph without advance
// with one anchor.
func isMark(g glyphs.Glyph) bool { return g.Advance == 0 && len(g.Anchors) == 1 }

// gdefInfo returns the GDEF table which says which glyphs are combining marks
// and which are base glyphs, or nil if the font has no combining marks.
func gdefInfo(gs []glyphs.Glyph) *gdef.Table {
	cls := classdef.Table{}
	marks := false
	for i, g := range gs {
		switch {
		case isMark(g):
			cls[glyph.ID(i)] = gdef.GlyphClassMark
			marks = true
		case len(g.Path) > 0:
			cls[glyph.ID(i)] = gdef.GlyphClassBase
		}
	}
	if !marks {
		return nil
	}
	return &gdef.Table{GlyphClass: cls}
}
