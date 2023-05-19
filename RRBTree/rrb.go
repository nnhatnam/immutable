package RRBTree

import "github.com/nnhatnam/immutable/slice"

type refValue[V any] struct {
	value V
}

func newRefValue[V any](value V) *refValue[V] {
	return &refValue[V]{
		value: value,
	}
}

type node[V any] struct {
	sizes []int // The size of each child.

	children []*node[V]     // The children of the node.
	values   []*refValue[V] // The values of the node.

	owner *RRBTree[V] // The transient owner of the node. for persistent, it's nil.
}

// descriptorClone doesn't really clone the node. It just returns a new node with the same children and values.
// Any modification to the children and values of the clone will affect the original node.
//func (n *node[V]) descriptorClone() *node[V] {
//
//	if n == nil {
//		return nil
//	}
//	clone := &node[V]{
//		cumulativeSize: n.cumulativeSize,
//		children:       n.children,
//		values:         n.values,
//	}
//
//	return clone
//}

func (n *node[V]) isBalancedNode() bool {
	return n.sizes == nil
}

// caller must not modify the returned slice.
func (n *node[V]) getSizes(h int) []int {
	if n.sizes != nil {
		return n.sizes
	}

	return cumulativeSumTable[h][:len(n.children)]

}

func (n *node[V]) getLength(h int) int {
	if n.sizes != nil {
		return n.sizes[len(n.sizes)-1]
	}

	return cumulativeSumTable[h][len(n.children)-1]

}

//func (n *node[V]) isFull() bool {
//
//	return len(n.children) == branchFactor
//}

//func (n *node[V]) size() int {
//	if len(n.children) == 0 {
//		return len(n.values)
//	}
//
//	return n.cumulativeSize[len(n.children)-1]
//
//}

func newNode[V any]() *node[V] {
	return &node[V]{
		//children: make([]*node[V], 0, branchFactor),
	}
}

type RRBTree[V any] struct {
	// The root node of the tree.
	root *node[V]

	h int // The current level of the trie. It's counted from the bottom of the trie.

	size int // The number of elements in the tree.

	head []*refValue[V] // The head of the tree.

	tail []*refValue[V] // The tail of the tree.

}

func NewRRBTree[V any]() RRBTree[V] {

	return RRBTree[V]{}

}

//func NewRRBTreeMutable[V any]() *RRBTree[V] {
//
//	rrb := &RRBTree[V]{
//		root: nil,
//		size: 0,
//		tail: nil,
//	}
//
//	return rrb
//
//}

//func (t *RRBTree[V]) newNode() *node[V] {
//	return &node[V]{
//		//children: make([]refChild[V], 0, branchFactor),
//
//		owner: t,
//	}
//}

//func (t *RRBTree[V]) getTailNode() *node[V] {
//
//	if t.root == nil {
//		return nil
//	}
//
//	n := t.root // start from root
//	for j := t.h; j > 0; j-- {
//		n = n.children[len(n.children)-1]
//	}
//
//	return n
//
//}

// https://github.com/golang/go/wiki/CodeReviewComments#receiver-type
// pass node as value to make sure it is always shallow copied.
// caller must make sure that n is not nil or lese it will panic with nil pointer dereference.
// it is very useful for persistent data structure.
func (t RRBTree[V]) pushTail(n node[V], h int, index int, tail []*refValue[V]) *node[V] {

	if h == 0 {

		return &node[V]{
			values: tail,
		}
	}

	if n.isBalancedNode() {
		position := t.size - len(t.tail)
		position = (position >> shiftTable[h]) & maskTable[h]

		if position > len(n.children) {
			//push
			if len(n.children) == maxBranches {

				parent := t.pushTail(node[V]{
					children: []*node[V]{&n},
				}, h+1, index, tail)

				return parent
			}

			n.children = slice.Push(n.children, t.pushTail(node[V]{}, h-1, index, tail))
		}

		n.children[position] = t.pushTail(*n.children[position], h-1, index, tail)
		return &n
	}

	position := findPosition(n.sizes, index)

	if position > len(n.children) {
		//push
		if len(n.children) == maxBranches {

			parent := t.pushTail(node[V]{
				children: []*node[V]{&n},
			}, h+1, index, tail)

			return parent
		}

		n.children = slice.Push(n.children, t.pushTail(node[V]{}, h-1, index, tail))
	}

	index -= n.sizes[position-1]

	if len(n.children) == maxBranches {

		parent := t.pushTail(node[V]{
			children: []*node[V]{&n},
		}, h+1, index, tail)

		return parent
	}

	n.children[position] = t.pushTail(*n.children[position], h-1, index, tail)
	return &n

}

