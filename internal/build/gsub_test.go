// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package build

import (
	"bytes"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"testing"

	"seehuhn.de/go/sfnt/header"
	"seehuhn.de/go/sfnt/name"
	"seehuhn.de/go/sfnt/opentype/gtab"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/font/glyphs/glyphstest"
)

// TestLigatureOrder: the ligatures of three letters are tried before those of
// two, so that ffi and ffl are not taken for ff followed by i and l.
func TestLigatureOrder(t *testing.T) {
	t.Parallel()
	gs := glyphstest.Set(design.Regular())
	info := gsubInfo(gs)
	if info == nil {
		t.Fatal("no GSUB table")
	}
	var liga *gtab.Gsub4_1
	for _, f := range info.FeatureList {
		if f.Tag == "liga" {
			liga = info.LookupList[f.Lookups[0]].Subtables[0].(*gtab.Gsub4_1)
		}
	}
	if liga == nil || len(liga.Repl) != 1 {
		t.Fatal("want one ligature set, for f")
	}
	var names []string
	prev := 99
	for _, l := range liga.Repl[0] {
		names = append(names, gs[l.Out].Name)
		if len(l.In) > prev {
			t.Errorf("the ligature %s with %d letters comes after one with %d", gs[l.Out].Name, len(l.In)+1, prev+1)
		}
		prev = len(l.In)
	}
	for _, want := range []string{"ff", "fi", "fl", "ffi", "ffl", "f_j"} {
		if !slices.Contains(names, want) {
			t.Errorf("the ligature %s is missing from %v", want, names)
		}
	}
}

// TestLigaturesInFile: a shaping engine sets the ligatures of the written
// font, the plain letters with the feature liga switched off, and no ligature
// before an accented i (with an ogonek too, which keeps its dot) or before i
// with a mark above it, where the dotless i takes the mark.
func TestLigaturesInFile(t *testing.T) {
	t.Parallel()
	path := needTool(t, "hb-shape")
	fname := buildTestFont(t)
	shape := func(text string, args ...string) string { return shapeNames(t, path, fname, text, args...) }
	for _, c := range []struct{ text, want, plain string }{
		{"fi", "fi", "f i"},
		{"fl", "fl", "f l"},
		{"ff", "ff", "f f"},
		{"ffi", "ffi", "f f i"},
		{"ffl", "ffl", "f f l"},
		{"fj", "f_j", "f j"},
		{"fff", "ff f", "f f f"},
		{"ffffi", "ff ffi", "f f f f i"},
		{"fí", "f iacute", "f iacute"},
		{"fï", "f idieresis", "f idieresis"},
		{"fi̊", "f dotlessi uni030A", "f dotlessi uni030A"},
		{"fi\u0328", "f iogonek", "f iogonek"},
	} {
		if got := shape(c.text); got != c.want {
			t.Errorf("%q: %s, want %s", c.text, got, c.want)
		}
		if got := shape(c.text, "--features=-liga"); got != c.plain {
			t.Errorf("%q without liga: %s, want %s", c.text, got, c.plain)
		}
	}
}

// glyphNumber matches the glyph number of an entry of the output of hb-shape.
var glyphNumber = regexp.MustCompile(`(\[|\|)\d+`)

// shapeNames returns the names of the glyphs which hb-shape (at path) sets
// for the text in the font file, separated by spaces.
func shapeNames(t *testing.T, path, fname, text string, args ...string) string {
	t.Helper()
	cmd := exec.Command(path, append(args, fname, text)...)
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	if err := cmd.Run(); err != nil {
		t.Fatalf("hb-shape: %v\n%s", err, errb.String())
	}
	res, err := parseHBShape(out.String())
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, g := range res {
		names = append(names, g.name)
	}
	return strings.Join(names, " ")
}

