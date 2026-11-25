// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

import (
	"reflect"
	"testing"

	"seehuhn.de/go/postscript/funit"
	"seehuhn.de/go/sfnt/opentype/mathtable"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/design/designtest"
)

// refRange is the range of a constant in a unit: the em ("em"), the x-height
// ("x") or the fraction rule ("rule"), or plain per cent ("%").
type refRange struct {
	unit     string
	min, max float64
}

// refRanges holds the ranges of the study.  The minima of the four stack
// shifts are the 50 units of the one font whose fault the study names.
var refRanges = map[string][]refRange{
	"ScriptPercentScaleDown":                   {{"%", 70, 80}},
	"ScriptScriptPercentScaleDown":             {{"%", 50, 60}},
	"DelimitedSubFormulaMinHeight":             {{"em", 1.120, 1.500}},
	"DisplayOperatorMinHeight":                 {{"em", 1.120, 1.800}},
	"MathLeading":                              {{"em", 0.140, 0.160}},
	"AxisHeight":                               {{"em", 0.250, 0.271}, {"x", 0.545, 0.606}},
	"AccentBaseHeight":                         {{"em", 0.440, 0.480}, {"x", 1.000, 1.044}},
	"FlattenedAccentBaseHeight":                {{"em", 0.420, 0.692}},
	"SubscriptShiftDown":                       {{"em", 0.210, 0.250}, {"x", 0.444, 0.573}},
	"SubscriptTopMax":                          {{"em", 0.313, 0.368}, {"x", 0.670, 0.839}},
	"SubscriptBaselineDropMin":                 {{"em", 0.045, 0.222}},
	"SuperscriptShiftUp":                       {{"em", 0.343, 0.430}, {"x", 0.748, 0.956}},
	"SuperscriptShiftUpCramped":                {{"em", 0.252, 0.360}, {"x", 0.533, 0.814}},
	"SuperscriptBottomMin":                     {{"em", 0.108, 0.192}, {"x", 0.249, 0.410}},
	"SuperscriptBaselineDropMax":               {{"em", 0.230, 0.394}},
	"SubSuperscriptGapMin":                     {{"em", 0.140, 0.250}, {"rule", 2.21, 5.00}},
	"SuperscriptBottomMaxWithSubscript":        {{"em", 0.328, 0.380}, {"x", 0.701, 0.886}},
	"SpaceAfterScript":                         {{"em", 0.040, 0.060}},
	"UpperLimitGapMin":                         {{"em", 0.100, 0.200}, {"rule", 1.68, 5.00}},
	"UpperLimitBaselineRiseMin":                {{"em", 0.100, 0.300}},
	"LowerLimitGapMin":                         {{"em", 0.100, 0.167}},
	"LowerLimitBaselineDropMin":                {{"em", 0.433, 0.670}},
	"StackTopShiftUp":                          {{"em", 0.050, 0.470}},
	"StackTopDisplayStyleShiftUp":              {{"em", 0.050, 0.780}},
	"StackBottomShiftDown":                     {{"em", 0.050, 0.385}},
	"StackBottomDisplayStyleShiftDown":         {{"em", 0.050, 0.701}},
	"StackGapMin":                              {{"em", 0.100, 0.348}, {"rule", 1.68, 5.90}},
	"StackDisplayStyleGapMin":                  {{"em", 0.140, 0.812}, {"rule", 2.31, 13.8}},
	"FractionNumeratorShiftUp":                 {{"em", 0.394, 0.600}, {"x", 0.872, 1.299}},
	"FractionNumeratorDisplayStyleShiftUp":     {{"em", 0.580, 0.725}, {"x", 1.352, 1.571}},
	"FractionDenominatorShiftDown":             {{"em", 0.250, 0.585}, {"x", 0.547, 1.237}},
	"FractionDenominatorDisplayStyleShiftDown": {{"em", 0.598, 0.710}, {"x", 1.329, 1.632}},
	"FractionNumeratorGapMin":                  {{"em", 0.040, 0.116}, {"rule", 1.00, 2.00}},
	"FractionNumDisplayStyleGapMin":            {{"em", 0.120, 0.198}, {"rule", 2.20, 3.05}},
	"FractionRuleThickness":                    {{"em", 0.040, 0.068}, {"x", 0.093, 0.152}},
	"FractionDenominatorGapMin":                {{"em", 0.040, 0.116}, {"rule", 1.00, 2.00}},
	"FractionDenomDisplayStyleGapMin":          {{"em", 0.120, 0.198}, {"rule", 2.20, 3.05}},
	"OverbarVerticalGap":                       {{"rule", 1.68, 3.00}},
	"OverbarRuleThickness":                     {{"rule", 0.62, 1.27}},
	"UnderbarVerticalGap":                      {{"rule", 1.68, 3.00}},
	"UnderbarRuleThickness":                    {{"rule", 0.62, 1.27}},
	"RadicalVerticalGap":                       {{"em", 0.050, 0.112}, {"x", 0.116, 0.236}, {"rule", 1.25, 1.73}},
	"RadicalDisplayStyleVerticalGap":           {{"em", 0.112, 0.174}, {"x", 0.236, 0.396}, {"rule", 1.73, 3.70}},
	"RadicalRuleThickness":                     {{"em", 0.040, 0.068}, {"rule", 0.62, 1.00}},
	"RadicalExtraAscender":                     {{"em", 0.040, 0.112}, {"rule", 1.00, 1.73}},
	"RadicalKernBeforeDegree":                  {{"em", 0.040, 0.300}},
	"RadicalKernAfterDegree":                   {{"em", -0.600, -0.335}},
	"RadicalDegreeBottomRaisePercent":          {{"%", 50, 65}},
}

