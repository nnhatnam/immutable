package RRBTree

import (
	"github.com/nnhatnam/immutable/slice"
)

// items stores items in a node.
type items[T any] []T

// insertAt inserts a value into the given index, pushing all subsequent values
// forward.
func (s *items[T]) insertAt(index int, item T) {
	var zero T
	*s = append(*s, zero)
	if index < len(*s) {
		copy((*s)[index+1:], (*s)[index:])
	}
	(*s)[index] = item
}

// removeAt removes a value at a given index, pulling all subsequent values
// back.
func (s *items[T]) removeAt(index int) T {
	item := (*s)[index]
	copy((*s)[index:], (*s)[index+1:])
	var zero T
	(*s)[len(*s)-1] = zero
	*s = (*s)[:len(*s)-1]
	return item
}

// pop removes and returns the last element in the list.
func (s *items[T]) pop() (out T) {
	index := len(*s) - 1
	out = (*s)[index]
	var zero T
	(*s)[index] = zero
	*s = (*s)[:index]
	return
}

// truncate truncates this instance at index so that it contains only the
// first index items. index must be less than or equal to length.
func (s *items[T]) truncate(index int) {
	var toClear items[T]
	*s, toClear = (*s)[:index], (*s)[index:]
	var zero T
	for i := 0; i < len(toClear); i++ {
		toClear[i] = zero
	}
}

func (s *items[T]) maybeTruncate(index int) {
	if index < len(*s) {
		s.truncate(index)
	}
}

// retain is opposite of truncate. It keeps this instance from index to the end, removing all prior items.
// index must be less than or equal to length.
func (s *items[T]) retain(index int) {
	var toClear items[T]
	*s, toClear = (*s)[index:], (*s)[:index]
	var zero T
	for i := 0; i < len(toClear); i++ {
		toClear[i] = zero
	}
}

func (s *items[T]) maybeRetain(index int) {
	if index < len(*s) {
		s.retain(index)
	}
}

func (s *items[T]) slice(i, j int) {
	_ = (*s)[i:j]

	var left, right items[T]
	left, *s, right = (*s)[:i], (*s)[i:j], (*s)[j:]

	var zero T
	for i := 0; i < len(left); i++ {
		left[i] = zero
	}
	for i := 0; i < len(right); i++ {
		right[i] = zero
	}
}

type refValue[V any] struct {
	value V
}

func newRefValue[V any](value V) *refValue[V] {
	return &refValue[V]{
		value: value,
	}
}

type copyOnWriteContext[V any] struct {
	rrb *RRBTree[V]
}

func (cow *copyOnWriteContext[V]) newNode() *node[V] {
	return &node[V]{
		cow: cow,
	}
}

func (cow *copyOnWriteContext[V]) createLeaf(items []*refValue[V]) *node[V] {
	return &node[V]{
		treeSize: len(items),
		values:   items,
		cow:      cow,
	}
}

func (cow *copyOnWriteContext[V]) createInternalNode(treeSize int, sizes []int, children ...*node[V]) *node[V] {
	return &node[V]{
		treeSize: treeSize,
		sizes:    sizes,
		children: children,
		cow:      cow,
	}
}

// There are three types of nodes in the tree:
// PartialNode is a node that has some children.
// FullNode is a node that is full children.
// LeafNode is a node that has no children.
type node[V any] struct {
	treeSize int // The size of the tree rooted at this node.

	sizes items[int] // The size of each child.

	children items[*node[V]]     // The children of the node.
	values   items[*refValue[V]] // The values of the node.

	//owner *RRBTree[V] // The transient owner of the node. for persistent, it's nil.

	cow *copyOnWriteContext[V] // The copy on write context.
}

func (n *node[V]) mutableFor(cow *copyOnWriteContext[V]) *node[V] {
	if n != nil && n.cow == cow {
		return n
	}

	m := cow.newNode()
	m.treeSize = n.treeSize
	m.sizes = slice.Copy(n.sizes)
	m.children = slice.Copy(n.children)
	m.values = slice.Copy(n.values)

	return m
}

type cloneType int

const (
	cloneWithOneExtraCap cloneType = iota
	cloneForPrepend
)

func (n *node[V]) mutableForWithCustomAllocs(cow *copyOnWriteContext[V], typ cloneType) *node[V] {

	m := cow.newNode()

	switch typ {
	case cloneWithOneExtraCap:
		m.treeSize = n.treeSize

		if len(n.sizes) != 0 {
			m.sizes = make([]int, len(n.sizes), len(n.sizes)+1)
			copy(m.sizes, n.sizes)
		}

		m.children = make([]*node[V], len(n.children), len(n.children)+1)
		copy(m.children, n.children)
	case cloneForPrepend:

		m.treeSize = n.treeSize
		if n.isRelaxedNode() {
			m.sizes = make([]int, len(n.sizes)+1)
			copy(m.sizes[1:], n.sizes)
		}
		m.children = make([]*node[V], len(n.children)+1)
		copy(m.children[1:], n.children)
	default:
		panic("unknown clone action")
	}

	return m

}

