// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package upper

import (
	"seehuhn.de/go/bedern/internal/outline"
)

// neighbourK returns the curvature which the edge has at the node at the
// start of segment si of the contour (before true) or at its end (before
// false), as seen from the segment on the other side of the node: that of the
// curve, or 0 for a line.
func neighbourK(c outline.Contour, si int, atStart bool) float64 {
	n := len(c.Segs)
	if atStart {
		j := (si - 1 + n) % n
		if !c.Segs[j].IsCurve {
			return 0
		}
		return curveAt(c, j).curvature(1)
	}
	j := (si + 1) % n
	if !c.Segs[j].IsCurve {
		return 0
	}
	return curveAt(c, j).curvature(0)
}

// mergeG2 returns the path with the two curves which meet at the node nearest
// to q replaced by one cubic: the one closest to them with their directions at
// its ends (fit), drawn again with the curvatures which the neighbouring curves
// have at its ends; the edge is continuous in curvature there; the
// curvature at an end which neighbours a line, or where the curvature given is
// not usable, is that of the fitted curve.
func mergeG2(path outline.Path, q Point) outline.Path {
	ci, si := segAt(path, q)
	c := path[ci]
	n := len(c.Segs)
	sj := (si + 1) % n
	a, b := curveAt(c, si), curveAt(c, sj)
	m, _, ok := fit(a, b)
	if !ok {
		panic("upper: no merged curve at the place of a refit")
	}
	k0, k3 := neighbourK(c, si, true), neighbourK(c, sj, false)
	if g, ok := g2(m[0], unit(sub(m[1], m[0])), k0, m[3], unit(sub(m[3], m[2])), k3); ok {
		m = g
	}
	nc := outline.Contour{Start: c.Start, Segs: append([]outline.Segment(nil), c.Segs...)}
	setCurve(&nc, si, m)
	nc.Segs = append(nc.Segs[:sj], nc.Segs[sj+1:]...)
	out := append(outline.Path(nil), path...)
	out[ci] = nc
	return out
}
