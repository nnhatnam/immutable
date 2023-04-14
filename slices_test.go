package immutable

import (
	"golang.org/x/exp/slices"
	"math"
	"strconv"
	"testing"
)

func panics(f func()) (b bool) {
	defer func() {
		if x := recover(); x != nil {
			b = true
		}
	}()
	f()
	return false
}

var setTests = []struct {
	s    []int
	i    int
	v    int
	want []int
}{
	{[]int{1, 2, 3}, 0, 4, []int{4, 2, 3}},
	{[]int{1, 2, 3}, 1, 4, []int{1, 4, 3}},
	{[]int{1, 2, 3}, 2, 4, []int{1, 2, 4}},

	{[]int{1}, 0, 4, []int{4}},
}

func TestSet(t *testing.T) {

	for _, test := range setTests {
		got := Set(test.s, test.i, test.v)
		if !slices.Equal(got, test.want) {
			t.Errorf("Set(%v, %d, %d) = %v, want %v", test.s, test.i, test.v, got, test.want)
		}
	}

	//panics tests
	for _, test := range []struct {
		name string
		s    []int
		i    int
		v    int
	}{
		{"with negative index", []int{42}, -1, 10},
		{"with out-of-bounds index", []int{42}, 1, 10},
		{"with nil slice", nil, 0, 10},
	} {
		if !panics(func() { Set(test.s, test.i, test.v) }) {
			t.Errorf("Delete %s: got no panic, want panic", test.name)
		}
	}

	// Test number of allocations
	for _, test := range setTests {

		allocs := testing.AllocsPerRun(100, func() {
			_ = Set(test.s, test.i, test.v)
		})
		if len(test.s) > 0 && allocs != 1 {
			t.Errorf("Set(%v) allocates %v times, want 1", test.s, allocs)
		} else if len(test.s) == 0 && allocs != 0 {
			t.Errorf("Set(%v) allocates %v times, want 0", test.s, allocs)
		}

	}

}

var copyTests = []struct {
	s    []int
	i    int
	v    []int
	want []int
}{
	{[]int{1, 2, 3}, 0, []int{4, 5, 6}, []int{4, 5, 6}},
	{[]int{1, 2, 3}, 1, []int{4, 5, 6}, []int{1, 4, 5}},
	{[]int{1, 2, 3}, 2, []int{4, 5, 6}, []int{1, 2, 4}},
	{[]int{1, 2, 3}, 3, []int{4, 5, 6}, []int{1, 2, 3}},
	{[]int{1, 2, 3}, 0, []int{}, []int{1, 2, 3}},
	{[]int{1, 2, 3}, 1, []int{}, []int{1, 2, 3}},
	{[]int{1, 2, 3}, 2, []int{}, []int{1, 2, 3}},
	{[]int{1, 2, 3}, 3, []int{}, []int{1, 2, 3}},
	{[]int{1, 2, 3}, 0, []int{4}, []int{4, 2, 3}},
	{[]int{1, 2, 3}, 1, []int{4}, []int{1, 4, 3}},
	{[]int{1, 2, 3}, 2, []int{4}, []int{1, 2, 4}},
	{[]int{1, 2, 3}, 3, []int{4}, []int{1, 2, 3}},
	{[]int{1, 2, 3}, 0, []int{4, 5}, []int{4, 5, 3}},
	{[]int{1, 2, 3}, 1, []int{4, 5}, []int{1, 4, 5}},
	{[]int{1, 2, 3}, 2, []int{4, 5}, []int{1, 2, 4}},
	{[]int{1}, 0, []int{4, 5, 6}, []int{4}},
}

func TestCopy(t *testing.T) {

	for _, test := range copyTests {
		got := Copy(test.s, test.v, test.i)
		if !slices.Equal(got, test.want) {
			t.Errorf("Copy(%v, %v, %v) = %v, want %v", test.s, test.i, test.v, got, test.want)
		}
	}

	//panics tests
	for _, test := range []struct {
		name string
		s    []int
		i    int
		v    []int
	}{
		{"with negative index", []int{42}, -1, []int{10}},
		{"with out-of-bounds index", []int{42}, 2, []int{10}},
	} {

		if !panics(func() { Copy(test.s, test.v, test.i) }) {
			t.Errorf("Override %s: got no panic, want panic", test.name)
		}
	}

	// Test number of allocations
	for _, test := range copyTests {

		allocs := testing.AllocsPerRun(100, func() {
			_ = Copy(test.s, test.v, test.i)
		})
		if len(test.s) > 0 && allocs != 1 {
			t.Errorf("Copy(%v) allocates %v times, want 1", test.s, allocs)
		} else if len(test.s) == 0 && allocs != 0 {
			t.Errorf("Copy(%v) allocates %v times, want 0", test.s, allocs)
		}

	}
}