func (n *node[V]) mutableChild(i int) *node[V] {
	m := n.mutableFor(n.cow)
	m.children[i] = m.children[i].mutableFor(n.cow)
	return m
}

func (n *node[V]) hasFullChildren() bool {
	return len(n.children) == maxBranches
}

func (n *node[V]) isFullNode() bool {
	return len(n.children) == maxBranches
}

func (n *node[V]) isEmpty() bool {
	return len(n.children) == 0
}

func (n *node[V]) insertAt(height, index int, child *node[V]) {
	n.treeSize = n.treeSize + child.treeSize
	n.children.insertAt(index, child)

	sizes := n.sizes

	if n.isRelaxedNode() {
		sizes.insertAt(index, child.treeSize)

		size := child.treeSize + n.readCumulativeSize(height, index-1)
		sizes[index] = size

		for i := index + 1; i < len(n.sizes); i++ {
			sizes[i] += child.treeSize
		}

	}

	if child.treeSize != cumulativeSumTable[height][0] {
		sizes = slice.Copy(cumulativeSumTable[height][:len(n.children)])
		offset := child.treeSize - sizes[index]
		sizes[index] = child.treeSize

		for i := index + 1; i < len(sizes); i++ {
			sizes[i] += +offset
		}
	}

	n.sizes = sizes

}

func (n *node[V]) addChild(child *node[V]) {

	n.treeSize = n.treeSize + child.treeSize
	n.children = append(n.children, child)

	if n.isRelaxedNode() {
		n.sizes = append(n.sizes, n.treeSize)
	}
}

func (n *node[V]) setChild(height, index int, child *node[V]) {

	if n.isBalancedNode() && child.treeSize != cumulativeSumTable[height][0] {

		// build the sizes table
		n.sizes = slice.Copy(cumulativeSumTable[height][:len(n.children)])
		n.sizes[len(n.children)-1] = n.treeSize
	}

	offset := n.readCumulativeSize(height, index) - n.readCumulativeSize(height, index-1)
	offset = child.treeSize - offset

	n.children[index] = child
	n.treeSize += offset

	for i := index; i < len(n.sizes); i++ {
		n.sizes[i] = n.sizes[i] + offset
	}

}

// readCumulativeSize returns the cumulative size of the node at the given index at the given height.
// readCumulativeSize find the cumulative size based of the `sizes` array inside the node for relaxed nodes.
// For balance nodes, it uses the cumulativeSumTable.
// Because the way it is calculated, it must be called when we are sure the sizes array is up-to-date.
func (n *node[V]) readCumulativeSize(height, index int) int {

	switch {
	case n.isRelaxedNode():
		return n.sizes[index]
	case index == len(n.children)-1:
		return n.treeSize
	case index < 0:
		return 0
	default:
		return cumulativeSumTable[height][index]
	}

}

type branchType int

const (
	oldBranch branchType = iota
	newBranch
)

func (n *node[V]) truncate(h, i int) (*node[V], items[*refValue[V]]) {

	if h == 0 {
		n.values.truncate(i)
		if i < maxBranches {
			return nil, n.values
		}
		n.treeSize = len(n.values)
		return n, nil
	}

	var slot int
	slot, i = navigate(n, h, i)

	child, tail := n.children[slot].mutableChild(slot).truncate(h-1, i)

	if child == nil {

		if slot == 0 {
			return nil, tail
		}

		n.treeSize = n.readCumulativeSize(h, slot-1)
		n.children.truncate(slot)
		n.sizes.maybeTruncate(slot)

		return n, tail

	}

	n.treeSize = n.readCumulativeSize(h, slot)
	n.children.truncate(slot + 1)
	n.sizes.maybeTruncate(slot + 1)

	// cumulative size at slot - 1
	n.setChild(h, slot, child)

	//n.children[slot] = child
	//
	//newSize := n.findCumulativeSize(h, slot-1) + child.treeSize
	//n.treeSize = newSize
	//n.sizes[slot] = newSize

	return n, tail
}

func (n *node[V]) retain(h, i int) (*node[V], items[*refValue[V]]) {

	if h == 0 {
		n.values.retain(i)
		if i < maxBranches {
			return nil, n.values
		}
		n.treeSize = len(n.values)
		return n, nil
	}

	var slot int
	slot, i = navigate(n, h, i)

	child, head := n.children[slot].mutableChild(slot).retain(h-1, i)

	if child == nil {
		if slot == len(n.children)-1 {
			return nil, head
		}

		slot = slot + 1
	}

	n.children.retain(slot)
	n.sizes.maybeRetain(slot)

	offset := child.treeSize - n.readCumulativeSize(h, 0)
	n.treeSize += offset

	for i := 0; i < len(n.sizes); i++ {
		n.sizes[i] = n.sizes[i] + offset
	}

	return n, head
}

