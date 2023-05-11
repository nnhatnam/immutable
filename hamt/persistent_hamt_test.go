package hamt

import (
	"fmt"
	"golang.org/x/exp/slices"
	"math/rand"
	"reflect"
	"sync/atomic"
	"testing"
	"unsafe"
)

//func dfs[K comparable, V any](t *testing.T, n *mapNode[K, V]) []*mapNode[K, V] {
//	t.Helper()
//	return dfsRef(t, n)
//}

func dfsRef[K comparable, V any](t *testing.T, n *mapNode[K, V], f func(n *mapNode[K, V]) bool) {
	t.Helper()
	if n == nil {
		return
	}

	if f(n) {
		return
	}

	for i := 0; i < len(n.contentArray)/2; i++ {
		nodeIdx := width*i + 1
		if n.contentArray[nodeIdx] != nil {
			//if f((*mapNode[K, V])(n.contentArray[nodeIdx])) {
			//	return
			//}
			dfsRef(t, (*mapNode[K, V])(n.contentArray[nodeIdx]), f)
		}
	}

	return
}

func TestNewPersistentHAMT(t *testing.T) {

	trie := NewPersistentHAMT[string, int](newHasher[string]())

	trie.Set("immutable", 1)

	if trie.Len() != 1 {
		t.Errorf("trie.Len() = %d, want %d", trie.Len(), 1)
	}

	result, ok := trie.Get("immutable")

	if !ok || result != 1 {
		t.Errorf("trie.Get() = %d, want %d", result, 1)
	}

}

func insertItem(t *testing.T, trie *PersistentHAMT[string, int], key string, value int, expectedLen int) {
	t.Helper()

	trie.Set(key, value)

	if trie.Len() != expectedLen {
		t.Errorf("trie.Len() = %d, want %d", trie.Len(), expectedLen)
	}

	result, ok := trie.Get(key)

	if !ok || result != value {
		t.Errorf("trie.Get() = %d, want %d", result, value)
	}
}

func TestPersistentHAMTInsertion(t *testing.T) {

	t.Run("Basic Insertion", func(t *testing.T) {
		trie := NewPersistentHAMT[string, int](newCollisionHasher[string]())

		insertItem(t, trie, "a", 1, 1)
		insertItem(t, trie, "b", 2, 2)
		insertItem(t, trie, "c", 3, 3)

		insertItem(t, trie, "d", 4, 4)
		insertItem(t, trie, "rehash2time_1", 5, 5)
		insertItem(t, trie, "rehash2time_2", 6, 6)

		insertItem(t, trie, "panic1", 7, 7)

		if !panics(func() { trie.Set("panic2", 8) }) {
			t.Errorf("trie.ReplaceOrInsert() should panic")
		}
	})

	t.Run("Random Insertion", func(t *testing.T) {
		trie := NewPersistentHAMT[string, int](newHasher[string]())

		inp := make(map[string]int)
		for i := 0; i < 10000; i++ {
			gen := generateUUID()
			inp[gen] = i
			insertItem(t, trie, gen, i, i+1)
		}

		for k, v := range inp {
			result, ok := trie.Get(k)
			if !ok || result != v {
				t.Errorf("trie.Get() = %d, want %d", result, v)
			}
		}

	})

}

func validateSharedNode[K comparable, V any](t *testing.T, trie *PersistentHAMT[K, V], clone *PersistentHAMT[K, V], expectedNodeShared int) []*mapNode[K, V] {
	t.Helper()

	var sharedNodes []*mapNode[K, V]
	var trieSharedNode []uintptr
	dfsRef[K, V](t, trie.root, func(n *mapNode[K, V]) bool {
		if n.refCount > 1 {
			sharedNodes = append(sharedNodes, n)
			trieSharedNode = append(trieSharedNode, uintptr(unsafe.Pointer(n)))
		}
		return false
	})

	var cloneSharedNode []uintptr

	dfsRef[K, V](t, clone.root, func(n *mapNode[K, V]) bool {
		if n.refCount > 1 {
			cloneSharedNode = append(cloneSharedNode, uintptr(unsafe.Pointer(n)))
		}
		return false
	})

	if slices.Compare(trieSharedNode, cloneSharedNode) != 0 {
		t.Errorf("the two tries don't shared node with each other = %v, want %v", trieSharedNode, cloneSharedNode)
		return sharedNodes
	}

	if len(sharedNodes) != expectedNodeShared {
		t.Errorf("sharedNodes = %d, want %d", len(sharedNodes), expectedNodeShared)
		return sharedNodes
	}

	return sharedNodes

}