var updateTests = []struct {
	s    []int
	i    int
	f    func(int) int
	want []int
}{
	{[]int{1, 2, 3}, 0, func(i int) int { return i + 1 }, []int{2, 2, 3}},
	{[]int{1, 2, 3}, 1, func(i int) int { return i + 1 }, []int{1, 3, 3}},
	{[]int{1, 2, 3}, 2, func(i int) int { return i + 1 }, []int{1, 2, 4}},
	{[]int{1, 2, 3}, 0, nil, []int{1, 2, 3}},
	{[]int{}, 1, nil, []int{}},
	{nil, 1, nil, []int{}},
}

func TestUpdate(t *testing.T) {

	for _, test := range updateTests {
		got := Update(test.s, test.i, test.f)
		if !slices.Equal(got, test.want) {
			t.Errorf("Update(%v, %v) = %v, want %v", test.s, test.i, got, test.want)
		}
	}

	//panics tests
	for _, test := range []struct {
		name string
		s    []int
		i    int
		f    func(int) int
	}{
		{"with negative index", []int{42}, -1, func(i int) int { return i + 1 }},
		{"with out-of-bounds index", []int{1, 2, 3}, 3, func(i int) int { return i + 1 }},
	} {
		if !panics(func() { Update(test.s, test.i, test.f) }) {
			t.Errorf("Update %s: got no panic, want panic", test.name)
		}
	}

	// Test number of allocations
	for _, test := range updateTests {

		allocs := testing.AllocsPerRun(100, func() {
			_ = Update(test.s, test.i, test.f)
		})
		if len(test.s) > 0 && allocs != 1 {
			t.Errorf("Update(%v) allocates %v times, want 1", test.s, allocs)
		} else if len(test.s) == 0 && allocs != 0 {
			t.Errorf("Update(%v) allocates %v times, want 0", test.s, allocs)
		}
	}

}

var insertTests = []struct {
	s    []int
	i    int
	add  []int
	want []int
}{
	{
		[]int{1, 2, 3},
		0,
		[]int{4},
		[]int{4, 1, 2, 3},
	},
	{
		[]int{1, 2, 3},
		1,
		[]int{4},
		[]int{1, 4, 2, 3},
	},
	{
		[]int{1, 2, 3},
		3,
		[]int{4},
		[]int{1, 2, 3, 4},
	},
	{
		[]int{1, 2, 3},
		2,
		[]int{4, 5},
		[]int{1, 2, 4, 5, 3},
	},
}

func TestInsert(t *testing.T) {
	s := []int{1, 2, 3}
	if got := Insert(s, 0); !slices.Equal(got, s) {
		t.Errorf("Insert(%v, 0) = %v, want %v", s, got, s)
	}

	for _, test := range insertTests {
		original := slices.Clone(test.s)
		clone := slices.Clone(test.s)
		if got := Insert(clone, test.i, test.add...); !slices.Equal(got, test.want) || !slices.Equal(test.s, original) {
			t.Errorf("Insert(%v, %d, %v...) = %v, want %v", test.s, test.i, test.add, got, test.want)
		}
	}

	//panics tests
	for _, test := range []struct {
		name string
		s    []int
		i    int
		add  []int
	}{
		{"with negative index", []int{42}, -1, []int{10}},
		{"with out-of-bounds index", []int{42}, 2, []int{10}},
	} {
		if !panics(func() { Insert(test.s, test.i, test.add...) }) {
			t.Errorf("Insert %s: got no panic, want panic", test.name)
		}
	}

	// Test number of allocations
	for _, test := range insertTests {

		allocs := testing.AllocsPerRun(100, func() {
			_ = Insert(test.s, test.i, test.add...)
		})
		if len(test.s) > 0 && allocs != 1 {
			t.Errorf("Insert(%v) allocates %v times, want 1", test.s, allocs)
		} else if len(test.s) == 0 && allocs != 0 {
			t.Errorf("Insert(%v) allocates %v times, want 0", test.s, allocs)
		}

	}
}

var removeTests = []struct {
	s    []int
	v    int
	want []int
}{
	{
		[]int{1, 2, 3},
		1,
		[]int{2, 3},
	},
	{
		[]int{1, 2, 3},
		2,
		[]int{1, 3},
	},
	{
		[]int{1, 2, 3},
		3,
		[]int{1, 2},
	},
	{
		[]int{1, 2, 3},
		4,
		[]int{1, 2, 3},
	},
	{
		[]int{},
		1,
		[]int{},
	},
	{
		nil,
		1,
		[]int{},
	},
}

func TestRemove(t *testing.T) {

	for _, test := range removeTests {
		original := slices.Clone(test.s)
		if got := Remove(test.s, test.v); !slices.Equal(got, test.want) || !slices.Equal(test.s, original) {
			t.Errorf("Remove(%v, %v) = %v, want %v", test.s, test.v, got, test.want)
		}
	}

	// Test number of allocations
	for _, test := range removeTests {

		allocs := testing.AllocsPerRun(100, func() {
			_ = Remove(test.s, test.v)
		})
		if allocs > 1 {
			t.Errorf("Remove(%v) allocates %v times, want <= 1", test.s, allocs)
		}
	}

}

