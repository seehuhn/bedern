// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

import (
	"fmt"
	"math"
	"sync"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The script forms of the signs and the primes.
//
// A sign is drawn by its construction with the script parameter set of the
// level (scriptParams), as the figures are: the rule of the signs, the
// stems and the round strokes grow by the factors of the level and the signs
// keep their ink width.  The
// form then takes the advance of its text form times the factor of its
// level, the white beyond its ink divided equally between its sides.  ∂ and ℓ
// are drawn with the pen of the italic by italic.BuildScript, as the letters
// of the maths italic are.
//
// A prime has its own placement: the script forms stand lower and larger than
// the raised prime of the text form; a prime in a superscript, shifted
// and scaled with it, ends near 750.

// scriptSignRunes are the signs which have script forms, at their code points
// in unicode-math: the ASCII signs, the signs of arithmetic and of the
// relations, the arrows, the set and logic symbols and ∂ and ∇.  ℓ, the
// script l of the maths italic, has script forms as well (scriptSymbolNames).
const scriptSignRunes = "+−=<>±×⋅∗()[]{}|,∶∞∈≤≥→⇒↦≠≡∼∂∇∀∃∅"

// scriptSymbolNames are the symbols drawn with the pen of the italic.
var scriptSymbolNames = []string{"uni2113", "uni2202"}

// scriptPrimeNames are the primes, which have a construction of their own
// in the script forms.
var scriptPrimeNames = []string{"minute", "second", "uni2034", "uni2057"}

// scriptPrimes gives, for each level, the lowest and the highest point of the
// ink of a prime in a script form: about 100 to 530 at the first level and
// 90 to 630 at the second; after the shift of 376 and the scaling by 0.70 the
// top of a prime in a superscript is near 750.
var scriptPrimes = [2]struct{ bottom, top float64 }{{100, 530}, {90, 630}}

// The weight and the lean of the script primes.  scriptPrimeWeights multiply
// the weight of the prime and scriptPrimeLean is its lean in degrees.
var (
	scriptPrimeWeights = [2]float64{1.31, 1.29}
	scriptPrimeLean    = [2]float64{22, 20}
)

// scriptPrime returns n primes of the parameters of a level: the tapered
// bar of the prime of the regular at the lean of the level, repeated at the
// pitch of the double prime, with the side spaces of the prime.
func scriptPrime(r *roman, level, n int) (outline.Path, float64) {
	q := r.p.Punct
	one, adv := r.get("minute")
	x0, _, x1, _ := one.BBox()
	h := q.StraightHeight
	wb, wt := q.PrimeBottomWidth, q.PrimeTopWidth
	top := r.p.Ascender
	lean := math.Tan(scriptPrimeLean[level-1]*math.Pi/180) * h
	var path outline.Path
	for i := range n {
		x := float64(i) * q.StraightPitch
		path = join(path, outline.Path{polygon(
			pt(x-wb/2, top-h), pt(x+wb/2, top-h), pt(x+lean+wt/2, top), pt(x+lean-wt/2, top))})
	}
	return withSides(path, x0, adv-x1)
}

// scriptSignNames returns the names of the glyphs of the signs which have
// script forms drawn by the constructions of the regular.
var scriptSignNames = sync.OnceValue(func() []string {
	byRune := map[rune]string{}
	for _, e := range glyphTable() {
		byRune[e.r] = e.name
	}
	for _, d := range derivations {
		byRune[d.r] = d.name
	}
	var out []string
	for _, r := range scriptSignRunes {
		n, ok := byRune[r]
		if !ok {
			panic(fmt.Sprintf("mathfont: no glyph for %U", r))
		}
		out = append(out, n)
	}
	return out
})

// drawSign returns the sign of the name as the roman r draws it: a derived
// sign by its construction, the others as the regular has them.
func drawSign(r *roman, name string) (outline.Path, float64) {
	for _, d := range derivations {
		if d.name == name {
			return d.draw(r)
		}
	}
	return r.get(name)
}

// scriptPrimeParams returns the parameters of the construction of the
// primes at a level: the script parameter set with the bars as high as the
// script prime and the pitch of a repeated prime as wide as the advance.
func scriptPrimeParams(level int) design.Params {
	q := scriptParams(level)
	s := scriptPrimes[level-1]
	q.Ascender = s.top
	q.Punct.StraightHeight = s.top - s.bottom
	q.Punct.StraightPitch = q.Punct.StraightPitch * scriptLevels[level-1].advance
	q.Punct.PrimeTopWidth *= scriptPrimeWeights[level-1]
	q.Punct.PrimeBottomWidth *= scriptPrimeWeights[level-1]
	return q
}

// signScriptForms are the forms of the signs, ℓ and the primes.
var signScriptForms = sync.OnceValues(func() (scriptSet, error) {
	return buildScripts(func(level int, put func(scriptForm), run func(func()), fail func(error)) {
		r := scriptRomans[scriptSigns][level-1]()
		for _, n := range scriptSignNames() {
			if n == "uni2202" {
				continue
			}
			run(func() {
				path, adv := drawSign(r, n)
				put(scriptForm{Name: n, Level: level, Path: path, Advance: adv})
			})
		}
		for _, n := range scriptSymbolNames {
			run(func() {
				g, stem, err := buildLower(n, level, true)
				if err != nil {
					fail(err)
					return
				}
				put(scriptForm{Name: n, Level: level, Path: g.Path, Advance: g.Advance, Stem: stem, Warnings: g.Warnings})
			})
		}
		pr := newRoman(scriptPrimeParams(level))
		for i, n := range scriptPrimeNames {
			run(func() {
				path, adv := scriptPrime(pr, level, i+1)
				put(scriptForm{Name: n, Level: level, Path: path, Advance: adv})
			})
		}
	})
})