func TestPersistentHAMTClone(t *testing.T) {

	t.Run("Basic Clone", func(t *testing.T) {
		trie := NewPersistentHAMT[string, int](newCollisionHasher[string]())

		insertItem(t, trie, "a", 1, 1)
		insertItem(t, trie, "b", 2, 2)
		insertItem(t, trie, "c", 3, 3)

		insertItem(t, trie, "d", 4, 4)
		insertItem(t, trie, "rehash2time_1", 5, 5)
		insertItem(t, trie, "rehash2time_2", 6, 6)

		insertItem(t, trie, "panic1", 7, 7)

		clone := trie.Clone()

		if clone.Len() != trie.Len() {
			t.Errorf("clone.Len() = %d, want %d", clone.Len(), trie.Len())
		}

		if clone.root != trie.root {
			t.Errorf("clone.root = %v, want %v", clone.root, trie.root)
		}

		if trie.root.refCount != 2 {
			t.Errorf("trie.root.refCount = %d, want %d", trie.root.refCount, 2)
		}

		sharedNodes := validateSharedNode(t, trie, clone, 1)

		if len(sharedNodes) != 1 {
			t.Errorf("sharedNodes count = %d, want %d", len(sharedNodes), 1)
		}

		//t.Log("sharedNodes count = ", len(sharedNodes))
		//clone.Set("a", 10)

		insertItem(t, clone, "a", 10, trie.Len())
		sharedNodes = validateSharedNode(t, trie, clone, 2)

		result, ok := trie.Get("a")
		if !ok || result != 1 {
			t.Errorf("trie.Get() = %d, want %d", result, 1)
		}

		result, ok = clone.Get("a")
		if !ok || result != 10 {
			t.Errorf("clone.Get() = %d, want %d", result, 10)
		}

		insertItem(t, clone, "b", 20, trie.Len())
		insertItem(t, clone, "c", 30, trie.Len())
		insertItem(t, clone, "d", 40, trie.Len())
		insertItem(t, clone, "rehash2time_1", 50, trie.Len())
		insertItem(t, clone, "rehash2time_2", 60, trie.Len())
		insertItem(t, clone, "panic1", 70, trie.Len())

		sharedNodes = validateSharedNode(t, trie, clone, 0)

		if len(sharedNodes) != 0 {
			t.Errorf("sharedNodes count = %d, want %d", len(sharedNodes), 0)
		}

		if !panics(func() { clone.Set("panic2", 8) }) {
			t.Errorf("trie.ReplaceOrInsert() should panic")
		}

		insertItem(t, clone, "e", 80, trie.Len()+1)
	})

	t.Run("Basic Clone with collision", func(t *testing.T) {
		trie := NewPersistentHAMT[string, int](newCollisionHasher[string]())

		insertItem(t, trie, "a", 1, 1)
		insertItem(t, trie, "b", 2, 2)
		insertItem(t, trie, "c", 3, 3)
		insertItem(t, trie, "d", 4, 4)

		clone := trie.Clone()

		insertItem(t, clone, "col_with_a", 8, trie.Len()+1)
		insertItem(t, clone, "col_with_b", 8, trie.Len()+2)
		insertItem(t, clone, "col_with_c", 8, trie.Len()+3)
		insertItem(t, clone, "col_with_d", 8, trie.Len()+4)

		validateSharedNode(t, trie, clone, 0)

		for _, v := range []string{"a", "b", "c", "d"} {
			r1, ok1 := trie.Get(v)
			r2, ok2 := clone.Get(v)

			if !ok1 || !ok2 || r1 != r2 {
				t.Errorf("trie.Get(%s) = %d, want %d", v, r1, r2)
			}
		}
	})

	inp2 := make(map[string]int)
	for i := 0; i < 10000; i++ {
		gen := generateUUID()
		inp2[gen] = i
	}

	inp2["extra"] = 10001

	t.Run("Bulk insert", func(t *testing.T) {
		trie := NewPersistentHAMT[string, int](newHasher[string]())

		inp := make(map[string]int)
		for i := 0; i < 10000; i++ {
			gen := generateUUID()
			inp[gen] = i
			insertItem(t, trie, gen, i, i+1)
		}

		clone := trie.Clone()

		for k, v := range inp {
			insertItem(t, clone, k, v, trie.Len())
		}

		validateSharedNode(t, trie, clone, 0)

		for k, v := range inp {

			if r, ok := trie.Get(k); !ok || r != v {
				t.Errorf("trie.Get(%s) = %d, want %d", k, r, v)
			}

		}

	})

	t.Run("Bulk insert 2", func(t *testing.T) {
		trie := NewPersistentHAMT[string, int](newHasher[string]())
		clone := trie.Clone()

		inp := make(map[string]int)
		for i := 0; i < 10000; i++ {
			gen := generateUUID()
			inp[gen] = i

			if i%2 == 0 {
				insertItem(t, trie, gen, i, trie.Len()+1)
			} else {
				insertItem(t, clone, gen, i, clone.Len()+1)
			}

		}

		if trie.Len() != 5000 {
			t.Errorf("trie.Len() = %d, want %d", trie.Len(), 5000)
		}

		if clone.Len() != 5000 {
			t.Errorf("clone.Len() = %d, want %d", clone.Len(), 5000)
		}

		validateSharedNode(t, trie, clone, 0)

		for k, v := range inp {

			if v%2 == 0 {
				if r, ok := trie.Get(k); !ok || r != v {
					t.Errorf("trie.Get(%s) = %d, want %d", k, r, v)
				}
			} else {
				if r, ok := clone.Get(k); !ok || r != v {
					t.Errorf("clone.Get(%s) = %d, want %d", k, r, v)
				}

			}

		}
	})

}

