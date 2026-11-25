// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/outline"
)

// The relations and binary operators.
// ∈ ⊂ ∪ ∨ ∼ ≺ ⊕ ∝ and ⋆ are new drawings with the rule of the signs as
// their stroke, monoline as the signs of the regular are; the others are
// made from them and from the signs of the regular by mirroring, by adding
// a bar of the rule, a wave or the solidus of the negations, or by running
// the construction with another size and the same stroke.  ∩ and ∧ are ∪ and
// ∨ mirrored in the axis, not turned.
//
// Side spaces: a sign whose ink is at least sideFloor wide and no wider than
// the plus is centred in the advance of the signs (MathAdvance), which leaves
// at most 80 units on each side; a wider or a narrower one has MathSide on
// each side.
const (
	// sideFloor is the narrowest ink centred in MathAdvance.
	sideFloor = 510
	// setHeight is the ink height of ⊂, ⊃ and ∈.
	setHeight = 540.0
	// cupWidth and cupHeight are the ink of ∪, ∩, ∨ and ∧.
	cupWidth  = 540.0
	cupHeight = 580.0
	// simShare is the ink height of ∼ over its width; its width is that of
	// the plus.
	simShare = 0.34
	// precBend is the turn of the edges of the strokes of ≺ from the vertex
	// to the ends, in degrees, as the barbs of the arrows turn.
	precBend = 50
	// precPoint is the flat at the vertex of ≺, half of that of < (36) and
	// the least which the clearance check allows.
	precPoint = 20
	// precWidth is the ink width of ≺ over that of <.
	precWidth = 1.06
	// oplusShare is the diameter of ⊕, ⊗ and ⊙ over the width of the plus.
	oplusShare = 1.15
	// circShare is the diameter of ∘ over the width of the plus.
	circShare = 0.54
	// starShare is the width of ⋆ over that of the plus; starInner is the
	// radius of its notches over that of its points.
	starShare = 0.72
	starInner = 0.45
	// proptoWidth and proptoHeight are the ink of ∝ over the width of the
	// plus.
	proptoWidth  = 1.10
	proptoHeight = 0.67
	proptoLoop   = 0.30
	// negOver is how far the solidus of the negations reaches beyond the
	// ink of the sign it crosses, above and below; it gives ≠ 524, ≁ 448,
	// ∉ 820, ≰ 906, ⊈ 960, ≢ 666 and ≮ 796.
	negOver = 140.0
	// subsetGap is the white between ⊂ and the bar of ⊆.
	subsetGap = 90.0
)

// newSigns are the new relations and binary operators.
var newSigns = append(append([]derivation{}, arrowDerivations...), relationDerivations...)

func init() {
	derivations = append(derivations, newSigns...)
}

// signSides returns the path placed in its advance by the rule of the side
// spaces above.
func signSides(r *roman, path outline.Path) (outline.Path, float64) {
	p := r.p
	x0, _, x1, _ := path.BBox()
	if w := x1 - x0; w >= sideFloor && w <= p.Punct.MathWidth+0.5 {
		return centredIn(path, p.MathAdvance()), p.MathAdvance()
	}
	return withSides(path, p.Punct.MathSide, p.Punct.MathSide)
}

// onAxisPath returns the path moved up to centre its ink on the axis.
func onAxisPath(r *roman, path outline.Path) outline.Path {
	_, y0, _, y1 := path.BBox()
	return shifted(path, 0, math.Round(r.p.Punct.MathAxis-(y0+y1)/2))
}

// sign wraps a construction which draws ink anywhere into a glyph centred on
// the axis with the side spaces of the rule.
func sign(draw func(r *roman) outline.Path) func(r *roman) (outline.Path, float64) {
	return func(r *roman) (outline.Path, float64) {
		return signSides(r, onAxisPath(r, draw(r)))
	}
}

