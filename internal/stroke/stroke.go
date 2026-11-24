// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

// Package stroke constructs glyph outlines from pen strokes.
//
// A stroke is a centre line given by a few nodes, each with a position, a
// direction, a curvature and the thickness of the pen there, and two ends.  The
// two edges of the stroke are derived at the nodes from these values by the
// formulas of the offset curve, and between two nodes each edge is one cubic
// curve with the points, directions and curvatures of the edge at both nodes.
// The edges are therefore tangent- and curvature-continuous at every node.  The
// edges are split at their horizontal and vertical extremes, pieces shorter
// than MinPiece are merged with a neighbour, and the ends add the cut, the
// point, the round, the ball, the wedge or the clip at the edge of another
// stroke.
package stroke

import (
	"fmt"
	"math"
	"sort"

	"seehuhn.de/go/bedern/internal/outline"
)

// Auto marks a value of a node which the rules supply.
var Auto = math.NaN()

func isAuto(x float64) bool { return math.IsNaN(x) }

// MinPiece is the shortest piece of an edge, in font units: the limit of the
// smoothness gate.
const MinPiece = 15.0

// A Node is a point of the centre line of a stroke.
type Node struct {
	P   Point   // position
	Dir float64 // direction of travel, in degrees
	K   float64 // signed curvature, positive to the left; Auto: by rule
	W   float64 // thickness of the stroke at right angles to it; Auto with a nib: from the nib
	DW  float64 // rate of change of W along the centre line; Auto: by rule
	DDW float64 // second derivative of W along the centre line

	// Skew moves the left edge point of the node by this distance along the
	// direction of travel and the right edge point back by as much: the pen
	// is held askew, as a broad nib touches its edges at points which are not
	// opposite each other.  It is meant for nodes on straight parts, where
	// the directions and curvatures of the edges do not change with it.
	Skew float64

	// Edge, if set, gives the curvatures of the two edges at the node instead
	// of K and DDW, which are then solved from them.
	Edge *EdgeK
}

// EdgeK holds the signed curvatures of the right and the left edge at a node.
type EdgeK struct{ Right, Left float64 }

// SuperK returns the curvatures at the two extremes of a quarter of the
// super-ellipse of the bowls (a cubic with the handles k a and k b, k = (8
// superness - 4) / 3) with the half axes a, along the extreme where it is
// horizontal or vertical, and b: the curvature at the extreme on the axis of a,
// and at the extreme on the axis of b.
func SuperK(a, b, superness float64) (atA, atB float64) {
	k := (8*superness - 4) / 3
	return 2 * b * (1 - k) / (3 * k * k * a * a), 2 * a * (1 - k) / (3 * k * k * b * b)
}

// solveEdgeK sets K and DDW of the node so that its edges have the curvatures
// of nd.Edge.
func solveEdgeK(nd *Node) error {
	kr, kl := nd.Edge.Right, nd.Edge.Left
	edgeK := func(k, ddw, sigma float64) float64 {
		m := *nd
		m.K, m.DDW = k, ddw
		e, err := edgeAt(m, sigma)
		if err != nil {
			return math.NaN()
		}
		return e.K
	}
	// For a fixed K the curvature of an edge is linear in DDW.
	ddwFor := func(k float64) float64 {
		a, b := edgeK(k, 0, 1), edgeK(k, 1, 1)
		return (kl - a) / (b - a)
	}
	f := func(k float64) float64 { return edgeK(k, ddwFor(k), -1) - kr }
	k := (kl + kr) / 2
	for i := 0; i < 100; i++ {
		fk := f(k)
		if math.Abs(fk) < 1e-13 {
			break
		}
		h := 1e-7
		d := (f(k+h) - fk) / h
		if d == 0 || math.IsNaN(d) {
			return fmt.Errorf("stroke: no centre curvature gives the edge curvatures at %v", nd.P)
		}
		k -= fk / d
	}
	if math.IsNaN(k) || math.Abs(f(k)) > 1e-9 {
		return fmt.Errorf("stroke: no centre curvature gives the edge curvatures at %v", nd.P)
	}
	nd.K, nd.DDW = k, ddwFor(k)
	return nil
}

// A Nib is a broad pen: an ellipse with the axis Major in the direction Angle
// (degrees) and the axis Minor across it.  It supplies the thickness of the
// stroke at the nodes which leave W to the rules: the width of the ellipse at
// right angles to the stroke.
type Nib struct {
	Major, Minor, Angle float64
}

// width returns the thickness of a stroke in the direction theta (degrees).
func (b Nib) width(theta float64) float64 {
	n := rot90(dirOf(theta))
	u := dirOf(b.Angle)
	v := rot90(u)
	a, c := dot(n, u)*b.Major/2, dot(n, v)*b.Minor/2
	return 2 * math.Hypot(a, c)
}

// EndKind is the kind of the end of a stroke.
type EndKind int

// The kinds of end.  A stroke which tapers to a point is a Cut end with a
// node of zero thickness.
const (
	Cut   EndKind = iota // a straight cut, at right angles or along the line of Angle
	Round                // a half circle across the end
	Ball                 // the ball terminal
	Wedge                // the edges meet in a point beyond the end
	Clip                 // the edges stop on the line through Q in the direction Angle, the edge of another stroke
)

// An End says how a stroke ends.
type End struct {
	Kind  EndKind
	Line  bool    // Cut: along the line of Angle through the end node; otherwise at right angles
	Angle float64 // Cut with Line, and Clip: the direction of the line, in degrees
	Q     Point   // Clip: a point of the line

	Ball BallSpec // Ball

	Length, Offset float64 // Wedge: the tip lies Length along the stroke beyond the end node and Offset to its left

	// Sever marks a cut which only bounds a part taken out of a letter: its
	// line is a JoinPiece, not a part of the shape.
	Sever bool

	// Hidden marks an end which lies inside another stroke of the letter:
	// its pieces must not survive the union (Letter.Build warns if they do).
	Hidden bool
}