func deleteItem(t *testing.T, trie *PersistentHAMT[string, int], key string, expectedLen int, expectedDeletion bool) {
	t.Helper()
	deleted := trie.Delete(key)

	if trie.Len() != expectedLen {
		t.Errorf("trie.Len() = %d, want %d", trie.Len(), expectedLen)
	}

	if deleted != expectedDeletion {
		t.Errorf("Expected the key %s to be deleted = %v, but receive %v", key, expectedDeletion, deleted)
	}

	if expectedDeletion {
		result, ok := trie.Get(key)
		if ok {
			t.Errorf("Expected the key %s to be deleted = %v, but receive %v %d", key, expectedDeletion, ok, result)
		}
	}

}

func TestPersistentHAMTDelete(t *testing.T) {
	t.Run("Basic Deletion", func(t *testing.T) {
		trie := NewPersistentHAMT[string, int](newCollisionHasher[string]())

		for k, v := range []string{"a", "b", "c", "d", "rehash2time_1", "rehash2time_2", "panic1"} {
			insertItem(t, trie, v, k+1, k+1)
		}

		for k, v := range []string{"a", "b", "c", "d", "rehash2time_1", "rehash2time_2", "panic1"} {

			deleteItem(t, trie, v, 6-k, true)
		}

		if trie.root != nil {
			t.Errorf("trie.root = %v, want %v", trie.root, nil)
		}

		for k, v := range []string{"a", "b", "c", "d", "rehash2time_1", "rehash2time_2", "panic1"} {
			insertItem(t, trie, v, k+1, k+1)
		}

		clone := trie.Clone()

		for k, v := range []string{"a", "b", "c", "d", "rehash2time_1", "rehash2time_2", "panic1"} {

			deleteItem(t, clone, v, 6-k, true)
		}
		if clone.root != nil {
			t.Errorf("clone.root = %v, want %v", clone.root, nil)
		}
		if trie.root == nil || trie.Len() != 7 {
			t.Errorf("trie.root is nil")
		}
		validateSharedNode(t, trie, clone, 0)
	})
}

// test cases from golang.org/x/tools/internal/persistent
type mapEntry struct {
	key   int
	value int
}

type validatedMap struct {
	impl     *PersistentHAMT[int, int]
	expected map[int]int      // current key-value mapping.
	deleted  map[mapEntry]int // maps deleted entries to their clock time of last deletion
	seen     map[mapEntry]int // maps seen entries to their clock time of last insertion
	clock    int
}

