// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

import (
	"fmt"
	"math"
	"sync"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/outline"

	"seehuhn.de/go/bedern/internal/font/italic"
	"seehuhn.de/go/bedern/internal/font/italic/assembly"
)

// The script forms of the figures and of the Greek.  The maths italic has its own forms
// (italic.go) and the double-struck letters theirs (doublestruck.go); the
// signs and the primes have theirs in scriptsigns.go.
//
// The figures and the Greek capitals are drawn by the constructions of the
// regular with a script parameter set of each level, which is derived from
// the parameters of the regular by the factors of the bold (design.BoldFrom)
// with the weights of scriptLevels: the stems and the round strokes heavier,
// the thin strokes with their factor thinParam, the widths by the share out,
// the serifs and the side spaces as they are.  The Greek lowercase is drawn
// by italic.BuildScript with the thin and the stem steps of buildLower, as
// the letters of the maths italic.  Each form then takes the advance of its
// text form times the factor of its level, the white beyond its ink divided
// equally between its sides, within the overhangs.
//
// The forms are derived from the regular parameter set, whatever set the
// font is built with.

// A scriptKind says which glyphs a script parameter set is for: the signs
// take the stems and the share-out of the letters, the figures and the
// capitals their own (scriptLevels).
type scriptKind int

const (
	scriptSigns scriptKind = iota
	scriptFigures
	scriptCapitals
)

// scriptFactors returns the factors of the script parameter set of a level
// (1 or 2) for the glyphs of a kind.
func scriptFactors(level int, kind scriptKind) design.BoldFactors {
	s := scriptLevels[level-1]
	stem, out := s.stem, s.out
	switch kind {
	case scriptFigures:
		stem = s.figStem
	case scriptCapitals:
		stem, out = s.capStem, s.capOut
	}
	f := design.BoldDefault()
	f.Stem, f.CapStem, f.Round, f.Small = stem, stem, stem, stem
	f.Ball, f.Dot, f.Sign = stem, stem, stem
	f.ArchJoin, f.BowlJoinTop, f.BowlJoinBottom = stem, stem, stem
	f.Thin = s.thinParam
	f.Jut, f.Side = 1, 1
	f.Out, f.RoundOut = out, out
	f.EBar, f.Eye, f.Mark, f.TildeWidth, f.TildeHeight = 1, 1, 1, 1, 1
	return f
}

// scriptParams returns the script parameter set of a level (1 or 2) for the
// signs.
func scriptParams(level int) design.Params {
	return design.BoldFrom(scriptFactors(level, scriptSigns))
}

// scriptRomans are the constructions of the regular with the script
// parameter sets of the two levels, for the signs, the figures and the
// capitals.
var scriptRomans = func() (out [3][2]func() *roman) {
	for kind := range out {
		for level := 1; level <= 2; level++ {
			out[kind][level-1] = sync.OnceValue(func() *roman {
				p := design.BoldFrom(scriptFactors(level, scriptKind(kind)))
				// The tips of the serifs have the floor of the thinnest
				// strokes of the level, as the pen of the letters has it.
				p.SerifThick = math.Max(p.SerifThick, scriptLevels[level-1].thin)
				if scriptKind(kind) == scriptFigures {
					p = widenOne(p, scriptLevels[level-1].oneWide)
				}
				return newRoman(p)
			})
		}
	}
	return out
}()

// widenOne returns the parameters with the flag and the foot serifs of the
// figure 1 longer to make its ink wide times that of the regular's: the
// stem and the round strokes of the other figures widen with the share-out of
// the weight, the 1 has no round stroke and keeps nearly its width (1.03
// where the others have 1.08 and 1.14).
func widenOne(p design.Params, wide float64) design.Params {
	ink := func(p design.Params) float64 {
		path, _ := newRoman(p).get("one")
		x0, _, x1, _ := path.BBox()
		return x1 - x0
	}
	if d := wide*ink(design.Regular()) - ink(p); d > 0 {
		p.Figures.OneFoot += math.Round(d / 2)
		p.Figures.OneFlag += math.Round(d / 2)
	}
	return p
}

// A scriptForm is one script form of a glyph as it is built, before it takes
// its advance.
type scriptForm struct {
	Name     string // the name of the text glyph
	Level    int
	Path     outline.Path
	Advance  float64
	Stem     float64  // the factor of the stems of the Greek lowercase, 0 for the others
	Warnings []string // the warnings of the stroke tool
}

// scriptNames returns the names of the figures and of the Greek capitals,
// which are drawn by the constructions of the regular.
func scriptNames() (figures []string, caps []derivation) {
	return romanNames()[:10], greekCapitals
}

// A scriptSet holds the script forms of the glyphs of one kind, by level and
// name.
type scriptSet [2]map[string]scriptForm