func (t RRBTree[V]) push(value V) RRBTree[V] {
	// Value receiver makes a copy of the type and pass it to the function.
	// The function stack now holds an equal object but at a different location on memory.
	// That means any changes done on the passed object will remain local to the method.
	// The original object will remain unchanged.

	if t.tail == nil {
		t.tail = make([]*refValue[V], 0, maxBranches)
	}

	t.tail = append(t.tail, newRefValue[V](value))
	t.size++

	if len(t.tail) == maxBranches {
		// make a new branch
		t.root = t.pushTail(*t.root, t.h, t.size-len(t.tail), t.tail)
		t.tail = nil
	}

	return t

}

func (t RRBTree[V]) popLast(n node[V], h int) (n1 *node[V], value *refValue[V], tail []*refValue[V]) {

	if h == 0 {
		n.values, value = slice.Pop(n.values)

		if len(n.values) < minBranches {
			tail = n.values
			return nil, value, n.values

		}
		return &n, value, nil
	}

	n1, value, tail = t.popLast(*n.children[len(n.children)-1], h-1)

	if n1 == nil {
		n.children, _ = slice.Pop(n.children)
		return &n, value, tail
	}

	n.children = slice.Set(n.children, len(n.children)-1, n1)
	return &n, value, tail

}

func (t RRBTree[V]) pop() (rrb *RRBTree[V], value *refValue[V], success bool) {

	if t.size == 0 {
		return
	}

	if len(t.tail) > 0 {
		t.tail, value = slice.Pop(t.tail)
		t.size--
		return &t, value, true
	}

	//case: empty tail
	t.root, value, t.tail = t.popLast(*t.root, t.h)

	return &t, value, true
}

// leftSlice doesn't include the index
func leftSlice[V any](n node[V], h int, index int) (left *node[V], tail []*refValue[V], newSize int) {

	if h == 0 {

		values := slice.SelectRange(n.values, 0, index)

		if len(values) < minBranches {
			return nil, values, 0
		}

		return &node[V]{
			values: values,
		}, nil, len(values)

	}

	var position int
	if n.isBalancedNode() {
		// the rest will be balanced
		position = index & maskTable[0]
		var n1 *node[V]
		n1, tail, newSize = leftSlice(*n.children[position], h-1, index>>shiftTable[1])

		if n1 == nil {
			n.children = slice.SelectRange(n.children, 0, position)
			return &n, tail, 0
		}

		n.children = slice.SelectRange(n.children, 0, position+1)
		n.children[position] = n1

		size := 1 << (mFactor * h)
		if newSize < size {

			n.sizes = make([]int, position)
			cumulative := size
			for i := 0; i < position; i++ {
				n.sizes[i] = cumulative
				cumulative += size
			}
			cumulative = cumulative + newSize
			n.sizes[position] = cumulative
			return &n, tail, 0

		}

		return &n, tail, newSize

	}

	// unbalanced node

	position = findPosition(n.sizes, index)

	var n1 *node[V]
	n1, tail, newSize = leftSlice(*n.children[position], h-1, position)

	if n1 == nil {
		n.children = slice.SelectRange(n.children, 0, position)
		n.sizes = slice.SelectRange(n.sizes, 0, position)
		return &n, tail, 0
	}

	n.children = slice.SelectRange(n.children, 0, position+1)
	n.sizes = slice.SelectRange(n.sizes, 0, position+1)

	n.children[position] = n1

	if len(n.sizes) > 1 {
		n.sizes[position] = n.sizes[position-1] + newSize
	} else {
		n.sizes[position] = newSize
	}

	return &n, tail, newSize

}