func TestSimpleMap(t *testing.T) {

	deletedEntries := make(map[mapEntry]int)
	seenEntries := make(map[mapEntry]int)

	inp := make(map[string]int)
	for i := 0; i < 10000; i++ {
		gen := generateUUID()
		inp[gen] = i
	}

	m1 := &validatedMap{
		impl:     NewPersistentHAMT[int, int](newIntHasher()),
		expected: make(map[int]int),
		deleted:  deletedEntries,
		seen:     seenEntries,
	}

	m3 := m1.clone()
	validateRefV3(t, m1, m3)

	m3.set(t, 8, 8)
	validateRefV3(t, m1, m3)

	m3.destroy()

	assertSameMap(t, entrySet(deletedEntries), map[mapEntry]struct{}{
		{key: 8, value: 8}: {},
	})

	validateRefV3(t, m1)

	m1.set(t, 1, 1)
	validateRefV3(t, m1)
	m1.set(t, 2, 2)
	validateRefV3(t, m1)
	m1.set(t, 3, 3)
	validateRefV3(t, m1)
	m1.remove(t, 2)
	validateRefV3(t, m1)
	m1.set(t, 6, 6)
	validateRefV3(t, m1)

	assertSameMap(t, entrySet(deletedEntries), map[mapEntry]struct{}{
		{key: 2, value: 2}: {},
		{key: 8, value: 8}: {},
	})

	//m1.impl.Range(func(k int, v int) bool {
	//	return false
	//})
	m2 := m1.clone()
	validateRefV3(t, m1, m2)
	m1.set(t, 6, 60)
	validateRefV3(t, m1, m2)
	m1.remove(t, 1)
	validateRefV3(t, m1, m2)

	keyhash1 := m1.impl.hash(100, 0)
	keyhash2 := m1.impl.hash(1, 0)
	gotAllocs := int(testing.AllocsPerRun(10, func() {
		// Can not test this on m1.impl.Delete because the hash function's allocs are not deterministic.
		m1.impl.root, _ = m1.impl.delete(m1.impl.root, 100, keyhash1, 0, false, 1)
		m1.impl.root, _ = m1.impl.delete(m1.impl.root, 1, keyhash2, 0, false, 1)
	}))
	wantAllocs := 0
	if gotAllocs != wantAllocs {
		t.Errorf("wanted %d allocs, got %d", wantAllocs, gotAllocs)
	}

	for i := 10; i < 14; i++ {
		m1.set(t, i, i)
		validateRefV3(t, m1, m2)
	}

	m1.set(t, 10, 100)
	validateRefV3(t, m1, m2)

	m1.remove(t, 12)
	validateRefV3(t, m1, m2)

	m2.set(t, 4, 4)
	validateRefV3(t, m1, m2)
	m2.set(t, 5, 5)
	validateRefV3(t, m1, m2)

	m1.destroy()

	assertSameMap(t, entrySet(deletedEntries), map[mapEntry]struct{}{
		{key: 2, value: 2}:    {},
		{key: 6, value: 60}:   {},
		{key: 8, value: 8}:    {},
		{key: 10, value: 10}:  {},
		{key: 10, value: 100}: {},
		{key: 11, value: 11}:  {},
		{key: 12, value: 12}:  {},
		{key: 13, value: 13}:  {},
	})

	m2.set(t, 7, 7)
	validateRefV3(t, m2)

	m2.destroy()

	assertSameMap(t, entrySet(seenEntries), entrySet(deletedEntries))
}

func toArray(m *PersistentHAMT[int, int]) []int {
	var result []int
	m.Range(func(k int, v int) bool {
		result = append(result, k)
		return false
	})
	return result
}

func TestRandomMap(t *testing.T) {
	deletedEntries := make(map[mapEntry]int)
	seenEntries := make(map[mapEntry]int)

	m := &validatedMap{
		impl:     NewPersistentHAMT[int, int](newIntHasher()),
		expected: make(map[int]int),
		deleted:  deletedEntries,
		seen:     seenEntries,
	}

	keys := make([]int, 0, 1000)
	for i := 0; i < 1000; i++ {
		key := rand.Intn(10000)
		m.set(t, key, key)
		keys = append(keys, key)

		if i%10 == 1 {
			index := rand.Intn(len(keys))
			last := len(keys) - 1
			key = keys[index]
			keys[index], keys[last] = keys[last], keys[index]
			keys = keys[:last]
			m.remove(t, key)

		}
	}

	m.destroy()
	assertSameMap(t, entrySet(seenEntries), entrySet(deletedEntries))
}

