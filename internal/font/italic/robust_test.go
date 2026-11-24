// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"fmt"
	"testing"
)

// robustSets are the parameter sets of the sweep of TestRobust: the stem at
// 72, 74, 76 and 78 (WithStem) and the x-height 10 lower, unchanged and 10
// higher (WithXHeight), the slant unchanged.
func robustSets() map[string]Params {
	out := map[string]Params{}
	for _, stem := range []float64{72, 74, 76, 78} {
		for _, dx := range []float64{-10, 0, 10} {
			p := Default().WithStem(stem)
			out[fmt.Sprintf("stem%v-x%+v", stem, dx)] = p.WithXHeight(p.XHeight + dx)
		}
	}
	return out
}

// notRobust are the letters which do not hold over the sweep, each with its
// cause.  The
// sweep reports them by name without failing.
var notRobust = map[string]string{
	"y": "a wiggle at the bottom of the turn at the stem of 72",
	"k": "the inner turn of the exit of the leg, a spike at the stem of 72: the leg has no stem whose edge places its lowest node",
	"z": "a wiggle of the top stroke at the stem of 72 and a kink at the lower left corner at the stem of 76",
	"o": "wiggles and short pieces next to the top and the bottom of the bowl, where the extremes of the edges fall a few units off the nodes, at the stem of 78",
	"c": "the counter under the stalk of the ball and the lower end, from the stem of 76",
	"s": "wiggles of the upper end and a spike of the spine, at the stems of 72 and 78",
	"e": "kinks where the bar turns into the right side of the eye, and a jump at the inner lower left",
	"a": "inflections of the bowl at the x-height 10 higher and wiggles at its top join from the stem of 76",
	"d": "the bowl as on a",
	"q": "the bowl, wiggles at its top join and its bottom",
	"g": "the bowl as on q; the tail holds",
	"b": "the turn into the bowl, corners of 23 to 24 degrees at the branch, and the free side",
	"p": "wiggles of the free side of the bowl",
	"v": "wiggles at the turn at the foot of the stem and at the ball of the free side",
	"w": "short pieces at the first turn where it runs into the second stem",
	"x": "warnings of the arcs near the meeting and wiggles at the thin ends",

	"germandbls": "wiggles of the lower bowl near the bottom",
	"oe":         "the joins of the o into the e at the bottom, kinks and wiggles",
	"ae":         "the bowl of a near its lower join, and the counter of the e as on e",
	"fl":         "the turn of the bridge into the stem of l, a jump at the stem of 76",
	"ff":         "a short piece at the crest of the second f",
	"ffl":        "the turn of the bridge into the stem of l, as on fl",
	"f_b":        "the turn of the bridge into the stem, and the bowl as on b",
	"f_h":        "the turn of the bridge into the stem, as on fl",
	"f_k":        "the turn of the bridge into the stem, and the leg as on k",

	"kgreenlandic": "the arm and the leg, as on k",
	"thorn":        "the free side of the bowl, as on p",

	"eth":     "the end of the arm, the closing end of the bowl and the crossing of the bar lie in an island of the default parameters (found by scanning the end and the bar), and the bottom of the bowl at the lower x-heights",
	"uni0259": "the turned e, with the findings of e at the bottom of the eye and at the turn into the bar",
}

// TestRobust asks every letter for a clean build, no warning of the stroke
// tool and no finding of the italic profile, on every set of the sweep, so
// that the values of the description follow from the parameters by rule and
// do not lie in islands which a change of the stem or the x-height leaves.
// It belongs to the full tier.
func TestRobust(t *testing.T) {
	if testing.Short() {
		t.Skip("the sweep of the parameters belongs to the full tier")
	}
	for set, p := range robustSets() {
		t.Run(set, func(t *testing.T) {
			t.Parallel()
			for _, name := range Names() {
				g, err := Build(p, name)
				var bad []string
				switch {
				case err != nil:
					bad = append(bad, err.Error())
				default:
					bad = append(bad, g.Warnings...)
					for _, f := range g.Findings {
						bad = append(bad, fmt.Sprintf("%s at %.0f,%.0f: %s", f.Kind, f.X, f.Y, f.Detail))
					}
				}
				if len(bad) == 0 {
					continue
				}
				if why, ok := notRobust[name]; ok {
					t.Logf("%s: %d findings and warnings, known (%s)", name, len(bad), why)
					continue
				}
				for _, b := range bad {
					t.Errorf("%s: %s", name, b)
				}
			}
		})
	}
}
