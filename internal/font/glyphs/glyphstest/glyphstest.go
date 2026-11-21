// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

// Package glyphstest supplies the glyphs of a parameter set to the tests of
// several packages.  The set is built once per test binary for each value of
// the parameters and shared, safely for parallel tests.  The caller gets copies
// and may change them.
package glyphstest

import (
	"slices"
	"sync"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/outline"
)

type entry struct {
	once sync.Once
	gs   []glyphs.Glyph
}

var (
	mu    sync.Mutex
	cache = map[design.Params]*entry{}
)

// Set returns what glyphs.Set returns for the parameters.
func Set(p design.Params) []glyphs.Glyph {
	mu.Lock()
	e := cache[p]
	if e == nil {
		e = new(entry)
		cache[p] = e
	}
	mu.Unlock()
	e.once.Do(func() { e.gs = glyphs.Set(p) })
	gs := make([]glyphs.Glyph, len(e.gs))
	for i, g := range e.gs {
		g.Path = clone(g.Path)
		gs[i] = g
	}
	return gs
}

// Glyph returns the glyph of Set(p) with the given name, or the zero glyph if
// there is none.
func Glyph(name string, p design.Params) glyphs.Glyph {
	for _, g := range Set(p) {
		if g.Name == name {
			return g
		}
	}
	return glyphs.Glyph{}
}

func clone(p outline.Path) outline.Path {
	if p == nil {
		return nil
	}
	q := make(outline.Path, len(p))
	for i, c := range p {
		q[i] = outline.Contour{Start: c.Start, Segs: slices.Clone(c.Segs)}
	}
	return q
}
