// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"strings"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
)

// Check accepts the regular parameters, and turns the panic of a construction
// which fails into an error which names the glyph.
func TestCheck(t *testing.T) {
	t.Parallel()
	if err := Check(design.Regular()); err != nil {
		t.Fatalf("regular parameters: %v", err)
	}
	p := design.Regular()
	p.BallStalk = p.BallDiameter // a stalk as thick as the ball
	err := Check(p)
	if err == nil {
		t.Fatal("Check accepts a stalk as thick as the ball")
	}
	if !strings.Contains(err.Error(), "glyph a:") {
		t.Errorf("the error %q does not name the glyph a", err)
	}
	if err := CheckGlyph(design.Regular(), "no-such-glyph"); err == nil {
		t.Error("CheckGlyph accepts a glyph which does not exist")
	}
}
