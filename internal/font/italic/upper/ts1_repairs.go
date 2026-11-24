// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package upper

import (
	"seehuhn.de/go/bedern/internal/outline"
)

// The refits of the glyphs of TS1 whose curves the shear distorts: each is
// the function of a source (see refit), found by looking at the sheared outline
// where the smoothness check reports a finding.

// straightAt returns the path with the curve which ends nearest to q made a
// line to the point to; the curve which follows starts there.
func straightAt(path outline.Path, q, to Point) outline.Path {
	ci, si := segAt(path, q)
	nc := outline.Contour{Start: path[ci].Start, Segs: append([]outline.Segment(nil), path[ci].Segs...)}
	nc.Segs[si] = outline.Segment{P3: to}
	if si == len(nc.Segs)-1 {
		nc.Start = to
	}
	out := append(outline.Path(nil), path...)
	out[ci] = nc
	return out
}

// collapseAt returns the path without the tiny segment which ends nearest to
// q: the edge before it ends where the segment ended (atEnd) or where it
// started, and the edge after it starts there.  The contour may start or end
// with the segment.
func collapseAt(path outline.Path, q Point, atEnd bool) outline.Path {
	ci, si := segAt(path, q)
	c := path[ci]
	n := len(c.Segs)
	tiny := c.Segs[si]
	var start Point
	if si == 0 {
		start = c.Start
	} else {
		start = c.Segs[si-1].P3
	}
	nc := outline.Contour{Start: c.Start}
	switch {
	case si == 0:
		nc.Segs = append(nc.Segs, c.Segs[1:]...)
		if atEnd {
			nc.Start = tiny.P3
			nc.Segs[n-2].P3 = tiny.P3
		}
	case si == n-1:
		nc.Segs = append(nc.Segs, c.Segs[:n-1]...)
		if !atEnd {
			nc.Start = start
		} else {
			nc.Segs[n-2].P3 = tiny.P3
		}
	default:
		nc.Segs = append(nc.Segs, c.Segs[:si]...)
		if atEnd {
			nc.Segs[si-1].P3 = tiny.P3
		}
		nc.Segs = append(nc.Segs, c.Segs[si+1:]...)
	}
	out := append(outline.Path(nil), path...)
	out[ci] = nc
	return out
}

// splitExtreme returns the path with the curve which ends nearest to q cut in
// two at its extreme in y (axis 1) or x (axis 0).
func splitExtreme(path outline.Path, q Point, axis int) outline.Path {
	ci, si := segAt(path, q)
	c := path[ci]
	cu := curveAt(c, si)
	get := func(p Point) float64 {
		if axis == 0 {
			return p.X
		}
		return p.Y
	}
	ts := outline.CubicExtrema(get(cu[0]), get(cu[1]), get(cu[2]), get(cu[3]))
	if len(ts) != 1 {
		panic("upper: the curve of a split has no single extreme")
	}
	a, b := cu.split(ts[0])
	nc := outline.Contour{Start: c.Start, Segs: append([]outline.Segment(nil), c.Segs[:si]...)}
	nc.Segs = append(nc.Segs, outline.Segment{IsCurve: true, P1: a[1], P2: a[2], P3: a[3]}, outline.Segment{IsCurve: true, P1: b[1], P2: b[2], P3: b[3]})
	nc.Segs = append(nc.Segs, c.Segs[si+1:]...)
	out := append(outline.Path(nil), path...)
	out[ci] = nc
	return out
}

// refitDollar draws the places of the old-style dollar again where the ticks of
// its stroke meet the S: the corners of the ticks are corners and the flat
// stretches between them are lines.
func refitDollar(p outline.Path) outline.Path {
	p = collapseAt(p, pt(248.9, 309.4), false)
	p = collapseAt(p, pt(291.5, 380.1), false)
	p = collapseAt(p, pt(336, 676), true)
	p = splitExtreme(p, pt(336, 676), 1)
	p = straightAt(p, pt(284, 292), pt(284, 292))
	p = straightAt(p, pt(186, -12), pt(186, -12))
	p = straightAt(p, pt(208, -12), pt(208, -12))
	return straightAt(p, pt(371, 640), pt(371, 640))
}

// figureRefits are the refits of the figures, by name: the short curve in the
// neck of the ball of the 3 is one curve with the neck, which starts with the
// curvature of the ball.
var figureRefits = map[string]func(outline.Path) outline.Path{
	"five.oldstyle":     func(p outline.Path) outline.Path { return mergeG2(p, pt(76.8, -157.6)) },
	"three.prop":        func(p outline.Path) outline.Path { return mergeG2(p, pt(199.9, 612.5)) },
	"three.oldstyle":    func(p outline.Path) outline.Path { return mergeG2(p, pt(135.6, 400.5)) },
	"three.taboldstyle": func(p outline.Path) outline.Path { return mergeG2(p, pt(146, 401)) },
}

// blankK is the curvature at the node at the lowest point of the underside of
// the bowl of the blank symbol, between those of the curves on its two sides.
const blankK = -0.0085

// refitBlank draws the curve before that node again, ending with blankK.
func refitBlank(p outline.Path) outline.Path {
	return setEnds(p, pt(222.5, 20), keep, blankK)
}

// refitFlorin draws the inner edge of the hook again, ending without curvature
// where it meets the stem.
func refitFlorin(p outline.Path) outline.Path {
	return setEnds(p, pt(207, 576), keep, 0)
}

// refitOmega makes the short curve at the foot of the leg a line and trims the
// curves which run beyond their nodes.
func refitOmega(p outline.Path) outline.Path {
	return trimOvershoots(straightAt(p, pt(274, 90), pt(273, 90)))
}

// refitLira joins the stem to the ends of the two bars in corners, without the
// tiny curve between.
func refitLira(p outline.Path) outline.Path {
	p = collapseAt(p, pt(163, 298.8), true)
	return collapseAt(p, pt(244.5, 300), false)
}
