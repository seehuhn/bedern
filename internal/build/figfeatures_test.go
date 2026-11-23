// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package build

import (
	"os/exec"
	"strconv"
	"strings"
	"testing"
)

// TestFigureFeatures: a shaping engine sets the lining tabular figures
// without a feature, the old-style figures with onum and the proportional
// ones with pnum, each keeping the other property; lnum and tnum undo onum
// and pnum, alone and together.  The tabular forms have one advance, the
// proportional 1 is narrower than it, and the proportional forms differ in
// their advances.
func TestFigureFeatures(t *testing.T) {
	t.Parallel()
	path := needTool(t, "hb-shape")
	fname := buildTestFont(t)
	figures := map[rune]string{'0': "zero", '1': "one", '2': "two", '4': "four", '6': "six", '8': "eight", '9': "nine"}
	const text = "1984 2026"
	want := func(suffix string) string {
		var names []string
		for _, r := range text {
			if r == ' ' {
				names = append(names, "space")
			} else {
				names = append(names, figures[r]+suffix)
			}
		}
		return strings.Join(names, " ")
	}
	for _, c := range []struct{ features, suffix string }{
		{"", ""},
		{"onum", ".taboldstyle"},
		{"pnum", ".prop"},
		{"onum,pnum", ".oldstyle"},
		{"lnum", ""},
		{"tnum", ""},
		{"lnum,tnum", ""},
		{"onum,lnum", ""},
		{"onum,tnum", ".taboldstyle"},
		{"pnum,lnum", ".prop"},
		{"pnum,tnum", ""},
		{"onum,pnum,lnum", ".prop"},
		{"onum,pnum,tnum", ".taboldstyle"},
		{"onum,pnum,lnum,tnum", ""},
	} {
		if got := shapeNames(t, path, fname, text, "--features="+c.features); got != want(c.suffix) {
			t.Errorf("%q with the features %q: %s, want %s", text, c.features, got, want(c.suffix))
		}
	}

	advances := func(features string) []int {
		out, err := exec.Command(path, "--features="+features, fname, "1984").Output()
		if err != nil {
			t.Fatalf("hb-shape: %v", err)
		}
		res, err := parseHBShape(string(out))
		if err != nil {
			t.Fatal(err)
		}
		var as []int
		for _, g := range res {
			a, err := strconv.Atoi(g.advance)
			if err != nil {
				t.Fatal(err)
			}
			as = append(as, a)
		}
		return as
	}
	tabular := advances("")[0]
	for _, features := range []string{"", "onum"} {
		for _, a := range advances(features) {
			if a != tabular {
				t.Errorf("1984 with the features %q: the advances are %v, want %d for each", features, advances(features), tabular)
				break
			}
		}
	}
	for _, features := range []string{"pnum", "onum,pnum"} {
		as := advances(features)
		if as[0] >= tabular {
			t.Errorf("1984 with the features %q: the advance of 1 is %d, want less than %d", features, as[0], tabular)
		}
		if as[0] >= as[1] || as[0] >= as[3] {
			t.Errorf("1984 with the features %q: the advances are %v, want that of 1 the least", features, as)
		}
	}
}
