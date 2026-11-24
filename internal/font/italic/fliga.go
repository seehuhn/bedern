// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"seehuhn.de/go/bedern/internal/fair"
	"seehuhn.de/go/bedern/internal/stroke"
)

// The drawn ligatures of f: the hook of f runs over into the top of the
// letter which follows, whose stem stands LigPitch right of the stem of f,
// the pitch of the stems of n.
//
// Before an ascender (l, b, h, k) the stroke of f, from the ball of its
// tail up the stem, runs over a crest at the height of the crest of f and
// down into the stem of the ascender in place of its entry, as the arch of n
// runs into its right stem, and on as that letter's stroke: the bridge.
// Before a second f the stroke of the first f runs over its crest into the
// stem of the second, below the second's own crest, and ends inside it; the
// second f keeps its hook.  The crossbar is that of f; after a second f it
// runs on from the first f through the stem of the second.

// liga is a drawn ligature: its name, its code point (0 for none) and the
// letters which follow the f, after the first.
type liga struct {
	r     rune
	parts []string
}

// ligatures are the drawn ligatures, by name.
var ligatures = map[string]liga{
	"fl":  {'ﬂ', []string{"l"}},
	"f_b": {0, []string{"b"}},
	"f_h": {0, []string{"h"}},
	"f_k": {0, []string{"k"}},
	"ff":  {'ﬀ', []string{"f"}},
	"ffl": {'ﬄ', []string{"f", "l"}},
}

// LigatureParts returns the letters for which the drawn ligature stands, and
// false if the name is not a drawn ligature.
func LigatureParts(name string) ([]string, bool) {
	l, ok := ligatures[name]
	if !ok {
		return nil, false
	}
	return append([]string{"f"}, l.parts...), true
}

func init() {
	why := "the leg of k in the ligature: it bows out from the stem and turns the other way into the exit, each edge inflecting once, as on k"
	intendedFindings["f_k"] = []intended{
		{Kind: fair.Inflection, X: 548, Y: 118, Dir: -90, Reason: why},
		{Kind: fair.Inflection, X: 616, Y: 82, Dir: 81, Reason: why},
	}
	for name, l := range ligatures {
		parts := l.parts
		letters[name] = struct {
			r    rune
			draw func(Params) letter
		}{l.r, func(p Params) letter { return fLigature(p, parts) }}
	}
}

// fRise returns the nodes of f from the ball of its tail up the stem whose
// left edge stands at e, to the crest at the middle between its centre line
// and that of the stem LigPitch to the right, and the start of the stroke.
func (p Params) fRise(e float64) ([]stroke.Node, stroke.End) {
	c := e + p.Stem/2
	tail, ball := p.tail(e)
	var out []stroke.Node
	for k := len(tail) - 1; k >= 0; k-- {
		nd := tail[k]
		nd.Dir += 180
		out = append(out, nd)
	}
	top := p.XHeight + p.AscenderEntryRise
	up := p.stemNode(c, top+p.TopLift-p.EntryJoin)
	up.Dir = 90
	crest := stroke.Node{P: pt(c+p.LigPitch*p.LigCrestAt, top+p.Overshoot-p.nibWidth(0)/2), Dir: 0, K: auto, W: auto, DW: 0}
	return append(out, up, crest), ball
}

// fBar returns the crossbar of the f whose stem's left edge stands at e,
// reaching left as far as on f and right by right beyond the right edge of
// the stem.
func (p Params) fBar(e, right float64) stroke.Stroke {
	y := p.XHeight - p.Hair/2
	return stroke.Stroke{Nodes: []stroke.Node{
		{P: pt(e-p.FBarLeft, y), Dir: 0, W: p.Hair},
		{P: pt(e+p.Stem+right, y), Dir: 0, W: p.Hair},
	}}
}

// fLigature draws f followed by the letters of parts, each f but the last
// bridged into the next letter.
func fLigature(p Params, parts []string) letter {
	e := p.FSideLeft
	var out letter
	for i, part := range parts {
		pitch := p.LigPitch
		if part == "b" || part == "h" || part == "k" {
			pitch = p.LigBowlPitch
		}
		next := e + pitch
		q := p
		q.LigPitch = pitch
		if part == "f" {
			q.LigCrestAt = p.LigFCrestAt
		}
		rise, ball := q.fRise(e)
		if part == "f" {
			// The first f runs into the stem of the second below its crest
			// and ends inside it.
			cn := next + p.Stem/2
			join, drop := p.LigFJoin, p.LigFDrop
			if i < len(parts)-1 {
				// The second f runs on into an ascender, and its top turns sooner.
				join, drop = p.LigFLJoin, p.LigFLDrop
			}
			end := stroke.Node{P: pt(cn+join, p.XHeight+p.AscenderEntryRise+p.Overshoot-p.BallStalk/2-drop), Dir: p.LigFDir, K: auto, W: auto, DW: auto}
			out.strokes = append(out.strokes, stroke.Stroke{Nib: p.nib(), Nodes: append(rise, end), Start: ball, End: hidden})
			if i == len(parts)-1 {
				lf := f(p)
				out.strokes = append(out.strokes, moved(lf.strokes[0], next-p.FSideLeft), p.fBar(e, p.LigPitch+p.FBarRight))
				out.advance = next - p.FSideLeft + lf.advance
			}
			e = next
			continue
		}
		// An ascender: its first stroke from the node where the entry joins
		// the stem.
		d := letters[part].draw(p)
		edge := ascenderEdge(p, part)
		dx := next - edge
		first := moved(d.strokes[0], dx)
		// Where the bridge turns down into the stem it runs vertically in the
		// italic, as the arch of n does at its shoulder.
		join := p.XHeight + p.AscenderEntryRise + p.TopLift - p.EntryJoin
		vert := p.vertNode(next+p.Stem/2, join, join+p.LigVert, p.ShoulderW)
		s := stroke.Stroke{Nib: p.nib(), Nodes: append(append(rise, vert), first.Nodes[3:]...), Start: ball, End: first.End}
		bar := p.fBar(e, p.FBarRight)
		if i > 0 {
			// The bar of the first f runs on through the stem of the second.
			bar = p.fBar(e-p.LigPitch, p.LigPitch+p.FBarRight)
		}
		out.strokes = append(out.strokes, s, bar)
		for _, t := range d.strokes[1:] {
			out.strokes = append(out.strokes, moved(t, dx))
		}
		out.advance = next - edge + d.advance
	}
	return out
}

// ascenderEdge returns the left edge of the stem of the ascender of l, b, h
// and k.
func ascenderEdge(p Params, name string) float64 {
	switch name {
	case "b":
		return p.SideStemLeft - p.BLeftLess
	case "h", "k":
		return p.SideStemLeft - p.HLeftLess
	}
	return p.SideStemLeft
}
