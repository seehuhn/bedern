// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"math"
	"sort"
	"sync"
)

// The italic Greek lowercase of the maths font: drawn with the pen, the parts and the slant of
// the italic, as one hand with the maths italic.  Each letter is
// registered by its file with addGreek, under its glyph name and the code
// point of the Greek block; the maths font maps it to the mathematical
// italic code point.

// greekLetters are the drawn Greek letters, by name.
var greekLetters = map[string]struct {
	r    rune
	draw func(Params) letter
}{}

// addGreek registers a Greek letter, its set width rounded to whole units.
func addGreek(name string, r rune, draw func(Params) letter) {
	greekLetters[name] = struct {
		r    rune
		draw func(Params) letter
	}{r, func(p Params) letter {
		l := draw(p)
		l.advance = math.Round(l.advance)
		return l
	}}
}

// GreekNames returns the names of the drawn Greek letters, sorted.
func GreekNames() []string {
	var out []string
	for n := range greekLetters {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

// GreekGlyphs returns the drawn Greek letters, built in parallel.
func GreekGlyphs(p Params) ([]*Glyph, error) {
	names := GreekNames()
	out := make([]*Glyph, len(names))
	errs := make([]error, len(names))
	var wg sync.WaitGroup
	for i, name := range names {
		wg.Add(1)
		go func() {
			defer wg.Done()
			out[i], errs[i] = Build(p, name)
		}()
	}
	wg.Wait()
	for _, err := range errs {
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}