func dumpMap(t *testing.T, prefix string, n *mapNode[int, int]) {

	if n == nil {
		t.Logf("%s nil", prefix)
		return
	}
	//t.Logf("%s {key: %v, value: %v (ref: %v), ref: %v, weight: %v}", prefix, n.key, n.value.value, n.value.refCount, n.refCount)
	for i := 0; i < len(n.contentArray)/2; i++ {
		recordIdx := width * i
		nodeIdx := width*i + 1

		if n.contentArray[recordIdx] != nil {
			//r := (*record[int, int])(n.contentArray[recordIdx])
			//t.Logf("%s %d {key : %v , value : %v}", prefix, i, r.key, r.value)
		}

		if n.contentArray[nodeIdx] != nil {
			dumpMap(t, prefix+"->", (*mapNode[int, int])(n.contentArray[nodeIdx]))
		}
	}

}

func entrySet(m map[mapEntry]int) map[mapEntry]struct{} {
	set := make(map[mapEntry]struct{})
	for k := range m {
		set[k] = struct{}{}
	}
	return set
}

//
//func validateRef(t *testing.T, maps ...*validatedMap) {
//	t.Helper()
//
//	actualCountByEntry := make(map[mapEntry]int32)
//	nodesByEntry := make(map[mapEntry]map[*mapNode[int, int]]struct{})
//	expectedCountByEntry := make(map[mapEntry]int32)
//	for i, m := range maps {
//		dfsRefV2(m.impl.root, actualCountByEntry, nodesByEntry, 1)
//		dumpMap(t, fmt.Sprintf("%d: root ->", i), m.impl.root)
//	}
//	for entry, nodes := range nodesByEntry {
//		expectedCountByEntry[entry] = int32(len(nodes))
//	}
//	assertSameMap(t, expectedCountByEntry, actualCountByEntry)
//}
//
//func dfsRefV2(node *mapNode[int, int], countByEntry map[mapEntry]int32, nodesByEntry map[mapEntry]map[*mapNode[int, int]]struct{}, trueCount int32) {
//	if node == nil {
//		return
//	}
//
//	if count := atomic.LoadInt32(&node.refCount); count > 1 {
//		trueCount += count - 1
//	}
//
//	for i := 0; i < len(node.contentArray)/2; i++ {
//		recordIdx := width * i
//		nodeIdx := width*i + 1
//
//		if node.contentArray[recordIdx] != nil {
//			r := (*record[int, int])(node.contentArray[recordIdx])
//			entry := mapEntry{key: r.key, value: r.value}
//
//			count := atomic.LoadInt32(&r.refCount)
//			countByEntry[entry] = trueCount + count - 1
//
//			nodes, ok := nodesByEntry[entry]
//			if !ok {
//				nodes = make(map[*mapNode[int, int]]struct{})
//				nodesByEntry[entry] = nodes
//			}
//			nodes[node] = struct{}{}
//			//if count := atomic.LoadInt32(&r.refCount); count > 1 {
//			//	countByEntry[entry] = trueCount + count - 1
//			//} else {
//			//	countByEntry[entry] = trueCount
//			//}
//
//		}
//		if node.contentArray[nodeIdx] != nil {
//			dfsRefV2((*mapNode[int, int])(node.contentArray[nodeIdx]), countByEntry, nodesByEntry, trueCount)
//		}
//
//	}
//
//}

func validateRefV3(t *testing.T, maps ...*validatedMap) {
	t.Helper()

	actualRefByEntry := make(map[mapEntry]map[*PersistentHAMT[int, int]]struct{})
	assumingCountByEntry := make(map[mapEntry]int32)

	expectedCountByEntry := make(map[mapEntry]int32)
	for i, m := range maps {
		//count := atomic.LoadInt32(&m.impl.root.refCount)

		dfsRefV3(m.impl, m.impl.root, assumingCountByEntry, actualRefByEntry, 1)
		dumpMap(t, fmt.Sprintf("%d: root ->", i), m.impl.root)
	}

	for entry, ref := range actualRefByEntry {
		expectedCountByEntry[entry] = int32(len(ref))
	}
	assertSameMap(t, expectedCountByEntry, assumingCountByEntry)
}

