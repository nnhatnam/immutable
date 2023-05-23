package RRBTree

import (
	"github.com/nnhatnam/immutable/slice"
)

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

func (n *node[V]) reachedMaxBranch() bool {
	return len(n.children) == maxBranches
}

// caller must not modify the returned slice.
func (n *node[V]) getSizeTable(h int) []int {
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

//func pushTailOnBalancedBranch[V any](n node[V], h int, index int, tail []*refValue[V]) (*node[V], int) {
//	if h == 0 {
//		if len(n.values) == 0 {
//			return &node[V]{
//				values: tail,
//			}, 0
//		}
//
//		sibling := node[V]{
//			values: tail,
//		}
//
//		return &node[V]{
//			children: []*node[V]{&n, &sibling},
//		}, 1
//	}
//
//	//var slot int // the slot to insert the new child
//	//slot, index = navigate[V](&n, h, index)
//	if index >= (1 << ((h + 1) * mFactor)) {
//		if h == 7 {
//			panic("index out of bound")
//		}
//		parent := &node[V]{
//			children: []*node[V]{&n, nil /*sibling*/},
//		}
//		parent.children[1], _ = pushTailOnBalancedBranch(node[V]{}, h, 0, tail)
//		return parent, h + 1
//	}
//
//	//var slot int // the slot to insert the new child
//	_, index = navigate[V](&n, h, index)
//
//	child, _ := pushTail(node[V]{}, h-1, index, tail)
//	n.children = slice.Push(n.children, child)
//	return &n, h
//}

//func checkSubTreePushable[V any](n *node[V], h int) bool {
//	if h == 1 {
//		if len(n.children) == maxBranches {
//			return false
//		}
//		return true
//	}
//
//	return checkSubTreePushable(n.children[len(n.children)-1], h-1)
//}

func walkFirstBranch[V any](n node[V], h int, nodes []*node[V]) {
	nodes[h] = &n
	if h != 0 {
		walkFirstBranch(*n.children[0], h-1, nodes)
	}
	return

}

func walkLastBranch[V any](n node[V], h int, nodes []*node[V]) {
	nodes[h] = &n
	if h != 0 {
		walkLastBranch(*n.children[len(n.children)-1], h-1, nodes)
	}
	return
}

//func pushHead[V any](n node[V], h int, index int, head []*refValue[V]) (*node[V], int) {
//
//	nodes := make([]*node[V], h+1)
//	walkLastBranch(n, h, nodes)
//
//
//}
//
//func pushTail1[V any](n node[V], h int, position int, tail []*refValue[V]) (*node[V], int) {
//
//	if h == 0 {
//		if len(n.values) == 0 {
//			return &node[V]{
//				values: tail,
//			}, h
//		}
//
//		sibling := node[V]{
//			values: tail,
//		}
//
//		parent := &node[V]{
//			children: []*node[V]{&n, &sibling},
//		}
//
//		if len(n.values) != maxBranches {
//			//build sizes table
//			parent.sizes = []int{len(n.values), len(n.values) + len(sibling.values)}
//
//		}
//
//		return parent, h + 1
//
//	}
//
//	if n.isBalancedNode() {
//
//		//position := t.size - len(t.tail)
//
//		var slot int // the slot to insert the tail
//		slot, position = navigate[V](&n, h, position)
//
//		//index := (position >> shiftTable[0]) & maskTable[0]
//		//position = position & maskTable[1]
//
//		if slot > len(n.children) {
//			//push
//			if len(n.children) == maxBranches {
//
//				parent, _ := pushTail(node[V]{
//					children: []*node[V]{&n},
//				}, h+1, position, tail)
//
//				return parent, h + 1
//			}
//
//			child, _ := pushTail(node[V]{}, h-1, position, tail)
//			n.children = slice.Push(n.children, child)
//		}
//
//		n.children[position], _ = pushTail(*n.children[position], h-1, position, tail)
//		return &n, h
//	}
//
//	index := findPosition(n.sizes, position)
//	position = position - n.sizes[index-1]
//
//	if index > len(n.children) {
//		//push
//		if len(n.children) == maxBranches {
//
//			parent, _ := pushTail(node[V]{
//				children: []*node[V]{&n},
//			}, h+1, position, tail)
//
//			return parent, h + 1
//		}
//
//		child, _ := pushTail(node[V]{}, h-1, index, tail)
//		n.children = slice.Push(n.children, child)
//	}
//
//	//index -= n.sizes[position-1]
//
//	if len(n.children) == maxBranches {
//
//		parent, _ := pushTail(node[V]{
//			children: []*node[V]{&n},
//		}, h+1, position, tail)
//
//		return parent, h + 1
//	}
//
//	n.children[position], _ = pushTail(*n.children[position], h-1, index, tail)
//	return &n, h
//
//}

func pushTail[V any](n node[V], h int, tail []*refValue[V]) (*node[V], int) {

	nodes := make([]*node[V], h+1)
	walkLastBranch(n, h, nodes)

	var i int
	nodes[0] = &node[V]{
		values: tail,
	}

	isBalance := len(tail) == maxBranches

	// find the deepest node (except leaf) that has less than maxBranches children, where we can push the new branch.
	for i = 1; i <= h; i++ {
		if len(nodes[i].children) < maxBranches {
			//found the node
			if !isBalance || !nodes[i].isBalancedNode() {
				sizes := nodes[i].getSizeTable(i) // get the sizes table of the node
				idx := len(sizes) - 1
				sizes = slice.Push(sizes, sizes[idx]+len(tail)) // push the new cumulative size to the sizes table

			}

			nodes[i].children = slice.Push(nodes[i].children, nodes[i-1]) // push the new node to the children of the current node
			break
		}

		// node is full, we can't push the new node into it. We must create a new branch.
		nodes[i] = &node[V]{
			children: []*node[V]{nodes[i-1]},
		}
	}

	if i > h {
		// all nodes are full, we must create a new root
		// i > h in case h == 0
		parent := &node[V]{
			children: []*node[V]{&n, nodes[i-1]},
		}

		if !n.isBalancedNode() {
			sizes := n.getSizeTable(h)
			idx := len(sizes) - 1
			parent.sizes = slice.Push(sizes, sizes[idx]+len(tail))
		}
		return parent, h + 1
	}

	// update the remaining nodes up to the root.
	for i = i + 1; i <= h; i++ {
		idx := len(nodes[i].children) - 1
		nodes[i].children = slice.Set(nodes[i].children, idx, nodes[i-1])

		if !nodes[i].isBalancedNode() || !nodes[i-1].isBalancedNode() {
			sizes := nodes[i].getSizeTable(i)
			idx = len(sizes) - 1
			sizes = slice.Push(sizes, sizes[idx]+len(tail))
		}
	}

	return nodes[h], h

}

// https://github.com/golang/go/wiki/CodeReviewComments#receiver-type
// pass node as value to make sure it is always shallow copied.
// caller must make sure that n is not nil or lese it will panic with nil pointer dereference.
// it is very useful for persistent data structure.
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

		if t.root == nil {
			t.root = &node[V]{
				values: t.tail,
			}
			t.tail = nil
			return t
		}

		t.root, t.h = pushTail(*t.root, t.h, t.tail)
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

//
//func walkFirstBranch[V any](n *node[V], h int, nodes []*node[V]) {
//
//	nodes[h-1] = n
//	walkFirstBranch(n.children[0], h-1, nodes)
//
//}

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
		copy(mergedSizes, left.getSizeTable(h))
		copy(mergedSizes[i:], right.getSizeTable(h))

		for j := i; j < len(mergedSizes); j++ {
			mergedSizes[j] += mergedSizes[i-1]
		}

		left.sizes = mergedSizes
		left.children = slice.Concat(left.children, right.children)

		return &left, nil

	}

	// merge into two nodes

	ls := left.getSizeTable(h)
	rs := right.getSizeTable(h)

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
			right.sizes = slice.SelectRange(right.getSizeTable(h), shiftedCount, len(right.children))
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
		//index := t.size - len(t.tail)

		if len(t.tail) > maxBranches {
			tail := slice.SelectRange(t.tail, 0, maxBranches)
			t.tail = slice.SelectRange(t.tail, maxBranches, len(t.tail))

			pushTail(*t.root, t.h, tail)
		}

		t.size += other.size
		return
	}
}

func navigate[V any](node *node[V], h, position int) (idx, nextPos int) {

	if node.isBalancedNode() {
		idx = (position >> shiftTable[h]) & maskTable[0]
		//nextPos = position & (1<<shiftTable[h] - 1)
		return idx, position & (1<<shiftTable[h] - 1)
	}

	idx = findPosition(node.sizes, position)
	nextPos = position - node.sizes[idx]
	return
}

func get[V any](t *RRBTree[V], i int) V {
	if i > t.size || i < 0 {
		panic("Index out of bounds")
	}

	idx := t.size - len(t.tail)

	if i >= idx {
		return t.tail[i-idx].value
	}

	if len(t.head) > 0 && i < len(t.head) {
		return t.head[i].value
	}

	n := t.root
	h := t.h
	var slot int
	for h > 0 {
		slot, i = navigate(n, h, i)
		n = n.children[slot]
		h--
	}

	return n.values[i].value
}

func (t RRBTree[V]) Get(i int) V {

	return get(&t, i)

}

func (t RRBTree[V]) Len() int {

	return t.size
}

func (t RRBTree[V]) Append(value V) RRBTree[V] {
	return t.push(value)
}
