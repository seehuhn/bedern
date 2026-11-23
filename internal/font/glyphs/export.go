// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"errors"
	"fmt"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// Union returns the outline of the ink which the closed contours of the paths
// cover together, with the orientation of doc.go.  The contours must cross
// where they meet, as for union.
func Union(paths ...outline.Path) outline.Path {
	var parts []chain
	for _, p := range paths {
		parts = append(parts, pathChains(p)...)
	}
	return chainsPath(union(parts...)...)
}

// CheckOutline applies the structural checks of Check to an outline which is
// not a glyph of the table, unrounded and rounded, with the bounding-box rule
// and the sharp crotches of the glyph like.  The glyph like must be in the
// table and must not be a composed letter.
func CheckOutline(p design.Params, like string, path outline.Path, adv float64) error {
	if _, ok := lookup(like); !ok || composites[like] != nil {
		return fmt.Errorf("glyphs: %s is not a glyph of the table which is drawn whole", like)
	}
	if problems := checkStages(p, like, path, adv, path.Round()); len(problems) > 0 {
		return errors.New(problems[0])
	}
	return nil
}
