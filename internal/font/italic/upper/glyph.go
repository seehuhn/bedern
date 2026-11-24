// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package upper

import (
	"fmt"
	"math"
	"runtime"
	"sort"
	"sync"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/outline"
)

// A Glyph is a glyph of the italic: its name, its code point, its advance and
// its outline, rounded to whole units.
type Glyph struct {
	Name    string
	Rune    rune
	Advance float64
	Path    outline.Path
	Raw     outline.Path // the outline before the rounding

	// Map is the affine map (x, y) -> (Map[0] x + Map[1] y + Map[2], y)
	// which takes the upright roman glyph to the italic, apart from the turn
	// of the rounds, the circles and the lattice.
	Map [3]float64
}

// A source says how a glyph of the italic is made from the roman.
type source struct {
	r        rune
	roman    string                                                     // the roman glyph, by default the glyph's name
	rounds   func(Params, design.Params) []round                        // the rounds whose stress turns
	draw     func(Params, design.Params) (outline.Path, float64)        // a drawing of its own instead of the roman's
	upright  bool                                                       // taken unchanged from the roman
	tabular  bool                                                       // a figure, which keeps the advance of the roman
	oldstyle bool                                                       // a tabular glyph which stands in the cell of the old-style figures
	italic   func(Params, design.Params) (outline.Path, float64, error) // a drawing of its own, in the italic
	dots     func(design.Params) []float64                              // the diameters of the dots which stay circles
	zone     float64                                                    // the zone of the balls, see circle; 0 for 1.5
	ball     func(design.Params) float64                                // the diameter of the balls which stay circles
	refit    func(outline.Path) outline.Path                            // the curves which the shear distorts, drawn again after it (see ts1_refit.go)
}

// dotZone is the zone of the dots and the heads of the comma and the quotes,
// see circle.
var dotZone = 1.5

// glyphsOutline is the outline of the roman code.
var glyphsOutline = glyphs.Outline

// sources are the glyphs which the package makes, by name.
var sources = map[string]source{}

// Names returns the names of the glyphs of the package, sorted.
func Names() []string {
	var out []string
	for n := range sources {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

// Build returns the glyph of the given name.
func Build(p Params, name string) (*Glyph, error) {
	src, ok := sources[name]
	if !ok {
		return nil, fmt.Errorf("upper: no glyph %q", name)
	}
	rp := p.Roman()
	if src.italic != nil {
		path, adv, err := src.italic(p, rp)
		if err != nil {
			return nil, err
		}
		return &Glyph{Name: name, Rune: src.r, Advance: adv, Path: path.Round(), Raw: path}, nil
	}
	var path outline.Path
	var adv float64
	if src.draw != nil {
		path, adv = src.draw(p, rp)
	} else {
		rn := src.roman
		if rn == "" {
			rn = name
		}
		if src.upright {
			rp = design.Regular()
		}
		path, adv, ok = glyphsOutline(rn, rp)
		if !ok {
			return nil, fmt.Errorf("upper: no roman glyph %q", rn)
		}
	}
	path = explicit(path)
	if src.upright {
		return &Glyph{Name: name, Rune: src.r, Advance: adv, Path: path.Round(), Raw: path, Map: [3]float64{1, 0, 0}}, nil
	}
	var rs []round
	if src.rounds != nil {
		rs = src.rounds(p, rp)
	}
	var cs []circle
	if src.ball != nil {
		cs = circlesOf(path, src.ball(rp)/2)
		for i := range cs {
			cs[i].Zone = src.zone
			if cs[i].Zone == 0 {
				cs[i].Zone = 1.5
			}
		}
	}
	if src.dots != nil {
		for _, d := range src.dots(rp) {
			for _, c := range circlesOf(path, d/2) {
				c.Zone = dotZone
				if src.zone != 0 {
					c.Zone = src.zone
				}
				cs = append(cs, c)
			}
		}
	}
	raw, adv, m := p.italicise(name, path, adv, rs, cs, rp)
	if src.refit != nil {
		raw = src.refit(raw)
	}
	return &Glyph{Name: name, Rune: src.r, Advance: adv, Path: raw.Round(), Raw: raw, Map: m}, nil
}

// italicise makes the italic glyph from the upright outline of the roman code:
// the stress turned, narrowed by Width with the lean of the letter, the stems
// on the lattice, sheared and moved, and with nodes at the extremes.
func (p Params) italicise(name string, path outline.Path, adv float64, rs []round, cs []circle, rp design.Params) (outline.Path, float64, [3]float64) {
	k := p.slant()
	path = turned(path, rs, p.Turn)
	path = shortened(path, (1-p.ObtuseSerif)*rp.Caps.SerifJut, rp.Caps.SerifThick+2, rp.CapHeight)
	ln := p.Lean[name]
	path = keptCircles(path, cs, p.Width, ln.By+k)
	// The narrowing keeps the side spaces, or, for the figures, the advance
	// and the centre of the cell.
	x0, _, x1, _ := path.BBox()
	o, newAdv := x0, adv-(1-p.Width)*(x1-x0)
	if sources[name].tabular {
		o, newAdv = adv/2, adv
	}
	path = transformed(path, p.Width, ln.By, (1-p.Width)*o-ln.By*ln.At)
	path = onLattice(path, p.SlantDen, []float64{0, rp.CapHeight, rp.Figures.Height, rp.XHeight})
	// The figures are sheared about half the figure height.
	sh := p.shift()
	if sources[name].tabular {
		sh = k * rp.Figures.Height / 2
		if sources[name].oldstyle {
			sh = k * rp.XHeight / 2
		}
	}
	path = transformed(path, 1, k, -sh)
	for i, c := range cs {
		x := p.Width*c.C.X + ln.By*(c.C.Y-ln.At) + (1-p.Width)*o + k*c.C.Y - sh
		cs[i].C = pt(x, c.C.Y)
	}
	path = atExtremes(path, k, p.SlantDen, cs)
	path = circleG2(path, cs)
	m := [3]float64{p.Width, ln.By + k, (1-p.Width)*o - ln.By*ln.At - sh}
	return path, math.Round(newAdv), m
}

// explicit returns the path with every contour closed by an explicit last
// segment which ends at its start.
func explicit(path outline.Path) outline.Path {
	out := make(outline.Path, 0, len(path))
	for _, c := range path {
		if len(c.Segs) == 0 {
			continue
		}
		nc := outline.Contour{Start: c.Start, Segs: append([]outline.Segment(nil), c.Segs...)}
		if last := nc.Segs[len(nc.Segs)-1].P3; dist(last, c.Start) > 1e-9 {
			nc.Segs = append(nc.Segs, outline.Segment{P3: c.Start})
		}
		out = append(out, nc)
	}
	return out
}

// Glyphs returns all glyphs of the package, built side by side, in the
// order of Names.
func Glyphs(p Params) ([]*Glyph, error) {
	names := Names()
	out := make([]*Glyph, len(names))
	errs := make([]error, len(names))
	sem := make(chan struct{}, runtime.GOMAXPROCS(0))
	var wg sync.WaitGroup
	for i, n := range names {
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			out[i], errs[i] = Build(p, n)
			<-sem
		}()
	}
	wg.Wait()
	for _, err := range errs {
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}
