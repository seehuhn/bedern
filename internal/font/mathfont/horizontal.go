// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

import (
	"fmt"
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/outline"
)

// The horizontal constructions: the wide accents with their wider sizes, the
// bars over and under with their assemblies, the accent arrows with theirs,
// the assemblies of the arrows, the minus and the equals, and the braces,
// brackets and parentheses over and under (spans.go).  A size or a part has
// its ink from the origin to its advance, as the accents of the font have;
// the parts of an assembly are listed from the left, and every connector
// lies on a straight run of its part.
const (
	// hConn is the length of every connector and hExt the length of every
	// extender.  Every assembly has fixed parts at both ends, and twice the
	// connector less MinConnectorOverlap is at least as long as the
	// extender.
	hConn = 330.0
	hExt  = 600.0
	// hEndRun is the straight run at the joining end of an end part of a
	// span, and hRun the straight shaft of a part which carries a head, a
	// bar or a hook, or of a plain end: the connector and a margin.
	hEndRun = hConn
	hRun    = hConn + 20
	// barWidth is the ink width of U+0305 and U+0332.
	barWidth = 400.0
	// vecBothWidth is the ink width of U+20E1.
	vecBothWidth = 500.0
)

// wideWidths are the ink widths of the wider sizes of the wide accents: the
// first 1.76 times the circumflex and each after it about 1.36 times the one
// before, the largest 2400.
var wideWidths = []float64{520, 700, 950, 1300, 1750, 2400}

// A wideAccent is an accent with wider sizes: the construction of a
// combining mark of the regular run at the widths of wideWidths, its stroke
// and its height grown with the logarithm of the width to thick and height
// times those of the base at the largest size.
type wideAccent struct {
	name          string // the glyph of the font
	r             rune   // its code point if the glyph is new, else 0
	mark          string // the combining mark of the regular whose construction it runs
	below         bool
	thick, height float64
}

// wideAccents are the wide accents.  The stroke of the circumflex and the
// caron is 1.05 times the base at the largest size, the height of the
// circumflex 1.44 and that of the caron 1.56; the stroke of the tilde is 1.20
// and its height 1.52; the stroke of the breve is 1.16 and its height 1.24.
var wideAccents = []wideAccent{
	{"uni0302", 0, "uni0302", false, 1.05, 1.44},
	{"uni030C", 0, "uni030C", false, 1.05, 1.56},
	{"tildecomb", 0, "tildecomb", false, 1.20, 1.52},
	{"uni0306", 0, "uni0306", false, 1.16, 1.24},
	{"uni0330", 0x0330, "tildecomb", true, 1.20, 1.52},
}

// markWidth returns the ink width of the mark in the parameters.
func markWidth(a design.Accents, mark string) float64 {
	switch mark {
	case "uni0302", "uni030C":
		return a.CircWidth
	case "tildecomb":
		return a.TildeWidth
	case "uni0306":
		return a.BreveWidth
	}
	panic("mathfont: no wide form of " + mark)
}

// The wider tildes, whose waves are flatter than that of the text, have their
// crests wideTildeCrest of the width from the ends of the centre line and
// fall wideTildeSteep times as steeply as the chord from the crest to the
// trough (design.Accents).
const (
	wideTildeCrest = 0.275
	wideTildeSteep = 2.0
)

// wideParams returns the parameters with the mark w wide, its strokes grown
// by tf and its height by hf; a tilde wider than the text's takes the crest
// and the fall of the wider tildes.
func wideParams(p design.Params, mark string, w, tf, hf float64) design.Params {
	q := p
	a := &q.Acc
	switch mark {
	case "uni0302", "uni030C":
		a.CircWidth, a.CircHeight, a.CircThin, a.CircThick = w, a.CircHeight*hf, a.CircThin*tf, a.CircThick*tf
	case "tildecomb":
		if w > a.TildeWidth {
			a.TildeCrest, a.TildeSteep = wideTildeCrest, wideTildeSteep
		}
		a.TildeWidth, a.TildeHeight, a.TildeThick, a.TildeThin = w, a.TildeHeight*hf, a.TildeThick*tf, a.TildeThin*tf
	case "uni0306":
		a.BreveWidth, a.BreveHeight, a.BreveThick, a.BreveThin = w, a.BreveHeight*hf, a.BreveThick*tf, a.BreveThin*tf
	}
	return q
}

// accentBase returns AccentBaseHeight: LuaTeX raises an accent by the
// height of its base above it.
func accentBase(p design.Params) float64 { return float64(Constants(p).AccentBaseHeight.V) }