func (n *node[V]) pushHead(h int, head []*refValue[V]) (*node[V], int) {

	var pushHeadRecursive func(*node[V], int) (*node[V], branchType)

	pushHeadRecursive = func(m *node[V], h int) (*node[V], branchType) {

		if h == 0 {
			return m.cow.createLeaf(head), newBranch
		}

		child, bType := pushHeadRecursive(m.children[0], h-1)

		if bType == oldBranch {
			m.setChild(h, 0, child)
			return m, oldBranch
		}

		// if we get here, it means we have a new branch
		// so we either push the child to the current node or create a new node if the current node is full
		if m.hasFullChildren() {
			if child.isRelaxedNode() {
				return m.cow.createInternalNode(child.treeSize, []int{child.treeSize}, child), newBranch
			}

			return m.cow.createInternalNode(child.treeSize, nil, child), newBranch
		}

		m.insertAt(h, 0, child)
		return m, oldBranch
	}

	m, bType := pushHeadRecursive(n, h)
	if bType == newBranch {
		var sizes []int
		if n.isRelaxedNode() || m.isRelaxedNode() {
			sizes = []int{m.treeSize, n.treeSize + m.treeSize}
		}
		return n.cow.createInternalNode(n.treeSize+m.treeSize, sizes, n, m), h + 1
	}
	return m, h
}

// pushTail pushes the tail into the tree rooted at the given node `n`.
func (n *node[V]) pushTail(h int, tail []*refValue[V]) (*node[V], int) {

	var pushTailRecursive func(*node[V], int) (*node[V], branchType)

	pushTailRecursive = func(m *node[V], h int) (*node[V], branchType) {

		if h == 0 {
			return m.cow.createLeaf(tail), newBranch
		}

		child, bType := pushTailRecursive(m.children[len(m.children)-1], h-1)

		if bType == oldBranch {
			out := m.mutableFor(m.cow)
			out.setChild(h, len(out.children)-1, child)
			return out, oldBranch
		}

		// if we get here, it means we have a new branch
		// so we either push the child to the current node or create a new node if the current node is full
		if m.hasFullChildren() {
			if child.isRelaxedNode() {
				return m.cow.createInternalNode(child.treeSize, []int{child.treeSize}, child), newBranch
			}

			return m.cow.createInternalNode(child.treeSize, nil, child), newBranch
		}

		out := m.mutableForWithCustomAllocs(m.cow, cloneWithOneExtraCap)
		out.addChild(child)
		return out, oldBranch
	}

	m, bType := pushTailRecursive(n, h)
	if bType == newBranch {
		var sizes []int
		if n.isRelaxedNode() || m.isRelaxedNode() {
			sizes = []int{n.treeSize, n.treeSize + m.treeSize}
		}
		return n.cow.createInternalNode(n.treeSize+m.treeSize, sizes, n, m), h + 1
	}
	return m, h
}

func (n *node[V]) pop(h int) (_ *node[V], _ int, value *refValue[V], tail items[*refValue[V]]) {

	var popRecursive func(*node[V], int) *node[V]

	popRecursive = func(m *node[V], h int) *node[V] {

		if h == 0 {

			value = m.values.pop()

			if len(m.values) < maxBranches {
				tail = m.values
				return nil
			}

			m.treeSize--
			return m
		}

		slot := len(m.children) - 1

		//child := m.mutableChild(slot)
		child := popRecursive(m.mutableChild(slot), h-1)

		if child == nil {

			if slot == 0 {
				return nil
			}

			m.children.pop()
			m.sizes.pop()
			m.treeSize = m.readCumulativeSize(h, slot-1)
			return m
		}

		m.setChild(h, slot, child)
		return m

	}

	m := popRecursive(n.mutableFor(n.cow), h)
	if m == nil {
		return nil, 0, value, tail
	}

	m, h = shrink(m, h)
	return m, h, value, tail

}

func (n *node[V]) shallowClone() *node[V] {
	clone := *n
	return &clone
}

func (n *node[V]) isBalancedNode() bool {
	return len(n.sizes) == 0
}

func (n *node[V]) isRelaxedNode() bool {

	if len(n.children) > 0 {
		return len(n.sizes) != 0
	}

	return len(n.children) != maxBranches

}

func (n *node[V]) isBalanced(h int) bool {
	if h == 0 {
		return len(n.values) == maxBranches
	}
	return len(n.sizes) == 0
}

func (n *node[V]) reachedMaxBranch() bool {
	return len(n.children) == maxBranches
}

// caller must not modify the returned slice.
func (n *node[V]) getSizeTable(h int) []int {
	if len(n.sizes) != 0 {
		return n.sizes
	}
	return cumulativeSumTable[h][:len(n.children)]
}