// mirroredSign returns the sign drawn by draw, mirrored in a vertical line.
func mirroredSign(draw func(r *roman) outline.Path) func(r *roman) (outline.Path, float64) {
	return sign(func(r *roman) outline.Path { return mirroredX(draw(r), 0) })
}

// mirroredDraw returns the construction draw mirrored in a vertical line, for a
// negation whose solidus keeps its lean.
func mirroredDraw(draw func(r *roman) outline.Path) func(r *roman) outline.Path {
	return func(r *roman) outline.Path { return mirroredX(draw(r), 0) }
}

// turnedDown returns the sign drawn by draw, mirrored in a horizontal line.
func turnedDown(draw func(r *roman) outline.Path) func(r *roman) (outline.Path, float64) {
	return sign(func(r *roman) outline.Path { return mirroredY(draw(r), 0) })
}

// subsetOf draws ⊂ of the given ink height: a stroke of the rule which runs
// from the end of the upper arm to the left, round through half a turn and
// back to the end of the lower arm, the arms straight and cut at right angles,
// the ink the width of the plus, centred on y = 0.
func subsetOf(r *roman, height, width float64) outline.Path {
	t := r.p.Punct.MathThick
	a := height/2 - t/2 // the centre line of the arms
	x := t/2 + a        // where the round leaves the arms
	return glyphs.Stroke([]glyphs.Station{
		{X: width, Y: a, Dir: 180, W: t, Flat: true},
		{X: x, Y: a, Dir: 180, W: t},
		{X: t / 2, Y: 0, Dir: 270, W: t},
		{X: x, Y: -a, Dir: 0, W: t, Flat: true},
		{X: width, Y: -a, Dir: 0, W: t},
	})
}

func subset(r *roman) outline.Path { return subsetOf(r, setHeight, r.p.Punct.MathWidth) }

// element draws ∈: the round of ⊂ as wide as < and a bar of the rule on its
// middle from the round to the ends of the arms.
func element(r *roman) outline.Path {
	t, w := r.p.Punct.MathThick, r.p.Punct.MathAngleWidth
	bar := outline.Path{box(t/2, -t/2, w, t/2)}
	return glyphs.Union(subsetOf(r, setHeight, w), bar)
}

// subsetEq draws ⊆: ⊂ over a bar of the rule as wide as its ink, subsetGap
// below it.
func subsetEq(r *roman) outline.Path {
	t, w := r.p.Punct.MathThick, r.p.Punct.MathWidth
	y := -setHeight/2 - subsetGap
	return join(subset(r), outline.Path{box(0, y-t, w, y)})
}

// negation returns the solidus of the negations across the ink of the path,
// centred on the middle of its ink: a parallelogram of the rule across,
// leaning by the angle of the slash, negOver higher than the ink above and
// below.
func negationOver(r *roman, path outline.Path) outline.Path {
	x0, y0, x1, y1 := path.BBox()
	return solidus(r, (x0+x1)/2, (y0+y1)/2, y1-y0+2*negOver)
}

// solidus returns a parallelogram of the rule across, leaning by the angle of
// the slash, h high and centred on (x, y).
func solidus(r *roman, x, y, h float64) outline.Path {
	p := r.p
	a := p.Punct.SlashAngle * math.Pi / 180
	w := p.Punct.MathThick / math.Cos(a)
	run := h * math.Tan(a)
	y0, y1 := y-h/2, y+h/2
	return outline.Path{polygon(
		pt(x-run/2-w/2, y0), pt(x-run/2+w/2, y0), pt(x+run/2+w/2, y1), pt(x+run/2-w/2, y1))}
}

