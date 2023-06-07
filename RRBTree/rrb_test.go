package RRBTree

import (
	"fmt"
	"golang.org/x/exp/slices"
	"math/rand"
	"testing"
	"time"
)

func convertToArray[V any](arr []*refValue[V]) []V {
	var result []V
	for _, v := range arr {
		result = append(result, v.value)
	}
	return result
}

func dumpObj(s string, print bool) {
	if print {
		fmt.Println(s)
	}
}

func toArr[V any](arr []*refValue[V]) []V {
	var result []V
	for _, v := range arr {
		result = append(result, v.value)
	}
	return result
}

func verifyTree[V any](t *testing.T, rrb *RRBTree[V], h int, dump bool) bool {
	t.Helper()

	var verify func(n *node[V], h int, path string, isLastBranch bool) (int, bool)

	verify = func(n *node[V], h int, path string, isLastBranch bool) (int, bool) {
		//fmt.Println("Calling path: ", path, isLastBranch, len(n.children))
		if n == nil {
			return 0, true
		}

		if h == 0 {
			dumpObj(fmt.Sprintf("Path: [%v] %v", path+" -> leaf ", toArr(n.values)), dump)

			if len(n.sizes) > 0 {
				t.Fatalf("Path: [%v] - Expected leaf's size to be empty, got %v", path+" -> leaf ", n.sizes)
				return len(n.values), false
			}

			//if !lastBranch && len(n.values) < minBranches {
			//	t.Errorf("Path: [%v] - Expected leaf's children to be at least %v for non-last branch, got %v", path+" -> leaf ", minBranches, len(n.values))
			//	return len(n.values), false
			//}

			if len(n.values) < minBranches {
				t.Fatalf("Path: [%v] - Expected leaf's children to be at least %v for non-last branch, got %v", path+" -> leaf ", minBranches, len(n.values))
				return len(n.values), false
			}

			//fmt.Println("Leaf", toArr(n.values), len(n.children))
			return len(n.values), true
		}

		dumpObj(fmt.Sprintf("Path: [%v] (treeSize : %v - sizes: %v)", path, n.treeSize, n.sizes), dump)

		cumulativeSize := 0
		for i, child := range n.children {

			trueCount, ok := verify(child, h-1, path+" -> "+fmt.Sprint(i), isLastBranch && i == len(n.children)-1)

			if !ok {
				return trueCount, false
			}

			if trueCount != child.treeSize {
				t.Fatalf("Path: [%v] - Mismatched treeSize and trueCount. Expected size of %v (treeSize), got %v (trueCount)", path+" -> "+fmt.Sprint(i), child.treeSize, trueCount)
			}

			cumulativeSize += trueCount

			cummulativeCalc := n.sizes

			if n.isBalancedNode() {
				cummulativeCalc = cumulativeSumTable[h][:len(n.children)]
				cummulativeCalc[len(n.children)-1] = n.treeSize
			}

			if cummulativeCalc[i] != cumulativeSize {
				t.Fatalf("Path: [%v] - Expected cumulative size at slot %v on level %v is  %v, got %v (%v) (treeSize: %v - isBalancedNode %v - n.sizes %v)", path, i, h, cumulativeSize, cummulativeCalc[i], cummulativeCalc, n.treeSize, n.isBalancedNode(), n.sizes)
				return trueCount, false
			}

			//if !n.isBalancedNode() {
			//	if cummulativeCalc[i] != cumulativeSize {
			//		t.Fatalf("Path: [%v] - Expected cumulative size at slot %v on level %v is  %v, got %v (%v)", path, i, h, cumulativeSize, cummulativeCalc[i], cummulativeCalc)
			//		return trueCount, false
			//	}
			//} else {
			//
			//	if !isLastBranch {
			//		cummulativeCalc = cumulativeSumTable[h][:len(n.children)]
			//		if cummulativeCalc[i] != cumulativeSize {
			//			fmt.Println("yooo: ", len(n.children)-1, i, n.sizes, len(n.children), isLastBranch)
			//			t.Fatalf("Path: [%v] - Expected cumulative size at slot %v on level %v is  %v, got %v (%v)", path, i, h, cumulativeSize, cummulativeCalc[i], cummulativeCalc)
			//			return trueCount, false
			//		}
			//	}
			//
			//}

		}
		return cumulativeSize, true
	}

	trueCount, ok := verify(rrb.root, h, "root", true)
	size := rrb.Len() - len(rrb.head) - len(rrb.tail)
	if trueCount != size {
		t.Fatalf("Expected size of %v, got %v", size, trueCount)
	}
	return ok

}

