// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"slices"
	"strings"
	"sync"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/design/designtest"
	"seehuhn.de/go/bedern/internal/outline"
)

// The structural checks do not depend on the shape of a letter.  Every glyph
// of the table is checked, unrounded and rounded, for every variant, every
// parameter set and, in the full tier, the bold parameters.  In the other
// perturbed sets than those of composedRoutine the tests leave out the derived
// glyphs (derived) and the glyphs whose checks have the same inputs as in the
// regular set (sameChecks), which give the same result.

type namedParams struct {
	name   string
	params design.Params
	glyphs []string // the glyphs to check and to build, all if empty
}

// routineSets is the number of perturbed parameter sets on which the routine
// tier (go test -short) runs the structure test and the tests of the letters,
// see designtest.Sets: the first nine, which move the lowercase proportions,
// the capitals, the weight and the counters, and the parameters of each group
// of letters and of the punctuation once.  The full tier runs them on all sets,
// and TestGolden does so in both tiers.
const routineSets = 9

// allParams returns the parameter sets on which the property tests run: the
// variants and, in the routine tier, the first routineSets perturbed sets; the
// full tier adds the bold parameters.
func allParams() []namedParams {
	sets := paramSets(designtest.Sets(testing.Short(), routineSets))
	if !testing.Short() {
		sets = append(sets, boldParams())
	}
	return sets
}

// boldExempt lists the tests which do not run on the bold parameters.
var boldExempt = map[string]bool{
	"TestCapBCountersAreOpen":            true,
	"TestCapCGWidthsAndOpening":          true,
	"TestCapRRightSideIsOpen":            true,
	"TestFigureEdges":                    true,
	"TestFigureFourFlat":                 true,
	"TestFigureOuterEdgesWindIn":         true,
	"TestFigureTwoBowlMiddle":            true,
	"TestGEar":                           true,
	"TestGLinkSwells":                    true,
	"TestHookInnerEdgeHasNoDent":         true,
	"TestQuestionBalance":                true,
	"TestQuestionHook":                   true,
	"TestRoundKeepsSmoothNodesSmooth":    true,
	"TestSBowlsAreRound":                 true,
	"TestSectionEye":                     true,
	"TestSectionJoints":                  true,
	"TestStalkIsEven":                    true,
	"TestSterlingBar":                    true,
	"TestSterlingBaseIsHeavyInTheMiddle": true,
	"TestSuperiorFigures":                true,
	"TestSymbolNodes":                    true,
}

// boldApplies says whether the test t runs on the parameter set np.
func boldApplies(t *testing.T, np namedParams) bool {
	name, _, _ := strings.Cut(t.Name(), "/")
	return np.name != "bold" || !boldExempt[name]
}

// structureRoutineSets is the number of perturbed parameter sets on which the
// routine tier runs the structure test, besides the variants; the full tier
// runs it on all sets and on the bold parameters.
const structureRoutineSets = 1

// structureParams returns the parameter sets on which the structure test runs.
func structureParams() []namedParams {
	if !testing.Short() {
		return allParams()
	}
	return paramSets(designtest.Sets(true, structureRoutineSets))
}

// derivedParams returns the parameter sets on which the tests of the derived
// glyphs run: the variants and, in the routine tier, the first
// composedRoutineSets perturbed sets.
func derivedParams() []namedParams {
	if !testing.Short() {
		return allParams()
	}
	return paramSets(designtest.Sets(true, composedRoutineSets))
}

// goldenParams returns every parameter set, whichever the tier, and in the
// full tier the bold parameters.
func goldenParams() []namedParams {
	sets := paramSets(designtest.Perturbed())
	if !testing.Short() {
		sets = append(sets, boldParams())
	}
	return sets
}

// boldParams returns the parameters of Bedern Bold, on which the full tier
// runs the fingerprints, the structural checks and the tests of the letters.
func boldParams() namedParams {
	return namedParams{name: "bold", params: design.Bold()}
}

func paramSets(perturbed []designtest.Set) []namedParams {
	var all []namedParams
	for _, v := range design.Variants() {
		all = append(all, namedParams{name: v.Name, params: v.Params})
	}
	for _, s := range perturbed {
		all = append(all, namedParams{name: s.Name, params: s.Params})
	}
	return all
}

