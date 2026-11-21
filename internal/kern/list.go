// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package kern

import (
	"fmt"
	"io"
	"sort"
	"text/tabwriter"
)

// Sorted returns the pairs with a kern, the largest first; pairs of equal size
// are in the order of the glyph list.
func (t *Table) Sorted() []*Pair {
	ps := append([]*Pair(nil), t.Pairs...)
	sort.SliceStable(ps, func(i, j int) bool {
		a, b := ps[i].Kern, ps[j].Kern
		if a < 0 {
			a = -a
		}
		if b < 0 {
			b = -b
		}
		return a > b
	})
	return ps
}

// WriteList writes the kerns as plain text, the largest first: the two glyphs,
// the kern, the mean white of the pair before and after it and the target
// (in units, over the zone of the pair), the closest distance of the outlines
// before and after, the pair which sets the target, and the classes of the
// two sides (the left one is that of the right side of the first glyph).  An
// F in the last column marks a kern set by the floor, an X one set by an
// exception.
func (t *Table) WriteList(w io.Writer) error {
	n1, n2 := 0, 0
	for _, row := range t.Values {
		n2 = len(row) - 1
	}
	n1 = len(t.Values) - 1
	fmt.Fprintf(w, "%d pairs with a kern, %d classes of right sides with kerns (%d in all), %d classes of left sides with kerns (%d in all), %d single pairs\n\n",
		len(t.Pairs), n1, t.NClass1, n2, t.NClass2, len(t.Single))
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(tw, "pair\tkern\twhite\tafter\ttarget\tdist\tafter\tzone\tref\tclass1\tclass2\t\t")
	for _, pr := range t.Sorted() {
		flag := ""
		if pr.Floor {
			flag = "F"
		}
		if pr.Reason != "" {
			flag = "X"
		}
		fmt.Fprintf(tw, "%s %s\t%d\t%.0f\t%.0f\t%.0f\t%.0f\t%.0f\t%d-%d\t%s\t%d\t%d\t%s\t\n",
			pr.Left.Name, pr.Right.Name, pr.Kern, pr.White0, pr.White1, pr.Target,
			pr.Dist0, pr.Dist1, int(float64(pr.Zone.Lo)*t.Font.P.Kern.Band+t.Font.lo), int(float64(pr.Zone.Hi)*t.Font.P.Kern.Band+t.Font.lo), pr.Ref, t.Class1[pr.Left.Index], t.Class2[pr.Right.Index], flag)
	}
	return tw.Flush()
}