// negated returns the sign drawn by draw crossed by the solidus of the
// negations.  Where an edge of the solidus cuts an edge of the sign less
// than minSeg from a node, the solidus is moved sideways by the least number
// of units, at most 20, which clears it.
func negated(draw func(r *roman) outline.Path) func(r *roman) outline.Path {
	return func(r *roman) outline.Path {
		path := draw(r)
		neg := negationOver(r, path)
		var first outline.Path
		for i := 0; i <= 40; i++ {
			dx := float64((i + 1) / 2)
			if i%2 == 1 {
				dx = -dx
			}
			u := glyphs.Union(path, shifted(neg, dx, 0))
			if i == 0 {
				first = u
			}
			if !hasShortSeg(u.Round()) {
				return u
			}
		}
		return first
	}
}

// minSeg is the shortest segment which the fairness analysis accepts
// (fair.Default.ShortSeg is 15), with a unit to spare.
const minSeg = 16

// hasShortSeg reports whether the path has a segment whose ends are less
// than minSeg apart.
func hasShortSeg(p outline.Path) bool {
	for _, c := range p {
		from := c.Start
		for _, s := range c.Segs {
			if math.Hypot(s.P3.X-from.X, s.P3.Y-from.Y) < minSeg {
				return true
			}
			from = s.P3
		}
	}
	return false
}

// subsetNeq draws ⊊: ⊆ with a short solidus across its bar, at the middle
// of the bar, reaching 0.75 of subsetGap above and below the bar.
func subsetNeq(r *roman) outline.Path {
	t, w := r.p.Punct.MathThick, r.p.Punct.MathWidth
	y := -setHeight/2 - subsetGap - t/2
	return glyphs.Union(subsetEq(r), solidus(r, w/2, y, t+1.5*subsetGap))
}

// subsetNeqSign places ⊊ where ⊆ stands, its solidus hanging below.
func subsetNeqSign(r *roman) (outline.Path, float64) {
	_, y0, _, y1 := subsetEq(r).BBox()
	return signSides(r, shifted(subsetNeq(r), 0, math.Round(r.p.Punct.MathAxis-(y0+y1)/2)))
}

// cup draws ∪: a stroke of the rule down the left stem, round through half a
// turn and up the right stem, the stems straight and cut level.
func cup(r *roman) outline.Path {
	t := r.p.Punct.MathThick
	a := cupWidth/2 - t/2 // the radius of the centre line of the round
	top, bottom := cupHeight/2, -cupHeight/2+t/2
	return glyphs.Stroke([]glyphs.Station{
		{X: t / 2, Y: top, Dir: 270, W: t, Flat: true},
		{X: t / 2, Y: bottom + a, Dir: 270, W: t},
		{X: cupWidth / 2, Y: bottom, Dir: 0, W: t},
		{X: cupWidth - t/2, Y: bottom + a, Dir: 90, W: t, Flat: true},
		{X: cupWidth - t/2, Y: top, Dir: 90, W: t},
	})
}

// vee draws ∨: two straight strokes of the rule from the top corners of its
// ink to a vertex at the bottom, cut flat MathVertexFlat wide as the vertex
// of <, the free ends cut level.
func vee(r *roman) outline.Path {
	t, f := r.p.Punct.MathThick, r.p.Punct.MathVertexFlat
	w, h := float64(cupWidth), float64(cupHeight)
	xm := w / 2
	a := math.Atan2(h, xm-f/2)
	c := t / math.Sin(a) // the width of a stroke along a level line
	yv := -h/2 + (c-f/2)*math.Tan(a)
	return outline.Path{polygon(
		pt(0, h/2), pt(xm-f/2, -h/2), pt(xm+f/2, -h/2), pt(w, h/2),
		pt(w-c, h/2), pt(xm, yv), pt(c, h/2))}
}

