// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package build

import (
	"bytes"
	"os/exec"
	"strconv"
	"testing"

	"seehuhn.de/go/sfnt"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/glyphs/glyphstest"
)

// TestKernFeatureInFile: the font as written carries the feature kern, and a
// shaping engine applies it: the first glyph of "AV", "To" and "ry" is
// narrower than in the font without kerning.
func TestKernFeatureInFile(t *testing.T) {
	t.Parallel()
	path := needTool(t, "hb-shape")
	p := design.Regular()
	gs := glyphstest.Set(p)
	kerned, err := regularFont()
	if err != nil {
		t.Fatal(err)
	}
	plain, err := FontWith(p, design.RegularStyle(), gs, testNow, Options{NoKern: true})
	if err != nil {
		t.Fatal(err)
	}
	hasKern := func(f *sfnt.Font) bool {
		if f.Gpos == nil {
			return false
		}
		for _, ft := range f.Gpos.FeatureList {
			if ft.Tag == "kern" {
				return true
			}
		}
		return false
	}
	if !hasKern(kerned) || hasKern(plain) {
		t.Fatalf("feature kern: kerned %v, unkerned %v", hasKern(kerned), hasKern(plain))
	}
	kf, pf := writeFont(t, kerned), writeFont(t, plain)
	shape := func(fname, text string) []glyphAdvance {
		cmd := exec.Command(path, fname, text)
		var out, errb bytes.Buffer
		cmd.Stdout, cmd.Stderr = &out, &errb
		if err := cmd.Run(); err != nil {
			t.Fatalf("hb-shape: %v\n%s", err, errb.String())
		}
		res, err := parseHBShape(out.String())
		if err != nil {
			t.Fatal(err)
		}
		return res
	}
	for _, s := range []string{"AV", "To", "fv"} {
		k, u := shape(kf, s), shape(pf, s)
		if len(k) != 2 || len(u) != 2 {
			t.Fatalf("%s: %v %v", s, k, u)
		}
		kw, err1 := strconv.Atoi(k[0].advance)
		uw, err2 := strconv.Atoi(u[0].advance)
		if err1 != nil || err2 != nil {
			t.Fatal(err1, err2)
		}
		if kw >= uw {
			t.Errorf("%s: the advance of the first glyph is %d with kerning and %d without", s, kw, uw)
		}
		if k[1] != u[1] {
			t.Errorf("%s: the second glyph changed: %v and %v", s, k[1], u[1])
		}
	}
}
