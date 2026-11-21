// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package build

import (
	"golang.org/x/text/language"

	"seehuhn.de/go/postscript/funit"

	"seehuhn.de/go/sfnt/glyph"
	"seehuhn.de/go/sfnt/opentype/classdef"
	"seehuhn.de/go/sfnt/opentype/coverage"
	"seehuhn.de/go/sfnt/opentype/gtab"

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