// A BallSpec is a ball terminal: the circle of the radius R about Centre.  The
// outer edge of the curl winds into the circle on a spiral whose curvature
// rises to that of the circle.  The inner edge meets the circle at the tip of
// the wedge, in the direction Tip (degrees) from the centre, at the angle Open
// (degrees) to the circle, with the curvature TipK (Auto: that of the inner
// edge at the end node).
type BallSpec struct {
	Centre    Point
	R         float64
	Tip, Open float64
	TipK      float64

	// Neck, if set, places the tip of the wedge instead of Tip: at the first
	// point of the circle, from where the outer edge runs into it back towards
	// the stalk, which lies Neck from the outer edge.
	Neck float64

	// InkSide, if set (-1 left, +1 right), places the ball by its side
	// space instead of by the abscissa of Centre: the leftmost or rightmost
	// ink of the ball and its curl lies at InkX.
	InkSide int
	InkX    float64

	// Power, if set, places the curl by the rule of the spiral (see
	// ballplace.go): the curvature of the outer edge rises as a power of
	// the length between Power[0] and Power[1], solved so that the spiral
	// from the end node runs into the circle; where the power would leave
	// that range the end node slides along its direction, and where the ball
	// lies outside the range of the rule it turns first.  The centre lies
	// on whole units.
	Power [2]float64

	// Depth, if set with Power, places the ball by the end node instead of
	// by Centre: its centre lies |Depth| from the tangent of the outer edge
	// at the end node, to the right of the stroke if Depth is positive and
	// to its left if negative, and as far ahead as the spiral puts it.
	Depth float64
}

// A Stroke is a centre line with a pen and two ends.  A closed stroke has no
// ends: its first node follows its last.
type Stroke struct {
	Nodes      []Node
	Closed     bool
	Nib        *Nib
	Start, End End

	// Slant, if set by ShearedRatio, is the slant as the integer ratio
	// num/den of its tangent; the edges of stems and the handles which leave
	// them are then placed on whole-unit points of their lines (lattice).
	Slant [2]int
}

// PieceKind says what a piece of an outline is.
type PieceKind int

// The kinds of piece.
const (
	EdgePiece   PieceKind = iota // part of an edge of the stroke
	EndPiece                     // part of an end which belongs to the shape (cut, round, ball, wedge)
	JoinPiece                    // the line along the edge of another stroke at a Clip end
	HiddenPiece                  // part of an end which another stroke covers
)

// A Piece is one curve of an outline.
type Piece struct {
	C    Cubic
	Kind PieceKind

	// Fixed marks a curve which runs out of a stem on the lattice: its
	// handles lie on whole-unit points of the stem's line, and it is not
	// split, merged or moved.
	Fixed bool
}

// A Chain is a list of pieces, each starting where the last ends.
type Chain []Piece

func (c Chain) start() Point { return c[0].C[0] }
func (c Chain) end() Point   { return c[len(c)-1].C[3] }

func (c Chain) reversed() Chain {
	out := make(Chain, len(c))
	for i, p := range c {
		out[len(c)-1-i] = Piece{C: p.C.Reversed(), Kind: p.Kind, Fixed: p.Fixed}
	}
	return out
}

// A Result is a stroke as outline: the right and the left edge in the
// direction of travel, and the contours, each a closed chain: the right edge,
// the end, the left edge backwards and the start; a closed stroke has the
// right and the left edge as two contours.
type Result struct {
	Right, Left Chain
	Contours    []Chain
	Warnings    []string
	WarnAt      []Point // the place of each warning about the shape of an edge, or NaN
}

// edgeData is an edge at a node: its point, unit direction and curvature.
type edgeData struct {
	P, T Point
	K    float64
	Lat  float64 // the slant of the lattice line the point lies on (see lattice), or 0
}

// resolved returns the nodes with the values of the rules filled in.
func (s Stroke) resolved() ([]Node, error) {
	return s.Resolved()
}

// Resolved returns the nodes of the stroke with the values which the rules
// supply filled in: the curvatures, the thicknesses from the pen and their
// rates of change.
func (s Stroke) Resolved() ([]Node, error) {
	ns := append([]Node(nil), s.Nodes...)
	n := len(ns)
	if n < 2 {
		return nil, fmt.Errorf("stroke: a stroke needs two nodes")
	}
	idx := func(i int) (int, bool) {
		if s.Closed {
			return (i + n) % n, true
		}
		return i, i >= 0 && i < n
	}
	// Thickness from the nib.
	for i := range ns {
		if isAuto(ns[i].W) {
			if s.Nib == nil {
				return nil, fmt.Errorf("stroke: node %d has no thickness and the stroke has no nib", i)
			}
		}
	}
	// Curvature: the mean of the curvatures at the node of the centre lines to
	// the neighbours drawn as cubics with the handles of a circular arc
	// (arcCubic between the two nodes, from their positions and directions).
	for i := range ns {
		if !isAuto(ns[i].K) {
			continue
		}
		var sum float64
		var cnt int
		if j, ok := idx(i - 1); ok {
			sum += arcHandles(ns[j], ns[i]).Curvature(1)
			cnt++
		}
		if j, ok := idx(i + 1); ok {
			sum += arcHandles(ns[i], ns[j]).Curvature(0)
			cnt++
		}
		ns[i].K = sum / float64(cnt)
	}
	if s.Nib != nil {
		for i := range ns {
			if !isAuto(ns[i].W) {
				continue
			}
			const h = 1e-3
			th := ns[i].Dir
			w0, wp, wm := s.Nib.width(th), s.Nib.width(th+h), s.Nib.width(th-h)
			rad := h * math.Pi / 180
			d1 := (wp - wm) / (2 * rad)
			d2 := (wp - 2*w0 + wm) / (rad * rad)
			ns[i].W = w0
			if isAuto(ns[i].DW) {
				ns[i].DW = d1 * ns[i].K
			}
			ns[i].DDW += d2 * ns[i].K * ns[i].K
		}
	}
	// Rate of change of the thickness: zero at a node where the thickness is
	// extreme, the weighted mean of the slopes of the two chords otherwise.
	for i := range ns {
		if !isAuto(ns[i].DW) {
			continue
		}
		jp, okp := idx(i - 1)
		jn, okn := idx(i + 1)
		var s1, s2, l1, l2 float64
		if okp {
			l1 = dist(ns[i].P, ns[jp].P)
			s1 = (ns[i].W - ns[jp].W) / l1
		}
		if okn {
			l2 = dist(ns[jn].P, ns[i].P)
			s2 = (ns[jn].W - ns[i].W) / l2
		}
		switch {
		case okp && okn && s1*s2 <= 0:
			ns[i].DW = 0
		case okp && okn:
			ns[i].DW = (s1*l2 + s2*l1) / (l1 + l2)
		case okp:
			ns[i].DW = s1
		default:
			ns[i].DW = s2
		}
	}
	for i := range ns {
		if ns[i].Edge != nil {
			if isAuto(ns[i].K) {
				ns[i].K = 0
			}
			if err := solveEdgeK(&ns[i]); err != nil {
				return nil, err
			}
		}
	}
	return ns, nil
}

