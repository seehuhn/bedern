// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"seehuhn.de/go/bedern/internal/fair"
	"seehuhn.de/go/bedern/internal/stroke"
)

func init() {
	intendedFindings["j"] = []intended{{
		Kind: fair.Wiggle, X: 134, Y: 366, Dir: -96,
		Reason: "the inner edge under the entry is that of i, the same curves to the unit moved by the side; the check pairs the extremes of its curvature with those of the straight stem, which on j runs on into the tail, and finds a wiggle which it does not find on i",
	}}
}

// j draws the j: one stroke from the entry of i down the stem through the
// baseline into the tail with its ball (tail), and the dot of i (dotCentre).
func j(p Params) letter {
	e := p.JSideLeft
	s := stroke.Stroke{Nodes: p.entry(e, p.XHeight), Start: round}
	nodes, end := p.tail(e)
	s.Nodes, s.End = append(s.Nodes, nodes...), end
	return letter{
		strokes: []stroke.Stroke{s},
		dots:    []dot{{C: p.dotCentre(e), D: p.Dot}},
		advance: p.JWidth,
	}
}
