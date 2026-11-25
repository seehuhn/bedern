// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

import (
	"math"

	"seehuhn.de/go/postscript/funit"
	"seehuhn.de/go/sfnt/opentype/mathtable"

	"seehuhn.de/go/bedern/internal/font/design"
)

// Constants returns the constants of the MATH table from the parameters of the
// regular: the x-height x, the em, the axis and the rule R, which is the
// stroke of the minus.  The shifts are fractions of x and the gaps multiples
// of R; the other values are fractions of the em.
func Constants(p design.Params) mathtable.Constants {
	x := p.XHeight
	em := p.UnitsPerEm
	axis := p.Punct.MathAxis
	rule := p.Punct.MathThick
	v := func(f float64) mathtable.Value { return mathtable.Value{V: funit.Int16(math.Round(f))} }
	u := func(f float64) funit.Uint16 { return funit.Uint16(math.Round(f)) }
	return mathtable.Constants{
		ScriptPercentScaleDown:       70,
		ScriptScriptPercentScaleDown: 50,
		DelimitedSubFormulaMinHeight: u(1.3 * em),
		DisplayOperatorMinHeight:     u(1.3 * em),

		MathLeading:               v(0.15 * em),
		AxisHeight:                v(axis),
		AccentBaseHeight:          v(x + p.OvershootLower),
		FlattenedAccentBaseHeight: v(p.Figures.Height),

		SubscriptShiftDown:                v(0.49 * x),
		SubscriptTopMax:                   v(0.8 * x),
		SubscriptBaselineDropMin:          v(0.15 * em),
		SuperscriptShiftUp:                v(0.84 * x),
		SuperscriptShiftUpCramped:         v(0.67 * x),
		SuperscriptBottomMin:              v(0.28 * x),
		SuperscriptBaselineDropMax:        v(0.23 * em),
		SubSuperscriptGapMin:              v(4 * rule),
		SuperscriptBottomMaxWithSubscript: v(0.8 * x),
		SpaceAfterScript:                  v(0.05 * em),

		UpperLimitGapMin:          v(2 * rule),
		UpperLimitBaselineRiseMin: v(0.2 * em),
		LowerLimitGapMin:          v(0.15 * em),
		LowerLimitBaselineDropMin: v(0.6 * em),

		StackTopShiftUp:                  v(0.99 * x),
		StackTopDisplayStyleShiftUp:      v(1.54 * x),
		StackBottomShiftDown:             v(0.77 * x),
		StackBottomDisplayStyleShiftDown: v(1.54 * x),
		StackGapMin:                      v(3 * rule),
		StackDisplayStyleGapMin:          v(7 * rule),

		StretchStackTopShiftUp:      v(2 * rule),
		StretchStackBottomShiftDown: v(0.59 * em),
		StretchStackGapAboveMin:     v(2 * rule),
		StretchStackGapBelowMin:     v(2 * rule),

		FractionNumeratorShiftUp:                 v(0.96 * x),
		FractionNumeratorDisplayStyleShiftUp:     v(1.43 * x),
		FractionDenominatorShiftDown:             v(0.80 * x),
		FractionDenominatorDisplayStyleShiftDown: v(1.54 * x),
		FractionNumeratorGapMin:                  v(rule),
		FractionNumDisplayStyleGapMin:            v(3 * rule),
		FractionRuleThickness:                    v(rule),
		FractionDenominatorGapMin:                v(rule),
		FractionDenomDisplayStyleGapMin:          v(3 * rule),
		SkewedFractionHorizontalGap:              v(0.35 * em),
		SkewedFractionVerticalGap:                v(rule),

		OverbarVerticalGap:     v(3 * rule),
		OverbarRuleThickness:   v(rule),
		OverbarExtraAscender:   v(rule),
		UnderbarVerticalGap:    v(3 * rule),
		UnderbarRuleThickness:  v(rule),
		UnderbarExtraDescender: v(rule),

		RadicalVerticalGap:             v(1.25 * rule),
		RadicalDisplayStyleVerticalGap: v(rule + x/4),
		RadicalRuleThickness:           v(rule),
		RadicalExtraAscender:           v(rule),
		// The kern after the degree is -8.5/18 em and the degree stands above
		// the peak of the radical.
		RadicalKernBeforeDegree: v(5.0 / 18 * em),
		RadicalKernAfterDegree:  v(-8.5 / 18 * em),

		RadicalDegreeBottomRaisePercent: 65,
	}
}