// lattice moves the edge points of the stem nodes onto whole-unit points of
// their lines.  Under a slant whose tangent is num/den, a line x = x0 + y
// num/den with an integer x0 has a whole-unit point at every height which is a
// multiple of den.  A stem node is a node without curvature and without
// change of thickness whose direction is that of the slant; each of its edge
// points moves along its edge to the nearest such height and sideways to the
// nearest integer x0, by at most half a unit, so that the stem's edge is a
// line of the lattice.  The curves which leave such a point put their handles
// on the lattice too (runOut), so that the rounding of the font changes
// nothing on the stem and the edge leaves it exactly in its direction.
func lattice(ns []Node, sides *[2][]edgeData, slant [2]int) {
	k := float64(slant[0]) / float64(slant[1])
	den := float64(slant[1])
	for i, nd := range ns {
		if nd.K != 0 || nd.DW != 0 || nd.DDW != 0 || isAuto(nd.K) {
			continue
		}
		t := dirOf(nd.Dir)
		if math.Abs(t.X-k*t.Y) > 1e-9 {
			continue
		}
		for j := range sides {
			e := &sides[j][i]
			x0 := math.Round(e.P.X - k*e.P.Y)
			y := math.Round(e.P.Y/den) * den
			e.P = Point{X: x0 + k*y, Y: y}
			e.Lat = k
		}
	}
}

// snapHandles moves the handles of c which lie on the line of its start, a
// lattice point of a line of the slant k, to the nearest whole-unit points of
// that line ahead of the start.
func snapHandles(c Cubic, k float64) Cubic {
	den := latticeDen(k)
	t := unit(sub(c[1], c[0]))
	if norm(t) == 0 || math.Abs(t.Y) < 1e-9 {
		return c
	}
	for i := 1; i <= 2; i++ {
		d := sub(c[i], c[0])
		if math.Abs(cross(t, d)) > 1e-6 {
			continue
		}
		q, s := onLattice(c[0], t, dot(d, t), den)
		if s <= 0 {
			q, _ = onLattice(c[0], t, den/math.Abs(t.Y), den)
		}
		c[i] = q
	}
	return c
}

// onLattice returns the point of the line through p (a lattice point) in the
// direction t which lies nearest to p + d t at a height which is a multiple of
// den, and the distance from p to it along t.
func onLattice(p, t Point, d, den float64) (Point, float64) {
	y := math.Round((p.Y+d*t.Y)/den) * den
	s := (y - p.Y) / t.Y
	return add(p, mul(t, s)), s
}

// feasibleAuto moves the curvature of an edge at a node whose curvature the
// rules supply, where a curve to a neighbour has no solution with it, to the
// value nearest to the rule's for which the curves to both neighbours exist:
// the candidates are the rule's value times 1 plus or minus multiples of 0.05,
// up to a factor of 3 either way, and zero.  Nodes are visited in order, twice.
func feasibleAuto(es []edgeData, s Stroke, closed bool) {
	n := len(es)
	ok := func(a, b edgeData) bool {
		if _, ok := hermite(a.P, a.T, a.K, b.P, b.T, b.K); ok {
			return true
		}
		_, ok := edgeCurves(a, b)
		return ok
	}
	both := func(k int) bool {
		good := true
		if j := k - 1; j >= 0 || closed {
			good = good && ok(es[(j+n)%n], es[k])
		}
		if j := k + 1; j < n || closed {
			good = good && ok(es[k], es[j%n])
		}
		return good
	}
	for pass := 0; pass < 2; pass++ {
		for k := range es {
			if !isAuto(s.Nodes[k].K) || s.Nodes[k].Edge != nil || both(k) {
				continue
			}
			k0 := es[k].K
			for step := 1; step <= 40; step++ {
				found := false
				for _, f := range []float64{1 + 0.05*float64(step), 1 - 0.05*float64(step)} {
					if f < 1.0/3 && f > 0 || f <= 0 && step < 40 {
						continue
					}
					if f <= 0 {
						f = 0 // the last candidate: no curvature
					}
					es[k].K = k0 * f
					if both(k) {
						found = true
						break
					}
				}
				if found {
					break
				}
				es[k].K = k0
			}
		}
	}
}

// foldMargin is the least ratio of the radius of the centre line to half the
// thickness of the pen: below it the inner edge folds back on itself.
const foldMargin = 1.15

// checkFold refuses a stroke whose centre line, drawn between two nodes as
// the cubic with the handles of a circular arc, turns anywhere with a radius
// less than foldMargin times half the thickness there (the thickness
// interpolated linearly between the nodes): the inner edge would fold back.
// The message names the place, the radius and the least radius the pen needs.
func checkFold(ns []Node, closed bool) error {
	n := len(ns)
	last := n - 1
	if closed {
		last = n
	}
	for i := 0; i < last; i++ {
		a, b := ns[i], ns[(i+1)%n]
		if isAuto(a.W) || isAuto(b.W) {
			continue
		}
		c := arcHandles(a, b)
		for j := 0; j <= 32; j++ {
			t := float64(j) / 32
			k := math.Abs(c.Curvature(t))
			w := a.W + t*(b.W-a.W)
			if k > 0 && 1/k < foldMargin*w/2 {
				q := c.At(t)
				return fmt.Errorf("stroke: the centre line turns with a radius of %.1f at %.0f,%.0f, where the pen is %.0f thick: it needs a radius of at least %.1f there, or a thinner pen", 1/k, q.X, q.Y, w, foldMargin*w/2)
			}
		}
	}
	return nil
}