func TestStructure(t *testing.T) {
	t.Parallel()
	regular := namedParams{name: "regular", params: design.Regular()}
	for _, g := range derivedGlyphs {
		if _, ok := lookup(g); !ok {
			t.Errorf("derivedGlyphs names %s, which is not in the table", g)
		}
	}
	for _, np := range structureParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			all := composedRoutine(np.name) || np.name == "bold"
			var names []string
			for _, e := range table {
				if len(np.glyphs) > 0 && !slices.Contains(np.glyphs, e.name) {
					continue
				}
				if !all && (derived(e.name) || sameChecks(regular, np, e.name)) {
					continue
				}
				names = append(names, e.name)
			}
			forEachParallel(len(names), func(i int) { testStructureGlyph(t, np, names[i]) })
		})
	}
}

// derivedGlyphs lists the glyphs which are not composed but are drawn from
// other glyphs scaled or with other weights: the superior figures, the
// fractions, the ordinals, © and ®.  Their structural checks and the tests of
// these glyphs and of the euro, which is the C scaled, run on derivedParams.
var derivedGlyphs = []string{
	"uni00B9", "uni00B2", "uni00B3", "onequarter", "onehalf", "threequarters",
	"ordfeminine", "ordmasculine", "copyright", "registered",
}

// derived says whether the glyph is composed of a base and a mark or is one of
// derivedGlyphs.
func derived(glyph string) bool {
	return composites[glyph] != nil || slices.Contains(derivedGlyphs, glyph)
}

// composedRoutineSets is the number of perturbed parameter sets on which the
// routine tier runs the tests of the composed letters and of the derived
// glyphs, besides the variants; the full tier runs them on all sets.
const composedRoutineSets = 3

// composedRoutine says whether the structure test of every glyph, where it runs,
// and the tests of the composed letters run on the parameter set of the given
// name in this tier.
func composedRoutine(set string) bool {
	if !testing.Short() || set == "regular" {
		return true
	}
	return composedRoutineNames()[set]
}

// composedRoutineNames holds the names of the perturbed sets of composedRoutine.
var composedRoutineNames = sync.OnceValue(func() map[string]bool {
	names := map[string]bool{}
	for _, s := range designtest.Sets(true, composedRoutineSets) {
		names[s.Name] = true
	}
	return names
})

// sameChecks says whether the structural checks of the glyph have the same
// inputs in the parameter sets a and b: the same unrounded and rounded
// outlines and advance, and the same values of the parameters which the checks
// read (the clearance floor, the ink margins of the sides, the overhang of the
// marks, the number of contours of the mark and the sharp crotches).
func sameChecks(a, b namedParams, glyph string) bool {
	pa, adva := cachedOutline(a, glyph)
	pb, advb := cachedOutline(b, glyph)
	if adva != advb || !pathsEqual(pa, pb) || !pathsEqual(cachedRounded(a, glyph).Path, cachedRounded(b, glyph).Path) {
		return false
	}
	p, q := a.params, b.params
	e, _ := lookup(glyph)
	la, ra := e.inkMargins(p)
	lb, rb := e.inkMargins(q)
	if clearanceFloor(p) != clearanceFloor(q) || la != lb || ra != rb || p.Acc.Overhang != q.Acc.Overhang {
		return false
	}
	if c := composites[glyph]; c != nil && len(markOutline(p, c.mark, formLower)) != len(markOutline(q, c.mark, formLower)) {
		return false
	}
	if f, ok := sharpCrotches[glyph]; ok && !slices.Equal(f(p), f(q)) {
		return false
	}
	return true
}

// testStructureGlyph runs the structural checks on one glyph in one parameter
// set, unrounded and rounded.
func testStructureGlyph(t *testing.T, np namedParams, glyph string) {
	path, adv := cachedOutline(np, glyph)
	rounded := cachedRounded(np, glyph)
	for _, msg := range checkStages(np.params, glyph, path, adv, rounded.Path) {
		t.Errorf("glyph %s, parameter set %s, %s", glyph, np.name, msg)
	}
}

