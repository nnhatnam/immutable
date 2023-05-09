package hamt

import (
	"math/bits"
	"sync/atomic"
	"unsafe"

	"github.com/nnhatnam/immutable/slice"
)

const (
	childPerNode   = 64
	arity          = 6
	width          = 2
	exhaustedLevel = childPerNode / arity

	maxDepth = 55 // 64-bit hash, 6-bit per level, 64/6 ~ 11, 11*5 = 55 (accepted 5 rehashes)
)

type Hasher[K comparable] interface {
	Hash(key K) uint64

	// Rehash is used to rehash the key when the key is already hashed for the given level
	Rehash(key K, prevHashCount int) uint64
}

type mapNode[K comparable, V any] struct {
	bitmap   uint64
	refCount int32

	contentArray []unsafe.Pointer // either suffix hash or []entry[K, V], *Value is stored in even slots, []entry[K, V] in odd slots.
}

//func newMapNode[K comparable, V any]() *mapNode[K, V] {
//	return &mapNode[K, V]{
//		contentArray: make([]unsafe.Pointer, 0),
//	}
//}

func newMapNodeWithRef[K comparable, V any]() *mapNode[K, V] {
	return &mapNode[K, V]{
		contentArray: make([]unsafe.Pointer, 0),
		refCount:     1,
	}
}

func (n *mapNode[K, V]) shallowCloneWithRef() *mapNode[K, V] {
	//atomic.AddInt32(&n.refCount, -1)
	n.decRef()

	n1 := &mapNode[K, V]{
		bitmap:   n.bitmap,
		refCount: 1,
	}

	n1.contentArray = make([]unsafe.Pointer, len(n.contentArray))
	for i := 0; i < len(n.contentArray)/2; i++ {

		recordIdx := i * 2
		nodeIdx := i*2 + 1
		n1.contentArray[recordIdx] = n.contentArray[recordIdx]
		if n.contentArray[nodeIdx] != nil {
			n1.contentArray[nodeIdx] = n.contentArray[nodeIdx]
			(*mapNode[K, V])(n1.contentArray[nodeIdx]).incRef()
		}

	}
	//copy(n1.contentArray, n.contentArray)

	return n1

}

//func (n *mapNode[K, V]) shallowClone() *mapNode[K, V] {
//	n1 := &mapNode[K, V]{
//		bitmap:   n.bitmap,
//		refCount: 1,
//	}
//
//	n1.contentArray = make([]unsafe.Pointer, len(n.contentArray))
//	copy(n1.contentArray, n.contentArray)
//
//	return n1
//}

func (n *mapNode[K, V]) incRef() *mapNode[K, V] {
	if n != nil {
		atomic.AddInt32(&n.refCount, 1)
	}
	return n
}

func (n *mapNode[K, V]) decRef() {
	if n == nil {
		return
	}

	if atomic.AddInt32(&n.refCount, -1) == 0 {
		// free the node
		for i := 0; i < len(n.contentArray)/2; i++ {
			recordIdx := i * 2
			nodeIdx := i*2 + 1

			if n.contentArray[recordIdx] != nil {
				// free the record
				n.contentArray[recordIdx] = nil
			}

			if n.contentArray[nodeIdx] != nil {
				// free the node
				n1 := (*mapNode[K, V])(n.contentArray[nodeIdx])
				n1.decRef()
			}
		}
		n.contentArray = nil
	}
}

func (n *mapNode[K, V]) contentBlockInfo(loc int) (mask uint64, blockIndex int) {
	mask = 1 << loc

	if mask > 0 {
		// find the block index
		blockIndex = bits.OnesCount64(n.bitmap & (mask - 1))
	}

	return
}

//func (n *mapNode[K, V]) getContentIdx(loc int) (recordIdx int, nodeIdx int) {
//	var mask uint64 = 1 << loc
//	blockIdx := 0
//
//	if mask > 0 {
//		// find the block index
//		blockIdx = bits.OnesCount64(n.bitmap & (mask - 1))
//	}
//
//	recordIdx = width * blockIdx
//	nodeIdx = recordIdx + 1
//	return recordIdx, nodeIdx
//}

func (n *mapNode[K, V]) getContentIndexFromMask(mask uint64) (recordIdx int, nodeIdx int) {

	blockIdx := 0

	if mask > 0 {
		// find the block index
		blockIdx = bits.OnesCount64(n.bitmap & (mask - 1))
	}

	recordIdx = width * blockIdx
	nodeIdx = recordIdx + 1
	return recordIdx, nodeIdx
}

// TryInsertBlock tries to set the record at the given location. If the location is already occupied, return false
//func (n *mapNode[K, V]) TryInsertBlock(loc int, r *record[K, V], n1 *mapNode[K, V]) bool {
//
//	mask, blockIdx := n.contentBlockInfo(loc)
//
//	// if the block is not empty, we can't set. return false
//	if n.bitmap&mask != 0 {
//		return false
//	}
//
//	recordIdx := width * blockIdx
//
//	n.contentArray = slices.Insert(n.contentArray, recordIdx, unsafe.Pointer(r), unsafe.Pointer(n1))
//	n.bitmap |= mask
//	return true
//
//}

