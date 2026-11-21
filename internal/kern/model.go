// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package kern

import (
	"math"
)

// ZoneOf returns the bands in which the pair a, b is seen, and the letters of
// the pair which sets the target.
//
// The zone is the band range from the baseline to the x-height for two
// lowercase letters, and from the baseline to the cap height as soon as a
// capital takes part.  The ascenders of lowercase letters do not widen the
// zone.  A pair with a punctuation mark is seen where the mark has ink (the
// dot of the full stop, the whole height of a quotation mark).
func (f *Font) ZoneOf(a, b *Glyph) (z Zone, ref string) {
	xrows := int(f.P.XHeight / f.P.Kern.Band)
	crows := int(f.P.CapHeight / f.P.Kern.Band)
	zero := f.Row(0)
	upper := a.Kind == Upper || b.Kind == Upper
	lower := a.Kind == Lower || b.Kind == Lower
	var m, o *Glyph
	switch {
	case a.Kind == Punct:
		m, o = a, b
	case b.Kind == Punct:
		m, o = b, a
	}
	top := zero + xrows
	if upper {
		top = zero + crows
	}
	switch {
	case upper && lower:
		ref = "Hn"
	case upper:
		ref = "HH"
	default:
		ref = "nn"
	}
	if m == nil {
		return Zone{zero, top}, ref
	}
	// The rows of the mark.
	lo := f.Row(math.Floor(m.Bottom/f.P.Kern.Band) * f.P.Kern.Band)
	hi := f.Row(math.Ceil(m.Top/f.P.Kern.Band) * f.P.Kern.Band)
	if m.Bottom > float64(xrows)*f.P.Kern.Band*0.8 {
		// A raised mark is seen against the whole height of the letter.
		if !upper {
			top = zero + crows
			ref = "ll"
		}
	}
	if lo < zero && (m.Bottom >= 0 || o.Bottom >= 0) {
		// The tail of a comma below a letter without a descender is left
		// out of the zone.
		lo = zero
	}
	if lo < 0 {
		lo = 0
	}
	if hi > top {
		hi = top
	}
	if lo > top {
		lo = top
	}
	return Zone{lo, hi}, ref
}

// white returns the mean white of the pair over the zone when b is moved by k.
func (f *Font) white(a, b *Glyph, z Zone, k float64) float64 {
	depth := f.P.Kern.Depth
	var sum float64
	for i := z.Lo; i < z.Hi; i++ {
		ra, lb := a.Right[i], b.Left[i]
		if math.IsNaN(ra) || math.IsNaN(lb) {
			sum += depth
			continue
		}
		sum += math.Max(0, math.Min(depth, ra+lb+k))
	}
	return sum / float64(z.Hi-z.Lo)
}

// glyphOf finds the glyph with the given letter.
func (f *Font) glyphOf(r rune) *Glyph {
	for _, g := range f.Glyphs {
		if g.Rune == r {
			return g
		}
	}
	return nil
}

// Target returns the mean white which the pair should have in the zone.
func (f *Font) Target(ref string, z Zone) float64 {
	g := func(s string) *Glyph { return f.glyphOf(rune(s[0])) }
	switch ref {
	case "Hn":
		H, n := g("H"), g("n")
		return (f.white(H, n, z, 0) + f.white(n, H, z, 0)) / 2
	default:
		a, b := g(ref[:1]), g(ref[1:])
		return f.white(a, b, z, 0)
	}
}

// Solve returns the shift k of b for which the mean white over the zone is the
// target: the smallest k for which it is not less.
func (f *Font) Solve(a, b *Glyph, z Zone, target float64) float64 {
	lo, hi := -600.0, 600.0
	if f.white(a, b, z, hi) < target {
		return hi
	}
	if f.white(a, b, z, lo) >= target {
		return lo
	}
	for i := 0; i < 40; i++ {
		mid := (lo + hi) / 2
		if f.white(a, b, z, mid) >= target {
			hi = mid
		} else {
			lo = mid
		}
	}
	return hi
}
