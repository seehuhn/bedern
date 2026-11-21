// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package design

import (
	"strings"
	"testing"
)

func TestVariants(t *testing.T) {
	t.Parallel()
	vs := Variants()
	if len(vs) == 0 {
		t.Fatal("no variants")
	}
	if vs[0].Name != "regular" || vs[0].Params != Regular() {
		t.Errorf("first variant is %q, want regular parameters", vs[0].Name)
	}
	seen := map[string]bool{}
	for _, v := range vs {
		if v.Name == "" || v.Name != strings.ToLower(v.Name) || strings.ContainsAny(v.Name, " \t") {
			t.Errorf("bad variant name %q", v.Name)
		}
		if seen[v.Name] {
			t.Errorf("duplicate variant name %q", v.Name)
		}
		seen[v.Name] = true
		p := v.Params
		if err := p.Validate(); err != nil {
			t.Errorf("%s: %v", v.Name, err)
		}
		if v.Note == "" || strings.Contains(v.Note, "\n") {
			t.Errorf("%s: note must be one non-empty line", v.Name)
		}
	}
}