// wave draws the wave of ∼ with its ink centred on (0, 0): a stroke of the
// rule which rises from the left, crests, falls through the middle, passes a
// trough and rises to the right, the same turned through half a turn about
// its centre, the ends cut at right angles where the wave is steepest.  The
// stations are fitted to make the ink as wide as the plus and simShare of
// that high.
func wave(r *roman) outline.Path {
	t, w := r.p.Punct.MathThick, r.p.Punct.MathWidth
	h := math.Round(simShare * w)
	sw, a := w, (h-t)/2 // the span of the stations and the amplitude of the centre line
	var path outline.Path
	for range 8 {
		// The centre line follows a sine wave of one period, y = -a sin θ for
		// θ from -0.9π to 0.9π: the middle is its inflection, and the ends
		// lie just inside the inflections, where the wave is nearly steepest.
		k := 2 * math.Pi / sw
		const e = 0.9 * math.Pi
		dir := func(th float64) float64 { return math.Atan(-a*math.Cos(th)*k) * 180 / math.Pi }
		sts := []glyphs.Station{
			{X: -e / k, Y: a * math.Sin(e), Dir: dir(-e), W: t, ArcEnd: true},
			{X: -sw / 4, Y: a, Dir: 0, W: t},
			{X: 0, Y: 0, Dir: dir(0), W: t},
			{X: sw / 4, Y: -a, Dir: 0, W: t},
			{X: e / k, Y: -a * math.Sin(e), Dir: dir(e), W: t, ArcEnd: true},
		}
		for _, i := range []int{1, 3} {
			// The crest and the trough have the curvature of the sine.
			sts[i].K = a * k * k
			if i == 1 {
				sts[i].K = -sts[i].K
			}
		}
		path = glyphs.Stroke(sts)
		x0, y0, x1, y1 := path.BBox()
		sw += w - (x1 - x0)
		a += (h - (y1 - y0)) / 2
	}
	x0, y0, x1, y1 := path.BBox()
	return shifted(path, -(x0+x1)/2, -(y0+y1)/2)
}

// withWave returns the equals or ≡ with its upper bar replaced by the wave:
// ≃ from the equals, ≅ from three of its bars, at the pitch of the equals.
func withWave(bars int) func(r *roman) outline.Path {
	return func(r *roman) outline.Path {
		p := r.p.Punct
		t, w := p.MathThick, p.MathWidth
		pitch := p.EqualsGap + t

		var ys []float64
		if bars == 1 {
			ys = []float64{-pitch / 2, pitch / 2}
		} else {
			ys = []float64{-pitch, 0, pitch}
		}
		path := shifted(wave(r), 0, ys[len(ys)-1])
		for _, y := range ys[:len(ys)-1] {
			path = append(path, box(-w/2, y-t/2, w/2, y+t/2))
		}
		// The ink is placed by the bars, which stand where those of the
		// equals and ≡ stand: the axis is the centre of the bars and the wave
		// together, as there.
		return path
	}
}

// approx draws ≈: two waves at the pitch of the bars of the equals.
func approx(r *roman) outline.Path {
	p := r.p.Punct
	pitch := p.EqualsGap + p.MathThick
	w := wave(r)
	return join(shifted(w, 0, -pitch/2), shifted(w, 0, pitch/2))
}

// prec draws ≺: the chevron of the arrows (arrows.go) with strokes of the
// rule, setHeight high and as wide as <, mirrored to open to the right.
func prec(r *roman) outline.Path {
	return precOf(r, setHeight)
}

func precOf(r *roman, height float64) outline.Path {
	p := r.p.Punct
	ch := chevronOfLength(height/2, math.Round(precWidth*p.MathAngleWidth), precBend*math.Pi/180, precPoint, p.MathThick)
	return mirroredX(ch.path(), 0)
}

// precEq draws ⪯: ≺ over a bar of the rule as wide as its ink, made as ≤ is
// made from <: the whole leqHeight times as high as ≺, the white between the
// bar and ≺ that of ±, and ≺ flatter with the same width and stroke.
func precEq(r *roman) outline.Path {
	t := r.p.Punct.MathThick
	total := math.Round(leqHeight * setHeight)
	gap := pmGap(r)
	pr := precOf(r, total-t-gap)
	x0, y0, x1, _ := pr.BBox()
	return join(pr, outline.Path{box(x0, y0-gap-t, x1, y0-gap)})
}

