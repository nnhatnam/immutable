package hamt

import (
	"golang.org/x/exp/slices"
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

//func TestNewPersistentHAMT(t *testing.T) {
//
//	trie := NewPersistentHAMT[string, int](newHasher[string]())
//
//	trie.Set("immutable", 1)
//
//	if trie.Len() != 1 {
//		t.Errorf("trie.Len() = %d, want %d", trie.Len(), 1)
//	}
//
//	result, ok := trie.Get("immutable")
//
//	if !ok || result != 1 {
//		t.Errorf("trie.Get() = %d, want %d", result, 1)
//	}
//
//}

func insertItemPersistent(t *testing.T, trie *PersistentHAMT[string, int], key string, value int, expectedLen int) {
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

		insertItemPersistent(t, trie, "a", 1, 1)
		insertItemPersistent(t, trie, "b", 2, 2)
		insertItemPersistent(t, trie, "c", 3, 3)

		insertItemPersistent(t, trie, "d", 4, 4)
		insertItemPersistent(t, trie, "rehash2time_1", 5, 5)
		insertItemPersistent(t, trie, "rehash2time_2", 6, 6)

		insertItemPersistent(t, trie, "panic1", 7, 7)

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
			insertItemPersistent(t, trie, gen, i, i+1)
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
	trie := NewPersistentHAMT[string, int](newCollisionHasher[string]())

	insertItemPersistent(t, trie, "a", 1, 1)
	insertItemPersistent(t, trie, "b", 2, 2)
	insertItemPersistent(t, trie, "c", 3, 3)

	insertItemPersistent(t, trie, "d", 4, 4)
	insertItemPersistent(t, trie, "rehash2time_1", 5, 5)
	insertItemPersistent(t, trie, "rehash2time_2", 6, 6)

	insertItemPersistent(t, trie, "panic1", 7, 7)

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

	insertItemPersistent(t, clone, "a", 10, trie.Len())
	sharedNodes = validateSharedNode(t, trie, clone, 2)

	result, ok := trie.Get("a")
	if !ok || result != 1 {
		t.Errorf("trie.Get() = %d, want %d", result, 1)
	}

	result, ok = clone.Get("a")
	if !ok || result != 10 {
		t.Errorf("clone.Get() = %d, want %d", result, 10)
	}

	insertItemPersistent(t, clone, "b", 20, trie.Len())
	insertItemPersistent(t, clone, "c", 30, trie.Len())
	insertItemPersistent(t, clone, "d", 40, trie.Len())
	insertItemPersistent(t, clone, "rehash2time_1", 50, trie.Len())
	insertItemPersistent(t, clone, "rehash2time_2", 60, trie.Len())
	insertItemPersistent(t, clone, "panic1", 70, trie.Len())

	sharedNodes = validateSharedNode(t, trie, clone, 0)

	if len(sharedNodes) != 0 {
		t.Errorf("sharedNodes count = %d, want %d", len(sharedNodes), 0)
	}

}
