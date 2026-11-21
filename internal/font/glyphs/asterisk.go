// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The asterisk: six arms at sixty degrees, one pointing up, each a wedge which
// is AsteriskInner wide at the centre and AsteriskOuter wide at its flat end.
// The diameter is AsteriskSize and the top AsteriskTop.  Each arm ends in a
// round cap, a circle of AsteriskOuter across.  The arms are one outer
// contour, whose crotches are where the edges of neighbouring arms meet; the
// caps are joined to it by the union (symunion.go).

// asteriskPath returns the asterisk with its centre on the abscissa 0.
func asteriskPath(p design.Params) outline.Path {
	q := p.Punct
	r := q.AsteriskSize / 2
	cy := q.AsteriskTop - r
	// An arm runs to the centre of its cap, at the distance reach; its edges
	// are the lines from the width AsteriskInner at the centre of the star to
	// AsteriskOuter at the distance r.
	capR := q.AsteriskOuter / 2
	reach := r - capR
	endW := q.AsteriskInner + (q.AsteriskOuter-q.AsteriskInner)*reach/r
	const arms = 6
	type arm struct{ u, n outline.Point }
	as := make([]arm, arms)
	for k := range as {
		th := math.Pi/2 + float64(k)*2*math.Pi/arms
		as[k] = arm{outline.Point{X: math.Cos(th), Y: math.Sin(th)}, outline.Point{X: -math.Sin(th), Y: math.Cos(th)}}
	}
	// The edges of an arm run from the centre to the end: the right edge on the
	// side of the smaller angle, the left edge on the other.
	edge := func(a arm, side float64) (from, dir outline.Point) {
		from = outline.Point{X: side * q.AsteriskInner / 2 * a.n.X, Y: cy + side*q.AsteriskInner/2*a.n.Y}
		to := outline.Point{X: reach*a.u.X + side*endW/2*a.n.X, Y: cy + reach*a.u.Y + side*endW/2*a.n.Y}
		return from, outline.Point{X: to.X - from.X, Y: to.Y - from.Y}
	}
	var path outline.Path
	var pts []outline.Point
	for k, a := range as {
		next := as[(k+1)%arms]
		_, _ = a, next
		rf, rd := edge(a, -1)
		lf, ld := edge(a, 1)
		pts = append(pts, outline.Point{X: rf.X + rd.X, Y: rf.Y + rd.Y}, outline.Point{X: lf.X + ld.X, Y: lf.Y + ld.Y})
		nf, nd := edge(next, -1)
		pts = append(pts, meetLines(lf, ld, nf, nd))
	}
	addPolygon(&path, pts...)
	parts := pathChains(path)
	for _, a := range as {
		parts = append(parts, circleChain(outline.Point{X: reach * a.u.X, Y: cy + reach*a.u.Y}, capR))
	}
	return chainsPath(union(parts...)...)
}

// asterisk draws the asterisk.
func asterisk(p design.Params) (outline.Path, float64) {
	return placeInk(asteriskPath(p), p.Punct.Asterisk)
}