func dfsRefV3(hmap *PersistentHAMT[int, int], node *mapNode[int, int], assumingCountByEntry map[mapEntry]int32, actualRefByEntry map[mapEntry]map[*PersistentHAMT[int, int]]struct{}, trueCount int32) {
	if node == nil {
		return
	}

	if count := atomic.LoadInt32(&node.refCount); count > 1 {
		trueCount += count - 1
	}

	for i := 0; i < len(node.contentArray)/2; i++ {
		recordIdx := width * i
		nodeIdx := width*i + 1

		if node.contentArray[recordIdx] != nil {
			r := (*record[int, int])(node.contentArray[recordIdx])
			entry := mapEntry{key: r.key, value: r.value}

			count := atomic.LoadInt32(&r.refCount)
			assumingCountByEntry[entry] = trueCount + count - 1

			refs, ok := actualRefByEntry[entry]
			if !ok {
				refs = make(map[*PersistentHAMT[int, int]]struct{})
				actualRefByEntry[entry] = refs
			}
			refs[hmap] = struct{}{}

		}
		//fmt.Println("node by entry sir: ", actualRefByEntry)
		if node.contentArray[nodeIdx] != nil {
			dfsRefV3(hmap, (*mapNode[int, int])(node.contentArray[nodeIdx]), assumingCountByEntry, actualRefByEntry, trueCount)
		}

	}

}

func (vm *validatedMap) set(t *testing.T, key int, value int) {
	entry := mapEntry{key, value}

	vm.clock++
	vm.seen[entry] = vm.clock

	vm.impl.Put(key, value, func(deletedKey int, deletedValue int) {
		if deletedKey != key || deletedValue != value {
			t.Fatalf("unexpected passed in deleted entry: %v/%v, expected: %v/%v", deletedKey, deletedValue, key, value)
		}
		// Not safe if closure shared between two validatedMaps.
		vm.deleted[entry] = vm.clock
	})
	vm.expected[key] = value

	gotValue, ok := vm.impl.Get(key)
	if !ok || gotValue != value {
		t.Fatalf("unexpected get result after insertion, key: %v, expected: %v, got: %v (%v)", key, value, gotValue, ok)
	}
}

func (vm *validatedMap) validate(t *testing.T) {
	t.Helper()

	validateNode(t, vm.impl.root)

	// Note: this validation may not make sense if maps were constructed using
	// SetAll operations. If this proves to be problematic, remove the clock,
	// deleted, and seen fields.
	for key, value := range vm.expected {
		entry := mapEntry{key: key, value: value}
		if deleteAt := vm.deleted[entry]; deleteAt > vm.seen[entry] {
			t.Fatalf("entry is deleted prematurely, key: %d, value: %d", key, value)
		}
	}

	actualMap := make(map[int]int, len(vm.expected))
	vm.impl.Range(func(key, value int) bool {
		if other, ok := actualMap[key]; ok {
			t.Fatalf("key is present twice, key: %d, first value: %d, second value: %d", key, value, other)
		}
		actualMap[key] = value
		return false
	})

	assertSameMap(t, actualMap, vm.expected)
}

func validateNode(t *testing.T, node *mapNode[int, int]) {
	if node == nil {
		return
	}

	if node.contentArray == nil {
		t.Fatalf("node content array is nil")
	}

	if len(node.contentArray)%2 != 0 {
		t.Fatalf("node content array length is not even")
	}

	if node.bitmap == 0 {
		t.Fatalf("node bitmap is 0")
	}

	hasValue := false
	for i := 0; i < len(node.contentArray)/2; i++ {

		recordIdx := width * i
		nodeIdx := width*i + 1

		if node.contentArray[recordIdx] != nil || node.contentArray[nodeIdx] != nil {
			hasValue = true
		}

		if node.contentArray[nodeIdx] != nil {
			validateNode(t, (*mapNode[int, int])(node.contentArray[nodeIdx]))
		}

	}

	if !hasValue {
		t.Fatalf("node has no value")
	}

}

func (vm *validatedMap) remove(t *testing.T, key int) {
	vm.clock++
	vm.impl.Delete(key)
	delete(vm.expected, key)
	vm.validate(t)

	gotValue, ok := vm.impl.Get(key)
	if ok {
		t.Fatalf("unexpected get result after removal, key: %v, got: %v", key, gotValue)
	}
}

func (vm *validatedMap) clone() *validatedMap {
	expected := make(map[int]int, len(vm.expected))
	for key, value := range vm.expected {
		expected[key] = value
	}

	return &validatedMap{
		impl:     vm.impl.Clone(),
		expected: expected,
		deleted:  vm.deleted,
		seen:     vm.seen,
	}
}

func (vm *validatedMap) destroy() {
	vm.impl.Destroy()
}

func assertSameMap(t *testing.T, map1, map2 interface{}) {
	t.Helper()

	if !reflect.DeepEqual(map1, map2) {
		t.Fatalf("different maps:\n%v\nvs\n%v", map1, map2)
	}
}