// arcHandles returns the cubic from the node a to the node b with their
// directions and the handles of a circular arc with the same turn, a third of
// the chord where the directions are parallel.
func arcHandles(a, b Node) Cubic {
	t0, t3 := dirOf(a.Dir), dirOf(b.Dir)
	l := dist(a.P, b.P)
	theta := math.Abs(math.Atan2(cross(t0, t3), dot(t0, t3)))
	h := l / 3
	if theta > 1e-6 {
		h = l * 2 / 3 * math.Tan(theta/4) / math.Sin(theta/2)
	}
	return Cubic{a.P, add(a.P, mul(t0, h)), sub(b.P, mul(t3, h)), b.P}
}

// edgeAt returns the edge on the side sigma (+1 left, -1 right) at the node.
// With the unit tangent t and normal n = t turned left, the edge is
// e = c + sigma W/2 n, and with c' = t, t' = K n, n' = -K t (the derivative of
// K taken as zero) it has
//
//	e'  = (1 - sigma K W/2) t + sigma W'/2 n
//	e'' = -sigma K W' t + ((1 - sigma K W/2) K + sigma W''/2) n.
func edgeAt(nd Node, sigma float64) (edgeData, error) {
	t := dirOf(nd.Dir)
	n := rot90(t)
	a := 1 - sigma*nd.K*nd.W/2
	b := sigma * nd.DW / 2
	c := -sigma * nd.K * nd.DW
	d := a*nd.K + sigma*nd.DDW/2
	if a <= 0.05 {
		return edgeData{}, fmt.Errorf("stroke: the pen is wider than the curve at %v", nd.P)
	}
	d1 := add(mul(t, a), mul(n, b))
	k := (a*d - b*c) / math.Pow(a*a+b*b, 1.5)
	return edgeData{P: add(add(nd.P, mul(n, sigma*nd.W/2)), mul(t, sigma*nd.Skew)), T: unit(d1), K: k}, nil
}

// Build returns the outline of the stroke.
func (s Stroke) Build() (*Result, error) {
	s, err := s.placeBalls()
	if err != nil {
		return nil, err
	}
	ns, sides, err := s.edges()
	if err != nil {
		return nil, err
	}
	r := &Result{}
	var chains [2]Chain
	for i := range sides {
		chains[i], err = r.edgeChain(sides[i], s.Closed)
		if err != nil {
			return nil, err
		}
	}
	return s.buildRest(r, ns, sides, chains)
}

// edges returns the resolved nodes and the data of the right and the left
// edge at each of them, with the curvatures of the rules and the lattice.
func (s Stroke) edges() ([]Node, [2][]edgeData, error) {
	var sides [2][]edgeData // right, left
	ns, err := s.resolved()
	if err != nil {
		return nil, sides, err
	}
	if err := checkFold(ns, s.Closed); err != nil {
		return nil, sides, err
	}
	for i, sigma := range []float64{-1, 1} {
		for _, nd := range ns {
			e, err := edgeAt(nd, sigma)
			if err != nil {
				return nil, sides, err
			}
			sides[i] = append(sides[i], e)
		}
	}
	// At a node whose curvature the rules supply, each edge takes its own
	// curvature from its own neighbours: the mean of the curvatures at the
	// node of the cubics with the handles of a circular arc from the edge
	// point at the node to those at its neighbours.  The two edges of a
	// curve differ in curvature by more than the offset formula says where the
	// thickness changes, and so they need not agree with the centre line.
	n := len(ns)
	for k := range ns {
		if !isAuto(s.Nodes[k].K) || s.Nodes[k].Edge != nil || s.Nib != nil && isAuto(s.Nodes[k].W) {
			continue
		}
		for i := range sides {
			es := sides[i]
			var sum float64
			var cnt int
			if j := k - 1; j >= 0 || s.Closed {
				j = (j + n) % n
				sum += arcHandles(Node{P: es[j].P, Dir: angleOf(es[j].T)}, Node{P: es[k].P, Dir: angleOf(es[k].T)}).Curvature(1)
				cnt++
			}
			if j := k + 1; j < n || s.Closed {
				j = j % n
				sum += arcHandles(Node{P: es[k].P, Dir: angleOf(es[k].T)}, Node{P: es[j].P, Dir: angleOf(es[j].T)}).Curvature(0)
				cnt++
			}
			sides[i][k].K = sum / float64(cnt)
		}
	}
	for i := range sides {
		feasibleAuto(sides[i], s, s.Closed)
	}
	if s.Slant[1] > 0 {
		lattice(ns, &sides, s.Slant)
	}
	return ns, sides, nil
}

// buildRest is the part of Build after the edge data.
func (s Stroke) buildRest(r *Result, ns []Node, sides [2][]edgeData, chains [2]Chain) (*Result, error) {
	var err error
	right, left := chains[0], chains[1]
	if s.Closed {
		r.Right, r.Left = gridNodes(r.merge(right, false, false), true), gridNodes(r.merge(left, false, false), true)
		r.checkNodes(ns, true)
		r.checkEdges(r.Right, "right")
		r.checkEdges(r.Left, "left")
		r.Contours = []Chain{r.Right, r.Left.reversed()}
		return r, nil
	}
	// Short pieces are merged before the ends too, so that a clip continues a
	// curve of some length.
	right, left = r.merge(right, false, false), r.merge(left, false, false)
	// The end, then the start as the end of the reversed stroke.
	right, left, endCap, err := r.endCap(right, left, s.End, ns[len(ns)-1], sides[0][len(ns)-1], sides[1][len(ns)-1])
	if err != nil {
		return nil, err
	}
	last := reverseNode(ns[0])
	rr, ll, startCap, err := r.endCap(left.reversed(), right.reversed(), s.Start, last,
		reverseEdge(sides[1][0]), reverseEdge(sides[0][0]))
	if err != nil {
		return nil, err
	}
	right, left = ll.reversed(), rr.reversed()
	cs, ce := s.Start.Kind != Ball, s.End.Kind != Ball
	right, left = gridNodes(r.merge(right, cs, ce), false), gridNodes(r.merge(left, cs, ce), false)
	r.Right, r.Left = right, left
	r.checkNodes(ns, s.Closed)
	r.checkEdges(r.Right, "right")
	r.checkEdges(r.Left, "left")
	var c Chain
	c = append(c, right...)
	c = append(c, endCap...)
	c = append(c, left.reversed()...)
	c = append(c, startCap...)
	r.Contours = []Chain{c}
	return r, nil
}

