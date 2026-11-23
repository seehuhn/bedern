// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package designtest

import "seehuhn.de/go/bedern/internal/font/design"

// figuresUp moves the parameters of the figures of the set caps-up: taller
// figures, heavier strokes, a larger ball set higher and further in on 6 and 9,
// longer serifs and wider ink.
func figuresUp(p *design.Params) {
	f := &p.Figures
	f.Height, f.OverTop, f.Advance = 700, 10, 515
	f.Stem, f.Round, f.Thin, f.Heavy, f.Bowl, f.Ball = 92, 104, 44, 88, 102, 96
	for i := range f.Ink {
		if f.Ink[i] > 0 {
			f.Ink[i] += 20
		}
	}
	f.OneFlag, f.OneFoot, f.OneFlagThick, f.OneRise = 132, 124, 44, 18
	f.TwoBowlMid, f.TwoShoulder, f.TwoRise, f.TwoDiagFoot = 445, 501, 95, 148
	f.ThreeWaist, f.ThreeUpperWidth, f.ThreeUpperRight, f.ThreeUpperBowl = 362, 395, 395, 98
	f.ThreeOpen, f.ThreeUpperBall = 66, 20
	f.ThreeLowerLeft, f.ThreeLowerBall = -30, 0
	f.FourBarLow, f.FourFoot = 180, 104
	f.FiveFlag, f.FiveBowlTop, f.FiveStroke = 330, 434, 50
	f.SixBowlTop, f.SixLeftY, f.SixArcWidth = 455, 316, 432
	f.CurlHeight, f.SixBallIn = 139, 44
	f.SevenDrop, f.SevenFootX, f.SevenFoot = 100, 186, 92
	f.EightWaist, f.EightUpperWidth, f.EightGap = 378, 374, 82
}

// figuresDown is the opposite of figuresUp, for the set caps-down.
func figuresDown(p *design.Params) {
	f := &p.Figures
	f.Height, f.OverTop, f.Advance = 640, 6, 470
	f.Stem, f.Round, f.Thin, f.Heavy, f.Bowl, f.Ball = 75, 86, 32, 70, 84, 84
	for i := range f.Ink {
		if f.Ink[i] > 0 {
			f.Ink[i] -= 16
		}
	}
	f.Ink[4] -= 2 // the ink of 4 is narrowed by 2 more
	f.OneFlag, f.OneFoot, f.OneFlagThick, f.OneRise = 118, 110, 34, 10
	f.TwoBowlMid, f.TwoShoulder, f.TwoRise, f.TwoDiagFoot = 402, 458, 80, 126
	f.ThreeWaist, f.ThreeUpperWidth, f.ThreeUpperRight, f.ThreeUpperBowl = 330, 380, 360, 80
	f.FourBarLow, f.FourFoot = 162, 90
	f.FiveFlag, f.FiveBowlTop, f.FiveStroke, f.FiveInnerLeft = 302, 396, 40, -74
	f.SixBowlTop, f.SixLeftY, f.SixArcWidth = 416, 288, 396
	f.SevenDrop, f.SevenFootX, f.SevenFoot, f.SevenSerifThick = 86, 168, 78, 40
	f.EightWaist, f.EightUpperWidth, f.EightGap = 346, 342, 72
}
