// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package designtest

import "seehuhn.de/go/bedern/internal/font/design"

// capDiagUp moves the parameters of the diagonal capitals A, V, W, X, Y, Z, K,
// M and N in the set caps-up: heavier strokes, longer serifs, wider ink, deeper
// vertices and apex, the crossings and junctions higher by half the change of
// the cap height.
func capDiagUp(p *design.Params) {
	d := &p.CapDiag
	d.Thick, d.Thin, d.WThick, d.XThick, d.MThick, d.NThick, d.KLeg, d.ZDiag = 100, 46, 94, 95, 92, 98, 96, 98
	d.ThinStem, d.MStem = 50, 100
	d.OverThick, d.OverThin, d.KOverLeg, d.KOverArm = 72, 86, 64, 76
	d.VertexFlat, d.VertexDepth, d.ApexFlat, d.ApexRise = 32, 14, 32, 24
	d.WApexFlat, d.WApexHeight = 36, p.CapHeight-6
	d.NFlat, d.NDepth, d.NJoin, d.MFlat, d.MHeight, d.MCrotchDepth = 28, 12, 62, 28, 44, 85
	d.ABar, d.XCross, d.YJunction, d.KJunction = d.ABar+12, d.XCross+17, d.YJunction+17, d.KJunction+17
	d.KArmAngle, d.KLegAngle, d.XThinOffset, d.XLeanDiff = 47, 35, -8, 0.035
	d.AInk, d.VInk, d.WInk, d.XInk, d.YInk, d.ZInk, d.KInk, d.MInk = 690, 715, 995, 705, 730, 510, 705, 845
	d.ZInkTop = 480
}

// capDiagDown is the opposite of capDiagUp, for the set caps-down.
func capDiagDown(p *design.Params) {
	d := &p.CapDiag
	d.Thick, d.Thin, d.WThick, d.XThick, d.MThick, d.NThick, d.KLeg, d.ZDiag = 86, 38, 82, 83, 80, 85, 83, 85
	d.ThinStem, d.MStem = 38, 88
	d.OverThick, d.OverThin, d.KOverLeg, d.KOverArm = 60, 74, 56, 64
	d.VertexFlat, d.VertexDepth, d.ApexFlat, d.ApexRise = 24, 7, 24, 16
	d.WApexFlat, d.WApexHeight = 26, p.CapHeight-3
	d.NFlat, d.NDepth, d.NJoin, d.MFlat, d.MHeight, d.MCrotchDepth = 20, 7, 48, 20, 28, 65
	d.ABar, d.XCross, d.YJunction, d.KJunction = d.ABar-12, d.XCross-17, d.YJunction-17, d.KJunction-17
	d.KArmAngle, d.KLegAngle, d.XThinOffset, d.XLeanDiff = 45, 33, -4, 0.02
	d.AInk, d.VInk, d.WInk, d.XInk, d.YInk, d.ZInk, d.KInk, d.MInk = 640, 665, 930, 655, 680, 470, 660, 790
	d.ZInkTop = 442
}