func reverseNode(nd Node) Node {
	return Node{P: nd.P, Dir: nd.Dir + 180, K: -nd.K, W: nd.W, DW: -nd.DW, DDW: nd.DDW, Skew: nd.Skew}
}

func reverseEdge(e edgeData) edgeData { return edgeData{P: e.P, T: mul(e.T, -1), K: -e.K, Lat: e.Lat} }

// edgeChain returns the edge through the given data, one cubic between two
// nodes, with the extremes near a node moved onto it and the curves split at
// the other extremes.
func (r *Result) edgeChain(es []edgeData, closed bool) (Chain, error) {
	n := len(es)
	segs := n - 1
	if closed {
		segs = n
	}
	var notes []string
	var ch Chain
	for i := 0; i < segs; i++ {
		a, b := es[i], es[(i+1)%n]
		cs, ok := edgeCurves(a, b)
		if !ok {
			notes = append(notes, fmt.Sprintf("edge from %.1f,%.1f to %.1f,%.1f: no curve with both curvatures; tangent-continuous only", a.P.X, a.P.Y, b.P.X, b.P.Y))
		}
		fixed := (a.Lat != 0) != (b.Lat != 0)
		if fixed {
			// The handles on the stem's line of the curve
			// which touches the stem move to whole-unit points of it.
			if a.Lat != 0 {
				cs[0] = snapHandles(cs[0], a.Lat)
			} else {
				l := len(cs) - 1
				cs[l] = snapHandles(cs[l].Reversed(), b.Lat).Reversed()
			}
		}
		for _, c := range cs {
			if fixed && !c.IsStraight() {
				ch = append(ch, Piece{C: c, Fixed: true})
				continue
			}
			ch = append(ch, splitExtremes(c)...)
		}
	}
	for _, s := range notes {
		r.warn("%s", s)
	}
	return ch, nil
}

// edgeCurves returns the edge between the data a and b: one cubic with the
// points, directions and curvatures of both (hermite), or, where there is
// none or it has a handle shorter than minHandle, two, which meet at the middle of the curve with the handles of a
// circular arc, with its point and direction there and the first of these
// curvatures for which both exist: the curvature of that curve there, the
// mean of the curvatures at the ends, twice the first.  It reports
// false if the edge is only tangent-continuous.
func edgeCurves(a, b edgeData) ([]Cubic, bool) {
	// An edge which leaves a stem on the lattice runs out of it with its
	// handles on the lattice.
	if a.Lat != 0 && b.Lat == 0 && math.Abs(a.K) < 1e-9 {
		if cs, ok := runOut(a, b); ok {
			return cs, true
		}
	}
	if b.Lat != 0 && a.Lat == 0 && math.Abs(b.K) < 1e-9 {
		if cs, ok := runOut(reverseEdge(b), reverseEdge(a)); ok {
			return []Cubic{cs[1].Reversed(), cs[0].Reversed()}, true
		}
	}
	c, ok := hermite(a.P, a.T, a.K, b.P, b.T, b.K)
	if ok && shortHandle(c) >= minHandle {
		return []Cubic{c}, true
	}
	if math.Abs(a.K) < 1e-9 && math.Abs(b.K) > 1e-9 {
		if cs, ok := runOut(a, b); ok {
			return cs, true
		}
	}
	if math.Abs(b.K) < 1e-9 && math.Abs(a.K) > 1e-9 {
		if cs, ok := runOut(reverseEdge(b), reverseEdge(a)); ok {
			return []Cubic{cs[1].Reversed(), cs[0].Reversed()}, true
		}
	}
	// Two curves which meet on the curve with the handles of a circular arc,
	// at its middle (or else at 0.35 or 0.65 of its parameter), with its
	// point and direction there and, of a fixed list of curvatures, the one
	// for which both curves exist and lie closest to that curve.
	best := math.Inf(1)
	var bestCs []Cubic
	for _, tm := range []float64{0.5, 0.35, 0.65} {
		m, mt, mc := c.At(tm), c.Tangent(tm), c.Curvature(tm)
		for _, mk := range []float64{mc, (a.K + b.K) / 2, a.K, b.K, 0.5 * mc, 1.5 * mc, 2 * mc} {
			c1, ok1 := hermite(a.P, a.T, a.K, m, mt, mk)
			c2, ok2 := hermite(m, mt, mk, b.P, b.T, b.K)
			if !ok1 || !ok2 || math.Min(shortHandle(c1), shortHandle(c2)) < math.Min(minHandle, shortHandle(c)) {
				continue
			}
			if d := math.Max(deviation(c1, []Cubic{c}), deviation(c2, []Cubic{c})); d < best {
				best, bestCs = d, []Cubic{c1, c2}
			}
		}
		if bestCs != nil {
			break
		}
	}
	if bestCs != nil {
		return bestCs, true
	}
	return []Cubic{c}, ok
}

// runOut returns the edge from a, where it has no curvature, to b as a cubic
// B-spline with one inner knot, written as two cubics, whose control polygon
// has its first three points on the tangent at a, at runLow and runHigh of the
// way to the corner X where the tangents at a and b meet, and its fourth on the
// tangent at b, as far from b as gives the curvature of b there.  The edge
// leaves a without curvature, which then grows evenly: the shoulder of the
// arch of n in the font is drawn so.
func runOut(a, b edgeData) ([]Cubic, bool) {
	cs, ok := runOutAt(a, b, runLow, runHigh)
	if a.Lat == 0 || ok && !shortRunOut(cs) {
		return cs, ok
	}
	// On the lattice the control points move to whole-unit points, and the
	// fractions of the shoulders may then give no curve, or a curve shorter
	// than the rounding of the font keeps (which would merge it with its
	// neighbour and take its handles off the lattice): the next of a fixed
	// list of fractions is tried, and where none serves, the edge is one
	// curve (edgeCurves).
	for _, f := range [][2]float64{{0.15, 0.5}, {0.35, 0.8}, {0.1, 0.4}, {0.2, 0.9}, {0.3, 0.6}, {0.2, 0.55}} {
		if cs, ok := runOutAt(a, b, f[0], f[1]); ok && !shortRunOut(cs) {
			return cs, ok
		}
	}
	return nil, false
}

