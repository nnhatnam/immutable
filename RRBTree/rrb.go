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
	cumulativeSize []int // The cumulative size of the node.

	sizes []int // The size of each child.

	children []*node[V]     // The children of the node.
	values   []*refValue[V] // The values of the node.

	owner *RRBTree[V] // The transient owner of the node. for persistent, it's nil.
}

// descriptorClone doesn't really clone the node. It just returns a new node with the same children and values.
// Any modification to the children and values of the clone will affect the original node.
func (n *node[V]) descriptorClone() *node[V] {

	if n == nil {
		return nil
	}
	clone := &node[V]{
		cumulativeSize: n.cumulativeSize,
		children:       n.children,
		values:         n.values,
	}

	return clone
}

func (n *node[V]) isBalancedNode() bool {
	return n.sizes == nil
}

func (n *node[V]) isFull() bool {

	return len(n.children) == branchFactor
}

func (n *node[V]) size() int {
	if len(n.children) == 0 {
		return len(n.values)
	}

	return n.cumulativeSize[len(n.children)-1]

}

func (n *node[V]) appendChild(child *node[V]) {

	cumulativeSum := 0

	n.children = append(n.children, child)
	n.cumulativeSize = append(n.cumulativeSize, cumulativeSum+n.size())

}

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

func NewRRBTree[V any]() *RRBTree[V] {

	rrb := &RRBTree[V]{
		root: nil,
		size: 0,
		tail: nil,
	}

	return rrb

}

func (t *RRBTree[V]) newNode() *node[V] {
	return &node[V]{
		//children: make([]refChild[V], 0, branchFactor),

		owner: t,
	}
}

func (t *RRBTree[V]) getTailNode() *node[V] {

	if t.root == nil {
		return nil
	}

	n := t.root // start from root
	for j := t.h; j > 0; j-- {
		n = n.children[len(n.children)-1]
	}

	return n

}

func (t *RRBTree[V]) push(value V) {

	if t.tail == nil {
		t.tail = make([]*refValue[V], 0, branchFactor)
	}

	t.tail = append(t.tail, newRefValue[V](value))
	t.size++

	if len(t.tail) == maxBranches {
		// make a new branch
		t.root = t.pushTailDown(t.root, t.h)
	}

}

func (t *RRBTree[V]) pushTailDown(root *node[V], h int) *node[V] {
	if h == 0 {
		// pushTail always create a new leaf
		leaf := newNode[V]()
		leaf.values = t.tail
		t.tail = nil // clear tail
		t.last = leaf
		return leaf
	}

	n := root.descriptorClone()

	if n.isBalancedNode() {

		//full node
		if len(n.children) == branchFactor {
			n1 := newNode[V]()
			n1.children = append(n1.children, n)
			t.pushTailDown(n1, h+1)
			return n1
		}

		n.children = slice.Push(n.children, nil)
		n.children[len(n.children)-1] = t.pushTailDown(n.children[len(n.children)-1], h-1)
		return n

	}

	//case 2: relaxed node

	//check if the last branch is full
	if len(t.last.children) == maxBranches {
		n1 := newNode[V]()
		n1.children = append(n1.children, n)
		t.pushTailDown(n1, h+1)
		return n1
	}

	//minBranches <= len(t.last.children) < maxBranches
	// insert the first value of tail into the last branch

	lastIndex := len(n.children) - 1

	var recursiveClone func(trieNode *node[V], h int) *node[V]

	recursiveClone = func(trieNode *node[V], h int) *node[V] {
		if h == 0 {
			leaf := trieNode.descriptorClone()
			var v *refValue[V]
			t.tail, v = slice.PopFront(t.tail)
			leaf.values = slice.Push(leaf.values, v)

			return leaf
		}

		idx := len(trieNode.children) - 1
		trie1 := trieNode.descriptorClone()
		trie1.children = slice.Set(trie1.children, idx, recursiveClone(trie1.children[idx], h-1))

		if (trie1.sizes[idx] + 1) == len(trie1.sizes)*maxBranches {
			trie1.sizes = nil
		} else {
			trie1.sizes = slice.Set(trie1.sizes, idx, trie1.sizes[idx]+1)
		}

		return trie1
	}

	n.children[lastIndex] = recursiveClone(n.children[lastIndex], h-1)
	return n

}

func (t *RRBTree[V]) findPosition(sizes []int, idx int) int {
	i := 0
	for i < len(sizes) && sizes[i] <= idx {
		i++
	}

	return i
}

func (t *RRBTree[V]) get(i int) V {

	if i > t.size {
		panic("Index out of bounds")
	}
	n := t.root

	for j := t.h; j > 0; i-- {

		idx := (i >> shiftTable[j]) & maskTable[j]
		if n.sizes != nil {
			idx = t.findPosition(n.sizes, i)
			i -= n.sizes[j-1]
		}
		n = n.children[idx]
		j--
	}

	return n.values[i].value

}

func (t *RRBTree[V]) Size() int {

	return t.size
}
