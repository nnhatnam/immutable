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

// truncate truncates items to the given length.
func (s *items[T]) truncate(length int) {
	var toClear items[T]
	*s, toClear = (*s)[:length], (*s)[length:]
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
	for k := 0; k < len(left); k++ {
		left[k] = zero
	}
	for k := 0; k < len(right); k++ {
		right[k] = zero
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
	owner *RRBTree[V]
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

	if n.cow != nil && n.cow == cow {
		return n
	}

	m := cow.newNode()
	m.treeSize = n.treeSize
	m.sizes = slice.Copy(n.sizes)
	m.children = slice.Copy(n.children)
	m.values = slice.Copy(n.values)

	return m
}

type mutableFlag int

const (
	mAPPEND mutableFlag = iota
	mPREPEND
	mPOP
	mUPDATE
	mDUPLICATE
)

func (n *node[V]) mutFor(cow *copyOnWriteContext[V], flag mutableFlag) *node[V] {
	switch flag {
	case mAPPEND, mPREPEND:
		if n.cow != nil && n.cow == cow {
			return n
		}
		m := cow.newNode()
		m.treeSize = n.treeSize

		if len(n.sizes) != 0 {
			m.sizes = make([]int, len(n.sizes), len(n.sizes)+1)
			copy(m.sizes, n.sizes)
		}

		m.children = make([]*node[V], len(n.children), len(n.children)+1)
		copy(m.children, n.children)
		return m
	case mUPDATE, mDUPLICATE:
		if n.cow != nil && n.cow == cow {
			return n
		}

		m := cow.newNode()
		m.treeSize = n.treeSize
		m.sizes = slice.Copy(n.sizes)
		m.children = slice.Copy(n.children)
		m.values = slice.Copy(n.values)
		return m

	default:
		panic("unreachable")
		return nil
	}
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

func (n *node[V]) isLeafNode() bool {
	return len(n.children) == 0
}

func (n *node[V]) isEmpty() bool {
	return len(n.children) == 0
}

func (n *node[V]) insertAt(h height, index int, child *node[V]) {
	n.treeSize = n.treeSize + child.treeSize
	n.children.insertAt(index, child)

	sizes := n.sizes

	if n.isRelaxedNode() {
		sizes.insertAt(index, child.treeSize)

		size := child.treeSize + n.readCumulativeSize(h, index-1)
		sizes[index] = size

		for i := index + 1; i < len(n.sizes); i++ {
			sizes[i] += child.treeSize
		}

	}

	if child.treeSize != cumulativeSumTable[h][0] {
		sizes = slice.Copy(cumulativeSumTable[h][:len(n.children)])
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

func (n *node[V]) setChild(h height, index int, child *node[V]) {

	if n.isBalancedNode() && child.treeSize != cumulativeSumTable[h][0] {

		// build the sizes table
		n.sizes = slice.Copy(cumulativeSumTable[h][:len(n.children)])
		n.sizes[len(n.children)-1] = n.treeSize
	}

	offset := n.readCumulativeSize(h, index) - n.readCumulativeSize(h, index-1)
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
func (n *node[V]) readCumulativeSize(h height, index int) int {

	switch {
	case n.isRelaxedNode():
		return n.sizes[index]
	case index == len(n.children)-1:
		return n.treeSize
	case index < 0:
		return 0
	default:
		return cumulativeSumTable[h][index]
	}

}

type branchType int
type height int

const (
	oldBranch branchType = iota
	newBranch
)

func (b branchType) String() string {
	switch b {
	case oldBranch:
		return "old branch"
	case newBranch:
		return "new branch"
	default:
		return "unknown"
	}
}

func (n *node[V]) truncateItemsFor(cow *copyOnWriteContext[V], at height, length int) (*node[V], height, items[*refValue[V]]) {
	if n.isLeafNode() {
		m := n.mutFor(cow, mDUPLICATE)
		m.values.truncate(length)
		if length < maxBranches {
			return nil, 0, m.values
		}
		m.treeSize = len(m.values)
		return m, 0, nil
	}

	var slot int
	i := length - 1
	slot, i = navigate(n, at, i)

	child, childHeight, tail := n.children[slot].truncateItemsFor(cow, at-1, i+1)

	if child == nil {

		if slot == 0 {
			return nil, at, tail
		}

		m := n.mutFor(cow, mDUPLICATE)

		n.treeSize = n.readCumulativeSize(at, slot-1)
		n.children.truncate(slot)
		n.sizes.maybeTruncate(slot)

		return m, at, tail
	}

	m := n.mutFor(cow, mDUPLICATE)
	m.treeSize = m.readCumulativeSize(at, slot)
	m.children.truncate(slot + 1)
	m.sizes.maybeTruncate(slot + 1)

	slotTreeSize := m.treeSize
	offset := child.treeSize - slotTreeSize

	m.treeSize += offset
	if m.isRelaxedNode() {
		m.sizes[slot] = m.treeSize
	}

	if at == childHeight+1 {
		m, at = shrink(m, at)
		return m, at, tail
	}

	return m, at, tail

}

func (n *node[V]) truncateFor(cow *copyOnWriteContext[V], h height, length int) (*node[V], items[*refValue[V]]) {
	m, _, tail := n.truncateItemsFor(cow, h, length)
	return m, tail
}

func (n *node[V]) truncate(h height, i int) (*node[V], items[*refValue[V]]) {

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

	return n, tail
}

func (n *node[V]) retainItemsFor(cow *copyOnWriteContext[V], at height, from int) (*node[V], height, items[*refValue[V]]) {
	if n.isLeafNode() {
		m := n.mutFor(cow, mDUPLICATE)
		m.values.retain(from)
		if len(m.values) < maxBranches {
			return nil, 0, m.values
		}
		m.treeSize = len(m.values)
		return m, 0, nil
	}

	var slot int
	slot, from = navigate(n, at, from)

	child, childHeight, head := n.children[slot].retainItemsFor(cow, at-1, from)

	if child == nil {
		if slot == len(n.children)-1 {
			return nil, at, head
		}

		slot = slot + 1
	}

	m := n.mutFor(cow, mDUPLICATE)

	slotSize := m.readCumulativeSize(at, slot)

	offset := child.treeSize - slotSize
	n.treeSize += offset
	m.children.retain(slot)
	m.sizes.maybeRetain(slot)

	for i := 0; i < len(n.sizes); i++ {
		n.sizes[i] = n.sizes[i] + offset
	}

	if at == childHeight+1 {
		m, at = shrink(m, at)
		return m, at, head
	}

	return m, at, head
}

func (n *node[V]) retainFor(cow *copyOnWriteContext[V], h height, i int) (*node[V], items[*refValue[V]]) {

	m, _, head := n.retainItemsFor(cow, h, i)
	return m, head

}

func (n *node[V]) retain(h height, i int) (*node[V], items[*refValue[V]]) {

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

	for i = 0; i < len(n.sizes); i++ {
		n.sizes[i] = n.sizes[i] + offset
	}

	return n, head
}

func (n *node[V]) pushFrontItemsFor(cow *copyOnWriteContext[V], isRoot bool, items []*refValue[V]) (*node[V], branchType, height) {
	if n.isLeafNode() {
		return cow.createLeaf(items), newBranch, 0
	}
	slotSize := n.children[0].treeSize

	child, bType, childHeight := n.children[0].pushFrontItemsFor(cow, false, items)

	pHeight := childHeight + 1

	if bType == oldBranch {
		m := n.mutFor(cow, mUPDATE)
		offset := child.treeSize - slotSize
		m.treeSize += offset

		if child.treeSize != cumulativeSumTable[pHeight][0] {
			m.sizes = slice.Copy(cumulativeSumTable[pHeight][:len(m.children)])
		}

		for i := 0; i < len(m.sizes); i++ {
			m.sizes[i] += offset
		}

		return m, oldBranch, childHeight + 1
	}

	// if we get here, it means we have a new branch
	// so we either push the child to the current node or create a new node if the current node is full
	if n.hasFullChildren() {

		// if h == childHeight+1, it means that the child is the root.
		// In this case, we need to create a new root with its children consisting of the child and the old root
		if isRoot {
			size := n.treeSize + child.treeSize
			var sizes []int
			if n.isRelaxedNode() || child.isRelaxedNode() {
				sizes = []int{child.treeSize, size}
			}
			return cow.createInternalNode(size, sizes, child, n), newBranch, pHeight + 1
		}

		if child.isRelaxedNode() {
			return cow.createInternalNode(child.treeSize, []int{child.treeSize}, child), newBranch, pHeight
		}

		return cow.createInternalNode(child.treeSize, nil, child), newBranch, pHeight
	}

	// if we get here, it means we are have space to prepend new child, and we are not at the root yet
	m := n.mutFor(cow, mPREPEND)
	offset := child.treeSize
	m.treeSize += offset
	m.children.insertAt(0, child)

	if child.treeSize != cumulativeSumTable[pHeight][0] {
		m.sizes = slice.Copy(cumulativeSumTable[pHeight][:len(m.children)])
		offset = child.treeSize - m.sizes[0]
	}

	for i := 0; i < len(m.sizes); i++ {
		m.sizes[i] += offset
	}
	return m, oldBranch, childHeight + 1

}

func (n *node[V]) pushFrontItems(cow *copyOnWriteContext[V], items []*refValue[V]) (*node[V], branchType, height) {
	return n.pushFrontItemsFor(cow, true, items)
}

func (n *node[V]) pushHead(h height, head []*refValue[V]) (*node[V], height) {

	var pushHeadRecursive func(*node[V], height) (*node[V], branchType)

	pushHeadRecursive = func(m *node[V], h height) (*node[V], branchType) {

		if h == 0 {
			return m.cow.createLeaf(head), newBranch
		}

		m.children[0] = m.children[0].mutableFor(m.cow)
		slotSize := m.children[0].treeSize

		child, bType := pushHeadRecursive(m.children[0], h-1)

		if bType == oldBranch {

			offset := child.treeSize - slotSize
			m.treeSize += offset

			if child.treeSize != cumulativeSumTable[h][0] {
				m.sizes = slice.Copy(cumulativeSumTable[h][:len(m.children)])
			}

			for i := 0; i < len(m.sizes); i++ {
				m.sizes[i] += offset
			}

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

		offset := child.treeSize
		m.treeSize += offset
		m.children.insertAt(0, child)

		if child.treeSize != cumulativeSumTable[h][0] {
			m.sizes = slice.Copy(cumulativeSumTable[h][:len(m.children)])
			offset = child.treeSize - m.sizes[0]
		}

		for i := 0; i < len(m.sizes); i++ {
			m.sizes[i] += offset
		}
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

func (n *node[V]) pushItemsFor(cow *copyOnWriteContext[V], isRoot bool, items []*refValue[V]) (*node[V], branchType, height) {
	//if debugId == 1087 {
	//	fmt.Println("pushItemsFor", isRoot, items)
	//}
	if len(n.children) == 0 {
		// we are in leaf node
		leaf := cow.createLeaf(items)
		if isRoot {
			treeSize := n.treeSize + leaf.treeSize
			if n.isRelaxedNode() || leaf.isRelaxedNode() {
				return cow.createInternalNode(treeSize, []int{n.treeSize, treeSize}, n, leaf), newBranch, 1
			}
			return cow.createInternalNode(treeSize, nil, n, leaf), newBranch, 1
		}
		return leaf, newBranch, 0
	}

	slot := len(n.children) - 1
	slotTreeSize := n.children[slot].treeSize

	child, bType, childHeight := n.children[slot].pushItemsFor(cow, false, items)

	pHeight := childHeight + 1

	//if debugId == 1087 {
	//	fmt.Println("received child", bType, childHeight, pHeight, child)
	//}

	if bType == oldBranch {
		m := n.mutFor(cow, mUPDATE)
		offset := child.treeSize - slotTreeSize
		m.treeSize += offset
		if m.isRelaxedNode() {
			m.sizes[slot] = m.treeSize
		}
		m.children[slot] = child
		return m, oldBranch, pHeight
	}

	// if we get here, it means we have a new branch
	// so we either push the child to the current node or create a new node if the current node is full
	if n.isFullNode() {

		m := cow.createInternalNode(child.treeSize, nil, child)

		if child.isRelaxedNode() {
			m.sizes = []int{child.treeSize}
		}

		if isRoot {
			size := n.treeSize + m.treeSize

			var sizes []int
			if n.isRelaxedNode() || m.isRelaxedNode() {
				sizes = []int{n.treeSize, size}
			}
			return cow.createInternalNode(size, sizes, n, m), newBranch, pHeight + 1
		}

		return m, newBranch, pHeight
	}

	// if we get here, it means we have a new branch and the current node is not full.
	// So we push the child to the current node

	m := n.mutFor(cow, mAPPEND)
	m.treeSize += child.treeSize
	m.children = append(m.children, child)

	if m.isRelaxedNode() {
		m.sizes = append(m.sizes, m.treeSize)
	} else if child.isRelaxedNode() {
		m.sizes = slice.Copy(cumulativeSumTable[pHeight][:slot+2])
		m.sizes[slot+1] = m.treeSize
	}

	return m, oldBranch, pHeight
}

func (n *node[V]) pushItems(cow *copyOnWriteContext[V], items []*refValue[V]) (*node[V], height) {
	m, _, h := n.pushItemsFor(cow, true, items)
	return m, h
}

// pushTail pushes the tail into the tree rooted at the given node `n`.
func (n *node[V]) pushTail(h height, tail []*refValue[V]) (*node[V], height) {

	var pushTailRecursive func(*node[V], height) (*node[V], branchType)

	pushTailRecursive = func(m *node[V], h height) (*node[V], branchType) {

		if h == 0 {
			return m.cow.createLeaf(tail), newBranch
		}

		slot := len(m.children) - 1
		slotTreeSize := m.children[slot].treeSize

		m.children[slot] = m.children[slot].mutableFor(m.cow)

		child, bType := pushTailRecursive(m.children[slot], h-1)

		if bType == oldBranch {
			//out := m.mutableFor(m.cow)
			//out.setChild(h, len(out.children)-1, child)
			offset := child.treeSize - slotTreeSize
			m.treeSize += offset
			if m.isRelaxedNode() {
				m.sizes[slot] = m.treeSize
			}

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

		m.treeSize += child.treeSize
		m.children = append(m.children, child)

		if m.isRelaxedNode() {
			m.sizes = append(m.sizes, m.treeSize)
		} else if child.isRelaxedNode() {
			m.sizes = slice.Copy(cumulativeSumTable[h][:slot+2])
			m.sizes[slot+1] = m.treeSize
		}

		//out := m.mutableForWithCustomAllocs(m.cow, cloneWithOneExtraCap)
		//out.addChild(child)
		return m, oldBranch
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

func (n *node[V]) popItem(cow *copyOnWriteContext[V], h height) (*node[V], height, *refValue[V], items[*refValue[V]]) {

	if n.isLeafNode() {

		m := n.mutFor(cow, mPOP)

		ref := m.values.pop()

		if len(m.values) < maxBranches {
			return nil, 0, ref, m.values
		}

		m.treeSize--
		return m, 0, ref, nil
	}

	slot := len(n.children) - 1
	slotTreeSize := n.children[slot].treeSize

	child, childHeight, value, tail := n.children[slot].popItem(cow, h)

	if child == nil {

		if slot == 0 {
			return nil, childHeight + 1, value, tail
		}

		m := n.mutFor(cow, mDUPLICATE)
		m.treeSize -= slotTreeSize
		m.sizes.pop()
		m.children.pop()

		return m, childHeight + 1, value, tail
	}

	offset := child.treeSize - slotTreeSize
	m := n.mutFor(cow, mDUPLICATE)
	m.treeSize += offset
	if m.isRelaxedNode() {
		m.sizes[slot] = m.treeSize
	}

	if h == childHeight+1 {
		m, h = shrink(m, h)
		return m, h, value, tail
	}

	return m, childHeight + 1, value, tail
}

func (n *node[V]) pop(h height) (_ *node[V], _ height, value *refValue[V], tail items[*refValue[V]]) {

	var popRecursive func(*node[V], height) *node[V]

	popRecursive = func(m *node[V], h height) *node[V] {
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
		slotTreeSize := m.children[slot].treeSize

		m.children[slot] = m.children[slot].mutableFor(m.cow)

		// this is the start of recursive call
		child := popRecursive(m.children[slot], h-1)

		if child == nil {

			if slot == 0 {
				return nil
			}

			m.treeSize -= slotTreeSize
			m.sizes.pop()
			m.children.pop()

			return m
		}

		offset := child.treeSize - slotTreeSize

		m.treeSize += offset
		if m.isRelaxedNode() {
			m.sizes[slot] = m.treeSize
		}

		return m

	}

	m := popRecursive(n, h)
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

	return len(n.values) != maxBranches

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

func (n *node[V]) walk(h height, i int) []*node[V] {
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

type RRBTree[V any] struct {
	// The root node of the tree.
	root *node[V]

	h height // The current level of the trie. It's counted from the bottom of the trie.

	size int // The number of elements in the tree.

	head items[*refValue[V]] // The head of the tree.

	tail items[*refValue[V]] // The tail of the tree.

	cow *copyOnWriteContext[V]
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

func shrink[V any](n *node[V], h height) (m *node[V], newH height) {

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

	t.tail = slice.Push(t.tail, newRefValue[V](value))
	t.size++

	if len(t.tail) == maxBranches {
		// make a new branch
		if t.root == nil {
			//t.cow = &copyOnWriteContext[V]{}
			t.root = (*copyOnWriteContext[V])(nil).createLeaf(t.tail)
			t.tail = nil
			return t
		}

		// pushTail may create a new root, so in some cases we don't need to clone the current root.
		// For that reason, clone the root will be decided inside pushTail.
		t.root = t.root.mutableFor(nil)
		t.root, t.h = t.root.pushTail(t.h, t.tail)
		t.tail = nil

	}

	return t

}

func (t RRBTree[V]) prepend(value V) RRBTree[V] {

	t.head = slice.PushFront(t.head, newRefValue[V](value))
	t.size++

	if len(t.head) == maxBranches {
		// make a new branch

		if t.root == nil {
			t.root = (*copyOnWriteContext[V])(nil).createLeaf(t.head)
			t.head = nil
			return t
		}

		t.root = t.root.mutableFor(nil)
		t.root, t.h = t.root.pushHead(t.h, t.head)
		t.head = nil
	}

	return t
}

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

func navigate[V any](node *node[V], h height, position int) (idx, nextPos int) {

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

	t.tail = slice.Push(t.tail, newRefValue[V](value))
	t.size++

	if len(t.tail) == maxBranches {
		// make a new branch
		if t.root == nil {
			//t.cow = &copyOnWriteContext[V]{}
			t.root = (*copyOnWriteContext[V])(nil).createLeaf(t.tail)
			t.tail = nil
			return t
		}

		//fmt.Println("before pushItems: ", t.root.treeSize, t.root.sizes, t.root.children, debugId)
		t.root, t.h = t.root.pushItems(nil, t.tail)
		//fmt.Println("after pushItems: ", t.root.treeSize, t.root.sizes, t.root.children)
		t.tail = nil
	}

	return t
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
		t.root = t.root.mutableFor(nil)
		t.root, t.h, v, t.tail = t.root.pop(t.h)
		t.size--

		if v == nil {
			return t, value, false
		}
		return t, v.value, true
	}

}

func (t RRBTree[V]) Slice(i, j int) RRBTree[V] {
	return t.slice(i, j)
}