// shortRunOut reports whether one of the two curves of a run-out is shorter
// than the segments which the rounding of the font merges (outline.MinPiece).
func shortRunOut(cs []Cubic) bool {
	for _, c := range cs {
		if c.Length() < outline.MinPiece {
			return true
		}
	}
	// The handle at the curved end gives the curvature there, which the
	// rounding to whole units changes the more the shorter it is: below
	// roundHandle the rounding can take the curvature of the edge to zero.
	return dist(cs[1][2], cs[1][3]) < roundHandle
}

// roundHandle is the shortest handle at the curved end of a run-out on the
// lattice: the rounding of the font turns a smooth node whose handles are at
// least this long by at most 2.5 degrees (TestRoundingKeepsSmoothNodesSmooth).
const roundHandle = 8.0

// runOutAt is runOut with the second and third control points at the
// fractions low and high of the way to the corner.
func runOutAt(a, b edgeData, low, high float64) ([]Cubic, bool) {
	den := cross(a.T, b.T)
	if math.Abs(den) < 1e-9 {
		return nil, false
	}
	// X = a.P + u a.T = b.P - v b.T.
	d := sub(b.P, a.P)
	u := cross(d, b.T) / den
	v := cross(a.T, d) / den
	if u <= 0 || v <= 0 {
		return nil, false
	}
	q1 := add(a.P, mul(a.T, low*u))
	q2 := add(a.P, mul(a.T, high*u))
	if a.Lat != 0 && math.Abs(a.T.Y) > 1e-9 {
		// On a lattice line: the second and third points of the control
		// polygon, and the middle of the two, on whole-unit points of it.
		den := latticeDen(a.Lat)
		var s1, s2 float64
		q1, s1 = onLattice(a.P, a.T, low*u, den)
		q2, s2 = onLattice(a.P, a.T, high*u, 2*den)
		if s1 <= 0 {
			q1, s1 = onLattice(a.P, a.T, s1+den/math.Abs(a.T.Y), den)
		}
		// The middle of q1 and q2 is on the lattice when their heights
		// differ by a multiple of 2 den.
		if d := math.Mod(math.Abs(q2.Y-q1.Y), 2*den); d > 1e-9 && math.Abs(d-2*den) > 1e-9 {
			q2, s2 = onLattice(a.P, a.T, s2+den/math.Abs(a.T.Y), den)
		}
		if s2 <= s1 || s2 >= u {
			return nil, false
		}
	}
	build := func(h float64) (Cubic, Cubic) {
		q3 := sub(b.P, mul(b.T, h))
		m12, m23 := lerp(q1, q2, 0.5), lerp(q2, q3, 0.5)
		m := lerp(m12, m23, 0.5)
		return Cubic{a.P, q1, m12, m}, Cubic{m, m23, q3, b.P}
	}
	k := func(h float64) float64 { _, c2 := build(h); return c2.Curvature(1) }
	// The curvature at b falls as the handle grows.
	lo, hi := 1e-3*v, 3*v
	if (k(lo)-b.K)*(k(hi)-b.K) > 0 {
		return nil, false
	}
	for i := 0; i < 80; i++ {
		mid := (lo + hi) / 2
		if (k(mid)-b.K)*(k(lo)-b.K) > 0 {
			lo = mid
		} else {
			hi = mid
		}
	}
	c1, c2 := build((lo + hi) / 2)
	// The curve turns one way only.
	for _, c := range []Cubic{c1, c2} {
		for i := 1; i <= 16; i++ {
			if c.Curvature(float64(i)/16)*b.K < -1e-6 {
				return nil, false
			}
		}
	}
	return []Cubic{c1, c2}, true
}

// runLow and runHigh place the second and the third point of the control
// polygon of runOut, as in the shoulders of the font.
const (
	runLow  = 0.25
	runHigh = 0.65
)

// minHandle is the shortest handle, in font units, below which edgeCurves
// tries two curves: the rounding to whole units turns a shorter handle by
// several degrees.
const minHandle = 12.0

// shortHandle returns the length of the shorter handle of the curve.
func shortHandle(c Cubic) float64 { return math.Min(dist(c[0], c[1]), dist(c[2], c[3])) }

// splitExtremes returns the curve as pieces cut at its extremes, except those
// within half a unit of an end or of the cut before.  A short piece which this
// leaves next to a node is merged with the curve beyond the node (merge), so
// that the extreme becomes a node.
func splitExtremes(c Cubic) Chain {
	if c.IsStraight() {
		return Chain{{C: c}}
	}
	total := c.Length()
	var ts []float64
	for _, t := range c.Extrema() {
		l := c.lengthTo(t)
		if l >= 0.5 && total-l >= 0.5 {
			if len(ts) > 0 {
				prev := ts[len(ts)-1]
				if l-c.lengthTo(prev) < 0.5 {
					continue
				}
			}
			ts = append(ts, t)
		}
	}
	var out Chain
	prev := 0.0
	rest := c
	for _, t := range ts {
		a, b := rest.Split((t - prev) / (1 - prev))
		out = append(out, Piece{C: a})
		rest, prev = b, t
	}
	return append(out, Piece{C: rest})
}

func (r *Result) warn(format string, args ...any) {
	r.warnAt(Point{X: math.NaN(), Y: math.NaN()}, format, args...)
}

// warnAt adds a warning about the shape of an edge at the place q, which
// Letter.Build drops where another stroke of the letter covers it.
func (r *Result) warnAt(q Point, format string, args ...any) {
	r.Warnings = append(r.Warnings, fmt.Sprintf(format, args...))
	r.WarnAt = append(r.WarnAt, q)
}

