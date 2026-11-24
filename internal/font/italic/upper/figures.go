// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package upper

import "seehuhn.de/go/bedern/internal/font/design"

// figureBall returns the diameter of the ball terminals of the figures.
func figureBall(rp design.Params) float64 { return rp.Figures.Ball }

// The figures are the lining figures of the roman made italic, in the common
// advance of the roman, 490, narrowed about the middle of the cell; the closed counters of 0, 6, 8 and
// 9 turn to the stress of the italic and the balls stay circles.
func init() {
	names := []string{"zero", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine"}
	for i, n := range names {
		src := source{r: rune('0' + i), ball: figureBall, tabular: true}
		switch n {
		case "zero", "six", "eight", "nine":
			src.rounds = allCounters(n)
		}
		// The zones of the balls, chosen by the smoothness check: wide where
		// the hook is long (6 and 9) and narrow where the bowl of 5 passes
		// close to its ball.
		switch n {
		case "six", "nine":
			src.zone = 3.5
		case "three":
			src.zone = 1
		case "five":
			src.zone = 1
		}
		sources[n] = src
	}
}
