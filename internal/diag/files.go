// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package diag

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/glyphs"
)

// FileName maps a glyph name to the stem of the file name of its drawing.
// The result cannot leave the directory, is not hidden, and differs from the
// result for every other name even on a case-insensitive file system:
//
//   - a capital letter is followed by an underscore ("A" becomes "A_"),
//   - one leading dot is dropped (".notdef" becomes "notdef"),
//   - every other character, except the lower-case letters a to z, the
//     digits, "-" and a "." which is not at the start, is written as % and
//     the two lower-case hexadecimal digits of each of its bytes.
//
// Since a literal underscore is written as %5f, the underscore in the result
// always marks a capital.  The names ".x" and "x" both give "x".
func FileName(glyph string) string {
	name := strings.TrimPrefix(glyph, ".")
	var sb strings.Builder
	for i := 0; i < len(name); i++ {
		c := name[i]
		switch {
		case 'a' <= c && c <= 'z', '0' <= c && c <= '9', c == '-':
			sb.WriteByte(c)
		case 'A' <= c && c <= 'Z':
			sb.WriteByte(c)
			sb.WriteByte('_')
		case c == '.' && i > 0:
			sb.WriteByte(c)
		default:
			fmt.Fprintf(&sb, "%%%02x", c)
		}
	}
	if sb.Len() == 0 {
		return "%"
	}
	return sb.String()
}

// GlyphSet returns the glyphs of a parameter set for WriteAll.  Tests replace
// it by a function which shares the builds.
var GlyphSet = glyphs.Set

// WriteAll writes dir/<variant>/<glyph>.svg for every glyph of every
// variant, where the file name is given by FileName.  The drawing of every
// variant except "regular" shows the glyph of the same name from the regular
// variant as an overlay.
func WriteAll(dir string, variants []design.Variant) error {
	sets := make([][]glyphs.Glyph, len(variants))
	var regular []glyphs.Glyph
	for i, v := range variants {
		sets[i] = GlyphSet(v.Params)
		if v.Name == "regular" {
			regular = sets[i]
		}
	}
	for i, v := range variants {
		vdir := filepath.Join(dir, v.Name)
		if err := os.MkdirAll(vdir, 0o777); err != nil {
			return err
		}
		for _, g := range sets[i] {
			var opt Options
			if v.Name != "regular" && regular != nil {
				// The regular glyph of the same name is drawn on top.
				for i := range regular {
					if regular[i].Name == g.Name {
						opt = Options{Overlay: &regular[i], OverlayLabel: "regular"}
					}
				}
			}
			var buf bytes.Buffer
			if err := SVG(&buf, g, v.Params, v.Name, opt); err != nil {
				return err
			}
			path := filepath.Join(vdir, FileName(g.Name)+".svg")
			if err := os.WriteFile(path, buf.Bytes(), 0o666); err != nil {
				return err
			}
		}
	}
	return nil
}