func rightSlice[V any](n node[V], h, index int) (right *node[V], newSize int) {
	if h == 0 {

		n.children = slice.SelectRange(n.children, index, len(n.values)+1)

		return &n, len(n.children)
	}

	var position int
	if n.isBalancedNode() {

		position = index & maskTable[0]
		var n1 *node[V]
		n1, newSize = rightSlice(*n.children[position], h-1, index>>shiftTable[1])

		n.children = slice.SelectRange(n.children, position, len(n.children)+1)
		n.children[0] = n1

		size := 1 << (mFactor * h)
		n.sizes = make([]int, len(n.children))

		cumulative := newSize
		for i := 0; i < len(n.children); i++ {
			n.sizes[i] = cumulative
			cumulative += size
		}

		return &n, newSize

	}

	// unbalanced node

	position = findPosition(n.sizes, index)

	var n1 *node[V]
	n1, newSize = rightSlice(*n.children[position], h-1, position)

	n.children = slice.SelectRange(n.children, position, len(n.children)+1)
	n.sizes = slice.SelectRange(n.sizes, position, len(n.sizes)+1)

	n.children[position] = n1

	for k, v := range n.sizes {
		n.sizes[k] = v - index
	}

	return &n, newSize
}

func (t RRBTree[V]) clone() RRBTree[V] {
	return t
}

func (t RRBTree[V]) split(i int) (_ RRBTree[V], _ RRBTree[V]) {
	if i > t.size {
		panic("Index out of bounds")
	}

	var newSize int
	right := t.clone()
	right.root, newSize = rightSlice(*right.root, right.h, i)
	right.size = t.size - newSize

	t.root, t.tail, newSize = leftSlice(*t.root, t.h, i)
	t.size = newSize + 1

	return t, right

}

//func (t RRBTree[V]) findLastNode() *node[V] {
//
//	n := t.root // start from root
//	for j := t.h; j > 0; j-- {
//		n = n.children[len(n.children)-1]
//	}
//	return n
//}

func findPosition(sizes []int, idx int) int {
	i := 0
	for i < len(sizes) && sizes[i] <= idx {
		i++
	}

	return i
}

func calcIndex(sizes []int, slot, idx int) int {
	i := slot
	for i < len(sizes) && sizes[i] <= idx {
		i++
	}

	return i
}

func cloneLastBranch[V any](n node[V], h int, nodes []*node[V]) {

	nodes[h-1] = &n
	cloneLastBranch(*n.children[len(n.children)-1], h-1, nodes)

}

func walkLastBranch[V any](n *node[V], h int, nodes []*node[V]) {

	nodes[h-1] = n
	walkLastBranch(n.children[len(n.children)-1], h-1, nodes)

}

func walkFirstBranch[V any](n *node[V], h int, nodes []*node[V]) {

	nodes[h-1] = n
	walkFirstBranch(n.children[0], h-1, nodes)

}

func cloneFirstBranch[V any](n node[V], h int, nodes []*node[V]) {

	nodes[h-1] = &n
	cloneFirstBranch(*n.children[0], h-1, nodes)

}

func redistributed[V any](left, right node[V], h int) (newL *node[V], newR *node[V]) {

	tot := len(left.children) + len(right.children)

	if tot <= maxBranches {
		// merge into one node
		if left.isBalancedNode() && right.isBalancedNode() {
			left.children = slice.Concat(left.children, right.children)
			return &left, nil
		}

		i := len(left.children)
		mergedSizes := make([]int, tot)
		copy(mergedSizes, left.getSizes(h))
		copy(mergedSizes[i:], right.getSizes(h))

		for j := i; j < len(mergedSizes); j++ {
			mergedSizes[j] += mergedSizes[i-1]
		}

		left.sizes = mergedSizes
		left.children = slice.Concat(left.children, right.children)

		return &left, nil

	}

	// merge into two nodes

	ls := left.getSizes(h)
	rs := right.getSizes(h)

	i := len(left.children)
	j := maxBranches - i

	mergedSizes := make([]int, maxBranches)
	copy(mergedSizes, ls)
	copy(mergedSizes[i:], rs)

	for k := i; k < len(mergedSizes); k++ {
		mergedSizes[k] += mergedSizes[i-1]
	}

	left.sizes = mergedSizes
	left.children = slice.Concat(left.children, right.children[:j])

	if !right.isBalancedNode() {
		right.sizes = slice.Apply(rs[j:], func(i, e int) int { return e - rs[j-1] })
	}

	right.children = right.children[j:]

	return &left, &right

}