var removeAtTests = []struct {
	s    []int
	i    int
	want []int
}{
	{
		[]int{1, 2, 3},
		0,
		[]int{2, 3},
	},
	{
		[]int{1, 2, 3},
		1,
		[]int{1, 3},
	},
	{
		[]int{1, 2, 3},
		2,
		[]int{1, 2},
	},
}

func TestRemoveAt(t *testing.T) {

	for _, test := range removeAtTests {
		original := slices.Clone(test.s)
		if got := RemoveAt(test.s, test.i); !slices.Equal(got, test.want) || !slices.Equal(test.s, original) {
			t.Errorf("Remove(%v, %d) = %v, want %v", test.s, test.i, got, test.want)
		}
	}

	//panics tests
	for _, test := range []struct {
		name string
		s    []int
		i    int
	}{
		{"with negative index", []int{42}, -1},
		{"with out-of-bounds index", []int{42}, 2},
	} {
		if !panics(func() { RemoveAt(test.s, test.i) }) {
			t.Errorf("Remove %s: got no panic, want panic", test.name)
		}
	}

	// Test number of allocations
	for _, test := range removeAtTests {

		allocs := testing.AllocsPerRun(100, func() {
			_ = RemoveAt(test.s, test.i)
		})
		if allocs > 1 {
			t.Errorf("RemoveAt(%v) allocates %v times, want <= 1", test.s, allocs)
		}
	}

}

var removeRangeTests = []struct {
	s    []int
	i    int
	j    int
	want []int
}{
	{
		[]int{1, 2, 3, 4, 5, 6},
		0,
		2,
		[]int{3, 4, 5, 6},
	},
	{
		[]int{1, 2, 3, 4, 5, 6},
		1,
		3,
		[]int{1, 4, 5, 6},
	},
	{
		[]int{1, 2, 3, 4, 5, 6},
		0,
		6,
		[]int{},
	},
	{
		[]int{1, 2, 3, 4, 5, 6},
		0,
		0,
		[]int{1, 2, 3, 4, 5, 6},
	},
	{
		[]int{1, 2, 3, 4, 5, 6},
		5,
		6,
		[]int{1, 2, 3, 4, 5},
	},
	{
		[]int{1, 2, 3, 4, 5, 6},
		5,
		5,
		[]int{1, 2, 3, 4, 5, 6},
	},
}

func TestRemoveRange(t *testing.T) {

	for _, test := range removeRangeTests {
		original := slices.Clone(test.s)
		if got := RemoveRange(test.s, test.i, test.j); !slices.Equal(got, test.want) || !slices.Equal(test.s, original) {
			t.Errorf("RemoveRange(%v, %d, %d) = %v, want %v", test.s, test.i, test.j, got, test.want)
		}
	}

	//panics tests
	for _, test := range []struct {
		name string
		s    []int
		i, j int
	}{
		{"with negative first index", []int{42}, -2, 1},
		{"with negative second index", []int{42}, 1, -1},
		{"with out-of-bounds first index", []int{42}, 2, 3},
		{"with out-of-bounds second index", []int{42}, 0, 2},
		{"with invalid i>j", []int{42}, 1, 0},
	} {
		if !panics(func() { RemoveRange(test.s, test.i, test.j) }) {
			t.Errorf("RemoveRange %s: got no panic, want panic", test.name)
		}
	}

	// Test number of allocations
	for _, test := range removeRangeTests {

		allocs := testing.AllocsPerRun(100, func() {
			_ = RemoveRange(test.s, test.i, test.j)
		})
		if allocs > 1 {
			t.Errorf("RemoveRange(%v) allocates %v times, want <= 1", test.s, allocs)
		}

	}

}

var pushTests = []struct {
	s    []int
	add  []int
	want []int
}{
	{
		[]int{1, 2, 3},
		[]int{4},
		[]int{1, 2, 3, 4},
	},
	{
		[]int{1, 2, 3},
		[]int{4, 5},
		[]int{1, 2, 3, 4, 5},
	},
	{
		[]int{1, 2, 3},
		[]int{},
		[]int{1, 2, 3},
	},
	{
		[]int{},
		[]int{1, 2, 3},
		[]int{1, 2, 3},
	},
	{
		[]int{},
		[]int{},
		[]int{},
	},
}

func TestPush(t *testing.T) {

	for _, test := range pushTests {
		original := slices.Clone(test.s)
		if got := Push(test.s, test.add...); !slices.Equal(got, test.want) || !slices.Equal(test.s, original) {
			t.Errorf("Append(%v, %v...) = %v, want %v", test.s, test.add, got, test.want)
		}
	}

	// Test number of allocations
	for _, test := range pushTests {

		allocs := testing.AllocsPerRun(100, func() {
			_ = Push(test.s, test.add...)
		})
		if allocs > 1 {
			t.Errorf("Push(%v) allocates %v times, want <= 1", test.s, allocs)
		}

	}

}

