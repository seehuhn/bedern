// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package design

// This file holds the parameters of the frame of .notdef.  The construction is
// in internal/glyphs/notdef.go.

// Notdef holds the parameters of .notdef: the frame and the diagonals are Thick
// thick, and the side space on both sides of the ink is Side.  The glyph has the
// advance and the height of the figures.
type Notdef struct {
	Thick float64
	Side  float64
}

// regularNotdef returns the parameters of .notdef of Bedern Regular.
func regularNotdef() Notdef {
	return Notdef{Thick: 38, Side: 40}
}