// merge replaces each edge piece shorter than MinPiece, together with an edge
// neighbour at a smooth join, by one curve with the points, directions and
// curvatures of the two at their outer ends.  Of the two neighbours the one is
// taken whose merged curve keeps the extremes at its ends and lies closer; at
// an end of the chain which is a corner (cornerStart, cornerEnd), an extreme
// may stay inside the curve within MinPiece of the corner.  Straight pieces on
// one line are joined first.
func (r *Result) merge(ch Chain, cornerStart, cornerEnd bool) Chain {
	ch = joinLines(ch)
	stuck := map[Point]bool{}
	for iter := 0; iter < 20; iter++ {
		j := -1
		for i, p := range ch {
			if p.Kind == EdgePiece && !p.Fixed && !p.C.IsStraight() && p.C.Length() < MinPiece && !stuck[p.C[0]] {
				j = i
				break
			}
		}
		if j < 0 {
			return ch
		}
		best, bestA := math.Inf(1), -1
		var bestC Cubic
		for _, a := range []int{j - 1, j} {
			b := a + 1
			if a < 0 || b >= len(ch) || ch[a].Kind != EdgePiece || ch[b].Kind != EdgePiece || ch[a].Fixed || ch[b].Fixed {
				continue
			}
			ca, cb := ch[a].C, ch[b].C
			if math.Abs(angleBetween(ca.Tangent(1), cb.Tangent(0))) > 1 {
				continue
			}
			m, okm := hermite(ca[0], ca.Tangent(0), ca.Curvature(0), cb[3], cb.Tangent(1), cb.Curvature(1))
			if !okm {
				continue
			}
			// At an end of the edge, next to a corner, an extreme may stay
			// inside the curve within MinPiece of the corner.
			lo, hi := 1.0, 1.0
			// So may an extreme next to a straight piece, which the edge
			// leaves without a turn: under a slant the extreme of the round
			// lies a little inside the curve.
			if a == 0 && cornerStart || a > 0 && ch[a-1].C.IsStraight() {
				lo = MinPiece
			}
			if b == len(ch)-1 && cornerEnd || b < len(ch)-1 && ch[b+1].C.IsStraight() {
				hi = MinPiece
			}
			if innerExtreme(m, lo, hi) {
				continue
			}
			dev := deviation(m, []Cubic{ca, cb})
			if dev < best {
				best, bestA, bestC = dev, a, m
			}
		}
		if bestA < 0 {
			r.warnAt(ch[j].C[0], "a piece of %.1f units at %.1f,%.1f cannot be merged", ch[j].C.Length(), ch[j].C[0].X, ch[j].C[0].Y)
			stuck[ch[j].C[0]] = true
			continue
		}
		if best > 1 {
			r.warn("merging a short piece at %.1f,%.1f moves the edge by %.2f", ch[j].C[0].X, ch[j].C[0].Y, best)
		}
		out := append(Chain(nil), ch[:bestA]...)
		out = append(out, Piece{C: bestC})
		ch = append(out, ch[bestA+2:]...)
	}
	return ch
}

// gridReach is the distance along the edge, in font units, within which a
// smooth node moves to the point nearest to the grid of whole units.
const gridReach = 4.0

// gridNodes moves every smooth node between two curved edge pieces which is
// not at an extreme along the edge, by at most gridReach, to the point of the
// edge nearest to a point of whole units, and draws the two pieces again with
// the points, directions and curvatures of the old ones at their ends.  The
// rounding of the font then keeps the node and moves only the handles, which
// turns the edge less there.  A closed chain also moves the node between its
// last and its first piece.
func gridNodes(ch Chain, closed bool) Chain {
	n := len(ch)
	last := n - 1
	if closed {
		last = n
	}
	for i := 0; i < last; i++ {
		a, b := &ch[i], &ch[(i+1)%n]
		if a.Kind != EdgePiece || b.Kind != EdgePiece || a.C.IsStraight() || b.C.IsStraight() || a.Fixed || b.Fixed {
			continue
		}
		if math.Abs(angleBetween(a.C.Tangent(1), b.C.Tangent(0))) > 0.01 {
			continue
		}
		if _, axis := snapAxis(a.C.Tangent(1)); !axis {
			continue // at an extreme the rounding keeps the handles level
		}
		off := func(q Point) float64 {
			return math.Max(math.Abs(q.X-math.Round(q.X)), math.Abs(q.Y-math.Round(q.Y)))
		}
		// Near the node the parameter runs evenly with the length.
		va, vb := norm(a.C.D(1)), norm(b.C.D(0))
		la, lb := a.C.Length(), b.C.Length()
		at := func(s float64) (Cubic, float64, bool) {
			switch {
			case s < 0 && -s < la/3:
				return a.C, 1 + s/va, true
			case s > 0 && s < lb/3:
				return b.C, s / vb, true
			}
			return Cubic{}, 0, false
		}
		// The candidates, nearest to the grid first; the first whose two
		// curves exist, lie close to the old ones and keep their extremes at
		// the ends is taken.
		type cand struct{ off, s float64 }
		var cs []cand
		for s := -gridReach; s <= gridReach; s += 0.02 {
			if c, t, ok := at(s); ok {
				if o := off(c.At(t)); o < off(a.C[3])-0.05 {
					cs = append(cs, cand{o, s})
				}
			}
		}
		sort.Slice(cs, func(i, j int) bool { return cs[i].off < cs[j].off })
		var ca, cb Cubic
		found := false
		for k, cd := range cs {
			if k >= 12 {
				break
			}
			c, t, _ := at(cd.s)
			m, mt, mk := c.At(t), c.Tangent(t), c.Curvature(t)
			var ok1, ok2 bool
			ca, ok1 = hermite(a.C[0], a.C.Tangent(0), a.C.Curvature(0), m, mt, mk)
			cb, ok2 = hermite(m, mt, mk, b.C[3], b.C.Tangent(1), b.C.Curvature(1))
			if ok1 && ok2 && deviation(ca, []Cubic{a.C, b.C}) <= 0.25 && deviation(cb, []Cubic{a.C, b.C}) <= 0.25 &&
				!innerExtreme(ca, 1, 1) && !innerExtreme(cb, 1, 1) {
				found = true
				break
			}
		}
		if !found {
			continue
		}
		a.C, b.C = ca, cb
	}
	return ch
}