//func (n *mapNode[K, V]) UpdateBlock(loc int, r *record[K, V], n1 *mapNode[K, V]) bool {
//
//	_, blockIdx := n.contentBlockInfo(loc)
//	recordIdx := width * blockIdx
//
//	_ = n.contentArray[recordIdx] // bounds check hint to compiler; see golang.org/issue/14808
//
//	//n.bitmap |= mask
//	n.contentArray[recordIdx] = unsafe.Pointer(r)
//	n.contentArray[recordIdx+1] = unsafe.Pointer(n1)
//
//	return true
//
//}

func (n *mapNode[K, V]) TryGetBlock(loc int) (*record[K, V], *mapNode[K, V]) {

	_, blockIdx := n.contentBlockInfo(loc)
	recordIdx := width * blockIdx
	nodeIdx := recordIdx + 1
	if recordIdx >= len(n.contentArray) {
		return nil, nil
	}

	//_ = n.contentArray[recordIdx:nodeIdx] // bounds check hint to compiler; see golang.org/issue/14808

	return (*record[K, V])(n.contentArray[recordIdx]), (*mapNode[K, V])(n.contentArray[nodeIdx])
}

type PersistentHAMT[K comparable, V any] struct {
	root *mapNode[K, V]
	len  int

	hasher Hasher[K]

	mutable bool // if true, the HAMT is mutable, otherwise it is immutable.
}

func NewPersistentHAMT[K comparable, V any](h Hasher[K]) *PersistentHAMT[K, V] {
	return &PersistentHAMT[K, V]{
		hasher: h,
		root:   newMapNodeWithRef[K, V](),
	}
}

func (m *PersistentHAMT[K, V]) hash(key K, depth int) uint64 {

	prevHashCount := depth / (exhaustedLevel + 1)
	if prevHashCount == 0 {
		return m.hasher.Hash(key)
	} else if depth > maxDepth {
		panic("hash depth too deep, which mean there are too many collisions in the chosen hash function. Please choose a better hash function.")
	}

	return m.hasher.Rehash(key, prevHashCount)
}

func (m *PersistentHAMT[K, V]) Len() int {
	return m.len
}

// subroutine of mInsert
func (m *PersistentHAMT[K, V]) mInsertDoubleRecord(n *mapNode[K, V], keyHash uint64, r1 *record[K, V], colHash uint64, r2 *record[K, V], depth int) *mapNode[K, V] {
	// The double record situation only happens when we have a collision at the previous level.
	// When a collision happens, we have to create a new node and insert the two records into the new node.
	// So we know that n is a new node for sure.

	level := depth % (exhaustedLevel + 1)
	shift := level * arity

	bitpos := bucket(keyHash, shift)
	colpos := bucket(colHash, shift)

	var mask uint64 = 1 << bitpos
	recordIdx, _ := n.getContentIndexFromMask(mask)

	if bitpos == colpos { // collision again
		// we have to create a new node

		n1 := newMapNodeWithRef[K, V]()
		n.contentArray = slice.Insert(n.contentArray, recordIdx, nil, unsafe.Pointer(n1))
		n.bitmap |= mask

		if level == exhaustedLevel {
			keyHash = m.hash(r1.key, depth+1)
			colHash = m.hash(r2.key, depth+1)
		}

		m.mInsertDoubleRecord(n1, keyHash, r1, colHash, r2, depth+1)
		return n
	}

	n.contentArray = slice.Insert(n.contentArray, recordIdx, unsafe.Pointer(r1), nil)
	n.bitmap |= mask
	mask = 1 << colpos

	recordIdx, _ = n.getContentIndexFromMask(mask)

	n.contentArray = slice.Insert(n.contentArray, recordIdx, unsafe.Pointer(r2), nil)
	n.bitmap |= mask
	m.len++
	return n
}