var pushFrontTests = []struct {
	s    []int
	add  []int
	want []int
}{
	{
		[]int{1, 2, 3},
		[]int{4},
		[]int{4, 1, 2, 3},
	},

	{
		[]int{1, 2, 3},
		[]int{4, 5},
		[]int{4, 5, 1, 2, 3},
	},
	{
		[]int{1, 2, 3},
		[]int{},
		[]int{1, 2, 3},
	},
	{
		[]int{},
		[]int{1, 2, 3},
		[]int{1, 2, 3},
	},
	{
		[]int{},
		[]int{},
		[]int{},
	},
}

func TestPushFront(t *testing.T) {

	for _, test := range pushFrontTests {
		original := slices.Clone(test.s)
		if got := PushFront(test.s, test.add...); !slices.Equal(got, test.want) || !slices.Equal(test.s, original) {
			t.Errorf("PushFront(%v, %v...) = %v, want %v", test.s, test.add, got, test.want)
		}
	}

	// Test number of allocations
	for _, test := range pushFrontTests {

		allocs := testing.AllocsPerRun(100, func() {
			_ = PushFront(test.s, test.add...)
		})
		if allocs > 1 {
			t.Errorf("PushFront(%v) allocates %v times, want <= 1", test.s, allocs)
		}

	}

}

var popTests = []struct {
	s    []int
	want []int
}{
	{
		[]int{1, 2, 3},
		[]int{1, 2},
	},
	{
		[]int{1, 2},
		[]int{1},
	},
	{
		[]int{1},
		[]int{},
	},
	{
		[]int{},
		[]int{},
	},
}

func TestPop(t *testing.T) {

	for _, test := range popTests {
		original := slices.Clone(test.s)
		last := 0
		if len(test.s) > 0 {
			last = test.s[len(test.s)-1]
		}

		if got, v := Pop(test.s); !slices.Equal(got, test.want) || !slices.Equal(test.s, original) || v != last {
			t.Errorf("Pop(%v) = %v, want %v", test.s, got, test.want)
		}
	}

	// Test number of allocations
	for _, test := range popTests {

		allocs := testing.AllocsPerRun(100, func() {
			_, _ = Pop(test.s)
		})
		if allocs > 1 {
			t.Errorf("Pop(%v) allocates %v times, want <= 1", test.s, allocs)
		}
	}

}

var popFrontTests = []struct {
	s    []int
	want []int
}{
	{
		[]int{1, 2, 3},
		[]int{2, 3},
	},
	{
		[]int{1, 2},
		[]int{2},
	},
	{
		[]int{1},
		[]int{},
	},
	{
		[]int{},
		[]int{},
	},
}

func TestPopFront(t *testing.T) {

	for _, test := range popFrontTests {
		original := slices.Clone(test.s)
		first := 0
		if len(test.s) > 0 {
			first = test.s[0]
		}

		if got, v := PopFront(test.s); !slices.Equal(got, test.want) || !slices.Equal(test.s, original) || v != first {
			t.Errorf("PopFront(%v) = %v, want %v", test.s, got, test.want)
		}
	}

	// Test number of allocations
	for _, test := range popFrontTests {

		allocs := testing.AllocsPerRun(100, func() {
			_, _ = PopFront(test.s)
		})
		if allocs > 1 {
			t.Errorf("PopFront(%v) allocates %v times, want <= 1", test.s, allocs)
		}
	}

}

