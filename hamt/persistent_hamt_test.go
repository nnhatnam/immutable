package hamt

import (
	"testing"
)

func dfs[K comparable, V any](t *testing.T, n *mapNode[K, V]) []*mapNode[K, V] {
	t.Helper()
	return dfsRef(t, n)
}

func dfsRef[K comparable, V any](t *testing.T, n *mapNode[K, V]) []*mapNode[K, V] {
	t.Helper()
	if n == nil {
		return nil
	}

	var nodes []*mapNode[K, V]

	nodes = append(nodes, n)

	for i := 0; i < len(n.contentArray)/2; i++ {
		nodeIdx := width*i + 1
		if n.contentArray[nodeIdx] != nil {
			nodes = append(nodes, dfsRef(t, (*mapNode[K, V])(n.contentArray[nodeIdx]))...)
		}
	}

	return nodes
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

func TestPersistentHAMTBasicInsertion(t *testing.T) {

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

}
