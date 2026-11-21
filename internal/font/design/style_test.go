// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package design

import (
	"strings"
	"testing"
)

func TestStyleValidate(t *testing.T) {
	t.Parallel()
	with := func(f func(s *Style)) Style {
		s := RegularStyle()
		f(&s)
		return s
	}
	cases := []struct {
		name  string
		style Style
		ok    bool
	}{
		{"regular", RegularStyle(), true},
		{"test variant", Style{Family: "Bedern Test", Subfamily: "Fit190", PostScript: "BedernTest-Fit190", Weight: 400}, true},
		{"bold", with(func(s *Style) { s.Bold = true; s.Weight = 700 }), true},
		{"empty family", with(func(s *Style) { s.Family = "" }), false},
		{"empty subfamily", with(func(s *Style) { s.Subfamily = "" }), false},
		{"empty PostScript name", with(func(s *Style) { s.PostScript = "" }), false},
		{"space in PostScript name", with(func(s *Style) { s.PostScript = "Bedern Regular" }), false},
		{"bracket in PostScript name", with(func(s *Style) { s.PostScript = "Bedern[1]" }), false},
		{"slash in PostScript name", with(func(s *Style) { s.PostScript = "Bedern/Regular" }), false},
		{"percent in PostScript name", with(func(s *Style) { s.PostScript = "Bedern%" }), false},
		{"non-ASCII PostScript name", with(func(s *Style) { s.PostScript = "Regularÿ" }), false},
		{"long PostScript name", with(func(s *Style) { s.PostScript = strings.Repeat("a", 64) }), false},
		{"63 character PostScript name", with(func(s *Style) { s.PostScript = strings.Repeat("a", 63) }), true},
		{"weight zero", with(func(s *Style) { s.Weight = 0 }), false},
		{"weight 1001", with(func(s *Style) { s.Weight = 1001 }), false},
		{"bold but light", with(func(s *Style) { s.Bold = true }), false},
		{"heavy but not bold", with(func(s *Style) { s.Weight = 600 }), false},
	}
	for _, c := range cases {
		err := c.style.Validate()
		if (err == nil) != c.ok {
			t.Errorf("%s: Validate() = %v, want ok=%t", c.name, err, c.ok)
		}
	}
}
