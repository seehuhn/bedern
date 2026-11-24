// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package upper

import (
	"time"

	"seehuhn.de/go/sfnt"

	"seehuhn.de/go/bedern/internal/build"
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/glyphs"
)

// TrialFile is the name of the trial font file of the package.
const TrialFile = "Bedern-Italic-Upper-Trial.otf"

// TrialStyle is the style of the trial font, named as a trial.
func TrialStyle() design.Style {
	return design.Style{
		Family:     "Bedern Italic Upper Trial",
		Subfamily:  "Italic",
		PostScript: "BedernItalicUpperTrial-Italic",
		Weight:     400,
		Italic:     true,
	}
}

// FontGlyphs returns the glyphs of the package as glyphs of the font.
func FontGlyphs(p Params) ([]glyphs.Glyph, error) {
	gs, err := Glyphs(p)
	if err != nil {
		return nil, err
	}
	out := make([]glyphs.Glyph, len(gs))
	for i, g := range gs {
		out[i] = glyphs.Glyph{Name: g.Name, Rune: g.Rune, Advance: g.Advance, Path: g.Path}
	}
	return out, nil
}

// TrialFont returns the trial font: the .notdef and the space of the roman
// and the glyphs of the package, without kerning.
func TrialFont(p Params, now time.Time) (*sfnt.Font, error) {
	rp := design.Regular()
	var gs []glyphs.Glyph
	for _, name := range []string{".notdef", "space"} {
		path, adv, _ := glyphs.Outline(name, rp)
		gs = append(gs, glyphs.Glyph{Name: name, Rune: map[string]rune{"space": ' '}[name], Advance: adv, Path: path.Round()})
	}
	its, err := FontGlyphs(p)
	if err != nil {
		return nil, err
	}
	gs = append(gs, its...)
	return build.FontWith(rp, TrialStyle(), gs, now, build.Options{NoKern: true})
}
