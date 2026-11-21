// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package design

// A Variant is a named set of parameters, built as a separate font.
type Variant struct {
	Name   string // short lower-case name, used in file names
	Note   string // one line saying what differs from the regular parameters
	Params Params
}

// Variants returns the parameter sets which are built as fonts: the regular
// parameters.
func Variants() []Variant {
	return []Variant{
		{Name: "regular", Note: "the regular parameters", Params: Regular()},
	}
}
