package hamt

import (
	crand "crypto/rand"
	"fmt"
	"testing"

	"github.com/cespare/xxhash"
	"github.com/spaolacci/murmur3"
)

// borrow test cases from github.com/armon/go-radix
func generateUUID() string {
	buf := make([]byte, 16)
	if _, err := crand.Read(buf); err != nil {
		panic(fmt.Errorf("failed to read random bytes: %v", err))
	}

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%12x",
		buf[0:4],
		buf[4:6],
		buf[6:8],
		buf[8:10],
		buf[10:16])
}

func panics(f func()) (b bool) {
	defer func() {
		if x := recover(); x != nil {
			b = true
		}
	}()
	f()
	return false
}

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

// https://softwareengineering.stackexchange.com/questions/49550/which-hashing-algorithm-is-best-for-uniqueness-and-speed/145633#145633
type mySimpleHasher[K comparable] struct{}

func newHasher[K comparable]() *mySimpleHasher[K] {
	return &mySimpleHasher[K]{}
}

func (hs *mySimpleHasher[K]) Hash(key string) uint64 {
	return murmur3.Sum64([]byte(key))
}

func (hs *mySimpleHasher[K]) Rehash(key string, level int) uint64 {

	key = key + fmt.Sprint(level)

	return xxhash.Sum64String(key)
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

//
//func TestNewMutableHamtV1(t *testing.T) {
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
//}
//
//type collisionHasher[K comparable] struct{}
//
//func newCollisionHasher[K comparable]() *collisionHasher[K] {
//	return &collisionHasher[K]{}
//}
//
//func (hs *collisionHasher[K]) Hash(key string) uint64 {
//	switch key {
//	case "a":
//
//		return 0b000000
//	case "b":
//		// return 64 bit binary 10
//		return 0b111111000000
//	case "c":
//		return 0b111110111111000000 // 0.63.62
//	case "d":
//		// collision with c
//		return 0b111110111111000000 // 0.63.62
//	case "rehash2time_1":
//		return 0b000100
//	case "rehash2time_2":
//		return 0b000100
//	case "panic1":
//		return 0b000010
//	case "panic2":
//		return 0b000010
//	}
//	panic("This Hash is created for testing purposes and test cases are limited on words above, so this case should not happen")
//}
//
//func (hs *collisionHasher[K]) Rehash(key string, level int) uint64 {
//
//	switch key {
//	case "a":
//
//		return 0b000001
//	case "b":
//		// return 64 bit binary 10
//		return 0b111111000001
//	case "c":
//		return 0b111110111111000000
//	case "d":
//		return 0b111110111111000001
//	case "rehash2time_1":
//		if level <= 1 {
//			return 0b000100
//		}
//		return 0b001100
//	case "rehash2time_2":
//		if level < 1 {
//			return 0b000100
//		}
//		return 0b000100
//	case "panic1":
//		return 0b000010
//	case "panic2":
//		return 0b000010
//	}
//	panic("test cases are limited on words above, so this case should not happen")
//
//}
//
//func insertItem(t *testing.T, trie *HAMT[string, int], key string, value int, expectedLen int) {
//	trie.mReplaceOrInsert(key, value)
//
//	if trie.Len() != expectedLen {
//		t.Errorf("trie.Len() = %d, want %d", trie.Len(), expectedLen)
//	}
//
//	result, ok := trie.Get(key)
//
//	if !ok || result != value {
//		t.Errorf("trie.Get() = %d, want %d", result, value)
//	}
//}
//
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
