// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"sync"

	"seehuhn.de/go/bedern/internal/outline"
)

// The searches for a spiral (searchWindSpiral, searchRiseSpiral) depend on
// their arguments alone.  newWindSpiral and newRiseSpiral remember the result
// for each list of arguments, across parameter sets.  The cache is safe for
// parallel use.

// A spiralKey is the list of arguments of a search; the last two fields hold
// the end of the spiral, and rise tells the searches apart.
type spiralKey struct {
	p0       outline.Point
	phi0, k0 float64
	end      outline.Point
	last     float64
	rise     bool
}

// A spiralResult is what a search returned.
type spiralResult struct {
	w   windSpiral
	err error
}

var spiralCache sync.Map // spiralKey to spiralResult

// cachedSpiral returns the result of search for the key, calling it once for
// each key.
func cachedSpiral(k spiralKey, search func() (windSpiral, error)) (windSpiral, error) {
	if v, ok := spiralCache.Load(k); ok {
		r := v.(spiralResult)
		return r.w, r.err
	}
	w, err := search()
	spiralCache.Store(k, spiralResult{w, err})
	return w, err
}

// newWindSpiral is searchWindSpiral with remembered results.
func newWindSpiral(p0 outline.Point, phi0, k0 float64, centre outline.Point, r float64) (windSpiral, error) {
	return cachedSpiral(spiralKey{p0, phi0, k0, centre, r, false}, func() (windSpiral, error) {
		return searchWindSpiral(p0, phi0, k0, centre, r)
	})
}

// newRiseSpiral is searchRiseSpiral with remembered results.
func newRiseSpiral(p0 outline.Point, phi0, k0 float64, p3 outline.Point, phi3 float64) (windSpiral, error) {
	return cachedSpiral(spiralKey{p0, phi0, k0, p3, phi3, true}, func() (windSpiral, error) {
		return searchRiseSpiral(p0, phi0, k0, p3, phi3)
	})
}
