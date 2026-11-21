// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package design

import (
	"math"
)

// This file holds what follows from the parameters of the letters f, t and j:
// the positions of the stems, the advances, the left ink of the letters whose
// leftmost ink is the end of a crossbar.

// FStemLeft returns the position of the left-hand edge of the stem of f: that
// of l, the lighter ascender stem standing in the cell of a stem of width Stem.
func FStemLeft(p Params) float64 { return NarrowStemLeft(p) + AscenderInset(p) }

// FCellRight returns the right-hand edge of the cell of the stem of f, from
// which the reach of the hook is measured.
func FCellRight(p Params) float64 { return NarrowStemLeft(p) + p.Stem }

// FAdvance returns the advance of the letter f: the stem with the cell on the
// left, the reach of the hook beyond it, and the space to the right of the
// hook, which is negative where the hook overhangs the advance.
func FAdvance(p Params) float64 { return FCellRight(p) + p.FReach + p.FSideRight }

// FApex returns the height of the highest point of the hook of f.
func FApex(p Params) float64 { return p.Ascender + p.FOvershoot }

// FInkLeft returns the distance from the left edge of the advance of f to the
// leftmost ink, which is the end of the crossbar or the tip of the left foot
// serif, whichever reaches further.
func FInkLeft(p Params) float64 { return FStemLeft(p) - math.Max(p.FBarLeft, p.SerifJut) }

// TStemLeft returns the position of the left-hand edge of the stem of t, the
// stem of n.
func TStemLeft(p Params) float64 { return p.SideStemLeft }

// TTop returns the height of the flat top of the stem of t.
func TTop(p Params) float64 { return p.XHeight + p.TRise }

// TAdvance returns the advance of the letter t: the stem with the left stem
// side, the reach of the tail beyond it and the space to the right of the tail.
func TAdvance(p Params) float64 { return p.SideStemLeft + p.Stem + p.TReach + p.TSideRight }

// TInkLeft returns the distance from the left edge of the advance of t to the
// leftmost ink, the end of the crossbar.
func TInkLeft(p Params) float64 { return p.SideStemLeft - p.TBarLeft }

// JStemRight returns the position of the right-hand edge of the stem of j.
func JStemRight(p Params) float64 { return p.JStemLeft + p.Stem }

// JAdvance returns the advance of the letter j: the stem, which stands JStemLeft
// from the left edge, and the right stem side of the narrow letters.
func JAdvance(p Params) float64 { return JStemRight(p) + NarrowSideRight(p) }

// JInkLeft returns the position of the leftmost ink of j, the end of the tail,
// which is negative where the tail overhangs the advance.
func JInkLeft(p Params) float64 { return p.JStemLeft - p.JReach }
