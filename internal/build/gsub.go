// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package build

import (
	"maps"
	"math"
	"slices"

	"golang.org/x/text/language"

	"seehuhn.de/go/postscript/funit"

	"seehuhn.de/go/sfnt/glyph"
	"seehuhn.de/go/sfnt/name"
	"seehuhn.de/go/sfnt/opentype/coverage"
	"seehuhn.de/go/sfnt/opentype/gdef"
	"seehuhn.de/go/sfnt/opentype/gtab"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/glyphs"
)

// Language systems of the script latn which have the feature locl: Turkish,
// Azerbaijani and Crimean Tatar.
var loclLanguages = []string{"tr-Latn", "az-Latn", "crh-Latn"}

// StylisticSetNames gives the names of the stylistic sets by the IDs which
// they have in the name table.
var StylisticSetNames = map[name.ID]string{
	ss01NameID: "Capital sharp s as two letters S",
}

// ss01NameID is the ID of the name of the first stylistic set.
const ss01NameID = 256

// gsubInfo returns the GSUB table for the scripts latn and DFLT, or nil if the
// font has none of its features.  The lookups are applied in this order.
//
//	locl: in the language systems of loclLanguages i becomes i.loclTRK (a
//	single substitution, type 1), which takes no part in a ligature.
//	ccmp: i and j followed by a combining mark which stands above the letter
//	become the dotless ı and ȷ (a chained context lookup, type 6, which calls a
//	single substitution, type 1).
//	liga: the ligatures (type 4), one lookup whose ligatures are tried in the
//	order of the number of their letters, the longest first.
//	ss01: the capital sharp s becomes the two letters S (type 1).
func gsubInfo(gs []glyphs.Glyph) *gtab.Info {
	index := make(map[string]glyph.ID, len(gs))
	for i, g := range gs {
		index[g.Name] = glyph.ID(i)
	}
	info := &gtab.Info{}
	single := func(from, to string) *gtab.LookupTable {
		a, okA := index[from]
		b, okB := index[to]
		if !okA || !okB {
			return nil
		}
		return &gtab.LookupTable{
			Meta: &gtab.LookupMetaInfo{LookupType: 1},
			Subtables: []gtab.Subtable{&gtab.Gsub1_2{
				Cov:                coverage.Table{a: 0},
				SubstituteGlyphIDs: []glyph.ID{b},
			}},
		}
	}
	add := func(l *gtab.LookupTable) gtab.LookupIndex {
		info.LookupList = append(info.LookupList, l)
		return gtab.LookupIndex(len(info.LookupList) - 1)
	}

	// The features of every language system, and those of the language
	// systems of loclLanguages alone.
	var common, local []gtab.FeatureIndex
	feature := func(list *[]gtab.FeatureIndex, f *gtab.Feature) {
		info.FeatureList = append(info.FeatureList, f)
		*list = append(*list, gtab.FeatureIndex(len(info.FeatureList)-1))
	}

	if l := single("i", "i.loclTRK"); l != nil {
		feature(&local, &gtab.Feature{Tag: "locl", Lookups: []gtab.LookupIndex{add(l)}})
	}

	// The coverage indices follow the order of the glyphs.
	dotless := gtab.Gsub1_2{Cov: coverage.Table{}}
	in := coverage.Set{}
	target := map[glyph.ID]glyph.ID{}
	for _, pr := range [][2]string{{"i", "dotlessi"}, {"i.loclTRK", "dotlessi"}, {"j", "uni0237"}} {
		a, okA := index[pr[0]]
		b, okB := index[pr[1]]
		if okA && okB {
			target[a] = b
		}
	}
	for _, a := range slices.Sorted(maps.Keys(target)) {
		dotless.Cov[a] = len(dotless.SubstituteGlyphIDs)
		dotless.SubstituteGlyphIDs = append(dotless.SubstituteGlyphIDs, target[a])
		in[a] = true
	}
	above := coverage.Set{}
	for i, g := range gs {
		if isMark(g) && g.Anchors[0].Class == "top" {
			above[glyph.ID(i)] = true
		}
	}
	if len(in) > 0 && len(above) > 0 {
		sub := add(&gtab.LookupTable{
			Meta:      &gtab.LookupMetaInfo{LookupType: 1},
			Subtables: []gtab.Subtable{&dotless},
		})
		chain := add(&gtab.LookupTable{
			Meta: &gtab.LookupMetaInfo{LookupType: 6},
			Subtables: []gtab.Subtable{&gtab.ChainedSeqContext3{
				Input:     []coverage.Set{in},
				Lookahead: []coverage.Set{above},
				Actions:   []gtab.SeqLookup{{SequenceIndex: 0, LookupListIndex: sub}},
			}},
		})
		feature(&common, &gtab.Feature{Tag: "ccmp", Lookups: []gtab.LookupIndex{chain}})
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
		l := add(&gtab.LookupTable{
			Meta:      &gtab.LookupMetaInfo{LookupType: 4},
			Subtables: []gtab.Subtable{sub},
		})
		feature(&common, &gtab.Feature{Tag: "liga", Lookups: []gtab.LookupIndex{l}})
	}

	if l := single("uni1E9E", "S_S"); l != nil {
		feature(&common, &gtab.Feature{
			Tag: "ss01", Lookups: []gtab.LookupIndex{add(l)},
			Params: &gtab.FeatureParamsStylisticSet{UINameID: ss01NameID},
		})
	}

	if len(info.FeatureList) == 0 {
		return nil
	}
	features := &gtab.Features{Required: 0xFFFF, Optional: common}
	info.ScriptList = gtab.ScriptListInfo{
		language.MustParse("und-Latn"): features,
		language.MustParse("und-Zzzz"): features,
	}
	if len(local) > 0 {
		all := slices.Sorted(slices.Values(append(slices.Clone(common), local...)))
		for _, tag := range loclLanguages {
			info.ScriptList[language.MustParse(tag)] = &gtab.Features{Required: 0xFFFF, Optional: all}
		}
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
