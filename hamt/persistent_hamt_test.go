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
	t.Run("Basic Delete", func(t *testing.T) {
		trie := NewPersistentHAMT[string, int](newCollisionHasher[string]())

		insertItem(t, trie, "a", 1, 1)
		insertItem(t, trie, "b", 2, 2)
		insertItem(t, trie, "c", 3, 3)

		insertItem(t, trie, "d", 4, 4)
		insertItem(t, trie, "rehash2time_1", 5, 5)
		insertItem(t, trie, "rehash2time_2", 6, 6)

		insertItem(t, trie, "panic1", 7, 7)

		deleteItem(t, trie, "a", 6, true)

		deleteItem(t, trie, "b", 5, true)

		deleteItem(t, trie, "c", 4, true)
		deleteItem(t, trie, "d", 3, true)

		deleteItem(t, trie, "rehash2time_1", 2, true)
		deleteItem(t, trie, "rehash2time_2", 1, true)

		deleteItem(t, trie, "panic1", 0, true)

		if trie.root != nil {
			t.Errorf("trie.root = %v, want %v", trie.root, nil)
		}

		//clone := trie.Clone()

	})
}
