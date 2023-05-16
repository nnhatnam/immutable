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

	tail []*refValue[V] // The tail of the tree.
	//sizeNonTail int            // The size of the tree without the tail.

	last *node[V] // The last node of the tree.
}

func NewRRBTree[V any]() RRBTree[V] {

	return RRBTree[V]{
		root: nil,
		size: 0,
		tail: nil,
	}

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

	//if root == nil {
	//	return nil
	//}

	if h == 0 {
		// pushTail always create a new leaf
		//leaf := newNode[V]()
		//leaf.values = t.tail
		//return leaf

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

func (t *RRBTree[V]) findLastNode() *node[V] {

	n := t.root // start from root
	for j := t.h; j > 0; j-- {
		n = n.children[len(n.children)-1]
	}
	return n
}

func findPosition(sizes []int, idx int) int {
	i := 0
	for i < len(sizes) && sizes[i] <= idx {
		i++
	}

	return i
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