func TestSimpleRRBTree(t *testing.T) {

	t.Run("Empty RRBTree", func(t *testing.T) {
		rrb := NewRRBTree[int]()
		if rrb.Len() != 0 {
			t.Errorf("Expected length of 0, got %v", rrb.Len())
		}

		rrb2 := RRBTree[int]{}

		if rrb2.Len() != 0 {
			t.Errorf("Expected length of 0, got %v", rrb2.Len())
		}
	})

	t.Run("Simple append", func(t *testing.T) {
		rrb := RRBTree[int]{}

		rrb = rrb.Append(1)

		if rrb.Len() != 1 {
			t.Errorf("Expected length of 1, got %v", rrb.Len())
		}

		if rrb.Get(0) != 1 {
			t.Errorf("Expected value of 1, got %v", rrb.Get(0))
		}

		rrb2 := rrb.Append(2)

		if rrb2.Len() != 2 {
			t.Errorf("Expected length of 2, got %v", rrb2.Len())
		}

		if !slices.Equal(convertToArray[int](rrb.tail), []int{1}) {
			t.Errorf("Expected tail of [1], got %v", convertToArray[int](rrb.tail))
		}

		if !slices.Equal(convertToArray[int](rrb2.tail), []int{1, 2}) {
			t.Errorf("Expected tail of [1, 2], got %v", convertToArray[int](rrb2.tail))
		}

	})

	t.Run("Append > buffer", func(t *testing.T) {
		rrb := RRBTree[int]{}

		history := make([]RRBTree[int], 0)
		for i := 0; i < 33; i++ {
			history = append(history, rrb)
			rrb = rrb.Append(i)

		}
		verifyTree(t, &rrb, rrb.h, false)

		if rrb.Len() != 33 {
			t.Errorf("Expected rrb length of 33, got %v", rrb.Len())
		}

		for i := 0; i < 33; i++ {

			if history[i].Len() != i {
				t.Errorf("Expected history length of %v, got %v", i, history[i].Len())
			}
			//fmt.Println("finding ", i)
			if rrb.Get(i) != i {
				t.Errorf("Expected find the value %v, got %v", i, rrb.Get(i))
			}
		}

		// 33 -> 65
		for i := 0; i < 32; i++ {
			history = append(history, rrb)
			rrb = rrb.Append(33 + i)
		}

		verifyTree(t, &rrb, rrb.h, false)

		if rrb.Len() != 65 {
			t.Errorf("Expected rrb length of 65, got %v", rrb.Len())
		}
		for i := 0; i < 65; i++ {
			if history[i].Len() != i {
				t.Errorf("Expected history length of %v, got %v", i, history[i].Len())
			}
			if result := rrb.Get(i); result != i {
				t.Errorf("Expected find the value %v, got %v", i, result)
			}
		}

	})

	t.Run("Append > buffer (extend)", func(t *testing.T) {
		rrb := RRBTree[int]{}

		history := make([]RRBTree[int], 0)

		nums := 1 << 16
		for i := 0; i < nums; i++ {
			history = append(history, rrb)
			debugId = i

			rrb = rrb.Append(i)

			// uncomment to debug
			//verifyTree(t, &rrb, rrb.h, false)

		}

		verifyTree(t, &rrb, rrb.h, false)

		if rrb.Len() != nums {
			t.Errorf("Expected rrb length of %v, got %v", nums, rrb.Len())
		}

		for i := 0; i < nums; i++ {
			if history[i].Len() != i {
				t.Errorf("Expected history length of %v, got %v", i, history[i].Len())
			}
			if rrb.Get(i) != i {
				t.Errorf("Expected find the value %v, got %v", i, rrb.Get(i))
			}
		}
	})

	t.Run("Simple Pop", func(t *testing.T) {
		rrb := RRBTree[int]{}

		count := 1 << 10 // reduce count for debugging
		for i := 0; i < count; i++ {
			rrb = rrb.Append(i)
		}

		for i := 0; i < count; i++ {

			var result int
			var ok bool

			rrb, result, ok = rrb.Pop()

			verifyTree(t, &rrb, rrb.h, false)

			if result != count-i-1 || !ok {
				t.Errorf("Expected pop %v, got %v", count-i-1, result)
			}

			if rrb.Len() > 0 && rrb.Get(rrb.Len()-1) != count-i-2 {
				t.Errorf("Expected last value %v, got %v", count-i-2, rrb.Get(rrb.Len()-1))
			}

		}

		if rrb.Len() != 0 && rrb.root != nil {
			t.Errorf("Expected length of 0, got %v", rrb.Len())
		}

	})

	t.Run("Random Append/Pop", func(t *testing.T) {

		count := 1 << 16 // reduce count for debugging
		ratio := 2
		for j := 0; j < 10; j++ {

			// this part adjust append/pop ratio
			//if j >= 10 {
			//	ratio++
			//}

			rrb := RRBTree[int]{}
			fmt.Println("Round ", j, rrb.Len())
			rand.Seed(time.Now().UnixNano())
			freeList := make([]bool, count)
			for i := 0; i < count; i++ {
				operation := rand.Intn(ratio) // round 1:  < ratio/2 = append, ratio = pop

				if operation < ratio-1 {
					rrb = rrb.Append(i)
					verifyTree(t, &rrb, rrb.h, false)
				} else {
					var result int
					var ok bool
					rrb, result, ok = rrb.Pop()
					verifyTree(t, &rrb, rrb.h, false)
					if ok {
						freeList[result] = true // mark as popped
					}
				}
			}

			verifyTree(t, &rrb, rrb.h, false)

			for i := 0; i < rrb.Len(); i++ {
				verifyTree(t, &rrb, rrb.h, false)
				if freeList[rrb.Get(i)] {
					t.Errorf("Didn't expected value %v to be in freelist", rrb.Get(i))
				}
				//fmt.Println("result", rrb.Get(i))
			}
		}

	})

	t.Run("Simple prepend", func(t *testing.T) {
		rrb := RRBTree[int]{}

		rrb = rrb.Prepend(1)

		if rrb.Len() != 1 {
			t.Errorf("Expected length of 1, got %v", rrb.Len())
		}

		if rrb.Get(0) != 1 {
			t.Errorf("Expected value of 1, got %v", rrb.Get(0))
		}

		rrb2 := rrb.Prepend(2)

		if rrb2.Len() != 2 {
			t.Errorf("Expected length of 2, got %v", rrb2.Len())
		}

		if !slices.Equal(convertToArray[int](rrb.head), []int{1}) {
			t.Errorf("Expected head of [1], got %v", convertToArray[int](rrb.head))
		}

		if !slices.Equal(convertToArray[int](rrb2.head), []int{2, 1}) {
			t.Errorf("Expected head of [2, 1], got %v", convertToArray[int](rrb2.head))
		}

	})

	t.Run("Prepend > buffer", func(t *testing.T) {
		rrb := RRBTree[int]{}

		history := make([]RRBTree[int], 0)
		for i := 0; i < 33; i++ {
			history = append(history, rrb)
			rrb = rrb.Prepend(-i)

		}
		verifyTree(t, &rrb, rrb.h, false)

		if rrb.Len() != 33 {
			t.Errorf("Expected rrb length of 33, got %v", rrb.Len())
		}

		for i := 0; i < 33; i++ {

			if history[i].Len() != i {
				t.Errorf("Expected history length of %v, got %v", i, history[i].Len())
			}
			if rrb.Get(i) != i-32 {
				t.Errorf("Expected find the value %v, got %v", i-32, rrb.Get(i))
			}
		}

		// 33 -> 65
		for i := 0; i < 32; i++ {
			history = append(history, rrb)
			rrb = rrb.Prepend(-(33 + i))
		}

		verifyTree(t, &rrb, rrb.h, false)

		if rrb.Len() != 65 {
			t.Errorf("Expected rrb length of 65, got %v", rrb.Len())
		}
		for i := 0; i < 65; i++ {
			if history[i].Len() != i {
				t.Errorf("Expected history length of %v, got %v", i, history[i].Len())
			}
			if result := rrb.Get(i); result != i-64 {
				t.Errorf("Expected find the value %v, got %v", i-64, result)
			}

		}

	})

	t.Run("Prepend > buffer (extend)", func(t *testing.T) {
		rrb := RRBTree[int]{}

		history := make([]RRBTree[int], 0)

		nums := 1 << 18
		for i := 0; i < nums; i++ {
			history = append(history, rrb)
			rrb = rrb.Prepend(nums - i - 1)
		}

		verifyTree(t, &rrb, rrb.h, false)

		//verifyTree(t, &rrb, rrb.h, false)

		if rrb.Len() != nums {
			t.Errorf("Expected rrb length of 33, got %v", rrb.Len())
		}

		for i := 0; i < nums; i++ {
			if history[i].Len() != i {
				t.Errorf("Expected history length of %v, got %v", i, history[i].Len())
			}
			if rrb.Get(i) != i {
				t.Errorf("Expected find the value %v, got %v", nums-i-1, rrb.Get(i))
			}
		}
	})

}
