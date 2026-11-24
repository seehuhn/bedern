// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"time"

	"seehuhn.de/go/sfnt"

	"seehuhn.de/go/bedern/internal/build"
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/glyphs"
)

// TrialFile is the name of the trial font file.
const TrialFile = "Bedern-Italic-Trial.otf"

// TrialStyle is the style of the trial font, named as a trial.
func TrialStyle() design.Style {
	return design.Style{
		Family:     "Bedern Italic Trial",
		Subfamily:  "Italic",
		PostScript: "BedernItalicTrial-Italic",
		Weight:     400,
		Italic:     true,
	}
}

// Glyphs returns the drawn letters, built.
func Glyphs(p Params) ([]*Glyph, error) {
	var out []*Glyph
	for _, name := range Names() {
		g, err := Build(p, name)
		if err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, nil
}

// TrialFont returns the trial font: the .notdef and the space of the roman
// and the italic letters drawn so far, without kerning.
func TrialFont(p Params, now time.Time) (*sfnt.Font, error) {
	rp := design.Regular()
	var gs []glyphs.Glyph
	for _, g := range glyphs.Set(rp) {
		if g.Name == ".notdef" || g.Name == "space" {
			gs = append(gs, glyphs.Glyph{Name: g.Name, Rune: g.Rune, Advance: g.Advance, Path: g.Path})
		}
	}
	its, err := Glyphs(p)
	if err != nil {
		return nil, err
	}
	for _, g := range its {
		gs = append(gs, glyphs.Glyph{Name: g.Name, Rune: g.Rune, Advance: g.Advance, Path: g.Path})
	}
	return build.FontWith(rp, TrialStyle(), gs, now, build.Options{NoKern: true})
}