func (n *node[V]) getLength(h int) int {
	if len(n.sizes) != 0 {
		return n.sizes[len(n.sizes)-1]
	}

	return cumulativeSumTable[h][len(n.children)-1]

}

func (n *node[V]) walk(h, i int) []*node[V] {
	nodes := make([]*node[V], h)
	root := n

	var slot int
	for ; h > 0; h-- {
		nodes[h] = root
		slot, i = navigate(root, h, i)
		root = root.children[slot]
	}
	nodes[0] = root
	return nodes
}

func (n *node[V]) walkLast(h int) []*node[V] {
	nodes := make([]*node[V], h)
	root := n

	for ; h > 0; h-- {
		nodes[h] = root
		root = root.children[len(root.children)-1]
	}
	nodes[0] = root
	return nodes
}

func (n *node[V]) walkFirst(h int) []*node[V] {
	nodes := make([]*node[V], h)
	root := n

	for ; h > 0; h-- {
		nodes[h] = root
		root = root.children[0]
	}
	nodes[0] = root
	return nodes
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

	head items[*refValue[V]] // The head of the tree.

	tail items[*refValue[V]] // The tail of the tree.

}

func NewRRBTree[V any]() RRBTree[V] {

	return RRBTree[V]{}

}

func walkFirstBranch[V any](n node[V], h int, nodes []*node[V]) {
	nodes[h] = &n
	if h != 0 {
		walkFirstBranch(*n.children[0], h-1, nodes)
	}
	return

}

func shrink[V any](n *node[V], h int) (m *node[V], newH int) {

	if n == nil || len(n.children) == 0 {
		return nil, 0
	}

	m = n

	for newH = h; newH > 0; newH-- {
		if len(m.children) != 1 {
			break
		}
		m = m.children[0]
	}

	return
}

// walk the tree and return the nodes at each level.
// the returned slice is ordered from bottom to top.
// the nodes returned are not copied.
// It depends on situation whether it's safe to modify the nodes.
// Most of the time, we don't modify the nodes, but in transients, we do.
func (t RRBTree[V]) walk(i int) []*node[V] {
	nodes := make([]*node[V], t.h)
	n := t.root

	var slot int
	for h := t.h; h > 0; h-- {
		nodes[h] = n
		slot, i = navigate(n, h, i)
		n = n.children[slot]
	}
	nodes[0] = n
	return nodes
}

func (t RRBTree[V]) walkLast() []*node[V] {
	nodes := make([]*node[V], t.h)
	n := t.root

	for h := t.h; h > 0; h-- {
		nodes[h] = n
		n = n.children[len(n.children)-1]
	}
	nodes[0] = n
	return nodes
}

func (t RRBTree[V]) walkFirst() []*node[V] {
	nodes := make([]*node[V], t.h)
	n := t.root

	for h := t.h; h > 0; h-- {
		nodes[h] = n
		n = n.children[0]
	}
	nodes[0] = n
	return nodes
}

func createLeaf[V any](values []*refValue[V]) *node[V] {
	return &node[V]{
		treeSize: len(values),
		values:   values,
	}
}

func createInternalNode[V any](treeSize int, sizes []int, children ...*node[V]) *node[V] {
	return &node[V]{
		treeSize: treeSize,
		sizes:    sizes,
		children: children,
	}
}

func setLastChildOnNode[V any](n *node[V], child *node[V]) *node[V] {

	m := n.shallowClone()

	idx := len(n.children) - 1

	m.treeSize += child.treeSize - m.children[idx].treeSize
	m.children = slice.Set(m.children, idx, child)

	if !m.isBalancedNode() {
		m.sizes = slice.Set(m.sizes, idx, m.treeSize)
	}

	return m

}

func walkLastBranch[V any](n node[V], h int, nodes []*node[V]) {
	nodes[h] = &n
	if h != 0 {
		walkLastBranch(*n.children[len(n.children)-1], h-1, nodes)
	}
	return
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
			t.root = (*copyOnWriteContext[V])(nil).createLeaf(t.tail)
			t.tail = nil
			return t
		}

		// pushTail may create a new root, so in some cases we don't need to clone the current root.
		// For that reason, clone the root will be decided inside pushTail.
		t.root, t.h = t.root.pushTail(t.h, t.tail)
		t.tail = nil

	}

	return t

}

