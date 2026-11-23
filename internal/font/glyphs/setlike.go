// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"fmt"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// SetLike returns the glyphs of the font for the parameters p with the nodes
// of the glyphs for the parameters ref: the outline of every glyph has the
// contours, the segments and the start of each contour which Set(ref) gives
// the glyph.  Each outline is rounded by outline.Path.RoundLike against the
// outline of the same glyph for ref, where Set rounds it by
// outline.Path.Round.
//
// A glyph whose outline cannot take the nodes of the model is returned
// as Set returns it, and failed holds the reason under its name; a composed
// letter fails with its base or its mark.  A glyph which cannot be drawn for
// p has no outline and fails too.
func SetLike(p, ref design.Params) (gs []Glyph, failed map[string]error) {
	gs = make([]Glyph, len(table))
	raws := make([]outline.Path, len(table))
	refRaws := make([]outline.Path, len(table))
	failed = map[string]error{}
	for i, e := range table {
		if composites[e.name] != nil {
			continue
		}
		refRaws[i], _ = e.draw(ref)
		path, adv, err := tryDraw(e, p)
		if err != nil {
			failed[e.name] = err
			gs[i] = Glyph{Name: e.name, Rune: e.r, Parts: LigatureParts(e.name)}
			continue
		}
		rounded, err := path.RoundLike(refRaws[i])
		if err != nil {
			failed[e.name] = err
			rounded = path.Round()
		}
		gs[i] = Glyph{Name: e.name, Rune: e.r, Advance: adv, Path: rounded, Parts: LigatureParts(e.name)}
		raws[i] = path
	}

	index := make(map[string]int, len(table))
	for i, e := range table {
		index[e.name] = i
	}
	type roundedMark struct {
		raw, ref, rounded outline.Path
		err               error
	}
	marks := map[markKey]roundedMark{}
	for i, e := range table {
		gs[i].Like = like(e.name)
		c := composites[e.name]
		if c == nil {
			if raws[i] != nil || failed[e.name] == nil {
				gs[i].Anchors = baseAnchors(p, e.name, raws[i])
			}
			continue
		}
		b := index[c.base]
		if raws[b] == nil && failed[c.base] != nil {
			failed[e.name] = fmt.Errorf("the base %s: %w", c.base, failed[c.base])
			gs[i] = Glyph{Name: e.name, Rune: e.r, Base: c.base, Like: gs[i].Like}
			continue
		}
		dx, dy, f := c.place(p, raws[b])
		rdx, rdy, _ := c.place(ref, refRaws[b])
		key := markKey{c.mark, f}
		m, ok := marks[key]
		if !ok {
			m.raw, m.ref = markOutline(p, c.mark, f), markOutline(ref, c.mark, f)
			m.rounded, m.err = m.raw.RoundLike(m.ref)
			if m.err != nil {
				m.rounded = m.raw.Round()
			}
			marks[key] = m
		}
		var path outline.Path
		err := failed[c.base]
		switch {
		case err != nil:
			err = fmt.Errorf("the base %s: %w", c.base, err)
		case c.joined():
			path, err = c.assemble(raws[b], m.raw, dx, dy).RoundLike(c.assemble(refRaws[b], m.ref, rdx, rdy))
		case m.err != nil:
			err = fmt.Errorf("the mark: %w", m.err)
		default:
			path = c.assemble(gs[b].Path, m.rounded, dx, dy)
		}
		if err != nil {
			failed[e.name] = err
			path = c.rounded(gs[b].Path, m.rounded, raws[b], m.raw, dx, dy)
		}
		gs[i] = Glyph{
			Name: e.name, Rune: e.r, Advance: gs[b].Advance,
			Path: path,
			Base: c.base, Like: gs[i].Like, Marks: shiftInk(m.rounded, dx, dy),
		}
	}
	for i, e := range table {
		if k, ok := combiningKind(e.name); ok {
			gs[i].Anchors = []Anchor{markAnchor(p, k)}
		}
		gs[i].Height = inklessHeight(p, e.name)
	}
	return gs, failed
}

// tryDraw draws the glyph of the entry for the parameters p; a panic of the
// glyph program is returned as an error.
func tryDraw(e entry, p design.Params) (path outline.Path, adv float64, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	path, adv = e.draw(p)
	return path, adv, nil
}
