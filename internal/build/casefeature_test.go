// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package build

import (
	"bytes"
	"math"
	"os/exec"
	"regexp"
	"strconv"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/font/glyphs/glyphstest"
)

// TestCaseInFile: with the feature case a shaping engine sets the capital
// forms of the spacing accents and of the combining marks above a letter,
// and leaves the marks below a letter and the letters as they are; without
// it, a combining mark above takes its capital form after a capital and
// after a letter with an ascender, and keeps its form after the other
// letters.
func TestCaseInFile(t *testing.T) {
	t.Parallel()
	path := needTool(t, "hb-shape")
	fname := buildTestFont(t)
	for _, c := range []struct{ text, args, want string }{
		{"`´ˆ˜¨", "--features=case", "grave.case acute.case circumflex.case tilde.case dieresis.case"},
		{"˝˚ˇ˘¯˙", "--features=case", "hungarumlaut.case ring.case caron.case breve.case macron.case dotaccent.case"},
		{"`´ˆ˜¨˝˚ˇ˘¯˙", "--features=-case", "grave acute circumflex tilde dieresis hungarumlaut ring caron breve macron dotaccent"},
		{"¸˛", "--features=case", "cedilla ogonek"},
		{"X́", "--features=case", "X acutecomb.case"},
		{"x́", "--features=case", "x acutecomb.case"},
		{"X̧̣", "--features=case", "X uni0327 dotbelowcomb"},
		{"ÉCOLE", "--features=case", "Eacute C O L E"},
		{"X́", "--features=-case", "X acutecomb.case"},
		{"M̀Q̂Z̃", "--features=-case", "M gravecomb.case Q uni0302.case Z tildecomb.case"},
		{"W̄X̆K̇", "--features=-case", "W uni0304.case X uni0306.case K uni0307.case"},
		{"Q̈X̊M̋F̌", "--features=-case", "Q uni0308.case X uni030A.case M uni030B.case F uni030C.case"},
		{"b̂ḱ", "--features=-case", "b uni0302.case k acutecomb.case"},
		{"x́q̂", "--features=-case", "x acutecomb q uni0302"},
		{"X̧", "--features=-case", "X uni0327"},
		{"X́", "--features=-ccmp", "X acutecomb"},
	} {
		if got := shapeNames(t, path, fname, c.text, c.args); got != c.want {
			t.Errorf("%q with %s: %s, want %s", c.text, c.args, got, c.want)
		}
	}
}

// markOffset matches the offset of a mark in the output of hb-shape.
var markOffset = regexp.MustCompile(`=\d+@(-?\d+),(-?\d+)\+0\]`)

// TestCapitalMarkPosition: a shaping engine sets the capital form of a
// combining mark with its lowest ink Acc.GapCap above the cap height over a
// capital and above the ascender height over a letter with an ascender, over
// the top anchor of the letter.
func TestCapitalMarkPosition(t *testing.T) {
	t.Parallel()
	path := needTool(t, "hb-shape")
	fname := buildTestFont(t)
	p := design.Regular()
	byName := map[string]glyphs.Glyph{}
	for _, g := range glyphstest.Set(p) {
		byName[g.Name] = g
	}
	mark := byName["uni0302.case"]
	mx0, my0, mx1, _ := mark.Path.BBox()
	for _, c := range []struct {
		text, base string
		top        float64
	}{
		{"X̂", "X", p.CapHeight},
		{"V̂", "V", p.CapHeight},
		{"k̂", "k", p.Ascender},
	} {
		cmd := exec.Command(path, fname, c.text)
		var out bytes.Buffer
		cmd.Stdout = &out
		if err := cmd.Run(); err != nil {
			t.Fatalf("hb-shape: %v", err)
		}
		m := markOffset.FindStringSubmatch(out.String())
		if m == nil {
			t.Errorf("%q: no mark with an offset in %s", c.text, out.String())
			continue
		}
		dx, _ := strconv.ParseFloat(m[1], 64)
		dy, _ := strconv.ParseFloat(m[2], 64)
		base := byName[c.base]
		if got, want := my0+dy, math.Round(c.top+p.Acc.GapCap); math.Abs(got-want) > 1 {
			t.Errorf("%q: the lowest ink of the mark is at %.0f, want %.0f", c.text, got, want)
		}
		var topX float64
		for _, a := range base.Anchors {
			if a.Class == "top" {
				topX = a.X
			}
		}
		if got := base.Advance + dx + (mx0+mx1)/2; math.Abs(got-topX) > 1 {
			t.Errorf("%q: the middle of the mark is at %.0f, and the top anchor of %s at %.0f", c.text, got, c.base, topX)
		}
	}
}