//// https://github.com/golang/go/wiki/CodeReviewComments#receiver-type
//// pass node as value to make sure it is always shallow copied.
//// caller must make sure that n is not nil or lese it will panic with nil pointer dereference.
//// it is very useful for persistent data structure.
//func (t RRBTree[V]) push(value V) RRBTree[V] {
//	// Value receiver makes a copy of the type and pass it to the function.
//	// The function stack now holds an equal object but at a different location on memory.
//	// That means any changes done on the passed object will remain local to the method.
//	// The original object will remain unchanged.
//
//	if t.tail == nil {
//		t.tail = make([]*refValue[V], 0, maxBranches)
//	}
//
//	t.tail = append(t.tail, newRefValue[V](value))
//	t.size++
//
//	if len(t.tail) == maxBranches {
//		// make a new branch
//
//		if t.root == nil {
//			t.root = createLeaf(t.tail)
//			t.tail = nil
//			return t
//		}
//
//		m, isNewBranch := pushTail(t.root, t.h, t.tail)
//		if isNewBranch {
//
//			t.h++
//
//			var sizes []int
//			if !t.root.isBalancedNode() || !m.isBalancedNode() || t.root.treeSize != cumulativeSumTable[t.h][0] {
//				sizes = []int{t.root.treeSize, t.root.treeSize + m.treeSize}
//			}
//
//			t.root = createInternalNode[V](t.root.treeSize+m.treeSize, sizes, t.root, m)
//
//		} else {
//			t.root = m
//		}
//
//		t.tail = nil
//	}
//
//	return t
//
//}

func prependChildToPartialNode[V any](n *node[V], h int, child *node[V]) *node[V] {

	m := n.shallowClone()
	if m.isBalancedNode() {
		m.children = slice.PushFront(m.children, child)
		m.treeSize += child.treeSize

		if !child.isBalancedNode() || child.treeSize != cumulativeSumTable[h][0] {
			// child has "invariant violation"

			idx := len(m.children) - 1

			m.sizes = slice.PushFront(cumulativeSumTable[h][:idx], child.treeSize)

			for i := 1; i < idx; i++ {
				m.sizes[i] += child.treeSize
			}

			m.sizes[idx] = m.treeSize
		}
		return m
	}

	// unbalanced node
	m.children = slice.PushFront(m.children, child)
	m.treeSize += child.treeSize

	idx := len(m.children) - 1
	m.sizes = slice.PushFront(m.sizes, child.treeSize)
	for i := 1; i < idx; i++ {
		m.sizes[i] += child.treeSize
	}
	m.sizes[idx] = m.treeSize

	return m

}

func setFirstChildOnNode[V any](n *node[V], child *node[V]) *node[V] {

	m := n.shallowClone()

	//idx := 0

	m.treeSize += child.treeSize - m.children[0].treeSize
	m.children = slice.Set(m.children, 0, child)

	if !m.isBalancedNode() {
		diff := child.treeSize - m.sizes[0]
		m.sizes = slice.Set(m.sizes, 0, child.treeSize)
		for i := 1; i < len(m.sizes); i++ {
			m.sizes[i] += diff
		}
	}

	return m

}

func pushHead[V any](n *node[V], h int, head []*refValue[V]) (m *node[V], isNewBranch bool) {

	if h == 0 {
		return createLeaf(head), true
	}

	var child *node[V]

	child, isNewBranch = pushHead(n.children[0], h-1, head)

	if len(n.children) == maxBranches && isNewBranch {
		var sizes []int

		if !child.isBalanced(h - 1) {
			sizes = []int{child.treeSize}
		}

		return createInternalNode(child.treeSize, sizes, child), true

	}

	if isNewBranch {
		return prependChildToPartialNode(n, h, child), false
	}

	return setFirstChildOnNode(n, child), false

}

func (t RRBTree[V]) prepend(value V) RRBTree[V] {

	t.head = slice.PushFront(t.head, newRefValue[V](value))
	//t.head = append(t.head, newRefValue[V](value))
	t.size++

	if len(t.head) == maxBranches {
		// make a new branch

		if t.root == nil {
			t.root = createLeaf(t.head)
			t.head = nil
			return t
		}

		m, isNewBranch := pushHead(t.root, t.h, t.head)

		if isNewBranch {
			t.h++

			var sizes []int
			if !t.root.isBalancedNode() || !m.isBalancedNode() || m.treeSize != cumulativeSumTable[t.h][0] {
				sizes = []int{m.treeSize, t.root.treeSize + m.treeSize}
			}

			t.root = createInternalNode[V](t.root.treeSize+m.treeSize, sizes, m, t.root)

		} else {
			t.root = m

		}

		t.head = nil
	}

	return t
}

//func (t RRBTree[V]) prepend(value V) RRBTree[V] {
//
//	t.head = slice.PushFront(t.head, newRefValue[V](value))
//	//t.head = append(t.head, newRefValue[V](value))
//	t.size++
//
//	if len(t.head) == maxBranches {
//		// make a new branch
//
//		if t.root == nil {
//			t.root = createLeaf(t.head)
//			t.head = nil
//			return t
//		}
//
//		m, isNewBranch := pushHead(t.root, t.h, t.head)
//
//		if isNewBranch {
//			t.h++
//
//			var sizes []int
//			if !t.root.isBalancedNode() || !m.isBalancedNode() || m.treeSize != cumulativeSumTable[t.h][0] {
//				sizes = []int{m.treeSize, t.root.treeSize + m.treeSize}
//			}
//
//			t.root = createInternalNode[V](t.root.treeSize+m.treeSize, sizes, m, t.root)
//
//		} else {
//			t.root = m
//
//		}
//
//		t.head = nil
//	}
//
//	return t
//}

