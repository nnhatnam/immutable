package RRBTree

var debugId = 0

// TODO: this file should be generated by a script.
const maxBranches = 32 // The maximum number of children of a node.
const minBranches = 30 // The minimum number of children of a node.
const mFactor = 5      // The m-factor of the tree. It's the number of bits used for indexing the tree.
const maxHeight = 6    // The maximum height of the tree. ( 7 levels from 0 -> 6). In RRBTree, height is counted bottom-up. The leafs are at height 0.

const minBranching = 32 // The number of children of a node. ( 2 ^ mFactor)
const maxBranching = 64 // The number of children of a node. ( 2 ^ mFactor)

// The shift table of the tree using for index calculation.
var shiftTable = [maxHeight + 1]int{0, 5, 10, 15, 20, 25, 30}
var maskTable = [maxHeight + 1]int{31, 31 << 5, 31 << 10, 31 << 15, 31 << 20, 31 << 25, 31 << 30}
var maskLengthTable = [maxHeight + 1]int{32, 32 * (1 << 1), 32 * (1 << 2), 32 * (1 << 3), 32 * (1 << 4), 32 * (1 << 5), 32 * (1 << 6)}

//var cumulativeSum = [maxBranches + 1]int{32, 32 * 2, 32 * 3, 32 * 4, 32 * 5, 32 * 6, 32 * 7, 32 * 8, 32 * 9, 32 * 10,
//	32 * 11, 32 * 12, 32 * 13, 32 * 14, 32 * 15, 32 * 16, 32 * 17, 32 * 18, 32 * 19, 32 * 20,
//	32 * 21, 32 * 22, 32 * 23, 32 * 24, 32 * 25, 32 * 26, 32 * 27, 32 * 28, 32 * 29, 32 * 30,
//	32 * 31, 32 * 32}

var cumulativeSumTable = [7][32]int{}

func buildCumulativeSum() {

	for i := 0; i <= maxHeight; i++ {
		//( 2 ^ (2m))
		//cumulativeSumTable[i] = (1 << i)
		for j := 0; j < 32; j++ {
			cumulativeSumTable[i][j] = (1 << (i * mFactor)) * (j + 1)
		}
	}
}

func init() {
	buildCumulativeSum()
}
