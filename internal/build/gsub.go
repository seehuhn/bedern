// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package build

import (
	"math"
	"slices"

	"golang.org/x/text/language"

	"seehuhn.de/go/postscript/funit"

	"seehuhn.de/go/sfnt/glyph"
	"seehuhn.de/go/sfnt/opentype/coverage"
	"seehuhn.de/go/sfnt/opentype/gdef"
	"seehuhn.de/go/sfnt/opentype/gtab"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/glyphs"
)

// gsubInfo returns the GSUB table for the scripts latn and DFLT, or nil if the
// font has neither ligatures nor dotless letters.  It has two features.
//
//	ccmp: i and j followed by a combining mark which stands above the letter
//	become the dotless ı and ȷ (a chained context lookup, type 6, which calls a
//	single substitution, type 1).
//	liga: the ligatures (type 4), one lookup whose ligatures are tried in the
//	order of the number of their letters, the longest first.
func gsubInfo(gs []glyphs.Glyph) *gtab.Info {
	index := make(map[string]glyph.ID, len(gs))
	for i, g := range gs {
		index[g.Name] = glyph.ID(i)
	}
	info := &gtab.Info{}

	dotless := gtab.Gsub1_2{Cov: coverage.Table{}}
	in := coverage.Set{}
	for _, pr := range [][2]string{{"i", "dotlessi"}, {"j", "uni0237"}} {
		a, okA := index[pr[0]]
		b, okB := index[pr[1]]
		if !okA || !okB {
			continue
		}
		dotless.Cov[a] = len(dotless.SubstituteGlyphIDs)
		dotless.SubstituteGlyphIDs = append(dotless.SubstituteGlyphIDs, b)
		in[a] = true
	}
	above := coverage.Set{}
	for i, g := range gs {
		if isMark(g) && g.Anchors[0].Class == "top" {
			above[glyph.ID(i)] = true
		}
	}
	if len(in) > 0 && len(above) > 0 {
		info.LookupList = append(info.LookupList,
			&gtab.LookupTable{
				Meta:      &gtab.LookupMetaInfo{LookupType: 1},
				Subtables: []gtab.Subtable{&dotless},
			},
			&gtab.LookupTable{
				Meta: &gtab.LookupMetaInfo{LookupType: 6},
				Subtables: []gtab.Subtable{&gtab.ChainedSeqContext3{
					Input:     []coverage.Set{in},
					Lookahead: []coverage.Set{above},
					Actions:   []gtab.SeqLookup{{SequenceIndex: 0, LookupListIndex: 0}},
				}},
			})
		info.FeatureList = append(info.FeatureList, &gtab.Feature{
			Tag: "ccmp", Lookups: []gtab.LookupIndex{1},
		})
	}

	type lig struct {
		first glyph.ID
		rule  gtab.Ligature
		n     int
	}
	var ligs []lig
	for i, g := range gs {
		if len(g.Parts) == 0 {
			continue
		}
		l := lig{n: len(g.Parts), rule: gtab.Ligature{Out: glyph.ID(i)}}
		ok := true
		for k, name := range g.Parts {
			id, found := index[name]
			ok = ok && found
			if k == 0 {
				l.first = id
			} else {
				l.rule.In = append(l.rule.In, id)
			}
		}
		if ok {
			ligs = append(ligs, l)
		}
	}
	if len(ligs) > 0 {
		slices.SortStableFunc(ligs, func(a, b lig) int { return b.n - a.n })
		sub := &gtab.Gsub4_1{Cov: coverage.Table{}}
		for _, l := range ligs {
			c, ok := sub.Cov[l.first]
			if !ok {
				c = len(sub.Repl)
				sub.Cov[l.first] = c
				sub.Repl = append(sub.Repl, nil)
			}
			sub.Repl[c] = append(sub.Repl[c], l.rule)
		}
		info.LookupList = append(info.LookupList, &gtab.LookupTable{
			Meta:      &gtab.LookupMetaInfo{LookupType: 4},
			Subtables: []gtab.Subtable{sub},
		})
		info.FeatureList = append(info.FeatureList, &gtab.Feature{
			Tag: "liga", Lookups: []gtab.LookupIndex{gtab.LookupIndex(len(info.LookupList) - 1)},
		})
	}
	if len(info.FeatureList) == 0 {
		return nil
	}
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

// ligatureCarets returns the caret list of the ligatures, or nil if there are
// none.
func ligatureCarets(p design.Params, gs []glyphs.Glyph) *gdef.LigCaretList {
	list := &gdef.LigCaretList{Cov: coverage.Table{}}
	for i, g := range gs {
		if len(g.Parts) == 0 {
			continue
		}
		var row []gdef.CaretValue
		for _, x := range glyphs.LigatureCarets(p, g.Parts) {
			row = append(row, gdef.CaretValue{Coordinate: funit.Int16(math.Round(x))})
		}
		list.Cov[glyph.ID(i)] = len(list.Carets)
		list.Carets = append(list.Carets, row)
	}
	if len(list.Carets) == 0 {
		return nil
	}
	return list
}