func popFromLeaf[V any](leaf *node[V]) (m *node[V], value *refValue[V], tail []*refValue[V]) {

	tail, value = slice.Pop(leaf.values)

	if len(tail) < maxBranches {
		return nil, value, tail
	}

	m = createLeaf(tail)

	return m, value, nil
}

func removeLastChildOnNode[V any](n *node[V]) *node[V] {
	if n == nil || len(n.children) == 1 {
		return nil
	}

	m := n.shallowClone()
	idx := len(m.children) - 1
	m.treeSize -= m.children[idx].treeSize
	m.children, _ = slice.Pop(m.children)
	m.sizes, _ = slice.Pop(m.sizes)

	return m
}

func popBack[V any](n *node[V], h int) (m *node[V], value *refValue[V], tail []*refValue[V]) {
	if h == 0 {
		return popFromLeaf(n)
	}

	var child *node[V]
	child, value, tail = popBack(n.children[len(n.children)-1], h-1)

	if child == nil {
		m = removeLastChildOnNode(n)
		return m, value, tail
	}

	return setLastChildOnNode(n, child), value, tail
}

//func greaterOrEqual[V any](root *node[V], h int, index int) (head []*refValue[V], right *node[V]) {
//	if h == 0 {
//		head = slice.Copy(root.values[index:])
//
//		if len(head) < maxBranches {
//			return
//		} else if len(head) > maxBranches {
//			idx := len(head) - maxBranches
//			return head[:idx], createLeaf(head[idx:])
//		}
//
//		return nil, createLeaf(head)
//	}
//
//	//nodes[h]
//	var slot int
//	var child *node[V]
//
//	slot, index = navigate(root, h, index)
//
//	head, child = greaterOrEqual(root.children[slot], h-1, index)
//
//	right = root.shallowClone()
//
//	if child == nil {
//		if slot == len(right.children)-1 {
//			return head, nil
//		}
//
//		right.children = slice.Slice(right.children, slot+1, len(right.children))
//
//		if !right.isBalancedNode() {
//
//			diff := root.sizes[slot]
//			right.sizes = slice.Slice(right.sizes, slot+1, len(right.sizes))
//			right.sizes[len(right.sizes)-1] = right.treeSize
//			right.treeSize -= diff
//
//			for i := 0; i < len(right.sizes); i++ {
//				right.sizes[i] -= diff
//			}
//
//		} else {
//			diff := cumulativeSumTable[h][slot]
//			right.treeSize -= diff
//		}
//
//		return head, right
//
//	}
//
//	if slot == 0 {
//		return head, setFirstChildOnNode(right, child)
//	}
//
//	right.children = slice.Slice(right.children, slot, len(right.children))
//	right.children[0] = child
//
//	if !right.isBalancedNode() {
//		offset := child.treeSize - root.sizes[slot]
//		right.sizes = slice.Slice(right.sizes, slot, len(right.sizes))
//		for i := 0; i < len(right.sizes); i++ {
//			right.sizes[i] += offset
//		}
//		right.treeSize = right.sizes[len(right.sizes)-1]
//
//	} else {
//
//		if child.treeSize != cumulativeSumTable[h][0] {
//
//			slotSize := cumulativeSumTable[h][slot]
//			right.sizes = slice.Copy(cumulativeSumTable[h][slot : len(right.sizes)+slot])
//			offset := child.treeSize - slotSize
//
//			for i := 0; i < len(right.sizes); i++ {
//				right.sizes[i] += offset
//			}
//
//			right.treeSize = cumulativeSumTable[h][slot]
//
//		}
//
//	}
//	return head, right
//
//}
//
//func lessOrEqual[V any](root *node[V], h int, index int) (left *node[V], tail []*refValue[V]) {
//	if h == 0 {
//
//		// consider to implement a case where we don't need to duplicate the slice whe it doesn't change the underlying array
//		if index == root.treeSize-1 {
//			return root, nil
//		}
//
//		tail = slice.Copy(root.values[:index+1])
//		if len(tail) < maxBranches {
//			return
//		} else if len(tail) > maxBranches {
//			return createLeaf(tail[:maxBranches]), tail[maxBranches:]
//		}
//
//		return createLeaf(tail), nil
//
//	}
//
//	//nodes[h]
//	var slot int
//	var child *node[V]
//
//	slot, index = navigate(root, h, index)
//
//	child, tail = lessOrEqual(root.children[slot], h-1, index)
//
//	left = root.shallowClone()
//
//	if child == nil {
//		if slot == 0 {
//			return nil, tail
//		}
//
//		left.children = slice.Slice(left.children, 0, slot)
//
//		if !left.isBalancedNode() {
//			left.sizes = slice.Slice(left.sizes, 0, slot)
//			left.treeSize = left.sizes[slot-1]
//		} else {
//			left.treeSize = cumulativeSumTable[h][slot-1]
//		}
//
//		return left, tail
//
//	}
//
//	if slot == len(left.children)-1 {
//		return setLastChildOnNode(left, child), tail
//	}
//
//	left.children = slice.Slice(left.children, 0, slot+1)
//	left.children[slot] = child
//
//	if !left.isBalancedNode() {
//		left.sizes = slice.Slice(left.sizes, 0, slot+1)
//		left.treeSize = left.sizes[slot]
//	} else {
//		left.treeSize = cumulativeSumTable[h][slot]
//	}
//	return left, tail
//
//}