// oplus draws ⊕: a ring of the rule oplusShare as wide as the plus and the
// bars of the plus across its counter, their ends in the ring.
func circled(kind string) func(r *roman) outline.Path {
	return func(r *roman) outline.Path {
		p := r.p.Punct
		t := p.MathThick
		rad := math.Round(oplusShare*p.MathWidth) / 2
		o := pt(0, 0)
		parts := []outline.Path{ring(o, rad, t)}
		reach := rad - t/2 // the bars end in the middle of the ring
		switch kind {
		case "plus":
			parts = append(parts, outline.Path{box(-reach, -t/2, reach, t/2)}, outline.Path{box(-t/2, -reach, t/2, reach)})
		case "times":
			bar := outline.Path{box(-reach, -t/2, reach, t/2)}
			parts = append(parts, rotated(bar, o, math.Pi/4), rotated(bar, o, -math.Pi/4))
		case "dot":
			return join(parts[0], dotAt(r, 0, 0))
		}
		return glyphs.Union(parts...)
	}
}

// circ draws ∘: the ring of ⊕ run with circShare of the width of the plus as
// its diameter and the same stroke.
func circ(r *roman) outline.Path {
	p := r.p.Punct
	return ring(pt(0, 0), math.Round(circShare*p.MathWidth)/2, p.MathThick)
}

// star draws ⋆: a five-pointed star, starShare as wide as the plus, the
// radius of its notches starInner of that of its points, the centre of the
// points on the axis.
func star(r *roman) outline.Path {
	w := starShare * r.p.Punct.MathWidth
	ro := w / 2 / math.Sin(2*math.Pi/5)
	ri := starInner * ro
	var pts []outline.Point
	for i := range 10 {
		a := math.Pi/2 + float64(i)*math.Pi/5
		rr := ro
		if i%2 == 1 {
			rr = ri
		}
		pts = append(pts, pt(rr*math.Cos(a), rr*math.Sin(a)))
	}
	return outline.Path{polygon(pts...)}
}

// starSign places ⋆ with the centre of its points, not of its ink, on the
// axis.
func starSign(r *roman) (outline.Path, float64) {
	return signSides(r, shifted(star(r), 0, r.p.Punct.MathAxis))
}

// propto draws ∝: one stroke of the rule from the end of the upper arm on the
// right through a crossing on the axis, round a loop on the left and through
// the crossing again to the end of the lower arm, the arms ending level and
// cut upright; the ink proptoWidth and proptoHeight of the width of the plus.
// The top and the bottom of the loop lie proptoLoop of the width from the
// left, and the crossing halfway between them and the ends of the arms, at
// the angle with which a stroke turns from level to the crossing and back to
// level in the same distance.
func propto(r *roman) outline.Path {
	p := r.p.Punct
	t := p.MathThick
	w, h := math.Round(proptoWidth*p.MathWidth), math.Round(proptoHeight*p.MathWidth)
	a := h/2 - t/2
	xl := proptoLoop * w
	d := (w - xl) / 2
	xc := xl + d
	phi := 2 * math.Atan(a/d) * 180 / math.Pi
	sts := []glyphs.Station{
		{X: w, Y: a, Dir: 180, W: t},
		{X: xc, Y: 0, Dir: 180 + phi, W: t},
		{X: xl, Y: -a, Dir: 180, W: t},
		{X: t / 2, Y: 0, Dir: 90, W: t},
		{X: xl, Y: a, Dir: 0, W: t},
		{X: xc, Y: 0, Dir: -phi, W: t},
		{X: w, Y: -a, Dir: 0, W: t},
	}
	return glyphs.Union(glyphs.Stroke(sts))
}

