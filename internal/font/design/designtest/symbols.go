// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package designtest

import "seehuhn.de/go/bedern/internal/font/design"

// symbolsUp moves the parameters of the question mark, the ampersand and the
// symbols of the set caps-up: taller signs, heavier strokes and thin parts.
func symbolsUp(p *design.Params) {
	s := &p.Sym
	s.QTop, s.QInk, s.QStem = 715, 390, 56
	s.AmpTop, s.AmpArm, s.AmpThick, s.AmpBowlThick, s.AmpTail = 730, 440, 106, 108, 44
	s.SectSpine, s.SectHookThick = 86, 76
	s.AtRing, s.AtStem, s.AtBowlThick = 46, 86, 94
	s.BraceArm, s.DegThick, s.TildeThick, s.CircThick = 82, 46, 68, 52
	s.NumStroke, s.NumBar, s.PctZeroThick = 56, 60, 66
	s.DollarSpine, s.DollarStroke, s.SterStem, s.SterBase, s.EuroThick = 104, 44, 96, 80, 108
}

// symbolsDown is the opposite of symbolsUp, for the set caps-down.
func symbolsDown(p *design.Params) {
	s := &p.Sym
	s.QTop, s.QInk, s.QStem = 670, 350, 44
	s.AmpTop, s.AmpArm, s.AmpThick, s.AmpBowlThick, s.AmpTail = 670, 400, 84, 92, 34
	s.SectSpine, s.SectHookThick = 68, 58
	s.AtRing, s.AtStem, s.AtBowlThick = 34, 68, 74
	s.BraceArm, s.DegThick, s.TildeThick, s.CircThick = 64, 34, 52, 40
	s.NumStroke, s.NumBar, s.PctZeroThick = 44, 48, 50
	s.DollarSpine, s.DollarStroke, s.SterStem, s.SterBase, s.EuroThick = 82, 32, 78, 64, 86
}