func (t RRBTree[V]) slice(i, j int) RRBTree[V] {
	start := 0 + len(t.head)
	end := t.size - len(t.tail)

	//fmt.Println("start , end , i , j", start, end, i, j)

	// [0 : head ) - [start : end) - [tail : size)

	// fmt.Println(t.size, "[0, ", len(t.head), ")", " [", start, " , ", end, ") [", end, t.size, ")", "-->", i)
	switch {
	case i > j || i < 0 || j > t.size+1:
		panic("Index out of bounds")
	case i >= end: // look into tail
		t.tail = slice.Slice(t.tail, i-end, j-end)
		t.size = len(t.tail)
		t.head = nil
		t.root = nil
		return t
	case j <= start: // look into head
		t.head = slice.Slice(t.head, i, j)
		t.size = len(t.head)
		t.tail = nil
		t.root = nil
		return t
	case i <= start && j >= end:
		t.head = slice.Slice(t.head, i, start)
		t.tail = slice.Slice(t.tail, 0, j-end)
		t.size = j - i
		return t
	case i == j:
		return RRBTree[V]{}
	default: // look into root
		t.size = j - i

		if i <= start {
			t.head = slice.Slice(t.head, i, start)
			t.root, t.tail = t.root.mutableFor((*copyOnWriteContext[V])(nil)).truncate(t.h, j-start)
			t.root, t.h = shrink(t.root, t.h)
			return t
		} else if j >= end {
			t.tail = slice.Slice(t.tail, 0, j-end)
			t.root, t.head = t.root.mutableFor((*copyOnWriteContext[V])(nil)).retain(t.h, i-start)
			t.root, t.h = shrink(t.root, t.h)
			return t
		}

		t.root, t.tail = t.root.mutableFor((*copyOnWriteContext[V])(nil)).truncate(t.h, j-start)
		t.root, t.h = shrink(t.root, t.h)
		t.root, t.head = t.root.mutableFor((*copyOnWriteContext[V])(nil)).retain(t.h, i-start)
		t.root, t.h = shrink(t.root, t.h)

		return t
	}
}

//func (t RRBTree[V]) slice(i, j int) RRBTree[V] {
//	start := 0 + len(t.head)
//	end := t.size - len(t.tail)
//
//	//fmt.Println("start , end , i , j", start, end, i, j)
//
//	// [0 : head ) - [start : end) - [tail : size)
//
//	// fmt.Println(t.size, "[0, ", len(t.head), ")", " [", start, " , ", end, ") [", end, t.size, ")", "-->", i)
//	switch {
//	case i > j || i < 0 || j > t.size+1:
//		panic("Index out of bounds")
//	case i >= end: // look into tail
//		t.tail = slice.Slice(t.tail, i-end, j-end)
//		t.size = len(t.tail)
//		t.head = nil
//		t.root = nil
//		return t
//	case j <= start: // look into head
//		t.head = slice.Slice(t.head, i, j)
//		t.size = len(t.head)
//		t.tail = nil
//		t.root = nil
//		return t
//	case i <= start && j >= end:
//		t.head = slice.Slice(t.head, i, start)
//		t.tail = slice.Slice(t.tail, 0, j-end)
//		t.size = j - i
//		return t
//	case i == j:
//		return RRBTree[V]{}
//	default: // look into root
//		t.size = j - i
//
//		if i <= start {
//			t.head = slice.Slice(t.head, i, start)
//			t.root, t.tail = lessOrEqual(t.root, t.h, j-start-1)
//			t.root, t.h = shrink(t.root, t.h)
//			return t
//		} else if j >= end {
//			t.tail = slice.Slice(t.tail, 0, j-end)
//			t.head, t.root = greaterOrEqual(t.root, t.h, i-start)
//			t.root, t.h = shrink(t.root, t.h)
//			return t
//		}
//
//		t.root, t.tail = lessOrEqual(t.root, t.h, j-start-1)
//		t.root, t.h = shrink(t.root, t.h)
//		t.head, t.root = greaterOrEqual(t.root, t.h, i-start)
//		t.root, t.h = shrink(t.root, t.h)
//		return t
//
//	}
//}

