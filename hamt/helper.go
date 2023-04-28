package hamt

// mask returns the mask for the given hash and shift.
// mask tell us which bit to look at in the hash
//func mask(hash uint64, shift int) int {
//	return int((hash >> shift) & ((1 << arity) - 1)) // (hash >> shift) & 0b111111
//}

// bucket returns the bucket index where the item with the given hash should be stored.
// hash is the hash of the key
// shift tells us the location of the bit to look at in the hash
func bucket(hash uint64, shift int) int {
	return int((hash >> shift) & ((1 << arity) - 1)) // (hash >> shift) & 0b111111
}

// blockIndex returns the index of the block in the content array.
//func blockIndex(bitmap uint64, mask uint64) int {
//	if mask == 0 {
//		return 0
//	}
//	return bits.OnesCount64(bitmap & (mask - 1))
//}
