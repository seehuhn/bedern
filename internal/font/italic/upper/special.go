// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package upper

import (
	"seehuhn.de/go/bedern/internal/font/design"
)

// letterBall returns the diameter of the ball terminals of the letters of the
// roman, which ẞ takes from the terminal of 3.
func letterBall(rp design.Params) float64 { return rp.BallDiameter }

// The special capitals of the roman made italic by the method of the
// capitals: ẞ, whose bowls hang on the stem and whose ball stays a circle.
func init() {
	sources["uni1E9E"] = source{r: 'ẞ', zone: specialZone, rounds: bowlCounters("uni1E9E"), ball: letterBall}
	sources["Thorn"] = source{r: 'Þ', rounds: bowlCounters("Thorn")}
	sources["Eng"] = source{r: 'Ŋ', ball: func(rp design.Params) float64 { return rp.Caps.CGSJQ.JBall }}
	sources["IJ"] = source{r: 'Ĳ', rounds: bottomRound("IJ", 250), ball: func(rp design.Params) float64 { return rp.Caps.CGSJQ.JBall }}
	sources["Ldot"] = source{r: 'Ŀ', dots: punctDots}
	sources["AE"] = source{r: 'Æ'}
	sources["OE"] = source{r: 'Œ', rounds: func(_ Params, rp design.Params) []round {
		path, _, _ := romanOutline("OE", rp)
		return counters(path, 110, 250, 25)
	}}
	judgedFindings["AE"] = []judged{{Kind: extremeKind, X: 1023, Y: 615, Dir: -90, Reason: beakExtreme}}
	judgedFindings["OE"] = []judged{{Kind: extremeKind, X: 966, Y: 615, Dir: -90, Reason: beakExtreme}}
}

// specialZone is the zone of the ball of ẞ, see circle.
const specialZone = 1.0