func (m *PersistentHAMT[K, V]) replaceOrInsert(n *mapNode[K, V], keyHash uint64, depth int, r *record[K, V], pathCopy bool) *mapNode[K, V] {

	if n.refCount > 1 || pathCopy {
		pathCopy = true
		n = n.shallowCloneWithRef()
		//fmt.Println("must shallow clone")
	}

	level := depth % (exhaustedLevel + 1)
	shift := level * arity
	loc := bucket(keyHash, shift)

	var mask uint64 = 1 << loc

	recordIdx, nodeIdx := n.getContentIndexFromMask(mask)

	// if the block is empty, we can insert the record directly
	if n.bitmap&mask == 0 {

		n.contentArray = slice.Insert(n.contentArray, recordIdx, unsafe.Pointer(r), nil)
		n.bitmap |= mask
		m.len++
		return n
	}

	// if the block is not empty, we have to check if there is a collision
	colRecord, n1 := (*record[K, V])(n.contentArray[recordIdx]), (*mapNode[K, V])(n.contentArray[nodeIdx])

	//fmt.Println("col and record: ", colRecord, n1)
	if colRecord == nil {

		if level == exhaustedLevel {
			keyHash = m.hash(r.key, depth+1)
		}
		n1 = m.replaceOrInsert(n1, keyHash, depth+1, r, pathCopy)

		n.contentArray[nodeIdx] = unsafe.Pointer(n1)
		return n
	} else if n1 == nil {
		// n1 == nil && colRecord != nil => update or expand
		if colRecord.key == r.key { // update

			n.contentArray[recordIdx] = unsafe.Pointer(r)
			//m.len++
			return n
		}

		n.contentArray[recordIdx] = nil
		n2 := newMapNodeWithRef[K, V]()

		colHash := m.hash(colRecord.key, depth)
		if level == exhaustedLevel {
			keyHash = m.hash(r.key, depth)
		}
		n.contentArray[recordIdx] = nil // remove the record
		n2 = m.mInsertDoubleRecord(n2, keyHash, r, colHash, colRecord, depth+1)

		n.contentArray[nodeIdx] = unsafe.Pointer(n2)

	}

	return n
}

func (m *PersistentHAMT[K, V]) delete(n *mapNode[K, V], k K, keyHash uint64, depth int, pathCopy bool) (*mapNode[K, V], bool) {

	if n.refCount > 1 || pathCopy {
		pathCopy = true
		n = n.shallowCloneWithRef()
		//fmt.Println("must shallow clone")
	}

	level := depth % (exhaustedLevel + 1)
	shift := level * arity
	loc := bucket(keyHash, shift)

	var mask uint64 = 1 << loc
	recordIdx, nodeIdx := n.getContentIndexFromMask(mask)
	// can't find the key to delete
	if n.bitmap&mask == 0 {
		return n, false
	}

	colRecord, n1 := (*record[K, V])(n.contentArray[recordIdx]), (*mapNode[K, V])(n.contentArray[nodeIdx])

	if colRecord == nil {
		if level == exhaustedLevel {
			keyHash = m.hash(k, depth+1)
		}
		var deleted bool
		n1, deleted = m.delete(n1, k, keyHash, depth+1, pathCopy)
		n.contentArray[nodeIdx] = unsafe.Pointer(n1)
		if n1 == nil {
			n.contentArray = slice.RemoveRange(n.contentArray, recordIdx, nodeIdx+1)
			n.bitmap ^= mask
			if n.bitmap == 0 {
				return nil, true
			}
		}

		return n, deleted
	}

	if colRecord.key == k {
		n.contentArray[recordIdx] = nil
		n.contentArray = slice.RemoveRange(n.contentArray, recordIdx, nodeIdx+1)
		n.bitmap ^= mask
		m.len--

		// if there is only one record left, remove the node
		if n.bitmap == 0 {
			return nil, true
		}
		return n, true
	}

	return n, false
}

func (m *PersistentHAMT[K, V]) get(n *mapNode[K, V], k K, keyHash uint64, depth int) (_ V, _ bool) {

	level := depth % (exhaustedLevel + 1)
	shift := level * arity

	loc := bucket(keyHash, shift)

	colRecord, n1 := n.TryGetBlock(loc)

	if colRecord == nil && n1 == nil {
		return
	}

	if colRecord != nil {
		if colRecord.key == k {
			return colRecord.value, true
		}
		return
	}

	if level == exhaustedLevel {
		keyHash = m.hash(k, depth+1)
	}
	return m.get(n1, k, keyHash, depth+1)

}

func (m *PersistentHAMT[K, V]) Set(k K, v V) {
	keyHash := m.hash(k, 0)

	if m.root == nil {
		m.root = newMapNodeWithRef[K, V]()
	}

	m.root = m.replaceOrInsert(m.root, keyHash, 0, newRecord[K, V](k, v), false)

}

func (m *PersistentHAMT[K, V]) Get(k K) (_ V, _ bool) {
	if m.root == nil {
		return
	}

	keyHash := m.hash(k, 0)
	return m.get(m.root, k, keyHash, 0)
}

func (m *PersistentHAMT[K, V]) Clone() *PersistentHAMT[K, V] {
	return &PersistentHAMT[K, V]{
		root:   m.root.incRef(),
		len:    m.len,
		hasher: m.hasher,
	}
}

func (m *PersistentHAMT[K, V]) Delete(k K) bool {
	if m.root == nil {
		return false
	}

	keyHash := m.hash(k, 0)
	var deleted bool
	m.root, deleted = m.delete(m.root, k, keyHash, 0, false)
	return deleted
}
