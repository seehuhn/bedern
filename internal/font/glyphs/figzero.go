// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// figZero draws the figure 0 as O is drawn, narrower: two super-ellipses with
// the squareness of the o, the outer one counter-clockwise and the counter
// clockwise.  The sides are Figures.Round thick, the weight of the round stroke
// of o, and the top and bottom Figures.Thin; the zero reaches OverTop above the
// height of the figures and OvershootUpper below the baseline.  It is
// Figures.Ink[0] wide, 0.64 of the width of O.
func figZero(p design.Params) (outline.Path, float64) {
	f := p.Figures
	x0 := design.FigureLeft(p, 0)
	x1 := x0 + f.Ink[0]
	y0, y1 := figBottom(p), figTop(p)
	var path outline.Path
	superEllipse(&path, x0, y0, x1, y1, p.SupernessOuter, true)
	superEllipse(&path, x0+f.Round, y0+f.Thin, x1-f.Round, y1-f.Thin, p.SupernessInner, false)
	return path, f.Advance
}
