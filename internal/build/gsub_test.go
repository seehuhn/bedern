// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package build

import (
	"bytes"
	"os/exec"
	"slices"
	"strings"
	"testing"

	"seehuhn.de/go/sfnt/opentype/gtab"

	"seehuhn.de/go/bedern/internal/font/design"
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
	shape := func(text string, args ...string) string {
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
