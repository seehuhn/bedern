// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"fmt"
	"io"
	"text/tabwriter"

	"seehuhn.de/go/bedern/internal/font/design"
)

var sideKindNames = [...]string{
	sideUnknown: "unknown", sideStem: "stem", sideRound: "round", sideStemD: "stemD",
	sideStemBP: "stemBP", sideBowlDQ: "bowlDQ", sideBowlBP: "bowlBP", sideBowlC: "bowlC",
	sideStemR: "stemR", sideStemQ: "stemQ", sideNarrow: "narrow", sideBowlA: "bowlA",
	sideTailA: "tailA", sideOpen: "open", sideOverhang: "overhang", sideBarF: "barF",
	sideBarT: "barT", sideLoopG: "loopG", sideEarG: "earG", sideDiag: "diag",
	sideDiagTail: "diagTail", sideZ: "z", sideK: "k",

	sideCapStem: "capStem", sideCapRound: "capRound", sideCapArmE: "capArmE",
	sideCapArmF: "capArmF", sideCapArmL: "capArmL", sideCapT: "capT",

	sideCapC: "capC", sideCapG: "capG", sideCapS: "capS", sideCapJ: "capJ",
	sideCapBowlStem: "capBowlStem", sideCapBowl: "capBowl",
	sideCapBowlB: "capBowlB", sideCapBowlD: "capBowlD", sideCapR: "capR",
	sideCapA: "capA", sideCapV: "capV", sideCapW: "capW", sideCapX: "capX", sideCapY: "capY",
	sideCapZ: "capZ", sideCapK: "capK", sideCapM: "capM", sideCapN: "capN",

	sideFigure: "figure", sideMark: "mark",
}

func (k sideKind) String() string {
	if name, ok := punctKindNames[k]; ok {
		return name
	}
	return sideKindNames[k]
}

// WriteTable writes the glyphs of the font made from p as plain text: for
// each glyph the advance, the left and right side bearing of the rounded
// outline (the distance of the ink from the edge of the advance, negative if
// the ink reaches beyond it) and the kinds of its two sides.
func WriteTable(w io.Writer, p design.Params) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(tw, "glyph\tadvance\tlsb\trsb\tleft\tright\t")
	for i, g := range Set(p) {
		lsb, rsb := "-", "-"
		if len(g.Path) > 0 {
			x0, _, x1, _ := g.Path.BBox()
			lsb, rsb = fmt.Sprint(x0), fmt.Sprint(g.Advance-x1)
		}
		fmt.Fprintf(tw, "%s\t%g\t%s\t%s\t%s\t%s\t\n", g.Name, g.Advance, lsb, rsb, table[i].left, table[i].right)
	}
	return tw.Flush()
}
