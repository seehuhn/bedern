// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package design

import (
	"errors"
	"fmt"
	"strings"
)

// Style identifies one font of a family.
//
// The full name of a font is the family name followed by the subfamily name,
// also for the regular style ("Bedern Regular").  The test variants use a
// free-text subfamily together with the regular flags.
type Style struct {
	Family     string // family name, for example "Bedern"
	Subfamily  string // style name, for example "Regular"
	PostScript string // PostScript name, for example "Bedern-Regular"
	Weight     int    // OS/2 weight class, 400 for regular
	Italic     bool
	Bold       bool
}

// RegularStyle returns the style of Bedern Regular.
func RegularStyle() Style {
	return Style{
		Family:     "Bedern",
		Subfamily:  "Regular",
		PostScript: "Bedern-Regular",
		Weight:     400,
	}
}

// Validate reports whether the style can be used for a font.
func (s Style) Validate() error {
	if s.Family == "" {
		return errors.New("style: empty family name")
	}
	if s.Subfamily == "" {
		return errors.New("style: empty subfamily name")
	}
	if s.PostScript == "" {
		return errors.New("style: empty PostScript name")
	}
	if len(s.PostScript) > 63 {
		return fmt.Errorf("style: PostScript name %q is longer than 63 characters", s.PostScript)
	}
	for _, c := range []byte(s.PostScript) {
		if c <= ' ' || c >= 0x7f || strings.IndexByte("[](){}<>/%", c) >= 0 {
			return fmt.Errorf("style: PostScript name %q contains the invalid character %q", s.PostScript, c)
		}
	}
	if s.Weight < 1 || s.Weight > 1000 {
		return fmt.Errorf("style: weight %d is outside 1..1000", s.Weight)
	}
	if s.Bold && s.Weight < 600 {
		return fmt.Errorf("style: bold style with weight %d below 600", s.Weight)
	}
	if !s.Bold && s.Weight >= 600 {
		return fmt.Errorf("style: weight %d needs the bold flag", s.Weight)
	}
	return nil
}
