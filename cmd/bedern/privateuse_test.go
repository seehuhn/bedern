// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"seehuhn.de/go/sfnt"
	"seehuhn.de/go/sfnt/glyph"

	"seehuhn.de/go/bedern/internal/font/glyphs"
)

// withoutGlyph names the entries of the table of the Private Use Area for
// which the fonts have no glyph.
var withoutGlyph = []string{
	"cwmascender", "cwmcapital", "died", "leaf",
	"tieaccentcapital", "tieaccentlowercase", "Orogate", "orogate",
	"threequartersemdash", "twelveudash", "born",
}

// TestPrivateUse checks the character map of every font against the table of
// the Private Use Area: a code point of that area has a glyph only if the
// table has it, with the glyph name of the table, and every entry of the
// table has its glyph, except the entries of withoutGlyph, which have none.
// The test names the entries without a glyph.
func TestPrivateUse(t *testing.T) {
	t.Parallel()
	table := glyphs.PrivateUseTable()
	byRune := map[rune]glyphs.PrivateUse{}
	names := map[string]bool{}
	for _, e := range table {
		if e.Rune < glyphs.PrivateUseFirst || e.Rune > glyphs.PrivateUseLast {
			t.Errorf("%s: the code point %U is outside the Private Use Area", e.Name, e.Rune)
		}
		if _, dup := byRune[e.Rune]; dup {
			t.Errorf("the table has the code point %U twice", e.Rune)
		}
		if names[e.Name] {
			t.Errorf("the table has the glyph %s twice", e.Name)
		}
		if e.Name == "" || e.Description == "" || !strings.HasPrefix(e.Command, `\`) {
			t.Errorf("%U: the entry lacks a name, a command or a description", e.Rune)
		}
		byRune[e.Rune] = e
		names[e.Name] = true
	}
	for _, name := range withoutGlyph {
		if !names[name] {
			t.Errorf("withoutGlyph names %s, which the table lacks", name)
		}
	}

	r := ran(t)
	for _, f := range fonts {
		fname := filepath.Join(t.TempDir(), f.file)
		if err := os.WriteFile(fname, r.files[f.file], 0o666); err != nil {
			t.Fatal(err)
		}
		font, err := sfnt.ReadFile(fname)
		if err != nil {
			t.Fatalf("%s: %v", f.file, err)
		}
		cm, err := font.CMapTable.GetBest()
		if err != nil {
			t.Fatalf("%s: %v", f.file, err)
		}
		glyphNames := map[string]bool{}
		for gid := range font.NumGlyphs() {
			glyphNames[font.GlyphName(glyph.ID(gid))] = true
		}
		for c := rune(glyphs.PrivateUseFirst); c <= glyphs.PrivateUseLast; c++ {
			gid := cm.Lookup(c)
			if gid == 0 {
				continue
			}
			e, ok := byRune[c]
			switch name := font.GlyphName(gid); {
			case !ok:
				t.Errorf("%s: the glyph %s at %U is not in the table", f.file, name, c)
			case name != e.Name:
				t.Errorf("%s: the glyph at %U is %s, and the table has %s", f.file, c, name, e.Name)
			}
		}
		for _, e := range table {
			has := cm.Lookup(e.Rune) != 0
			listed := slices.Contains(withoutGlyph, e.Name)
			switch {
			case has && listed:
				t.Errorf("%s: %s (%U) has a glyph and is named in withoutGlyph", f.file, e.Name, e.Rune)
			case !has && !listed:
				t.Errorf("%s: no glyph for %s (%U) of the table", f.file, e.Name, e.Rune)
			case !has && glyphNames[e.Name]:
				t.Errorf("%s: the glyph %s is not at its code point %U", f.file, e.Name, e.Rune)
			case !has:
				t.Logf("%s: no glyph for %s (%U)", f.file, e.Name, e.Rune)
			}
		}
	}
}

// TestPrivateUseText checks the text of the table: one line for each entry,
// with the code point, the glyph name, the command and the description
// separated by tabs.
func TestPrivateUseText(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	if err := glyphs.WritePrivateUse(&buf); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
	table := glyphs.PrivateUseTable()
	if len(lines) != len(table) {
		t.Fatalf("%d lines for %d entries", len(lines), len(table))
	}
	for i, e := range table {
		want := []string{fmt.Sprintf("U+%04X", e.Rune), e.Name, e.Command, e.Description}
		if got := strings.Split(lines[i], "\t"); !slices.Equal(got, want) {
			t.Errorf("line %d: %q, want %q", i+1, got, want)
		}
	}
}
