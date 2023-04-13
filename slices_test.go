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

}

var overrideTests = []struct {
	s    []int
	i    int
	v    []int
	want []int
}{
	{[]int{1, 2, 3}, 0, []int{4, 5, 6}, []int{4, 5, 6}},
	{[]int{1, 2, 3}, 1, []int{4, 5, 6}, []int{1, 4, 5, 6}},
	{[]int{1, 2, 3}, 2, []int{4, 5, 6}, []int{1, 2, 4, 5, 6}},
	{[]int{1, 2, 3}, 3, []int{4, 5, 6}, []int{1, 2, 3, 4, 5, 6}},
	{[]int{1, 2, 3}, 0, []int{}, []int{1, 2, 3}},
	{[]int{1, 2, 3}, 1, []int{}, []int{1, 2, 3}},
	{[]int{1, 2, 3}, 2, []int{}, []int{1, 2, 3}},
	{[]int{1, 2, 3}, 3, []int{}, []int{1, 2, 3}},
	{[]int{1, 2, 3}, 0, []int{4}, []int{4, 2, 3}},
	{[]int{1, 2, 3}, 1, []int{4}, []int{1, 4, 3}},
	{[]int{1, 2, 3}, 2, []int{4}, []int{1, 2, 4}},
	{[]int{1, 2, 3}, 3, []int{4}, []int{1, 2, 3, 4}},
	{[]int{1, 2, 3}, 0, []int{4, 5}, []int{4, 5, 3}},
	{[]int{1, 2, 3}, 1, []int{4, 5}, []int{1, 4, 5}},
	{[]int{1, 2, 3}, 2, []int{4, 5}, []int{1, 2, 4, 5}},
}

func TestOverride(t *testing.T) {

	for _, test := range overrideTests {
		got := Override(test.s, test.i, test.v...)
		if !slices.Equal(got, test.want) {
			t.Errorf("SetOverride(%v, %v, %v) = %v, want %v", test.s, test.i, test.v, got, test.want)
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

		if !panics(func() { Override(test.s, test.i, test.v...) }) {
			t.Errorf("Override %s: got no panic, want panic", test.name)
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
		if got := Concat(test.s1, test.s2); !slices.Equal(got, test.want) || cap(got) > len(got) {
			t.Errorf("Concat(%v, %v) = %v (cap %v - len %v), want %v", test.s1, test.s2, got, cap(got), len(got), test.want)
		}
	}
}