// coloneq draws ≔, or ≕ if mirror is true: the dots of ∶ and the equals, the
// white between them the side space of ∶.
func coloneq(mirror bool) func(r *roman) outline.Path {
	return func(r *roman) outline.Path {
		p := r.p.Punct
		dots, _ := ratio(r)
		dx0, _, dx1, _ := dots.BBox()
		dots = shifted(dots, -dx1, -p.MathAxis)
		gap := dx0 // the side space of the colon
		t, w := p.MathThick, p.MathWidth
		h := p.EqualsGap/2 + t
		path := join(dots, outline.Path{box(gap, p.EqualsGap/2, gap+w, h), box(gap, -h, gap+w, -p.EqualsGap/2)})
		if mirror {
			path = mirroredX(path, 0)
		}
		return path
	}
}

// leqSlant draws ⩽: the < of ≤ over a bar of the rule parallel to its lower
// stroke, the white between them, taken upright, that of ≤.
func leqSlant(r *roman) outline.Path {
	p := r.p
	t, f := p.Punct.MathThick, p.Punct.MathVertexFlat
	less, _ := r.get("less")
	_, ly0, _, ly1 := less.BBox()
	total := math.Round(leqHeight * (ly1 - ly0))
	gap := pmGap(r)
	flat := lessWithHeight(p, total-t-gap)
	x0, y0, x1, y1 := flat.BBox()
	ym := (y0 + y1) / 2
	// The outer edge of the lower stroke runs from the lower end of the flat
	// of the vertex to the lower corner of the free end.
	slope := (y0 - (ym - f/2)) / (x1 - x0)
	tv := t * math.Sqrt(1+slope*slope) // the stroke of the rule, taken upright
	top0, top1 := ym-f/2-gap, y0-gap
	bar := outline.Path{polygon(pt(x0, top0-tv), pt(x1, top1-tv), pt(x1, top1), pt(x0, top0))}
	return join(flat, bar)
}

// lessSign returns < or > of the regular.
func lessSign(name string) func(r *roman) outline.Path {
	return func(r *roman) outline.Path { p, _ := r.get(name); return p }
}

// derivedPath returns the path of a derived sign.
func derivedPath(draw func(r *roman) (outline.Path, float64)) func(r *roman) outline.Path {
	return func(r *roman) outline.Path { p, _ := draw(r); return p }
}

// barNegation returns the construction of ∤ (one bar) or ∦ (two bars of the
// parallel): the bars crossed at the axis by the solidus of the negations
// laid across them, mirrored in the diagonal: it meets the upright
// bars at the angle at which the solidus of ≠ meets the level bars, and
// extends negOver beyond the bars on each side; with MathSide on each
// side.
func barNegation(parallelBars bool) func(r *roman) (outline.Path, float64) {
	return func(r *roman) (outline.Path, float64) {
		bars, _ := r.get("bar")
		if parallelBars {
			bars, _ = parallel(r)
		}
		x0, _, x1, _ := bars.BBox()
		sl := solidus(r, 0, 0, x1-x0+2*negOver)
		sl = mapPoints(sl, func(q outline.Point) outline.Point { return pt(q.Y, q.X) })
		for i := range sl {
			sl[i] = reversed(sl[i])
		}
		path := glyphs.Union(bars, shifted(sl, (x0+x1)/2, r.p.Punct.MathAxis))
		return withSides(path, r.p.Punct.MathSide, r.p.Punct.MathSide)
	}
}

// nmid draws ∤: see barNegation.
func nmid(r *roman) (outline.Path, float64) { return barNegation(false)(r) }

