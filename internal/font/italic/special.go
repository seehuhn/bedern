// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import "seehuhn.de/go/bedern/internal/stroke"

// The findings which ĸ and ĳ take from k and j, at their places.
func init() {
	intendedFindings["kgreenlandic"] = intendedFindings["k"]
	j := intendedFindings["j"][0]
	j.X += 239
	intendedFindings["ij"] = []intended{j}
	// µ is the μ of the Greek lowercase.
	intendedFindings["micro"] = intendedFindings["mu"]
}

// The special letters which are made of the strokes of the basic ones: ſ, ĸ,
// þ, ĳ, ŀ and ŋ.

// longs draws the long s: the f without its crossbar, with the advance of f.
func longs(p Params) letter {
	l := f(p)
	return letter{strokes: l.strokes[:1], advance: l.advance}
}

// kgreenlandic draws the kra: the k with its stem cut at the x-height and
// begun with the entry of n.
func kgreenlandic(p Params) letter {
	l := k(p)
	e := p.SideStemLeft - p.HLeftLess
	l.strokes[0] = p.stem(e, p.XHeight, false)
	return l
}

// thorn draws the thorn: the p with its stem carried up to the ascender,
// where it begins with the entry of l.
func thorn(q Params) letter {
	l := p(q)
	e := q.SideStemLeft
	s := stroke.Stroke{Nodes: q.entry(e, q.XHeight+q.AscenderEntryRise), Start: round, End: hidden}
	s.Nodes = append(s.Nodes, q.footNodes(e+q.Stem/2)...)
	l.strokes[0] = s
	return l
}

// ij draws the ligature ĳ: the i and the j, the j IJCloser closer than its
// advance after i.
func ij(p Params) letter {
	li, lj := i(p), j(p)
	dx := li.advance - p.IJCloser
	out := letter{strokes: li.strokes, dots: li.dots, advance: dx + lj.advance}
	for _, s := range lj.strokes {
		out.strokes = append(out.strokes, moved(s, dx))
	}
	for _, d := range lj.dots {
		out.dots = append(out.dots, dot{C: pt(d.C.X+dx, d.C.Y), D: d.D})
	}
	return out
}

// ldot draws the l with the middle dot: the l and the dot of i at LDotY of
// the x-height, its ink LDotRight right of the right edge of the stem on
// the slant, and the advance LDotMore wider than l.
func ldot(p Params) letter {
	l := l(p)
	x := p.SideStemLeft + p.Stem + p.LDotRight + p.Dot/2
	l.dots = []dot{{C: pt(x, p.LDotY*p.XHeight), D: p.Dot}}
	l.advance += p.LDotMore
	return l
}

// eng draws the eng: the n whose right stem runs on below the baseline into
// the tail of j.
func eng(p Params) letter {
	l := n(p)
	e1 := p.SideStemLeft
	e2 := e1 + p.Stem + p.NCounter
	a := stroke.Stroke{Nib: p.nib(), Nodes: p.archNodes(e1, e2), Start: hidden}
	nodes, end := p.tail(e2)
	a.Nodes, a.End = append(a.Nodes, nodes...), end
	l.strokes[1] = a
	return l
}