// buildScripts builds the forms of both levels by running build for each
// level, in parallel with at most three builds at a time.
func buildScripts(build func(level int, put func(scriptForm), run func(func()), fail func(error))) (scriptSet, error) {
	var out scriptSet
	var mu sync.Mutex
	var firstErr error
	fail := func(err error) {
		mu.Lock()
		if firstErr == nil {
			firstErr = err
		}
		mu.Unlock()
	}
	put := func(f scriptForm) {
		mu.Lock()
		out[f.Level-1][f.Name] = f
		mu.Unlock()
	}
	for i := range out {
		out[i] = map[string]scriptForm{}
	}
	gate := make(chan struct{}, 3)
	var wg sync.WaitGroup
	run := func(f func()) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			gate <- struct{}{}
			defer func() { <-gate }()
			f()
		}()
	}
	for level := 1; level <= 2; level++ {
		build(level, put, run, fail)
	}
	wg.Wait()
	return out, firstErr
}

// uprightScriptForms are the forms of the figures and the Greek capitals,
// drawn by the constructions of the regular with the script parameter sets.
var uprightScriptForms = sync.OnceValues(func() (scriptSet, error) {
	figures, caps := scriptNames()
	return buildScripts(func(level int, put func(scriptForm), run func(func()), fail func(error)) {
		fr, cr := scriptRomans[scriptFigures][level-1](), scriptRomans[scriptCapitals][level-1]()
		for _, n := range figures {
			run(func() {
				path, adv := fr.get(n)
				put(scriptForm{Name: n, Level: level, Path: path, Advance: adv})
			})
		}
		for _, d := range caps {
			run(func() {
				path, adv := d.draw(cr)
				put(scriptForm{Name: d.name, Level: level, Path: path, Advance: adv})
			})
		}
	})
})

// lowerScriptForms are the forms of the Greek lowercase, by name of the
// glyph of the maths font.
var lowerScriptForms = sync.OnceValues(func() (scriptSet, error) {
	return buildScripts(func(level int, put func(scriptForm), run func(func()), fail func(error)) {
		for _, n := range italic.GreekNames() {
			run(func() {
				g, stem, err := buildLower(n, level, true)
				if err != nil {
					fail(err)
					return
				}
				r := mathItalicGreek(g.Rune)
				if r == 0 {
					fail(fmt.Errorf("mathfont: no mathematical italic for %s", n))
					return
				}
				put(scriptForm{Name: fmt.Sprintf("u%X", r), Level: level, Path: g.Path, Advance: g.Advance, Stem: stem, Warnings: g.Warnings})
			})
		}
	})
})

// fitScriptForm returns the form with the advance of the text glyph times
// the factor, the white beyond its ink divided equally between its sides
// besides those of the text glyph, within the overhangs.
func fitScriptForm(f scriptForm, text Glyph, factor float64) (outline.Path, float64) {
	tx0, _, tx1, _ := text.Path.BBox()
	x0, _, x1, _ := f.Path.BBox()
	e := factor*text.Advance - (x1 - x0) - tx0 - (text.Advance - tx1)
	path := shifted(f.Path, tx0+e/2-x0, 0)
	adv := math.Round(factor * text.Advance)
	return withinOverhang(path, adv)
}

// addScripts adds the script forms of the figures, the Greek capitals, the
// signs, the primes and the Greek lowercase to the glyphs and names them in the glyphs which they
// belong to.
// With lower unset the Greek lowercase is left out.
func addScripts(gs []Glyph, lower bool) ([]Glyph, error) {
	forms, err := uprightScriptForms()
	if err != nil {
		return nil, err
	}
	sets := []func() (scriptSet, error){signScriptForms}
	if lower {
		sets = append(sets, lowerScriptForms)
	}
	for _, f := range sets {
		more, err := f()
		if err != nil {
			return nil, err
		}
		for i := range forms {
			m := map[string]scriptForm{}
			for k, v := range forms[i] {
				m[k] = v
			}
			for k, v := range more[i] {
				m[k] = v
			}
			forms[i] = m
		}
	}
	var greek scriptSet // the Greek lowercase, which has its own advance
	if lower {
		if greek, err = lowerScriptForms(); err != nil {
			return nil, err
		}
	}
	n := len(gs)
	for i := range n {
		g := &gs[i]
		if _, ok := forms[0][g.Name]; !ok || g.Rune == 0 {
			continue
		}
		var scripts [2]Glyph
		for level := 1; level <= 2; level++ {
			f := forms[level-1][g.Name]
			factor := scriptLevels[level-1].advance
			if _, ok := greek[level-1][g.Name]; ok {
				factor = scriptLevels[level-1].greekAdvance
			}
			path, adv := fitScriptForm(f, *g, factor)
			s := Glyph{Name: fmt.Sprintf("%s.s%d", g.Name, level), Advance: adv, Path: path, Class: g.Class, Like: g.Like, Mirror: g.Mirror}
			if g.Attached {
				s.Italic = assembly.ItalicCorrection(glyphs.Glyph{Name: g.Name, Advance: adv, Path: path}, icPad)
				s.Attach, s.Attached = accentAttachment(g.Name, path), true
			}
			scripts[level-1] = s
		}
		g.Script = [2]string{scripts[0].Name, scripts[1].Name}
		gs = append(gs, scripts[:]...)
	}
	return gs, nil
}