// accentGap returns the white between AccentBaseHeight and the ink of the
// combining marks above it, which LuaTeX leaves between a base and its
// accent.
func accentGap(r *roman) float64 {
	path, _ := r.get("uni0302")
	_, y0, _, _ := path.BBox()
	return math.Round(y0 - accentBase(r.p))
}

// below moves an accent to hang below the baseline with the white gap: its
// top at -gap.
func below(path outline.Path, gap float64) outline.Path {
	_, _, _, y1 := path.BBox()
	return shifted(path, 0, math.Round(-gap-y1))
}

// form returns the size of the wide accent w wide: the base for the width
// of the mark of the regular.
func (a wideAccent) form(r *roman, w float64) outline.Path {
	p := r.p
	w0 := markWidth(p.Acc, a.mark)
	k := math.Log(w/w0) / math.Log(wideWidths[len(wideWidths)-1]/w0)
	q := wideParams(p, a.mark, w, math.Pow(a.thick, k), math.Pow(a.height, k))
	var path outline.Path
	if (a.mark == "uni0302" || a.mark == "uni030C") && w > w0 {
		// The mark of the regular cuts its feet level, which leaves a
		// sliver where a flat stroke meets the cut; the wider sizes cut
		// them across the stroke.
		base, _ := r.get(a.mark)
		_, y0, _, _ := base.BBox()
		path = shifted(wideCircumflex(w, q.Acc.CircHeight, q.Acc.CircThin, q.Acc.CircThick, a.mark == "uni030C"), 0, y0)
	} else {
		var ok bool
		path, _, ok = glyphs.Outline(a.mark, q)
		if !ok {
			panic("mathfont: the regular has no glyph " + a.mark)
		}
	}
	if a.below {
		path = below(path, accentGap(r))
	}
	return path
}

// wideCircumflex returns the circumflex, or the caron if caron is set, w
// wide and h high, standing on the height 0 and centred on the abscissa 0:
// two straight strokes, thin on the left of the circumflex and thick on its
// right, whose outer edges meet in the apex, whose inner edges meet in the
// crotch and whose feet are cut across the stroke through the inner corner,
// which stands on the height 0.  The caron is the circumflex turned through
// half a turn about the middle of its box.
func wideCircumflex(w, h, thin, thick float64, caron bool) outline.Path {
	// The stroke t thick whose outer edge leaves the apex at the angle th
	// The stroke t thick whose outer edge leaves the apex at the angle th
	// to the level reaches the abscissa -(h - t cos th) / tan th with its
	// outer foot; th solves this = -w/2.
	angle := func(t float64) float64 {
		lo, hi := 0.001, math.Pi/2-0.001
		for range 60 {
			th := (lo + hi) / 2
			if (h-t*math.Cos(th))/math.Tan(th) > w/2 {
				lo = th
			} else {
				hi = th
			}
		}
		return (lo + hi) / 2
	}
	// foot returns the inner and the outer corner of the foot of the left
	// stroke.
	foot := func(th, t float64) (in, out outline.Point) {
		s := (h - t*math.Cos(th)) / math.Sin(th)
		in = pt(-s*math.Cos(th)+t*math.Sin(th), 0)
		out = pt(in.X-t*math.Sin(th), t*math.Cos(th))
		return in, out
	}
	tl, tr := angle(thin), angle(thick)
	li, lo := foot(tl, thin)
	ri, ro := foot(tr, thick)
	ri.X, ro.X = -ri.X, -ro.X
	// The crotch, where the inner edges meet: li + a (cos tl, sin tl) =
	// ri + b (-cos tr, sin tr).
	a := (ri.X - li.X) / (math.Cos(tl) + math.Sin(tl)*math.Cos(tr)/math.Sin(tr))
	crotch := pt(li.X+a*math.Cos(tl), li.Y+a*math.Sin(tl))
	pts := []outline.Point{lo, li, crotch, ri, ro, pt(0, h)}
	if caron {
		for i, q := range pts {
			pts[i] = pt(-q.X, h-q.Y)
		}
	}
	return outline.Path{polygon(pts...)}
}

// A construction is the horizontal construction of a glyph.
type construction struct {
	Variants []string
	Parts    []Part
}

