// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package upper

import "seehuhn.de/go/bedern/internal/font/design"

// capOCounter returns the round of the counter of O, in the roman parameters.
func capOCounter(rp design.Params) round {
	c := rp.Caps
	cx := c.SideRoundLeft + c.OWidth/2
	cy := rp.CapHeight / 2
	return full(pt(cx, cy), c.OWidth/2-c.Round, rp.CapHeight/2+rp.OvershootUpper-c.RoundThin)
}

// The capitals are the roman capitals made italic; the rounds of O, Q, C, G,
// the bowls of B, D, P and R, the bottoms of U and J and the two halves of S
// turn to the stress of the italic.
func init() {
	for _, n := range "AEFHIKLMNTVWXYZ" {
		sources[string(n)] = source{r: n}
	}
	sources["O"] = source{r: 'O', rounds: allCounters("O")}
	sources["Q"] = source{r: 'Q', rounds: allCounters("Q")}
	sources["C"] = source{r: 'C', rounds: openRound(50, 310)}
	sources["G"] = source{r: 'G', rounds: openRound(50, 300)}
	for _, n := range "BDPR" {
		sources[string(n)] = source{r: n, rounds: bowlCounters(string(n))}
	}
	sources["U"] = source{r: 'U', rounds: bottomRound("U", 250)}
	sources["J"] = source{r: 'J', rounds: bottomRound("J", 250), ball: func(rp design.Params) float64 { return rp.Caps.CGSJQ.JBall }}
	sources["S"] = source{r: 'S', rounds: sRounds("S")}
}
