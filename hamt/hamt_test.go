package hamt

import (
	"fmt"
	"testing"

	"github.com/cespare/xxhash"
	"github.com/spaolacci/murmur3"
)

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

func TestNewMutableHamt(t *testing.T) {

	trie := New[string, int](newHasher[string]())

	trie.ReplaceOrInsert("immutable", 1)

	if trie.Len() != 1 {
		t.Errorf("trie.Len() = %d, want %d", trie.Len(), 1)
	}

	result, ok := trie.Get("immutable")

	if !ok {
		t.Errorf("trie.Get() = %d, want %d", result, 1)
	}

	// replace the value
	trie.ReplaceOrInsert("immutable", 2)

	if trie.Len() != 1 {
		t.Errorf("trie.Len() = %d, want %d", trie.Len(), 1)
	}

	result, ok = trie.Get("immutable")

	if !ok || result != 2 {
		t.Errorf("trie.Get() = %d, want %d", result, 2)
	}

}

type collisionHasher[K comparable] struct{}

func newCollisionHasher[K comparable]() *collisionHasher[K] {
	return &collisionHasher[K]{}
}

func (hs *collisionHasher[K]) Hash(key string) uint64 {
	switch key {
	case "a":

		return 0b000000
	case "b":
		// return 64 bit binary 10
		return 0b111111000000
	case "c":
		return 0b111110111111000000 // 0.63.62
	case "d":
		// collision with c
		return 0b111110111111000000 // 0.63.62
	}
	panic("This Hash is created for testing purposes and test cases are limited on words above, so this case should not happen")
}

func (hs *collisionHasher[K]) Rehash(key string, level int) uint64 {

	switch key {
	case "a":

		return 0b000001
	case "b":
		// return 64 bit binary 10
		return 0b111111000001
	case "c":
		return 0b111110111111000000
	case "d":
		return 0b111110111111000001
	}
	panic("test cases are limited on words above, so this case should not happen")

}

func TestHAMTInsertion(t *testing.T) {
	trie := New[string, int](newCollisionHasher[string]())

	trie.ReplaceOrInsert("a", 1)
	trie.ReplaceOrInsert("b", 2)

	if trie.Len() != 2 {
		t.Errorf("trie.Len() = %d, want %d", trie.Len(), 2)
	}

	aResult, ok := trie.Get("a")

	if !ok || aResult != 1 {
		t.Errorf("trie.Get() = %d, want %d", aResult, 1)
	}

	bResult, ok := trie.Get("b")

	if !ok || bResult != 2 {
		t.Errorf("trie.Get() = %d, want %d", bResult, 2)
	}

	trie.ReplaceOrInsert("c", 3)

	if trie.Len() != 3 {
		t.Errorf("trie.Len() = %d, want %d", trie.Len(), 3)
	}

	cResult, ok := trie.Get("c")

	if !ok || cResult != 3 {
		t.Errorf("trie.Get() = %d, want %d", cResult, 3)
	}

	trie.ReplaceOrInsert("d", 4)

	if trie.Len() != 4 {
		t.Errorf("trie.Len() = %d, want %d", trie.Len(), 4)
	}

	dResult, ok := trie.Get("d")

	if !ok || dResult != 4 {
		t.Errorf("trie.Get() = %d, want %d", dResult, 4)
	}

}
