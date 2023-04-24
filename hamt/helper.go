package hamt

import "math/bits"

// mask returns the mask for the given hash and shift.
func mask(hash uint64, shift int) uint64 {
	return (hash >> shift) & ((1 << arity) - 1)
}

func index(bitmap uint64, bitpos int) int {
	return bits.OnesCount64(bitmap & (uint64(bitpos) - 1))
}
