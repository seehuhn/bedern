// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

import (
	"testing"

	"seehuhn.de/go/bedern/internal/font/italic"
)

// TestItalicScriptFormsAreClean asks every script form of the lowercase of
// the maths italic for a clean build at the first try, with the stem and the
// pen of the level and the width of the level (no fallback to a lighter stem
// or a narrower letter): no warning of the stroke tool and no finding of the
// italic profile of the smoothness check, as the forms at text size are in
// TestItalicIsClean.  The script forms are built from parameters made for
// them (italic.Script.ParamsFor).  The full tier.
func TestItalicScriptFormsAreClean(t *testing.T) {
	if testing.Short() {
		t.Skip("the script forms are checked in the full tier")
	}
	for _, f := range forms(t) {
		if f.Level == 0 || f.From < "a" || f.From > "z" && f.From != "dotlessi" && f.From != "uni0237" {
			continue
		}
		if want := scriptLevels[f.Level-1].stem; f.Stem < want-1e-9 {
			t.Errorf("%s: drawn with the stem factor %.2f, not %.2f", f.Name, f.Stem, want)
		}
		for _, w := range f.Warnings {
			t.Errorf("%s (%s/%d): warning: %s", f.Name, f.From, f.Level, w)
		}
		for _, g := range f.Findings {
			t.Errorf("%s (%s/%d): %s", f.Name, f.From, f.Level, g)
		}
	}
}

// TestGreekScriptFormsAreClean asks every script form of the Greek lowercase
// for a clean build at the first try, with the stem of the level (no
// fallback to a lighter stem): no warning of the stroke tool and no finding
// of the italic profile of the smoothness check, as the text forms are in
// italic.TestGreekLettersAreSmooth.  The full tier.
func TestGreekScriptFormsAreClean(t *testing.T) {
	if testing.Short() {
		t.Skip("the script forms are checked in the full tier")
	}
	for _, name := range italic.GreekNames() {
		for level := 1; level <= 2; level++ {
			g, stem, err := buildLower(name, level, false)
			if err != nil {
				t.Errorf("%s at level %d: %v", name, level, err)
				continue
			}
			if want := scriptLevels[level-1].stem; stem < want-1e-9 {
				t.Errorf("%s at level %d is drawn with the stem factor %.2f, not %.2f", name, level, stem, want)
			}
			for _, w := range g.Warnings {
				t.Errorf("%s at level %d: warning: %s", name, level, w)
			}
			for _, f := range g.Findings {
				t.Errorf("%s at level %d: %s at %.0f,%.0f: %s", name, level, f.Kind, f.X, f.Y, f.Detail)
			}
		}
	}
}
