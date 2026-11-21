// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"runtime"
	"slices"
	"sync"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// A glyph is built once per test binary and shared.  The cache is keyed by the
// glyph name and the value of the parameters, is filled when a glyph is first
// asked for, and is safe for parallel tests.
// The paths which cachedOutline and cachedRounded return are shared and must
// not be changed by the tests; outlineOf and setOf return copies.

type buildKey struct {
	glyph  string
	params design.Params
}

type builtGlyph struct {
	once    sync.Once
	ok      bool
	rune    rune
	path    outline.Path
	advance float64

	roundOnce sync.Once
	rounded   outline.Path
}

// gate bounds the goroutines which a test starts itself.
var gate = make(chan struct{}, runtime.GOMAXPROCS(0))

var (
	buildMu sync.Mutex
	builds  = map[buildKey]*builtGlyph{}
)

// built returns the shared build of the glyph in the parameters p.
func built(glyph string, p design.Params) *builtGlyph {
	k := buildKey{glyph, p}
	buildMu.Lock()
	b := builds[k]
	if b == nil {
		b = new(builtGlyph)
		builds[k] = b
	}
	buildMu.Unlock()
	b.once.Do(func() {
		if e, ok := lookup(glyph); ok {
			b.path, b.advance = e.draw(p)
			b.rune, b.ok = e.r, true
		}
	})
	return b
}

// roundedPath returns the rounded outline of the shared build.
func (b *builtGlyph) roundedPath() outline.Path {
	b.roundOnce.Do(func() { b.rounded = b.path.Round() })
	return b.rounded
}

// clonePath returns a deep copy of the path.
func clonePath(p outline.Path) outline.Path {
	if p == nil {
		return nil
	}
	q := make(outline.Path, len(p))
	for i, c := range p {
		q[i] = outline.Contour{Start: c.Start, Segs: slices.Clone(c.Segs)}
	}
	return q
}

// outlineOf is Outline with shared builds: the result is a copy which the
// caller may change.
func outlineOf(name string, p design.Params) (path outline.Path, advance float64, ok bool) {
	b := built(name, p)
	if !b.ok {
		return nil, 0, false
	}
	return clonePath(b.path), b.advance, true
}

// roundedOf returns the glyph which Set(p) holds under the name, as a copy.
func roundedOf(name string, p design.Params) Glyph {
	b := built(name, p)
	return Glyph{Name: name, Rune: b.rune, Advance: b.advance, Path: clonePath(b.roundedPath())}
}

// setOf is Set with shared builds.
func setOf(p design.Params) []Glyph {
	gs := make([]Glyph, len(table))
	for i, e := range table {
		gs[i] = roundedOf(e.name, p)
	}
	return gs
}

// cachedOutline returns what Outline returns for the glyph in the parameter
// set np.  The path is shared and must not be changed.
func cachedOutline(np namedParams, glyph string) (outline.Path, float64) {
	b := built(glyph, np.params)
	return b.path, b.advance
}

// cachedRounded returns the glyph of Set(np.params) with the given name.  The
// path is shared and must not be changed.
func cachedRounded(np namedParams, glyph string) Glyph {
	b := built(glyph, np.params)
	return Glyph{Name: glyph, Rune: b.rune, Advance: b.advance, Path: b.roundedPath()}
}