// The relations and binary operators, in the order of the font.
var relationDerivations = []derivation{
	{"uni2208", 0x2208, Rel, "plus", sign(element)},
	{"uni220B", 0x220B, Rel, "plus", mirroredSign(element)},
	{"uni2209", 0x2209, Rel, "plus", sign(negated(element))},
	{"uni2282", 0x2282, Rel, "plus", sign(subset)},
	{"uni2283", 0x2283, Rel, "plus", mirroredSign(subset)},
	{"uni2286", 0x2286, Rel, "plus", sign(subsetEq)},
	{"uni2287", 0x2287, Rel, "plus", mirroredSign(subsetEq)},
	{"uni228A", 0x228A, Rel, "plus", subsetNeqSign},
	{"uni2288", 0x2288, Rel, "plus", sign(negated(subsetEq))},
	{"uni222A", 0x222A, Bin, "plus", sign(cup)},
	{"uni2229", 0x2229, Bin, "plus", turnedDown(cup)},
	{"uni2228", 0x2228, Bin, "plus", sign(vee)},
	{"uni2227", 0x2227, Bin, "plus", turnedDown(vee)},
	{"uni223C", 0x223C, Rel, "plus", sign(wave)},
	{"uni2241", 0x2241, Rel, "plus", sign(negated(wave))},
	{"uni2248", 0x2248, Rel, "plus", sign(approx)},
	{"uni2243", 0x2243, Rel, "plus", waveSign(1)},
	{"uni2245", 0x2245, Rel, "plus", waveSign(2)},
	{"uni227A", 0x227A, Rel, "plus", sign(prec)},
	{"uni227B", 0x227B, Rel, "plus", mirroredSign(prec)},
	{"uni2AAF", 0x2AAF, Rel, "plus", sign(precEq)},
	{"uni2AB0", 0x2AB0, Rel, "plus", mirroredSign(precEq)},
	{"uni2295", 0x2295, Bin, "plus", sign(circled("plus"))},
	{"uni2297", 0x2297, Bin, "plus", sign(circled("times"))},
	{"uni2299", 0x2299, Bin, "plus", sign(circled("dot"))},
	{"uni2218", 0x2218, Bin, "plus", sign(circ)},
	{"uni22C6", 0x22C6, Bin, "plus", starSign},
	{"uni221D", 0x221D, Rel, "plus", sign(propto)},
	{"uni2254", 0x2254, Rel, "plus", sign(coloneq(false))},
	{"uni2255", 0x2255, Rel, "plus", sign(coloneq(true))},
	{"uni2A7D", 0x2A7D, Rel, "plus", sign(leqSlant)},
	{"uni2A7E", 0x2A7E, Rel, "plus", mirroredSign(leqSlant)},
	{"uni2270", 0x2270, Rel, "plus", sign(negated(derivedPath(leq(false))))},
	{"uni2271", 0x2271, Rel, "plus", sign(negated(derivedPath(leq(true))))},
	{"uni226E", 0x226E, Rel, "plus", sign(negated(lessSign("less")))},
	{"uni226F", 0x226F, Rel, "plus", sign(negated(lessSign("greater")))},
	{"uni2262", 0x2262, Rel, "plus", sign(negated(derivedPath(equiv)))},
	{"uni2224", 0x2224, Rel, "plus", nmid},
	{"uni2226", 0x2226, Rel, "plus", barNegation(true)},
	{"uni220C", 0x220C, Rel, "plus", sign(negated(mirroredDraw(element)))},
	{"uni2284", 0x2284, Rel, "plus", sign(negated(subset))},
	{"uni2285", 0x2285, Rel, "plus", sign(negated(mirroredDraw(subset)))},
	{"uni2289", 0x2289, Rel, "plus", sign(negated(mirroredDraw(subsetEq)))},
	{"uni2249", 0x2249, Rel, "plus", sign(negated(approx))},
}

// waveSign places ≃ (bars 1) or ≅ (bars 2) with its bars and wave at the
// pitch of the equals, as the bars of the equals stand (≃) or with
// the middle bar on the axis (≅), centred on the axis by them and
// not by the ink.
func waveSign(bars int) func(r *roman) (outline.Path, float64) {
	return func(r *roman) (outline.Path, float64) {
		return signSides(r, shifted(withWave(bars)(r), 0, r.p.Punct.MathAxis))
	}
}