func merge[V any](left node[V], right node[V], h int) (newL, newR *node[V]) {

	if h == 1 {
		newL, newR = redistributed(left, right, h)
		return
	}

	//left.children[len(left.children) - 1], right.children[0] = merge(left, right , h -1)
	newL, newR = merge(left, right, h-1)

	if newL.isBalancedNode() {
		i := len(left.children) - 1
		if !left.isBalancedNode() {
			left.sizes = slice.Set(left.sizes, i, left.sizes[i]+newL.getLength(h))
		}

		left.children = slice.Set(left.children, i, newL)

	}

	if newR == nil {

		//if !right.isBalancedNode() {
		//	right.sizes = slice.Apply(right.sizes[1:], func(i, e int) int { return e - right.sizes[0] })
		//}
		right.sizes = slice.Apply(right.sizes[1:], func(i, e int) int { return e - right.sizes[0] })
		right.children, _ = slice.Pop(right.children)

	}

	if !newR.isBalancedNode() {
		shiftedCount := len(right.children) - len(newR.children)

		if !right.isBalancedNode() {
			right.sizes = slice.Apply(right.sizes[shiftedCount:], func(i, e int) int { return e - right.sizes[shiftedCount-1] })
		} else {
			right.sizes = slice.SelectRange(right.getSizes(h), shiftedCount, len(right.children))
		}

		right.children = slice.SelectRange(right.children, shiftedCount, len(right.children))

		return &left, &right
	}

	return &left, newR

}

func mergeLeaf[V any](left node[V], right node[V]) (mergedRoot *node[V]) {
	tot := len(left.values) + len(right.values)
	switch {
	case tot < maxBranches:
		mergedRoot = &node[V]{
			sizes:  []int{tot},
			values: slice.Concat(left.values, right.values),
		}
	case tot == maxBranches:
		mergedRoot = &node[V]{
			values: slice.Concat(left.values, right.values),
		}
	case tot < 2*maxBranches:
		l := make([]*refValue[V], maxBranches)
		copy(l, left.values)
		remain := maxBranches - len(left.values)
		copy(l[len(left.values):], right.values[:remain])
		r := make([]*refValue[V], len(right.values)-remain)
		copy(r, right.values[remain:])

		left.values = l
		right.values = r

		mergedRoot = &node[V]{
			sizes:    []int{len(left.values), len(right.values)},
			children: []*node[V]{&left, &right},
		}
	case tot == 2*maxBranches:
		mergedRoot = &node[V]{
			children: []*node[V]{&left, &right},
		}
	default:
		panic("Not implemented")

	}
	return
}

func (t RRBTree[V]) concatenate(other RRBTree[V]) {
	if other.root == nil {
		t.tail = slice.Concat(t.tail, other.tail)
		index := t.size - len(t.tail)

		if len(t.tail) > maxBranches {
			tail := slice.SelectRange(t.tail, 0, maxBranches)
			t.tail = slice.SelectRange(t.tail, maxBranches, len(t.tail))

			t.pushTail(*t.root, t.h, index, tail)
		}

		t.size += other.size
		return
	}
}

func get[V any](t *RRBTree[V], i int) V {
	if i > t.size {
		panic("Index out of bounds")
	}
	n := t.root

	for j := t.h; j > 0; i-- {

		idx := (i >> shiftTable[j]) & maskTable[j]
		if n.sizes != nil {
			idx = findPosition(n.sizes, i)
			i -= n.sizes[j-1]
		}
		n = n.children[idx]
		j--
	}

	return n.values[i].value
}

func (t RRBTree[V]) Get(i int) V {

	if i > t.size {
		panic("Index out of bounds")
	}
	n := t.root

	for j := t.h; j > 0; i-- {

		idx := (i >> shiftTable[j]) & maskTable[j]
		if n.sizes != nil {
			idx = findPosition(n.sizes, i)
			i -= n.sizes[j-1]
		}
		n = n.children[idx]
		j--
	}

	return n.values[i].value

}

func (t RRBTree[V]) Size() int {

	return t.size
}