// horizontalGlyphs returns the new glyphs of the horizontal constructions,
// the wider sizes and the parts, and the construction of each glyph which
// has one.
func horizontalGlyphs(p design.Params) ([]Glyph, map[string]construction, error) {
	r := newRoman(p)
	var gs []Glyph
	cons := map[string]construction{}
	accent := func(name string, rn rune, path outline.Path) Glyph {
		g := Glyph{Name: name, Rune: rn, Class: Accent, Like: "acutecomb"}
		g.Path, g.Advance = withSides(path.Round(), 0, 0)
		return g
	}
	for _, a := range wideAccents {
		if a.r != 0 {
			gs = append(gs, accent(a.name, a.r, a.form(r, markWidth(p.Acc, a.mark))))
		}
		var c construction
		for k, w := range wideWidths {
			n := fmt.Sprintf("%s.h%d", a.name, k+1)
			gs = append(gs, accent(n, 0, a.form(r, w)))
			c.Variants = append(c.Variants, n)
		}
		cons[a.name] = c
	}
	bg, bc := barGlyphs(r)
	gs = append(gs, bg...)
	for n, c := range bc {
		cons[n] = c
	}
	vg, vc := vecGlyphs(r)
	gs = append(gs, vg...)
	for n, c := range vc {
		cons[n] = c
	}
	ag, ac := arrowAssemblies(r)
	gs = append(gs, ag...)
	for n, c := range ac {
		cons[n] = c
	}
	sg, sc, err := spanGlyphs(p)
	if err != nil {
		return nil, nil, err
	}
	gs = append(gs, sg...)
	for n, c := range sc {
		cons[n] = c
	}
	return gs, cons, nil
}

// barAssembly returns the glyphs and the parts of a straight assembly whose
// ink, drawn by draw from 0 to a length, is the same at every abscissa: an
// end hRun long at both ends and the extender between.
func barAssembly(name string, class Class, like string, draw func(l float64) outline.Path) ([]Glyph, []Part) {
	end, ext := name+".end", name+".ext"
	gs := []Glyph{
		{Name: end, Advance: hRun, Path: draw(hRun), Class: class, Like: like},
		{Name: ext, Advance: hExt, Path: draw(hExt), Class: class, Like: like},
	}
	ps := []Part{
		{Name: end, EndConn: hConn, FullAdvance: hRun},
		{Name: ext, StartConn: hConn, EndConn: hConn, FullAdvance: hExt, Extender: true},
		{Name: end, StartConn: hConn, FullAdvance: hRun},
	}
	return gs, ps
}

// barGlyphs returns U+0305 and U+0332, the bars of \overline and \underline
// barGlyphs returns U+0305 and U+0332, the bars of \overline and \underline
// as accents: a bar of the rule barWidth wide, U+0305 OverbarVerticalGap
// above AccentBaseHeight and U+0332 UnderbarVerticalGap below the baseline,
// with their assemblies.
func barGlyphs(r *roman) ([]Glyph, map[string]construction) {
	p := r.p
	c := Constants(p)
	t := float64(c.OverbarRuleThickness.V)
	over := accentBase(p) + float64(c.OverbarVerticalGap.V)
	under := -float64(c.UnderbarVerticalGap.V)
	var gs []Glyph
	cons := map[string]construction{}
	for _, b := range []struct {
		name   string
		r      rune
		y0, y1 float64
	}{{"uni0305", 0x0305, over, over + t}, {"uni0332", 0x0332, under - t, under}} {
		draw := func(l float64) outline.Path { return outline.Path{box(0, b.y0, l, b.y1)} }
		gs = append(gs, Glyph{Name: b.name, Rune: b.r, Advance: barWidth, Path: draw(barWidth), Class: Accent, Like: "acutecomb"})
		pg, ps := barAssembly(b.name, Accent, "acutecomb", draw)
		gs = append(gs, pg...)
		cons[b.name] = construction{Parts: ps}
	}
	return gs, cons
}

// A shaft is a run of the shaft of an arrow across y: from y0 to y1.
type shaft [2]float64

// headPart returns the part of an assembly which carries the head ch at its
// right end, on the shafts: its tip at the advance and the shafts hRun long
// up to the barbs.
func headPart(ch chevron, shafts []shaft, upper [2]float64) (outline.Path, float64) {
	x0, _, _, _ := ch.path().BBox()
	w := math.Round(hRun - x0)
	parts := []outline.Path{shifted(ch.path(), w, 0)}
	end := w + ch.shaftEnd(upper[0], upper[1])
	for _, s := range shafts {
		parts = append(parts, outline.Path{box(0, s[0], end, s[1])})
	}
	return glyphs.Union(parts...), w
}

// shaftsPath returns the shafts l long from the origin.
func shaftsPath(shafts []shaft, l float64) outline.Path {
	var path outline.Path
	for _, s := range shafts {
		path = append(path, box(0, s[0], l, s[1]))
	}
	return path
}

// mirroredPart returns the part mirrored in the vertical through its middle.
func mirroredPart(path outline.Path, w float64) outline.Path {
	return mirroredX(path, w/2)
}

