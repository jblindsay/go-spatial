// Kdtree is a very simple K-D tree implementation.
// This implementation uses a fixed value for K.  The intention
// is to copy the code locally, change K to your needs, and
// change T.Data's type to suit your needs too.
package structures

import (
	"bytes"
	"fmt"
	"sort"
)

// K is the dimensionality of the points in this package's K-D trees.
const K = 2

// A Point is a location in K-dimensional space.
type Point [K]float64

// SqDist returns the square distance between two points.
func (a *Point) sqDist(b *Point) float64 {
	sqDist := 0.0
	for i, x := range a {
		diff := x - b[i]
		sqDist += diff * diff
	}
	return sqDist
}

// A T is a the node of a K-D tree.  A *T is the root of a K-D tree,
// and nil is an empty K-D tree.
type T struct {
	// Point is the K-dimensional point associated with the
	// data of this node.
	Point
	// Data is auxiliary data associated with the point of this node.
	Data interface{}

	split       int
	left, right *T
}

func (t T) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("Point Coordinates: ")
	buffer.WriteString(fmt.Sprintf("(%v, %v)\n", t.Point[0], t.Point[1]))
	buffer.WriteString("Point Data: ")
	s := fmt.Sprintf("%v", t.Data) //t.Data.(string)
	buffer.WriteString(s + "\n")
	return buffer.String()
}

// Insert returns a new K-D tree with the given node inserted.
// Inserting a node that is already a member of a K-D tree
// invalidates that tree.
func (t *T) Insert(n *T) *T {
	return t.insert(0, n)
}

func (t *T) insert(depth int, n *T) *T {
	if t == nil {
		n.split = depth % K
		n.left, n.right = nil, nil
		return n
	}
	if n.Point[t.split] < t.Point[t.split] {
		t.left = t.left.insert(depth+1, n)
	} else {
		t.right = t.right.insert(depth+1, n)
	}
	return t
}

// InRange appends all nodes in the K-D tree that are within a given
// distance from the given point to the given slice, which may be nil.
// To  avoid allocation, the slice can be pre-allocated with a larger
// capacity and re-used across multiple calls to InRange.
func (t *T) InRange(pt Point, dist float64, nodes []*T) []*T {
	if dist < 0 {
		return nodes
	}
	return t.inRange(&pt, dist, nodes)
}

func (t *T) inRange(pt *Point, r float64, nodes []*T) []*T {
	if t == nil {
		return nodes
	}

	diff := pt[t.split] - t.Point[t.split]

	thisSide, otherSide := t.right, t.left
	if diff < 0 {
		thisSide, otherSide = t.left, t.right
		diff = -diff // abs
	}
	nodes = thisSide.inRange(pt, r, nodes)
	if diff <= r {
		if t.Point.sqDist(pt) < r*r {
			nodes = append(nodes, t)
		}
		nodes = otherSide.inRange(pt, r, nodes)
	}

	return nodes
}

// Height returns the height of the K-D tree.
func (t *T) Height() int {
	if t == nil {
		return 0
	}
	ht := t.left.Height()
	if rht := t.right.Height(); rht > ht {
		ht = rht
	}
	return ht + 1
}

// New returns a new K-D tree built using the given nodes.
// Building a new tree with nodes that are already members of
// K-D trees invalidates those trees.
func New(nodes []*T) *T {
	if len(nodes) == 0 {
		return nil
	}
	return buildTree(0, preSort(nodes))
}

// BuildTree returns a new tree, built up from the given slice of nodes.
func buildTree(depth int, nodes *preSorted) *T {
	split := depth % K
	switch nodes.Len() {
	case 0:
		return nil
	case 1:
		nd := nodes.cur[0][0]
		nd.split = split
		nd.left, nd.right = nil, nil
		return nd
	}
	cur, left, right := nodes.splitMed(split)
	cur.split = split
	cur.left = buildTree(depth+1, &left)
	cur.right = buildTree(depth+1, &right)
	return cur
}

// PreSorted holds the nodes pre-sorted on each dimension.
type preSorted struct {
	// Cur is the currently sorted set of *Ts.
	cur [K][]*T

	// Next contains slices that will be used in the results
	// of splitting a preSorted.
	next [K][]*T
}

// PreSort returns the nodes pre-sorted on each dimension.
func preSort(nodes []*T) *preSorted {
	p := new(preSorted)
	for i := range p.cur {
		p.cur[i] = make([]*T, len(nodes))
		p.next[i] = make([]*T, len(nodes))
		copy(p.cur[i], nodes)
		sort.Sort(&nodeSorter{i, p.cur[i]})
	}
	return p
}

// Len returns the number of nodes.
func (p *preSorted) Len() int {
	return len(p.cur[0])
}

// SplitMed returns the median node on the split dimension and two
// preSorted structs that contain the nodes (still sorted on each
// dimension) that are less than and greater than or equal to the
// median node value on the given splitting dimension.
//
// The target of splitMed becomes invalid after the split, as its memory
// is hijacked by the two returned partitions.
func (p *preSorted) splitMed(dim int) (med *T, left, right preSorted) {
	m := len(p.cur[dim]) / 2
	for m > 0 && p.cur[dim][m-1] == p.cur[dim][m] {
		m--
	}
	med = p.cur[dim][m]
	pivot := med.Point[dim]
	nleft := leftSize(pivot, dim, p.cur)
	for d := range p.cur {
		// Use p's next slices as left and right's cur slices.
		left.cur[d] = p.next[d][:0]
		right.cur[d] = p.next[d][nleft+1 : nleft+1]

		for _, n := range p.cur[d] {
			if n == med {
				continue
			}
			if n.Point[dim] <= pivot {
				left.cur[d] = append(left.cur[d], n)
			} else {
				right.cur[d] = append(right.cur[d], n)
			}
		}

		// Re-use p's cur slice as left and right's next slices.
		left.next[d] = p.cur[d][:nleft]
		if nleft+1 < len(p.cur[d])-1 {
			right.next[d] = p.cur[d][nleft+1:]
		} else {
			right.next[d] = nil
		}
	}
	return
}

func leftSize(pivot float64, d int, nodes [K][]*T) int {
	var nleft int
	for _, n := range nodes[d] {
		if n.Point[d] <= pivot {
			nleft++
		}
	}
	// Minus 1 because the median point isn't placed down the left branch.
	return nleft - 1
}

// A nodeSorter implements sort.Interface, sortnig the nodes
// in ascending order of their point values on the split dimension.
type nodeSorter struct {
	split int
	nodes []*T
}

func (n *nodeSorter) Len() int {
	return len(n.nodes)
}

func (n *nodeSorter) Swap(i, j int) {
	n.nodes[i], n.nodes[j] = n.nodes[j], n.nodes[i]
}

func (n *nodeSorter) Less(i, j int) bool {
	return n.nodes[i].Point[n.split] < n.nodes[j].Point[n.split]
}