func findPosition(sizes []int, idx int) int {
	i := 0
	for i < len(sizes) && sizes[i] <= idx {
		i++
	}

	return i
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

		right.sizes = slice.Apply(right.sizes[1:], func(i, e int) int { return e - right.sizes[0] })
		right.children, _ = slice.Pop(right.children)

	}

	if !newR.isBalancedNode() {
		shiftedCount := len(right.children) - len(newR.children)

		if !right.isBalancedNode() {
			right.sizes = slice.Apply(right.sizes[shiftedCount:], func(i, e int) int { return e - right.sizes[shiftedCount-1] })
		} else {
			right.sizes = slice.Slice(right.getSizeTable(h), shiftedCount, len(right.children))
		}

		right.children = slice.Slice(right.children, shiftedCount, len(right.children))

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

//func (t RRBTree[V]) concatenate(other RRBTree[V]) (rrb RRBTree[V]) {
//
//	if other.root == nil {
//
//		tail := make([]*refValue[V], len(t.tail)+len(other.head)+len(other.tail))
//		copy(tail, t.tail)
//		copy(tail[len(t.tail):], other.head)
//		copy(tail[len(t.tail)+len(other.head):], other.tail)
//
//		t.tail = tail
//
//		if len(t.tail) > maxBranches {
//			tail = t.tail[:maxBranches]
//			t.tail = t.tail[maxBranches:]
//
//			m, isNewBranch := pushTail(t.root, t.h, tail)
//
//			if isNewBranch {
//
//				t.h++
//
//				var sizes []int
//				if !t.root.isBalancedNode() || !m.isBalancedNode() || t.root.treeSize != cumulativeSumTable[t.h][0] {
//					sizes = []int{t.root.treeSize, t.root.treeSize + m.treeSize}
//				}
//
//				t.root = createInternalNode[V](t.root.treeSize+m.treeSize, sizes, t.root, m)
//
//			} else {
//				t.root = m
//			}
//		}
//
//		t.size += other.size
//		return
//	}
//
//	m, isNewBranch := pushTail(t.root, t.h, t.tail)
//	if isNewBranch {
//
//		t.h++
//
//		var sizes []int
//		if !t.root.isBalancedNode() || !m.isBalancedNode() || t.root.treeSize != cumulativeSumTable[t.h][0] {
//			sizes = []int{t.root.treeSize, t.root.treeSize + m.treeSize}
//		}
//
//		t.root = createInternalNode[V](t.root.treeSize+m.treeSize, sizes, t.root, m)
//
//	} else {
//		t.root = m
//	}
//
//	t.tail = nil
//	return
//}

func navigate[V any](node *node[V], h, position int) (idx, nextPos int) {

	if node.isBalancedNode() {
		idx = (position >> shiftTable[h]) & maskTable[0]
		//nextPos = position & (1<<shiftTable[h] - 1)
		return idx, position & (1<<shiftTable[h] - 1)
	}

	idx = findPosition(node.sizes, position)
	if idx == 0 {
		nextPos = position
		return
	}
	nextPos = position - node.sizes[idx-1]
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

	start := 0 + len(t.head)
	end := t.size - len(t.tail)

	// [0 : head ) - [start : end) - [tail : size)

	//fmt.Println(t.size, "[0, ", len(t.head), ")", " [", start, " , ", end, ") [", end, t.size, ")", "-->", i)
	switch {
	case i < 0 || i >= t.size:
		panic("Index out of bounds")
	case i >= end: // look into tail
		return t.tail[i-end].value
	case i < start: // look into head
		return t.head[i].value

	default: // look into root
		n := t.root
		i -= start
		var slot int
		for h := t.h; h > 0; h-- {
			slot, i = navigate(n, h, i)
			n = n.children[slot]
		}
		return n.values[i].value

	}

}

func (t RRBTree[V]) Len() int {

	return t.size
}

func (t RRBTree[V]) Append(value V) RRBTree[V] {
	return t.push(value)
}

func (t RRBTree[V]) Prepend(value V) RRBTree[V] {
	return t.prepend(value)
}

func (t RRBTree[V]) Pop() (rrb RRBTree[V], value V, ok bool) {

	switch {
	case t.size == 0:
		return
	case len(t.tail) > 0:
		var v *refValue[V]
		t.tail, v = slice.Pop(t.tail)
		t.size--
		return t, v.value, true
	default:
		var v *refValue[V]
		//t.root, v, t.tail, t.h = pop(*t.root, t.h)
		t.root, v, t.tail = popBack(t.root, t.h)
		t.root, t.h = shrink(t.root, t.h)
		t.size--
		return t, v.value, true
	}

}

func (t RRBTree[V]) Slice(i, j int) RRBTree[V] {
	return t.slice(i, j)
}
