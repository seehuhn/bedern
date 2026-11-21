// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package design

import (
	"math"
)

// This file holds what follows from the parameters of the letter g: the
// positions of its bowl, its loop and its ear and the advance.  The construction
// is described in internal/glyphs/g.go.

// GLoopLeft returns the position of the left extreme of the lower loop of g, the
// leftmost ink of the letter: the left side space GSideLeft of g.
func GLoopLeft(p Params) float64 { return p.GSideLeft }

// GLoopRight returns the position of the right extreme of the lower loop.
func GLoopRight(p Params) float64 { return GLoopLeft(p) + p.GLoopWidth }

// GBowlLeft returns the position of the left extreme of the upper bowl of g, which
// lies GBowlInset to the right of that of the loop.
func GBowlLeft(p Params) float64 { return GLoopLeft(p) + p.GBowlInset }

// GBowlRight returns the position of the right extreme of the upper bowl.
func GBowlRight(p Params) float64 { return GBowlLeft(p) + p.GBowlWidth }

// GBowlTop returns the height of the top of the upper bowl, which is that of the o.
func GBowlTop(p Params) float64 { return p.XHeight + p.OvershootLower }

// GBowlBottom returns the height of the bottom of the upper bowl.
func GBowlBottom(p Params) float64 { return GBowlTop(p) - p.GBowlHeight }

// GEarEnd returns the position of the right extreme of the ball of the ear.
func GEarEnd(p Params) float64 { return GBowlRight(p) + p.GEarReach }

// GEarCrest returns the height of the highest point of the ear.
func GEarCrest(p Params) float64 { return p.XHeight + p.GEarRise }

// GAdvance returns the advance of the letter g: the space GSideRight to the right
// of the rightmost ink, which is the right extreme of the loop unless the ear
// reaches further.
func GAdvance(p Params) float64 {
	return math.Max(GLoopRight(p), GEarEnd(p)) + p.GSideRight
}