// The checks must fail on shapes which break them.
func TestStructureChecksCatchBreakage(t *testing.T) {
	t.Parallel()
	rect := func(x0, y0, x1, y1 float64, ccw bool) outline.Contour {
		var p outline.Path
		rect(&p, x0, y0, x1, y1, ccw)
		return p[0]
	}
	var bulge outline.Path // a curve which runs above both its end points
	bulge.MoveTo(0, 0)
	bulge.CurveTo(0, 50, 100, 50, 100, 0)
	bulge.LineTo(50, -30)
	var twice outline.Path
	twice.MoveTo(0, 0)
	twice.LineTo(100, 0)
	twice.LineTo(100, 0)
	twice.LineTo(100, 100)
	twice.LineTo(0, 100)

	// Each case is a shape which breaks one check.  The glyph is "none", which is
	// not in the table, unless the case names another.
	cases := []struct {
		name    string
		check   string
		path    outline.Path
		glyph   string
		advance float64
		want    string // a part of the message, if not empty
	}{
		{name: "zero-length segment", check: "contours", path: twice},
		{name: "inner contour counter-clockwise", check: "orientation",
			path: outline.Path{rect(0, 0, 100, 100, true), rect(20, 20, 80, 80, true)}},
		{name: "overlapping contours", check: "intersections",
			path: outline.Path{rect(0, 0, 100, 100, true), rect(50, 50, 150, 150, true)}},
		{name: "winding number two", check: "scans",
			path: outline.Path{rect(0, 0, 100, 100, true), rect(20, 20, 80, 80, true)}},
		{name: "curve beyond its end points", check: "extremes", path: bulge},
		{name: "counter too close to the outer contour", check: "clearance",
			path: outline.Path{rect(0, 0, 200, 200, true), rect(8, 8, 192, 192, false)}},
		{name: "stroke too thin", check: "clearance",
			path: outline.Path{rect(0, 0, 200, 12, true)}},
		{name: "glyph without a rule", check: "bbox",
			path: outline.Path{rect(50, 0, 150, 100, true)}, want: "no bounding-box rule for glyph none"},
		{name: "n which touches the edge of its advance", check: "bbox", glyph: "n", advance: 540,
			path: outline.Path{rect(0, 0, 540, 100, true)}},
		{name: "o which is too wide for its advance", check: "bbox", glyph: "o", advance: 486,
			path: outline.Path{rect(10, 0, 480, 100, true)}},
		{name: "f whose hook overhangs the advance by more than MaxOverhang", check: "bbox", glyph: "f", advance: 300,
			path: outline.Path{rect(24, 0, 361, 100, true)}},
		{name: "f whose bar starts too near the edge", check: "bbox", glyph: "f", advance: 300,
			path: outline.Path{rect(20, 0, 300, 100, true)}},
		{name: "j whose tail overhangs the advance by more than MaxOverhang", check: "bbox", glyph: "j", advance: 281,
			path: outline.Path{rect(-61, 0, 191, 100, true)}},
		{name: "t whose bar starts too near the edge", check: "bbox", glyph: "t", advance: 330,
			path: outline.Path{rect(8, 0, 320, 100, true)}},
	}
	for _, c := range cases {
		glyph, advance := c.glyph, c.advance
		if glyph == "" {
			glyph, advance = "none", 200
		}
		for _, ch := range structureChecks {
			if ch.name != c.check {
				continue
			}
			problems := ch.run(structureCase{params: design.Regular(), glyph: glyph, path: c.path, advance: advance})
			if len(problems) == 0 {
				t.Errorf("%s: check %s does not fail", c.name, c.check)
			} else if c.want != "" && !strings.Contains(problems[0], c.want) {
				t.Errorf("%s: check %s says %q, want %q", c.name, c.check, problems[0], c.want)
			}
		}
	}
	if len(checkIntegers(structureCase{path: outline.Path{rect(0.5, 0, 10, 10, true)}})) == 0 {
		t.Error("a fractional coordinate is not found")
	}
	// The overhang kinds allow what they say: the hook of f may reach 60 beyond
	// the advance and the tail of j 60 to the left of it, the bar of f and t come
	// to within the ink margin of the left edge.
	for _, c := range []struct {
		glyph   string
		advance float64
		path    outline.Path
	}{
		{"f", 300, outline.Path{rect(24, 0, 360, 100, true)}},
		{"j", 281, outline.Path{rect(-60, 0, 191, 100, true)}},
		{"t", 350, outline.Path{rect(21, 0, 340, 100, true)}},
		{"v", 508, outline.Path{rect(17, -12, 497, 448, true)}},
		{"w", 712, outline.Path{rect(17, -12, 701, 448, true)}},
		{"y", 508, outline.Path{rect(17, -200, 497, 448, true)}},
	} {
		if problems := checkBBox(structureCase{params: design.Regular(), glyph: c.glyph, path: c.path, advance: c.advance}); len(problems) > 0 {
			t.Errorf("glyph %s: the bounding-box check rejects ink which the side kinds allow: %s", c.glyph, problems[0])
		}
	}
}

// inTier says whether the tests which run over all glyphs check the glyph in
// the parameter set in this tier: a composed letter in the sets of
// composedRoutine, every other glyph in all sets.
func inTier(np namedParams, glyph string) bool {
	return composites[glyph] == nil || composedRoutine(np.name)
}