// TestNoDottedLigaturesInTurkish: in Turkish, Azerbaijani and Crimean Tatar a
// shaping engine forms no ligature of f with i, sets the other ligatures, and
// sets the dotless i under a mark above the letter; in the other languages
// the ligatures are formed.  The form of i of these languages has the outline,
// the advance, the kerning and the anchors of i.
func TestNoDottedLigaturesInTurkish(t *testing.T) {
	t.Parallel()
	path := needTool(t, "hb-shape")
	fname := buildTestFont(t)
	for _, lang := range []string{"tr", "az", "crh"} {
		for _, c := range []struct{ text, want string }{
			{"fi", "f i.loclTRK"},
			{"ffi", "ff i.loclTRK"},
			{"fl", "fl"},
			{"ffl", "ffl"},
			{"fj", "f_j"},
			{"fı", "f dotlessi"},
			{"i\u030A", "dotlessi uni030A"},
			{"i\u0323", "i.loclTRK dotbelowcomb"},
		} {
			if got := shapeNames(t, path, fname, c.text, "--language="+lang); got != c.want {
				t.Errorf("%q in the language %s: %s, want %s", c.text, lang, got, c.want)
			}
		}
	}
	for _, lang := range []string{"en", "de", "ro", "nl"} {
		if got := shapeNames(t, path, fname, "fi ffi", "--language="+lang); got != "fi space ffi" {
			t.Errorf("\"fi ffi\" in the language %s: %s, want fi space ffi", lang, got)
		}
	}

	// A text has the same advances and offsets with both forms of i: the
	// kerns and the places of the marks are those of i.
	positions := func(text string, args ...string) string {
		args = append(args, "--no-glyph-names", "--no-clusters", fname, text)
		out, err := exec.Command(path, args...).Output()
		if err != nil {
			t.Fatalf("hb-shape: %v", err)
		}
		return glyphNumber.ReplaceAllString(strings.TrimSpace(string(out)), "$1")
	}
	for _, text := range []string{"Tifi", "Vi", "fi\u0323", "f\u012F"} {
		tr := positions(text, "--language=tr")
		en := positions(text, "--language=en", "--features=-liga")
		if tr != en {
			t.Errorf("%q: positions %s in Turkish, %s with the plain i", text, tr, en)
		}
	}

	gs := glyphstest.Set(design.Regular())
	var plain, local *glyphs.Glyph
	for i := range gs {
		switch gs[i].Name {
		case "i":
			plain = &gs[i]
		case "i.loclTRK":
			local = &gs[i]
		}
	}
	if plain == nil || local == nil {
		t.Fatal("the font lacks i or i.loclTRK")
	}
	if !reflect.DeepEqual(plain.Path, local.Path) || plain.Advance != local.Advance || !reflect.DeepEqual(plain.Anchors, local.Anchors) {
		t.Error("i.loclTRK differs from i in its outline, its advance or its anchors")
	}
}

// TestStylisticSet: with the first stylistic set a shaping engine sets the
// two letters S for the capital sharp s, and without it the capital sharp s;
// the feature names its entry of the name table, which every name table of
// the written file has.
func TestStylisticSet(t *testing.T) {
	t.Parallel()
	path := needTool(t, "hb-shape")
	fname := buildTestFont(t)
	for _, c := range []struct{ args, want string }{
		{"--features=ss01", "G R O S_S E"},
		{"--features=-ss01", "G R O uni1E9E E"},
		{"--language=de", "G R O uni1E9E E"},
	} {
		if got := shapeNames(t, path, fname, "GROẞE", c.args); got != c.want {
			t.Errorf("GROẞE with %s: %s, want %s", c.args, got, c.want)
		}
	}
	if got := shapeNames(t, path, fname, "Maß", "--features=ss01"); got != "M a germandbls" {
		t.Errorf("Maß with ss01: %s, want M a germandbls", got)
	}

	font, err := regularFont()
	if err != nil {
		t.Fatal(err)
	}
	var id name.ID
	for _, f := range font.Gsub.FeatureList {
		if f.Tag != "ss01" {
			continue
		}
		params, ok := f.Params.(*gtab.FeatureParamsStylisticSet)
		if !ok {
			t.Fatal("the feature ss01 has no parameters")
		}
		id = name.ID(params.UINameID)
	}
	if id < 256 {
		t.Fatalf("the feature ss01 names the entry %d of the name table, want one from 256", id)
	}
	got := roundTrip(t, font)
	found := false
	for _, f := range got.Gsub.FeatureList {
		if params, ok := f.Params.(*gtab.FeatureParamsStylisticSet); ok && f.Tag == "ss01" && name.ID(params.UINameID) == id {
			found = true
		}
	}
	if !found {
		t.Error("the written file has no feature ss01 with the name of the set")
	}
	data, err := os.ReadFile(fname)
	if err != nil {
		t.Fatal(err)
	}
	tables := nameTables(t, data)
	if len(tables) == 0 {
		t.Fatal("the written file has no name table")
	}
	for key, tab := range tables {
		if tab.Extra[id] != StylisticSetNames[id] || tab.Extra[id] == "" {
			t.Errorf("the name table %s has %q for the stylistic set, want %q", key, tab.Extra[id], StylisticSetNames[id])
		}
	}
}

// nameTables returns the name tables of the font file data, by platform and
// language.
func nameTables(t *testing.T, data []byte) map[string]*name.Table {
	t.Helper()
	r := bytes.NewReader(data)
	h, err := header.Read(r)
	if err != nil {
		t.Fatal(err)
	}
	b, err := h.ReadTableBytes(r, "name")
	if err != nil {
		t.Fatal(err)
	}
	info, err := name.Decode(b)
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]*name.Table{}
	for lang, tab := range info.Mac {
		out["Mac "+lang] = tab
	}
	for lang, tab := range info.Windows {
		out["Windows "+lang] = tab
	}
	return out
}