var shilftLeftTests = []struct {
	s    []int
	j    int //shift by j
	want []int
}{
	{
		[]int{1, 2, 3},
		1,
		[]int{2, 3, 0},
	},
	{
		[]int{2, 3, 0},
		1,
		[]int{3, 0, 0},
	},
	{
		[]int{2, 0, 0},
		1,
		[]int{0, 0, 0},
	},
	{
		[]int{},
		1,
		[]int{},
	},
	{
		[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		5,
		[]int{6, 7, 8, 9, 10, 0, 0, 0, 0, 0},
	},
	{
		[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		10,
		[]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	},
	{
		[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		20,
		[]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	},
	{
		[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		0,
		[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	},
}

func TestShiftLeft(t *testing.T) {

	for _, test := range shilftLeftTests {
		original := slices.Clone(test.s)
		if got := ShiftLeft(test.s, test.j); !slices.Equal(got, test.want) || !slices.Equal(test.s, original) {
			t.Errorf("ShiftLeft(%v, %v) = %v, want %v", test.s, test.j, got, test.want)
		}
	}

	//panics tests
	for _, test := range []struct {
		name string
		s    []int
		j    int
	}{
		{"with negative index", []int{1, 2, 3, 4, 5}, -1},
	} {
		if !panics(func() { ShiftLeft(test.s, test.j) }) {
			t.Errorf("ShiftLeft %s: got no panic, want panic", test.name)
		}
	}

	// Test number of allocations
	for _, test := range shilftLeftTests {

		allocs := testing.AllocsPerRun(100, func() {
			_ = ShiftLeft(test.s, test.j)
		})
		if allocs > 1 {
			t.Errorf("ShiftLeft(%v, %v) allocates %v times, want <= 1", test.s, test.j, allocs)
		}
	}

}

var shilftRightTests = []struct {
	s    []int
	j    int //shift by j
	want []int
}{
	{
		[]int{1, 2, 3},
		1,
		[]int{0, 1, 2},
	},
	{
		[]int{0, 1, 2},
		1,
		[]int{0, 0, 1},
	},
	{
		[]int{0, 0, 1},
		1,
		[]int{0, 0, 0},
	},
	{
		[]int{},
		1,
		[]int{},
	},
	{
		[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		5,

		[]int{0, 0, 0, 0, 0, 1, 2, 3, 4, 5},
	},
	{
		[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		10,
		[]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	},
	{
		[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		0,
		[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	},
	{
		[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		20,
		[]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	},
}

func TestShiftRight(t *testing.T) {

	for _, test := range shilftRightTests {
		original := slices.Clone(test.s)
		if got := ShiftRight(test.s, test.j); !slices.Equal(got, test.want) || !slices.Equal(test.s, original) {
			t.Errorf("ShiftRight(%v, %v) = %v, want %v", test.s, test.j, got, test.want)
		}
	}

	//panics tests
	for _, test := range []struct {
		name string
		s    []int
		j    int
	}{
		{"with negative index", []int{1, 2, 3, 4, 5}, -1},
	} {
		if !panics(func() { ShiftRight(test.s, test.j) }) {
			t.Errorf("ShiftRight %s: got no panic, want panic", test.name)
		}
	}

	// Test number of allocations
	for _, test := range shilftRightTests {

		allocs := testing.AllocsPerRun(100, func() {
			_ = ShiftRight(test.s, test.j)
		})
		if allocs > 1 {
			t.Errorf("ShiftRight(%v, %v) allocates %v times, want <= 1", test.s, test.j, allocs)
		}
	}

}

var rotateLeftTests = []struct {
	s    []int
	j    int //rotate left by j
	want []int
}{
	{
		[]int{1, 2, 3},
		1,
		[]int{2, 3, 1},
	},
	{
		[]int{1, 2, 3},
		2,
		[]int{3, 1, 2},
	},
	{
		[]int{1, 2, 3},
		3,
		[]int{1, 2, 3},
	},
	{
		[]int{1, 2, 3},
		4,
		[]int{2, 3, 1},
	},
	{
		[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		5,
		[]int{6, 7, 8, 9, 10, 1, 2, 3, 4, 5},
	},
	{
		[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		10,
		[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	},
	{
		[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		20,
		[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	},
	{
		[]int{},
		1,
		[]int{},
	},
}

func TestRotateLeft(t *testing.T) {

	for _, test := range rotateLeftTests {
		original := slices.Clone(test.s)
		if got := RotateLeft(test.s, test.j); !slices.Equal(got, test.want) || !slices.Equal(test.s, original) || cap(got) > len(got) {
			t.Errorf("RotateLeft(%v, %v) = %v (cap %v - len %v), want %v", test.s, test.j, got, cap(got), len(got), test.want)
		}
	}

	//panics tests
	for _, test := range []struct {
		name string
		s    []int
		j    int
	}{
		{"with negative index", []int{1, 2, 3, 4, 5}, -1},
	} {
		if !panics(func() { RotateLeft(test.s, test.j) }) {
			t.Errorf("RotateLeft %s: got no panic, want panic", test.name)
		}
	}

	// Test number of allocations
	for _, test := range rotateLeftTests {

		allocs := testing.AllocsPerRun(100, func() {
			_ = RotateLeft(test.s, test.j)
		})
		if allocs > 1 {
			t.Errorf("RotateLeft(%v, %v) allocates %v times, want <= 1", test.s, test.j, allocs)
		}
	}
}

var rotateRightTests = []struct {
	s    []int
	j    int //rotate right by j
	want []int
}{
	{
		[]int{1, 2, 3},
		1,
		[]int{3, 1, 2},
	},
	{
		[]int{1, 2, 3},
		2,
		[]int{2, 3, 1},
	},
	{
		[]int{1, 2, 3},
		3,
		[]int{1, 2, 3},
	},
	{
		[]int{1, 2, 3},
		4,
		[]int{3, 1, 2},
	},
	{
		[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		5,
		[]int{6, 7, 8, 9, 10, 1, 2, 3, 4, 5},
	},
	{
		[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		10,
		[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	},
	{
		[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		20,
		[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	},
	{
		[]int{},
		1,
		[]int{},
	},
}

func TestRotateRight(t *testing.T) {

	for _, test := range rotateRightTests {
		original := slices.Clone(test.s)
		if got := RotateRight(test.s, test.j); !slices.Equal(got, test.want) || !slices.Equal(test.s, original) || cap(got) > len(got) {
			t.Errorf("RotateRight(%v, %v) = %v (cap %v - len %v), want %v", test.s, test.j, got, cap(got), len(got), test.want)
		}
	}

	//panics tests
	for _, test := range []struct {
		name string
		s    []int
		j    int
	}{
		{"with negative index", []int{1, 2, 3, 4, 5}, -1},
	} {
		if !panics(func() { RotateRight(test.s, test.j) }) {
			t.Errorf("RotateRight %s: got no panic, want panic", test.name)
		}
	}

	// Test number of allocations
	for _, test := range rotateRightTests {

		allocs := testing.AllocsPerRun(100, func() {
			_ = RotateRight(test.s, test.j)
		})
		if len(test.s) > 0 && allocs != 1 {
			t.Errorf("RotateRight(%v) allocates %v times, want 1", test.s, allocs)
		} else if len(test.s) == 0 && allocs != 0 {
			t.Errorf("RotateRight(%v) allocates %v times, want 0", test.s, allocs)
		}

	}
}

var reverseTests = []struct {
	s    []int
	want []int
}{
	{
		[]int{1, 2, 3},
		[]int{3, 2, 1},
	},
	{
		[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		[]int{10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
	},
	{
		[]int{},
		[]int{},
	},
	{
		[]int{1},
		[]int{1},
	},
	{
		[]int{1, 1},
		[]int{1, 1},
	},
	{
		nil,
		[]int{},
	},
}

func TestReverse(t *testing.T) {

	for _, test := range reverseTests {
		original := slices.Clone(test.s)
		if got := Reverse(test.s); !slices.Equal(got, test.want) || !slices.Equal(test.s, original) || cap(got) > len(got) {
			t.Errorf("Reverse(%v) = %v (cap %v - len %v), want %v", test.s, got, cap(got), len(got), test.want)
		}
	}

	// Test number of allocations
	for _, test := range reverseTests {
		allocs := testing.AllocsPerRun(1000, func() {
			Reverse(test.s)
		})
		if len(test.s) > 0 && allocs != 1 {
			t.Errorf("Reverse(%v) allocates %v times, want 1", test.s, allocs)
		} else if len(test.s) == 0 && allocs != 0 {
			t.Errorf("Reverse(%v) allocates %v times, want 0", test.s, allocs)
		}
	}
}

var concatTests = []struct {
	s1   []int
	s2   []int
	want []int
}{
	{
		[]int{1, 2, 3},
		[]int{4, 5, 6},
		[]int{1, 2, 3, 4, 5, 6},
	},
	{
		[]int{1, 2, 3},
		[]int{},
		[]int{1, 2, 3},
	},
	{
		[]int{},
		[]int{4, 5, 6},
		[]int{4, 5, 6},
	},
	{
		[]int{},
		[]int{},
		[]int{},
	},
	{
		nil,
		[]int{4, 5, 6},
		[]int{4, 5, 6},
	},
	{
		[]int{1, 2, 3},
		nil,
		[]int{1, 2, 3},
	},
	{
		nil,
		nil,
		[]int{},
	},
}

func TestConcat(t *testing.T) {

	for _, test := range concatTests {
		original1 := slices.Clone(test.s1)
		original2 := slices.Clone(test.s2)
		if got := Concat(test.s1, test.s2); !slices.Equal(got, test.want) || !slices.Equal(original1, test.s1) || !slices.Equal(original2, test.s2) || cap(got) > len(got) {
			t.Errorf("Concat(%v, %v) = %v (cap %v - len %v), want %v", test.s1, test.s2, got, cap(got), len(got), test.want)
		}
	}

	// Test number of allocations
	for _, test := range concatTests {

		allocs := testing.AllocsPerRun(1000, func() {
			Concat(test.s1, test.s2)
		})
		if allocs > 1 {
			t.Errorf("Concat(%v, %v) allocates %v times, want <= 1", test.s1, test.s2, allocs)
		}

	}
}

var concatAllTests = []struct {
	s    [][]int
	want []int
}{
	{
		[][]int{

			[]int{1, 2, 3},
			[]int{4, 5, 6},
			[]int{7, 8, 9},
		},
		[]int{1, 2, 3, 4, 5, 6, 7, 8, 9},
	},
	{
		[][]int{
			[]int{1, 2, 3},
			[]int{},
			[]int{7, 8, 9},
		},
		[]int{1, 2, 3, 7, 8, 9},
	},
	{
		[][]int{
			[]int{1, 2, 3},
			[]int{4, 5, 6},
			[]int{},
		},
		[]int{1, 2, 3, 4, 5, 6},
	},
	{
		[][]int{
			[]int{},
			[]int{},
			[]int{},
		},
		[]int{},
	},
	{
		[][]int{
			[]int{1, 2, 3},
			[]int{4, 5, 6},
			[]int{7, 8, 9},
			[]int{10, 11, 12},
			[]int{13, 14, 15},
			nil,
			[]int{16, 17, 18},
			nil,
			[]int{19, 20, 21},
		},
		[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21},
	},
	{
		[][]int{
			nil,
			nil,
			nil,
		},
		[]int{},
	},
}

func TestConcatAll(t *testing.T) {

	for _, test := range concatAllTests {
		if got := ConcatAll(test.s...); !slices.Equal(got, test.want) || cap(got) > len(got) {
			t.Errorf("ConcatAll(%v) = %v (cap %v - len %v), want %v", test.s, got, cap(got), len(got), test.want)
		}
	}

	// Test number of allocations
	for _, test := range concatAllTests {

		allocs := testing.AllocsPerRun(1000, func() {
			ConcatAll(test.s...)
		})
		if allocs > 1 {
			t.Errorf("ConcatAll(%v) allocates %v times, want <= 1", test.s, allocs)
		}

	}
}

var repeatTests = []struct {
	s    []int
	n    int
	want []int
}{
	{
		[]int{1, 2, 3},
		3,
		[]int{1, 2, 3, 1, 2, 3, 1, 2, 3},
	},
	{
		[]int{1, 2, 3},
		0,
		[]int{},
	},
	{
		[]int{1, 2, 3},
		1,
		[]int{1, 2, 3},
	},
	{
		[]int{1, 2, 3},
		-1,
		[]int{},
	},
	{
		[]int{},
		3,
		[]int{},
	},
	{
		nil,
		3,
		[]int{},
	},
}

func TestRepeat(t *testing.T) {

	for _, test := range repeatTests {
		original := slices.Clone(test.s)
		if got := Repeat(test.s, test.n); !slices.Equal(got, test.want) || !slices.Equal(original, test.s) || cap(got) > len(got) {
			t.Errorf("Repeat(%v, %v) = %v (cap %v - len %v), want %v", test.s, test.n, got, cap(got), len(got), test.want)
		}
	}

	//panics test
	for _, test := range []struct {
		name string
		s    []rune
		n    int
	}{
		{"overflow length", []rune("abc"), math.MaxInt},
	} {
		if !panics(func() { Repeat(test.s, test.n) }) {
			t.Errorf("Repeat(%v, %v) should panic", test.s, test.n)
		}
	}

	// Test number of allocations
	for _, test := range repeatTests {

		allocs := testing.AllocsPerRun(1000, func() {
			Repeat(test.s, test.n)
		})
		if allocs > 1 {
			t.Errorf("Repeat(%v, %v) allocates %v times, want <= 1", test.s, test.n, allocs)
		}

	}
}

var mapTests = []struct {
	s    []int
	f    func(int, int) string
	want []string
}{
	{
		[]int{1, 2, 3},
		func(i int, v int) string {
			return strconv.Itoa(i)
		},
		[]string{"0", "1", "2"},
	},
	{
		[]int{1, 2, 3},
		func(i int, v int) string { return strconv.Itoa(i * 2) },
		[]string{"0", "2", "4"},
	},
	{
		[]int{},
		func(i int, v int) string { return strconv.Itoa(i * 3) },
		[]string{},
	},
	{
		nil,
		func(i int, v int) string { return strconv.Itoa(i * 3) },
		[]string{},
	},
	{
		[]int{4, 5, 6},
		func(i int, v int) string { return strconv.Itoa(v) },
		[]string{"4", "5", "6"},
	},
	{
		[]int{1, 2, 3},
		nil,
		[]string{},
	},
	{
		nil,
		nil,
		[]string{},
	},
}

func TestMap(t *testing.T) {

	for _, test := range mapTests {
		original := slices.Clone(test.s)
		if got := Map(test.s, test.f); !slices.Equal(got, test.want) || !slices.Equal(original, test.s) || cap(got) > len(got) {
			t.Errorf("Map(%v) = %v (cap %v - len %v), want %v", test.s, got, cap(got), len(got), test.want)
		}
	}

	// Test number of allocations
	for _, test := range mapTests {

		allocs := testing.AllocsPerRun(1000, func() {
			Map(test.s, test.f)
		})
		if allocs > 1 {
			t.Errorf("Map(%v) allocates %v times, want <= 1", test.s, allocs)
		}

	}
}

//var flatTests = []struct {
//	s     []any
//	depth int
//	want  []any
//}{
//	{
//		[]any{1, 2, []int{3, 4}},
//		1,
//		[]any{1, 2, 3, 4},
//	},
//	{
//		[]any{1, 2, []any{3, 4, []int{5, 6}}},
//		1,
//		[]any{1, 2, 3, 4, []int{5, 6}},
//	},
//	{
//		[]any{1, 2, []any{3, 4, []int{5, 6}}},
//		2,
//		[]any{1, 2, 3, 4, 5, 6},
//	},
//	{
//		[]any{1, 2, []any{3, 4, []int{5, 6}}},
//		3,
//		[]any{1, 2, 3, 4, 5, 6},
//	},
//	{
//		[]any{1, 2, []any{3, 4, []int{}}},
//		2,
//		[]any{1, 2, 3, 4},
//	},
//}
//
//func TestFlat(t *testing.T) {
//
//	for _, test := range flatTests {
//		original := slices.Clone(test.s)
//		if got := Flat(test.s, test.depth); !slices.Equal(got, test.want) || !slices.Equal(original, test.s) || cap(got) > len(got) {
//			t.Errorf("Flat(%v, %v) = %v (cap %v - len %v), want %v", test.s, test.depth, got, cap(got), len(got), test.want)
//		}
//	}
//
//	//panics test
//	for _, test := range []struct {
//		name  string
//		s     []any
//		depth int
//	}{
//		{"negative depth", []any{1, 2, []any{3, 4, []int{5, 6}}}, -1},
//	} {
//		if !panics(func() { Flat(test.s, test.depth) }) {
//			t.Errorf("Flat(%v, %v) should panic", test.s, test.depth)
//		}
//	}
//
//	// Test number of allocations
//	for _, test := range flatTests {
//
//		allocs := testing.AllocsPerRun(1000, func() {
//			Flat(test.s, test.depth)
//		})
//		if allocs > 1 {
//			t.Logf("Flat(%v, %v) allocates %v times, want <= 1", test.s, test.depth, allocs)
//		}
//
//	}
//
//}

var filterTests = []struct {
	s    []string
	f    func(int, string) bool
	want []string
}{
	{
		[]string{"a", "b", "c"},
		func(i int, v string) bool { return i == 1 },
		[]string{"b"},
	},
	{
		[]string{"a", "b", "c"},
		func(i int, v string) bool { return v == "a" },
		[]string{"a"},
	},
	{
		[]string{"a", "b", "c"},
		func(i int, v string) bool { return v == "d" },
		[]string{},
	},
	{
		[]string{"a", "b", "c"},
		func(i int, v string) bool { return i == 3 },
		[]string{},
	},
	{
		[]string{"a", "b", "c"},
		func(i int, v string) bool { return i == 1 || v == "c" },
		[]string{"b", "c"},
	},
	{
		[]string{},
		func(i int, v string) bool { return i == 1 },
		[]string{},
	},
	{
		nil,
		func(i int, v string) bool { return i == 1 },

		[]string{},
	},
	{
		[]string{"a", "b", "c"},
		nil,
		[]string{},
	},
	{
		nil,
		nil,
		[]string{},
	},
	{
		[]string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"},
		func(i int, v string) bool { return v < "m" },
		[]string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l"},
	},
}

func TestFilter(t *testing.T) {

	for _, test := range filterTests {
		original := slices.Clone(test.s)
		if got := Filter(test.s, test.f); !slices.Equal(got, test.want) || !slices.Equal(original, test.s) {
			t.Errorf("Filter(%v) = %v (cap %v - len %v), want %v", test.s, got, cap(got), len(got), test.want)
		}
	}

	// Test number of allocations
	for _, test := range filterTests {

		allocs := testing.AllocsPerRun(1000, func() {
			Filter(test.s, test.f)
		})
		if allocs > 1 {
			t.Logf("Filter(%v) allocates %v times", test.s, allocs)
		}

	}
}

var partitionTests = []struct {
	s    []string
	f    func(int, string) bool
	want [][]string
}{
	{
		[]string{"a", "b", "c"},
		func(i int, v string) bool { return i == 1 },
		[][]string{{"b"}, {"a", "c"}},
	},
	{
		[]string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"},
		func(i int, v string) bool { return v < "m" },
		[][]string{{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l"}, {"m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"}},
	},
	// edges cases
	{
		[]string{"a", "b", "c"},
		func(i int, v string) bool { return i == 3 },
		[][]string{{}, {"a", "b", "c"}},
	},
	{
		[]string{"a", "b", "c"},
		func(i int, v string) bool { return i == 1 || v == "c" },
		[][]string{{"b", "c"}, {"a"}},
	},
	{
		[]string{},
		func(i int, v string) bool { return i == 1 },
		[][]string{{}, {}},
	},
	{
		nil,
		func(i int, v string) bool { return i == 1 },
		[][]string{{}, {}},
	},
	{
		[]string{"a", "b", "c"},
		nil,
		[][]string{{}, {}},
	},
	{
		nil,
		nil,
		[][]string{{}, {}},
	},
}

func TestPartition(t *testing.T) {

	for _, test := range partitionTests {
		original := slices.Clone(test.s)
		if matched, unmatched := Partition(test.s, test.f); !slices.Equal(matched, test.want[0]) || !slices.Equal(unmatched, test.want[1]) || !slices.Equal(original, test.s) {
			t.Errorf("Partition(%v) = (%v, %v), want (%v, %v)", test.s, matched, unmatched, test.want[0], test.want[1])
		}
	}

	// Test number of allocations
	for _, test := range partitionTests {

		allocs := testing.AllocsPerRun(1000, func() {
			Partition(test.s, test.f)
		})
		if allocs > 1 {
			t.Logf("Partition(%v) allocates %v times", test.s, allocs)
		}

	}

}
