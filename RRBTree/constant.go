package RRBTree

const branchFactor = 32 // The branch factor of the tree.

const maxBranches = 32 // The maximum number of children of a node.
const minBranches = 31 // The minimum number of children of a node.
const maxHeight = 6    // The maximum height of the tree. ( 7 levels from 0 -> 6). In RRBTree, height is counted bottom-up. The leafs are at height 0.

// The shift table of the tree using for index calculation.
var shiftTable = [maxHeight + 1]int{0, 5, 10, 15, 20, 25, 30}
var maskTable = [maxHeight + 1]int{31, 31 << 5, 31 << 10, 31 << 15, 31 << 20, 31 << 25, 31 << 30}
