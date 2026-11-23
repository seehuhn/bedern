// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package build

import (
	"bytes"
	"strings"

	"seehuhn.de/go/sfnt/header"
	"seehuhn.de/go/sfnt/name"
)

// Undated returns the font file data with the date removed from the unique
// identifier (name ID 3) of every name table, which then reads "full name;
// version".  The other tables are copied unchanged.
func Undated(data []byte) ([]byte, error) {
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
			if i := strings.LastIndex(t.Identifier, "; "); i >= 0 {
				t.Identifier = t.Identifier[:i]
			}
		}
	}
	tables["name"] = info.Encode(1)
	var buf bytes.Buffer
	if _, err := header.Write(&buf, h.ScalerType, tables); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