// value returns the constant of the given name as a number.
func value(c mathtable.Constants, name string) (float64, bool) {
	f := reflect.ValueOf(c).FieldByName(name)
	if !f.IsValid() {
		return 0, false
	}
	switch v := f.Interface().(type) {
	case mathtable.Value:
		return float64(v.V), true
	case funit.Uint16:
		return float64(v), true
	case funit.Int16:
		return float64(v), true
	case int16:
		return float64(v), true
	case uint16:
		return float64(v), true
	}
	return 0, false
}

// TestConstantsInRange requires every listed constant to lie inside
// its range, in each unit given, on the regular parameters and on the
// perturbed sets which move the x-height and the weight.
func TestConstantsInRange(t *testing.T) {
	sets := append([]designtest.Set{{Name: "regular", Params: design.Regular()}}, designtest.Sets(true, 3)...)
	for _, s := range sets {
		p := s.Params
		c := Constants(p)
		rule := p.Punct.MathThick
		for name, rs := range refRanges {
			v, ok := value(c, name)
			if !ok {
				t.Fatalf("no constant %s", name)
			}
			for _, r := range rs {
				var q float64
				switch r.unit {
				case "em":
					q = v / p.UnitsPerEm
				case "x":
					q = v / p.XHeight
				case "rule":
					q = v / rule
				case "%":
					q = v
				}
				// The ranges are given to the third decimal; half a unit of
				// the font is allowed for the rounding of the value.
				tol := 0.5 / p.UnitsPerEm
				if r.unit == "x" {
					tol = 0.5 / p.XHeight
				} else if r.unit == "rule" {
					tol = 0.5 / rule
				}
				if q < r.min-tol || q > r.max+tol {
					t.Errorf("%s: %s is %.0f, %.3f %s, outside the range %.3f to %.3f", s.Name, name, v, q, r.unit, r.min, r.max)
				}
			}
		}
	}
}

// TestConstantsKeepTheRelationsOfTheCraft checks the relations which the
// study found in every reference: the numerator shift 2.5 to 5.6 rules above
// the axis, the superscript above the subscript top, the gaps of the display
// at least those of the text.
func TestConstantsKeepTheRelationsOfTheCraft(t *testing.T) {
	sets := append([]designtest.Set{{Name: "regular", Params: design.Regular()}}, designtest.Sets(true, 3)...)
	for _, s := range sets {
		p := s.Params
		c := Constants(p)
		rule := float64(c.FractionRuleThickness.V)
		if k := (float64(c.FractionNumeratorShiftUp.V) - float64(c.AxisHeight.V)) / rule; k < 2.5 || k > 5.6 {
			t.Errorf("%s: the numerator stands %.2f rules above the axis, want 2.5 to 5.6", s.Name, k)
		}
		if c.SuperscriptBottomMin.V >= c.SuperscriptShiftUp.V {
			t.Errorf("%s: SuperscriptBottomMin %d is not below SuperscriptShiftUp %d", s.Name, c.SuperscriptBottomMin.V, c.SuperscriptShiftUp.V)
		}
		if c.FractionNumDisplayStyleGapMin.V < c.FractionNumeratorGapMin.V || c.StackDisplayStyleGapMin.V < c.StackGapMin.V {
			t.Errorf("%s: a gap of the display is smaller than that of the text", s.Name)
		}
		if c.FractionNumeratorDisplayStyleShiftUp.V <= c.FractionNumeratorShiftUp.V {
			t.Errorf("%s: the numerator of the display is not above that of the text", s.Name)
		}
	}
}
