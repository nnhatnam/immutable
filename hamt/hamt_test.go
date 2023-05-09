package hamt

import (
	"testing"
)

// for testing only
func (m *HAMT[K, V]) mReplaceOrInsert(k K, v V) (_ V, _ bool) {
	keyHash := m.hash(k, 0)
	if m.root == nil {
		m.root = newNode[K, V]()
	}

	ret, ok := m.mInsertRecord(m.root, keyHash, 0, newRecord[K, V](k, v))
	if !ok {
		m.len++
	}

	return ret, ok

}

func replaceOrInsertItem(t *testing.T, m *HAMT[string, int], key string, value int, expectedLen int) *HAMT[string, int] {
	trie := m.Set(key, value)

	if trie.Len() != expectedLen {
		t.Errorf("trie.Len() = %d, want %d", trie.Len(), expectedLen)
	}

	result, ok := trie.Get(key)

	if !ok || result != value {
		t.Errorf("trie.Get() = %d, want %d", result, value)
	}

	return trie
}

func TestNewMutableHamt(t *testing.T) {

	trie := New[string, int](newHasher[string]())

	trie1 := replaceOrInsertItem(t, trie, "immutable", 1, 1)

	trie2 := replaceOrInsertItem(t, trie1, "immutable", 2, 1)

	if trie.Len() != 0 || trie1.Len() != 1 || trie2.Len() != 1 {
		t.Errorf("trie.Len() = %d, want %d", trie.Len(), 0)
	}

	result, ok := trie.Get("immutable")

	if ok {
		t.Errorf("Should not find any item. trie.Get() = %d, want %d", result, 0)
	}

	result, ok = trie1.Get("immutable")

	if !ok || result != 1 {
		t.Errorf("trie.Get() = %d, want %d", result, 1)
	}

	result, ok = trie2.Get("immutable")

	if !ok || result != 2 {
		t.Errorf("trie.Get() = %d, want %d", result, 2)
	}

}

// func TestNewMutableHamtV1(t *testing.T) {
//
//	trie := New[string, int](newHasher[string]())
//
//	trie.mReplaceOrInsert("immutable", 1)
//
//	if trie.Len() != 1 {
//		t.Errorf("trie.Len() = %d, want %d", trie.Len(), 1)
//	}
//
//	result, ok := trie.Get("immutable")
//
//	if !ok {
//		t.Errorf("trie.Get() = %d, want %d", result, 1)
//	}
//
//	// replace the value
//	trie.mReplaceOrInsert("immutable", 2)
//
//	if trie.Len() != 1 {
//		t.Errorf("trie.Len() = %d, want %d", trie.Len(), 1)
//	}
//
//	result, ok = trie.Get("immutable")
//
//	if !ok || result != 2 {
//		t.Errorf("trie.Get() = %d, want %d", result, 2)
//	}
//
// }

func insertItem(t *testing.T, trie *HAMT[string, int], key string, value int, expectedLen int) {
	trie.mReplaceOrInsert(key, value)

	if trie.Len() != expectedLen {
		t.Errorf("trie.Len() = %d, want %d", trie.Len(), expectedLen)
	}

	result, ok := trie.Get(key)

	if !ok || result != value {
		t.Errorf("trie.Get() = %d, want %d", result, value)
	}
}

//func TestHAMTBasicInsertion(t *testing.T) {
//	trie := New[string, int](newCollisionHasher[string]())
//
//	insertItem(t, trie, "a", 1, 1)
//	insertItem(t, trie, "b", 2, 2)
//	insertItem(t, trie, "c", 3, 3)
//	insertItem(t, trie, "d", 4, 4)
//	insertItem(t, trie, "rehash2time_1", 5, 5)
//	insertItem(t, trie, "rehash2time_2", 6, 6)
//	insertItem(t, trie, "panic1", 7, 7)
//
//	if !panics(func() { trie.mReplaceOrInsert("panic2", 8) }) {
//		t.Errorf("trie.ReplaceOrInsert() should panic")
//	}
//
//}
//
//func TestHAMTBulkInsertion(t *testing.T) {
//	trie := New[string, int](newHasher[string]())
//
//	var min, max string
//	inp := make(map[string]int)
//	for i := 0; i < 10000; i++ {
//		gen := generateUUID()
//		inp[gen] = i
//		if gen < min || i == 0 {
//			min = gen
//		}
//		if gen > max || i == 0 {
//			max = gen
//		}
//	}
//
//	for k, v := range inp {
//		trie.mReplaceOrInsert(k, v)
//	}
//
//	if trie.Len() != len(inp) {
//		t.Errorf("Got %v expected %v", trie.Len(), len(inp))
//	}
//
//	for k, v := range inp {
//		out, ok := trie.Get(k)
//
//		if !ok {
//			t.Fatalf("missing key: %v", k)
//		}
//		if out != v {
//			t.Fatalf("value mis-match: %v %v", out, v)
//		}
//	}
//
//}
