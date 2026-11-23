// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package build

import (
	"reflect"
	"slices"
	"testing"

	"seehuhn.de/go/sfnt/glyph"
	"seehuhn.de/go/sfnt/opentype/gtab"

	"seehuhn.de/go/bedern/internal/font/glyphs"
)

// featureSingles returns, for the feature tag of the table, the single
// substitutions of its lookups by the names of the glyphs, and whether the
// table has the feature.
func featureSingles(info *gtab.Info, gs []glyphs.Glyph, tag string) (map[string]string, bool) {
	out := map[string]string{}
	found := false
	for _, f := range info.FeatureList {
		if f.Tag != tag {
			continue
		}
		found = true
		for _, l := range f.Lookups {
			for _, s := range info.LookupList[l].Subtables {
				if s, ok := s.(*gtab.Gsub1_2); ok {
					for from, i := range s.Cov {
						out[gs[from].Name] = gs[s.SubstituteGlyphIDs[i]].Name
					}
				}
			}
		}
	}
	return out, found
}

// TestFigureFeatureRules: each of the features lnum, onum, pnum and tnum
// changes one property of a figure and keeps the other, for the four forms
// of a figure, and a font which has only the figures at the code points has
// none of the four features.
func TestFigureFeatureRules(t *testing.T) {
	t.Parallel()
	gs := []glyphs.Glyph{{Name: ".notdef"}}
	for _, suffix := range []string{"", ".prop", ".taboldstyle", ".oldstyle"} {
		for _, f := range figureNames {
			gs = append(gs, glyphs.Glyph{Name: f + suffix, Advance: 500})
		}
	}
	info := gsubInfo(gs)
	if info == nil {
		t.Fatal("no GSUB table")
	}
	for tag, want := range map[string]map[string]string{
		"lnum": {"seven.taboldstyle": "seven", "seven.oldstyle": "seven.prop"},
		"onum": {"seven": "seven.taboldstyle", "seven.prop": "seven.oldstyle"},
		"pnum": {"seven": "seven.prop", "seven.taboldstyle": "seven.oldstyle"},
		"tnum": {"seven.prop": "seven", "seven.oldstyle": "seven.taboldstyle"},
	} {
		got, ok := featureSingles(info, gs, tag)
		if !ok {
			t.Errorf("no feature %s", tag)
			continue
		}
		if len(got) != 2*len(figureNames) {
			t.Errorf("%s: %d substitutions, want %d", tag, len(got), 2*len(figureNames))
		}
		for from, to := range want {
			if got[from] != to {
				t.Errorf("%s: %s becomes %q, want %s", tag, from, got[from], to)
			}
		}
	}
	if plain := gsubInfo(gs[:1+len(figureNames)]); plain != nil {
		t.Errorf("a font of the figures alone has a GSUB table: %v", plain.FeatureList)
	}
}

// TestCaseFeatureRules: the feature case gives every glyph which has a form
// with the suffix .case that form, and ccmp gives a combining mark above its
// capital form after a letter whose top anchor stands higher than that of n,
// and after no other.
func TestCaseFeatureRules(t *testing.T) {
	t.Parallel()
	top := func(y float64) []glyphs.Anchor { return []glyphs.Anchor{{Class: "top", Y: y}} }
	gs := []glyphs.Glyph{
		{Name: ".notdef"},
		{Name: "n", Advance: 500, Anchors: top(430)},
		{Name: "o", Advance: 500, Anchors: top(430)},
		{Name: "A", Advance: 700, Anchors: top(683)},
		{Name: "acute", Advance: 300},
		{Name: "acute.case", Advance: 300},
		{Name: "acutecomb", Anchors: top(430)},
		{Name: "acutecomb.case", Anchors: top(683)},
		{Name: "uni0327", Anchors: []glyphs.Anchor{{Class: "bottom"}}},
		{Name: "orphan.case", Advance: 300},
	}
	info := gsubInfo(gs)
	if info == nil {
		t.Fatal("no GSUB table")
	}
	got, ok := featureSingles(info, gs, "case")
	want := map[string]string{"acute": "acute.case", "acutecomb": "acutecomb.case"}
	if !ok || !reflect.DeepEqual(got, want) {
		t.Errorf("case: %v, want %v", got, want)
	}
	var rule *gtab.ChainedSeqContext3
	for _, f := range info.FeatureList {
		if f.Tag != "ccmp" {
			continue
		}
		for _, l := range f.Lookups {
			if s, ok := info.LookupList[l].Subtables[0].(*gtab.ChainedSeqContext3); ok && len(s.Backtrack) == 1 {
				rule = s
			}
		}
	}
	if rule == nil {
		t.Fatal("ccmp has no rule for a mark after a high letter")
	}
	names := func(set map[glyph.ID]bool) []string {
		var out []string
		for i := range set {
			out = append(out, gs[i].Name)
		}
		slices.Sort(out)
		return out
	}
	if got := names(rule.Backtrack[0]); !slices.Equal(got, []string{"A"}) {
		t.Errorf("the high letters are %v, want A alone", got)
	}
	if got := names(rule.Input[0]); !slices.Equal(got, []string{"acutecomb"}) {
		t.Errorf("the marks are %v, want acutecomb alone", got)
	}
	sub := info.LookupList[rule.Actions[0].LookupListIndex].Subtables[0].(*gtab.Gsub1_2)
	if len(sub.Cov) != 1 || gs[sub.SubstituteGlyphIDs[0]].Name != "acutecomb.case" {
		t.Errorf("the rule does not give acutecomb.case")
	}
}

// TestTieFeatureRule: the second stylistic set gives the two ties their
// rounder forms and names its entry of the name table.
func TestTieFeatureRule(t *testing.T) {
	t.Parallel()
	gs := []glyphs.Glyph{
		{Name: ".notdef"},
		{Name: "tieaccentlowercase", Advance: 300},
		{Name: "tieaccentcapital", Advance: 300},
		{Name: "tieaccentlowercase.new", Advance: 300},
		{Name: "tieaccentcapital.new", Advance: 300},
	}
	info := gsubInfo(gs)
	if info == nil {
		t.Fatal("no GSUB table")
	}
	got, ok := featureSingles(info, gs, "ss02")
	want := map[string]string{
		"tieaccentlowercase": "tieaccentlowercase.new",
		"tieaccentcapital":   "tieaccentcapital.new",
	}
	if !ok || !reflect.DeepEqual(got, want) {
		t.Errorf("ss02: %v, want %v", got, want)
	}
	for _, f := range info.FeatureList {
		if f.Tag != "ss02" {
			continue
		}
		params, ok := f.Params.(*gtab.FeatureParamsStylisticSet)
		if !ok || StylisticSetNames[ss02NameID] == "" || int(params.UINameID) != ss02NameID {
			t.Error("the feature ss02 does not name its entry of the name table")
		}
	}
}
