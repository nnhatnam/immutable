package RRBTree

import (
	"golang.org/x/exp/slices"
	"testing"
)

func convertToArray[V any](arr []*refValue[V]) []V {
	var result []V
	for _, v := range arr {
		result = append(result, v.value)
	}
	return result
}

func TestNewRRBTree(t *testing.T) {

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

		if rrb.Len() != 33 {
			t.Errorf("Expected rrb length of 33, got %v", rrb.Len())
		}

		for i := 0; i < 33; i++ {

			if history[i].Len() != i {
				t.Errorf("Expected history length of %v, got %v", i, history[i].Len())
			}
			if rrb.Get(i) != i {
				t.Errorf("Expected find the value %v, got %v", i, rrb.Get(i))
			}
		}

		// 33 -> 65
		for i := 0; i < 32; i++ {
			history = append(history, rrb)
			rrb = rrb.Append(33 + i)
		}

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
			if i > 32 {
				return
			}
		}

	})

	t.Run("Append > buffer (extend)", func(t *testing.T) {
		rrb := RRBTree[int]{}

		history := make([]RRBTree[int], 0)

		nums := 1<<10 + 33
		for i := 0; i < nums; i++ {
			history = append(history, rrb)
			rrb = rrb.Append(i)
		}

		if rrb.Len() != nums {
			t.Errorf("Expected rrb length of 33, got %v", rrb.Len())
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

}
