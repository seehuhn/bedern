// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package design

import (
	"math"
	"testing"
)

// TestValidate checks that Validate accepts the regular parameters and rejects
// parameter sets with invalid values.
func TestValidate(t *testing.T) {
	t.Parallel()
	if err := Regular().Validate(); err != nil {
		t.Fatalf("regular parameters: %v", err)
	}
	for name, f := range map[string]func(p *Params){
		"stem zero":                func(p *Params) { p.Stem = 0 },
		"counter negative":         func(p *Params) { p.NCounter = -1 },
		"ascender stem above stem": func(p *Params) { p.AscenderStem = p.Stem + 1 },
		"NaN":                      func(p *Params) { p.DotShift = math.NaN() },
		"infinity":                 func(p *Params) { p.Hair = math.Inf(1) },
		"thin not below curve":     func(p *Params) { p.ThinLower = p.CurveLower },
		"hairline not below stem":  func(p *Params) { p.Hair = p.Stem },
		"x-height above cap":       func(p *Params) { p.XHeight = p.CapHeight },
		"x-height above ascender":  func(p *Params) { p.Ascender = p.XHeight },
		"join above notch":         func(p *Params) { p.ArchJoin = p.ArchNotch },
		"notch above x-height":     func(p *Params) { p.ArchNotch = p.XHeight },
		"outer superness low":      func(p *Params) { p.SupernessOuter = 0.69 },
		"outer superness high":     func(p *Params) { p.SupernessOuter = 0.81 },
		"inner superness low":      func(p *Params) { p.SupernessInner = 0.69 },
		"inner superness high":     func(p *Params) { p.SupernessInner = 0.81 },
	} {
		p := Regular()
		f(&p)
		if p.Validate() == nil {
			t.Errorf("%s: Validate accepts the set", name)
		}
	}
}
