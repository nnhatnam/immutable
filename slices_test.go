package immutable

import (
	"golang.org/x/exp/slices"
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

func TestSet(t *testing.T) {

	s := []int{1, 2, 3}
	s2 := Set(s, 1, 4)

	if s2[1] != 4 {
		t.Errorf("s2 is not modified, expected %v but got %v", 4, s2[1])
	}

	if s[1] != 2 {
		t.Errorf("s is modified, expected %v but got %v", 2, s[1])
	}

	//panics tests
	for _, test := range []struct {
		name string
		s    []int
		i    int
		v    int
	}{
		{"with negative index", []int{42}, -1, 10},
		{"with out-of-bounds index", []int{42}, 2, 10},
	} {
		if !panics(func() { Set(test.s, test.i, test.v) }) {
			t.Errorf("Delete %s: got no panic, want panic", test.name)
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
}

var removeTests = []struct {
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

func TestRemove(t *testing.T) {

	for _, test := range removeTests {
		original := slices.Clone(test.s)
		if got := Remove(test.s, test.i); !slices.Equal(got, test.want) || !slices.Equal(test.s, original) {
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
		if !panics(func() { Remove(test.s, test.i) }) {
			t.Errorf("Remove %s: got no panic, want panic", test.name)
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

	//panics tests are not needed, because Push is just a wrapper around Insert

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

	//panics tests are not needed, because PushFront is just a wrapper around Insert

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

	//panics tests are not needed, because Pop is just a wrapper around Remove

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

	//panics tests are not needed, because PopFront is just a wrapper around Remove

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

}