// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package build

import (
	"maps"
	"math"
	"slices"
	"strings"

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
	ss02NameID: "Rounder ties",
}

// ss01NameID and ss02NameID are the IDs of the names of the first and the
// second stylistic set.
const (
	ss01NameID = 256
	ss02NameID = 257
)

// figureNames are the names of the figures at their code points, which are
// lining and tabular.  The other forms of a figure have its name with a
// suffix: .prop (lining, proportional), .taboldstyle (old-style, tabular) and
// .oldstyle (old-style, proportional).
var figureNames = []string{"zero", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine"}

// caseSuffix is the suffix of the name of the capital form of a glyph.
const caseSuffix = ".case"

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
//	case: a glyph which has a capital form, named with the suffix .case,
//	becomes that form (type 1): the spacing accents and the combining marks
//	which stand above a letter.
//	onum, pnum, lnum, tnum: the figures become old-style, proportional,
//	lining and tabular (type 1 each); each lookup keeps the other property
//	of a figure, and lnum and tnum, which come last, undo onum and pnum.
//	ss01: the capital sharp s becomes the two letters S (type 1).
//	ss02: the ties become their rounder forms, named with the suffix .new
//	(type 1).
//
// The lookup of ccmp also gives a combining mark above its capital form after
// a letter whose top anchor stands higher than that of n.
func gsubInfo(gs []glyphs.Glyph) *gtab.Info {
	index := make(map[string]glyph.ID, len(gs))
	for i, g := range gs {
		index[g.Name] = glyph.ID(i)
	}
	info := &gtab.Info{}
	// singles returns the single substitution of the pairs of glyph names
	// whose glyphs the font has, or nil if it has none of them.  The coverage
	// indices follow the order of the glyphs.
	singles := func(pairs ...[2]string) *gtab.LookupTable {
		target := map[glyph.ID]glyph.ID{}
		for _, pr := range pairs {
			a, okA := index[pr[0]]
			b, okB := index[pr[1]]
			if okA && okB {
				target[a] = b
			}
		}
		if len(target) == 0 {
			return nil
		}
		sub := &gtab.Gsub1_2{Cov: coverage.Table{}}
		for _, a := range slices.Sorted(maps.Keys(target)) {
			sub.Cov[a] = len(sub.SubstituteGlyphIDs)
			sub.SubstituteGlyphIDs = append(sub.SubstituteGlyphIDs, target[a])
		}
		return &gtab.LookupTable{
			Meta:      &gtab.LookupMetaInfo{LookupType: 1},
			Subtables: []gtab.Subtable{sub},
		}
	}
	single := func(from, to string) *gtab.LookupTable { return singles([2]string{from, to}) }
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
	var ccmp []gtab.LookupIndex
	if len(in) > 0 && len(above) > 0 {
		sub := add(&gtab.LookupTable{
			Meta:      &gtab.LookupMetaInfo{LookupType: 1},
			Subtables: []gtab.Subtable{&dotless},
		})
		ccmp = append(ccmp, add(&gtab.LookupTable{
			Meta: &gtab.LookupMetaInfo{LookupType: 6},
			Subtables: []gtab.Subtable{&gtab.ChainedSeqContext3{
				Input:     []coverage.Set{in},
				Lookahead: []coverage.Set{above},
				Actions:   []gtab.SeqLookup{{SequenceIndex: 0, LookupListIndex: sub}},
			}},
		}))
	}

	// The capital forms: of the marks, which ccmp sets after a high letter,
	// and of every glyph which has one, which case sets.
	var caseMarks, caseAll [][2]string
	for _, g := range gs {
		base, ok := strings.CutSuffix(g.Name, caseSuffix)
		if !ok {
			continue
		}
		j, ok := index[base]
		if !ok {
			continue
		}
		caseAll = append(caseAll, [2]string{base, g.Name})
		if isMark(g) && isMark(gs[j]) && above[j] && g.Anchors[0].Class == "top" {
			caseMarks = append(caseMarks, [2]string{base, g.Name})
		}
	}
	if l := singles(caseMarks...); l != nil {
		high := coverage.Set{}
		if n, ok := index["n"]; ok {
			if top, ok := topAnchor(gs[n]); ok {
				for i, g := range gs {
					if y, ok := topAnchor(g); ok && !isMark(g) && y > top {
						high[glyph.ID(i)] = true
					}
				}
			}
		}
		marks := coverage.Set{}
		for _, pr := range caseMarks {
			marks[index[pr[0]]] = true
		}
		if len(high) > 0 {
			sub := add(l)
			ccmp = append(ccmp, add(&gtab.LookupTable{
				Meta: &gtab.LookupMetaInfo{LookupType: 6},
				Subtables: []gtab.Subtable{&gtab.ChainedSeqContext3{
					Backtrack: []coverage.Set{high},
					Input:     []coverage.Set{marks},
					Actions:   []gtab.SeqLookup{{SequenceIndex: 0, LookupListIndex: sub}},
				}},
			}))
		}
	}
	if len(ccmp) > 0 {
		feature(&common, &gtab.Feature{Tag: "ccmp", Lookups: ccmp})
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

	if l := singles(caseAll...); l != nil {
		feature(&common, &gtab.Feature{Tag: "case", Lookups: []gtab.LookupIndex{add(l)}})
	}

	// The forms of the figures: a lookup changes one property of a figure,
	// lining or old-style, tabular or proportional, and keeps the other.
	figures := func(fromA, toA, fromB, toB string) *gtab.LookupTable {
		var pairs [][2]string
		for _, f := range figureNames {
			pairs = append(pairs, [2]string{f + fromA, f + toA}, [2]string{f + fromB, f + toB})
		}
		return singles(pairs...)
	}
	for _, f := range []struct {
		tag                    string
		fromA, toA, fromB, toB string
	}{
		{"onum", "", ".taboldstyle", ".prop", ".oldstyle"},
		{"pnum", "", ".prop", ".taboldstyle", ".oldstyle"},
		{"lnum", ".taboldstyle", "", ".oldstyle", ".prop"},
		{"tnum", ".prop", "", ".oldstyle", ".taboldstyle"},
	} {
		if l := figures(f.fromA, f.toA, f.fromB, f.toB); l != nil {
			feature(&common, &gtab.Feature{Tag: f.tag, Lookups: []gtab.LookupIndex{add(l)}})
		}
	}

	if l := single("uni1E9E", "S_S"); l != nil {
		feature(&common, &gtab.Feature{
			Tag: "ss01", Lookups: []gtab.LookupIndex{add(l)},
			Params: &gtab.FeatureParamsStylisticSet{UINameID: ss01NameID},
		})
	}
	ties := singles(
		[2]string{"tieaccentlowercase", "tieaccentlowercase.new"},
		[2]string{"tieaccentcapital", "tieaccentcapital.new"})
	if ties != nil {
		feature(&common, &gtab.Feature{
			Tag: "ss02", Lookups: []gtab.LookupIndex{add(ties)},
			Params: &gtab.FeatureParamsStylisticSet{UINameID: ss02NameID},
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

// topAnchor returns the height of the anchor of the glyph for the marks
// above a letter.
func topAnchor(g glyphs.Glyph) (float64, bool) {
	for _, a := range g.Anchors {
		if a.Class == "top" {
			return a.Y, true
		}
	}
	return 0, false
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
