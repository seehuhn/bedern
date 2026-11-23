// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package tex

import (
	"bytes"
	"fmt"
	"unicode"

	"seehuhn.de/go/bedern/internal/tfm"
)

// DesignSize is the design size of the fonts in points.
const DesignSize = 10

// fix returns the fix word of a length of v thousandths of the em.
func fix(v float64) tfm.Fix { return tfm.FixFromFloat(v / 1000) }

// TFM returns the metric file of m.  The design size is 10 pt, and one em is
// one design size.  The face byte names the weight and the slope.
//
// The sixteen font parameters are those of the EC fonts: the slant, the
// interword space, its stretch, its shrink, the x-height, the quad and the
// extra space, which TeX82 defines, and then the height of the capitals, the
// height of the ascenders, the height of the accented capitals (the greatest
// height of a capital beyond ASCII), the depth of the descenders, the greatest
// height and the greatest depth of a character, the width of a figure (that of
// the zero), the stem of the capitals and the distance of baselines.  The
// program of a character has its ligatures before its kerns, and characters
// with the same program share it.  The tables of the heights, the depths and
// the italic corrections hold no more values than the format allows, which
// Reduce brings about.
func (m *Metrics) TFM() ([]byte, error) {
	f := tfm.NewFont()
	f.DesignSize = DesignSize * tfm.Unity
	f.CodingScheme = m.Encoding.Scheme
	f.Family = "BEDERN"
	if m.Bold {
		f.Face += 2
	}
	if m.Italic {
		f.Face++
	}
	f.Params = []tfm.Fix{
		tfm.FixFromFloat(m.Slant), fix(m.Space), fix(m.Stretch), fix(m.Shrink),
		fix(m.XHeight), fix(m.Quad), fix(m.ExtraSpace),
	}
	var maxHeight, maxDepth, accented, digit float64
	for c, ch := range m.Chars {
		if ch == nil {
			continue
		}
		maxHeight = max(maxHeight, ch.Height)
		maxDepth = max(maxDepth, ch.Depth)
		if r := m.Encoding.Slots[c].Rune; r > 0x7F && unicode.IsUpper(r) {
			accented = max(accented, ch.Height)
		}
		if c == '0' {
			digit = ch.Width
		}
	}
	f.Params = append(f.Params, fix(m.CapHeight), fix(m.Ascender), fix(max(accented, m.CapHeight)),
		fix(m.Descender), fix(maxHeight), fix(maxDepth), fix(digit), fix(m.CapStem), fix(m.BaselineSkip))
	progs := make(map[int][]tfm.Step)
	for c, ch := range m.Chars {
		if ch == nil {
			continue
		}
		f.Chars[c] = &tfm.Char{
			Width:  fix(ch.Width),
			Height: fix(ch.Height),
			Depth:  fix(ch.Depth),
			Italic: fix(ch.Italic),
		}
		var steps []tfm.Step
		for _, l := range ch.Ligs {
			steps = append(steps, tfm.Step{Next: uint8(l.Next), Op: tfm.OpLig, Result: uint8(l.Result)})
		}
		for _, k := range ch.Kerns {
			steps = append(steps, tfm.Step{Next: uint8(k.Next), Op: tfm.OpKern, Kern: fix(k.Amount)})
		}
		if len(steps) > 0 {
			progs[c] = steps
		}
	}
	f.LigTable = tfm.Programs(progs)
	var buf bytes.Buffer
	report, err := f.Write(&buf)
	if err != nil {
		return nil, err
	}
	if len(report.Reductions) > 0 {
		r := report.Reductions[0]
		return nil, fmt.Errorf("the table of the %ss has %d values, more than the format holds", r.Table, r.From)
	}
	return buf.Bytes(), nil
}