// vecGlyphs returns the accent arrows U+20D6, U+20E1, U+20EE and U+20EF,
// drawn as U+20D7 is (arrows.go), and the assemblies of these and of U+20D7:
// the tail, a shaft hRun long, the extender and the part with the head, or
// the head mirrored for its tail.  The assembly with no extender is wider
// than the base, and each has one wider size, that assembly drawn whole.
// LuaTeX takes the assembly only for an argument at least as wide
// as it.  U+20D6 is U+20D7 mirrored, U+20E1 has a head at each end and is
// vecBothWidth wide, and U+20EE and U+20EF are U+20D6 and U+20D7 hung below
// the baseline with the white of the accents above their base (accentGap).
func vecGlyphs(r *roman) ([]Glyph, map[string]construction) {
	ch, t, y := vecFrame(r)
	gap := accentGap(r)
	shafts := []shaft{{-t / 2, t / 2}}
	head, hw := headPart(ch, shafts, [2]float64{0, t / 2})
	// both returns the arrow with a head at each end, w wide.
	both := func(w float64) outline.Path {
		tip := shifted(ch.path(), w, 0)
		se := ch.shaftEnd(0, t/2) // where the shaft ends inside the head, from the tip
		return glyphs.Union(tip, mirroredX(tip, w/2), outline.Path{box(-se, -t/2, w+se, t/2)})
	}
	base := vecArrow(ch, t, vecWidth)
	single, double := hRun+hw-hConn, 2*hw-hConn // the widths of the assemblies with no extender
	var gs []Glyph
	cons := map[string]construction{}
	accent := func(name string, rn rune, path outline.Path) {
		g := Glyph{Name: name, Rune: rn, Class: Accent, Like: "acutecomb"}
		g.Path, g.Advance = withSides(path.Round(), 0, 0)
		gs = append(gs, g)
	}
	part := func(name string, path outline.Path, w float64) {
		gs = append(gs, Glyph{Name: name, Advance: w, Path: path, Class: Accent, Like: "acutecomb"})
	}
	for _, v := range []struct {
		under         bool
		rights, lefts string // U+20D7 and U+20D6 or their forms below
		r, l, b       rune
	}{{false, "uni20D7", "uni20D6", 0, 0x20D6, 0x20E1}, {true, "uni20EF", "uni20EE", 0x20EF, 0x20EE, 0}} {
		place := func(path outline.Path) outline.Path {
			path = shifted(path, 0, y)
			if v.under {
				// The base, hung below.
				_, _, _, top := shifted(base, 0, y).BBox()
				path = shifted(path, 0, math.Round(-gap-top))
			}
			return path.Round()
		}
		if v.r != 0 {
			accent(v.rights, v.r, place(base))
		}
		accent(v.lefts, v.l, place(mirroredX(base, 0)))
		end, ext, right, left := v.rights+".end", v.rights+".ext", v.rights+".right", v.lefts+".left"
		accent(v.rights+".h1", 0, place(vecArrow(ch, t, single)))
		accent(v.lefts+".h1", 0, place(mirroredX(vecArrow(ch, t, single), 0)))
		part(end, place(shaftsPath(shafts, hRun)), hRun)
		part(ext, place(shaftsPath(shafts, hExt)), hExt)
		part(right, place(head), hw)
		part(left, place(mirroredPart(head, hw)), hw)
		extPart := Part{Name: ext, StartConn: hConn, EndConn: hConn, FullAdvance: hExt, Extender: true}
		cons[v.rights] = construction{Variants: []string{v.rights + ".h1"}, Parts: []Part{
			{Name: end, EndConn: hConn, FullAdvance: hRun}, extPart, {Name: right, StartConn: hConn, FullAdvance: hw}}}
		cons[v.lefts] = construction{Variants: []string{v.lefts + ".h1"}, Parts: []Part{
			{Name: left, EndConn: hConn, FullAdvance: hw}, extPart, {Name: end, StartConn: hConn, FullAdvance: hRun}}}
		if v.b != 0 {
			accent("uni20E1", v.b, place(both(vecBothWidth)))
			accent("uni20E1.h1", 0, place(both(double)))
			cons["uni20E1"] = construction{Variants: []string{"uni20E1.h1"}, Parts: []Part{
				{Name: left, EndConn: hConn, FullAdvance: hw}, extPart, {Name: right, StartConn: hConn, FullAdvance: hw}}}
		}
	}
	return gs, cons
}