// snapAxis returns the axis direction nearest to t, and whether t is more
// than 0.05 degrees from it.
func snapAxis(t Point) (Point, bool) {
	a := angleOf(t)
	q := math.Round(a/90) * 90
	return dirOf(q), math.Abs(a-q) > 0.05
}

// joinLines returns the chain with consecutive straight edge pieces on one
// line joined into one.
func joinLines(ch Chain) Chain {
	var out Chain
	for _, pc := range ch {
		if n := len(out); n > 0 && pc.Kind == EdgePiece && out[n-1].Kind == EdgePiece &&
			pc.C.IsStraight() && out[n-1].C.IsStraight() {
			a := out[n-1].C
			j := Cubic{a[0], lerp(a[0], pc.C[3], 1.0/3), lerp(a[0], pc.C[3], 2.0/3), pc.C[3]}
			if j.IsStraight() && math.Abs(angleBetween(a.Tangent(1), pc.C.Tangent(0))) < 1e-6 {
				out[n-1].C = j
				continue
			}
		}
		out = append(out, pc)
	}
	return out
}

// maxTurn is the largest turn of the centre line, in degrees, between two
// nodes: a quarter turn and what a slant of 15 degrees adds to it, which a
// cubic with the end data of the edges draws evenly.  A larger turn needs a
// node in between.
const maxTurn = 110.0

// checkNodes warns where two neighbouring nodes are too far apart in
// direction for the rules: where the centre line turns by more than maxTurn
// between them, or where it turns one way at the nodes and the other between
// them (the chord leaves both nodes on the same side), which asks for an
// inflection that the nodes do not place.
func (r *Result) checkNodes(ns []Node, closed bool) {
	n := len(ns)
	last := n - 1
	if closed {
		last = n
	}
	for i := 0; i < last; i++ {
		a, b := ns[i], ns[(i+1)%n]
		turn := math.Abs(angleBetween(dirOf(a.Dir), dirOf(b.Dir)))
		if turn > maxTurn {
			r.warn("the centre turns by %.0f degrees between the nodes at %.0f,%.0f and %.0f,%.0f: add a node between them", turn, a.P.X, a.P.Y, b.P.X, b.P.Y)
		}
	}
}

// Wiggle limits of checkEdges: curvature below wiggleFloor (1/unit) has no
// sign, and a dip of the curvature between two higher values counts when it
// falls below wiggleDip times the smaller of them.
const (
	wiggleFloor = 0.0005
	wiggleDip   = 0.6
)

// checkEdges warns where a curve of the edge, drawn by the rules, has an
// inflection although its ends curve the same way, or a dip of its curvature
// between two higher values: the shape the nodes ask for needs another node
// there, or other values at the nodes.  The warning names the place.
func (r *Result) checkEdges(ch Chain, side string) {
	for _, pc := range ch {
		if pc.Kind != EdgePiece || pc.C.IsStraight() {
			continue
		}
		const n = 48
		ks := make([]float64, n+1)
		for i := range ks {
			ks[i] = pc.C.Curvature(float64(i) / n)
		}
		k0, k1 := ks[0], ks[n]
		if k0*k1 > 0 && math.Abs(k0) > wiggleFloor && math.Abs(k1) > wiggleFloor {
			for i := 1; i < n; i++ {
				if ks[i]*k0 < 0 && math.Abs(ks[i]) > wiggleFloor {
					q := pc.C.At(float64(i) / n)
					r.warnAt(q, "the %s edge inflects at %.0f,%.0f between nodes which curve the same way: add a node or change the curvatures", side, q.X, q.Y)
					break
				}
			}
		}
		// A dip: an interior minimum of |k| well below the maxima on both
		// sides of it.
		for i := 1; i < n; i++ {
			a := math.Abs(ks[i])
			if !(a <= math.Abs(ks[i-1]) && a <= math.Abs(ks[i+1])) {
				continue
			}
			var l, rr float64
			for j := 0; j < i; j++ {
				l = math.Max(l, math.Abs(ks[j]))
			}
			for j := i + 1; j <= n; j++ {
				rr = math.Max(rr, math.Abs(ks[j]))
			}
			if m := math.Min(l, rr); m > 4*wiggleFloor && a < wiggleDip*m && (l > math.Abs(ks[0])+wiggleFloor || rr > math.Abs(ks[n])+wiggleFloor) {
				q := pc.C.At(float64(i) / n)
				r.warnAt(q, "the curvature of the %s edge dips at %.0f,%.0f: add a node there or change the curvatures", side, q.X, q.Y)
				break
			}
		}
	}
}

// innerExtreme reports whether the curve has an extreme more than lo from its
// start and more than hi from its end.
func innerExtreme(c Cubic, lo, hi float64) bool {
	l := c.Length()
	for _, t := range c.Extrema() {
		if s := c.lengthTo(t); s > lo && l-s > hi {
			return true
		}
	}
	return false
}

// angleBetween returns the angle from u to v in degrees.
func angleBetween(u, v Point) float64 {
	return math.Atan2(cross(u, v), dot(u, v)) * 180 / math.Pi
}

// deviation returns the largest distance of points of m from the curves cs.
func deviation(m Cubic, cs []Cubic) float64 {
	var pts []Point
	for _, c := range cs {
		for i := 0; i <= 64; i++ {
			pts = append(pts, c.At(float64(i)/64))
		}
	}
	var worst float64
	for i := 0; i <= 64; i++ {
		q := m.At(float64(i) / 64)
		best := math.Inf(1)
		for k := 1; k < len(pts); k++ {
			best = math.Min(best, segDist(q, pts[k-1], pts[k]))
		}
		worst = math.Max(worst, best)
	}
	return worst
}

// segDist returns the distance of q from the segment ab.
func segDist(q, a, b Point) float64 {
	d := sub(b, a)
	l2 := dot(d, d)
	if l2 == 0 {
		return dist(q, a)
	}
	t := math.Max(0, math.Min(1, dot(sub(q, a), d)/l2))
	return dist(q, add(a, mul(d, t)))
}
