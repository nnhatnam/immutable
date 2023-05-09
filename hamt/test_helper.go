package hamt

import (
	crand "crypto/rand"
	"fmt"
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
	case "rehash2time_1":
		return 0b000100
	case "rehash2time_2":
		return 0b000100
	case "panic1":
		return 0b000010
	case "panic2":
		return 0b000010
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
	case "rehash2time_1":
		if level <= 1 {
			return 0b000100
		}
		return 0b001100
	case "rehash2time_2":
		if level < 1 {
			return 0b000100
		}
		return 0b000100
	case "panic1":
		return 0b000010
	case "panic2":
		return 0b000010
	}
	panic("test cases are limited on words above, so this case should not happen")

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