// arrowAssemblies returns the parts of the assemblies of the arrows, the
// minus and the equals, with their shafts or bars where the glyph has them:
// for each arrow its tail, the extender and its head, where the tail of →
// and ⇒ is the end of the minus or the equals, that of ↔ and ⇔ the head
// mirrored, the tail of ↦ its bar with the shaft hRun long and that of ↪
// its hook with the shaft hRun long beyond it; for the minus and the equals
// an end hRun long at either side of the extender.
func arrowAssemblies(r *roman) ([]Glyph, map[string]construction) {
	p := r.p
	a := newArrowParts(r)
	axis := p.Punct.MathAxis
	t := a.t
	single := []shaft{{-t / 2, t / 2}}
	double := []shaft{{-a.gap/2 - t, -a.gap / 2}, {a.gap / 2, a.gap/2 + t}}
	on := func(path outline.Path) outline.Path { return shifted(path, 0, axis).Round() }
	var gs []Glyph
	add := func(name string, path outline.Path, w float64) {
		gs = append(gs, Glyph{Name: name, Advance: w, Path: on(path), Class: Rel, Like: "acutecomb"})
	}
	cons := map[string]construction{}
	sHead, sw := headPart(a.single, single, [2]float64{0, t / 2})
	dHead, dw := headPart(a.double, double, [2]float64{a.gap / 2, a.gap/2 + t})
	add("minus.ext", shaftsPath(single, hExt), hExt)
	add("minus.end", shaftsPath(single, hRun), hRun)
	add("equal.ext", shaftsPath(double, hExt), hExt)
	add("equal.end", shaftsPath(double, hRun), hRun)
	add("uni2192.right", sHead, sw)
	add("uni2190.left", mirroredPart(sHead, sw), sw)
	add("uni21D2.right", dHead, dw)
	add("uni21D0.left", mirroredPart(dHead, dw), dw)
	// The tail of ↦: its bar, as high as the head, and the shaft.
	h := a.single.h
	bw := math.Round(t + hRun)
	add("uni21A6.left", glyphs.Union(outline.Path{box(0, -h, t, h)}, shifted(shaftsPath(single, bw-t/2), t/2, 0)), bw)
	// The tail of ↪: its hook, drawn as hookArrow draws it, and the shaft.
	rad := (hookTop - t/2) / 2
	x := t/2 + hookStretch*rad
	hkw := math.Round(x + hookEnd*rad + hRun)
	hook := glyphs.Stroke([]glyphs.Station{
		{X: hkw, Y: 0, Dir: 180, W: t, Flat: true},
		{X: x, Y: 0, Dir: 180, W: t},
		{X: t / 2, Y: rad, Dir: 90, W: t},
		{X: x, Y: 2 * rad, Dir: 0, W: t, Flat: true},
		{X: x + hookEnd*rad, Y: 2 * rad, Dir: 0, W: t},
	})
	add("uni21AA.left", hook, hkw)

	ext := func(n string) Part {
		return Part{Name: n, StartConn: hConn, EndConn: hConn, FullAdvance: hExt, Extender: true}
	}
	first := func(n string, w float64) Part { return Part{Name: n, EndConn: hConn, FullAdvance: w} }
	last := func(n string, w float64) Part { return Part{Name: n, StartConn: hConn, FullAdvance: w} }
	cons["uni2192"] = construction{Parts: []Part{first("minus.end", hRun), ext("minus.ext"), last("uni2192.right", sw)}}
	cons["uni2190"] = construction{Parts: []Part{first("uni2190.left", sw), ext("minus.ext"), last("minus.end", hRun)}}
	cons["uni2194"] = construction{Parts: []Part{first("uni2190.left", sw), ext("minus.ext"), last("uni2192.right", sw)}}
	cons["uni21D2"] = construction{Parts: []Part{first("equal.end", hRun), ext("equal.ext"), last("uni21D2.right", dw)}}
	cons["uni21D0"] = construction{Parts: []Part{first("uni21D0.left", dw), ext("equal.ext"), last("equal.end", hRun)}}
	cons["uni21D4"] = construction{Parts: []Part{first("uni21D0.left", dw), ext("equal.ext"), last("uni21D2.right", dw)}}
	cons["uni21A6"] = construction{Parts: []Part{first("uni21A6.left", bw), ext("minus.ext"), last("uni2192.right", sw)}}
	cons["uni21AA"] = construction{Parts: []Part{first("uni21AA.left", hkw), ext("minus.ext"), last("uni2192.right", sw)}}
	for _, n := range []string{"minus", "equal"} {
		cons[n] = construction{Parts: []Part{first(n+".end", hRun), ext(n + ".ext"), last(n+".end", hRun)}}
	}
	return gs, cons
}
