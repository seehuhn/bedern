// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package build

import (
	"bytes"
	"io"
	"maps"
	"strings"

	"seehuhn.de/go/sfnt"
	"seehuhn.de/go/sfnt/header"
	"seehuhn.de/go/sfnt/name"
	"seehuhn.de/go/sfnt/opentype/gtab"
)

// Undated returns the font file data with the date removed from the unique
// identifier (name ID 3) of every name table, which then reads "full name;
// version".  The other tables are copied unchanged.
func Undated(data []byte) ([]byte, error) {
	return editNames(data, func(t *name.Table) {
		if i := strings.LastIndex(t.Identifier, "; "); i >= 0 {
			t.Identifier = t.Identifier[:i]
		}
	})
}

// Write writes the font file of f to w.  If the font has a stylistic set, the
// name tables of the file hold the names of StylisticSetNames.
func Write(w io.Writer, f *sfnt.Font) (int64, error) {
	var buf bytes.Buffer
	if _, err := f.Write(&buf); err != nil {
		return 0, err
	}
	data := buf.Bytes()
	if hasStylisticSet(f) {
		var err error
		data, err = editNames(data, func(t *name.Table) {
			if t.Extra == nil {
				t.Extra = map[name.ID]string{}
			}
			maps.Copy(t.Extra, StylisticSetNames)
		})
		if err != nil {
			return 0, err
		}
	}
	n, err := w.Write(data)
	return int64(n), err
}

// hasStylisticSet says whether the GSUB table of f has a stylistic set with a
// name.
func hasStylisticSet(f *sfnt.Font) bool {
	if f.Gsub == nil {
		return false
	}
	for _, feat := range f.Gsub.FeatureList {
		if _, ok := feat.Params.(*gtab.FeatureParamsStylisticSet); ok {
			return true
		}
	}
	return false
}

// editNames returns the font file data with edit applied to every name
// table.  The other tables are copied unchanged.
func editNames(data []byte, edit func(*name.Table)) ([]byte, error) {
	r := bytes.NewReader(data)
	h, err := header.Read(r)
	if err != nil {
		return nil, err
	}
	tables := make(map[string][]byte, len(h.Toc))
	for tag := range h.Toc {
		b, err := h.ReadTableBytes(r, tag)
		if err != nil {
			return nil, err
		}
		tables[tag] = b
	}
	info, err := name.Decode(tables["name"])
	if err != nil {
		return nil, err
	}
	for _, ts := range []name.Tables{info.Mac, info.Windows} {
		for _, t := range ts {
			edit(t)
		}
	}
	tables["name"] = info.Encode(1)
	var buf bytes.Buffer
	if _, err := header.Write(&buf, h.ScalerType, tables); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
