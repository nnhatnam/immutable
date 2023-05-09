package slice

// Set creates a new slice that is identical to `s` except that the element at index `i` is set to `v`.
// Set panics if `i` is out of range.
// In the returned slice, r[i] == v.
// This function is O(len(s)).
func Set[S ~[]E, E any](s S, i int, v E) S {
	s2 := make([]E, len(s))
	copy(s2, s)
	s2[i] = v
	return s2
}

// Copy is smiliar to `copy` but returns a new slice instead of modifying the destination slice.
func Copy[S ~[]E, E any](dst, src S, i int) S {

	s2 := make([]E, len(dst))
	copy(s2, dst)
	copy(s2[i:], src)
	return s2
}

// Update creates a new slice that is identical to `s` except that the element at index `i` is set to `f(s[i])`.
// Update panics if `i` is out of range.
// If `f` is nil, this function is equivalent to Set(s, i, s[i]).
// Call Update on empty slices is a no-op.
// In the returned slice, r[i] == f(s[i]).
// This function is O(len(s)).
func Update[S ~[]E, E any](s S, i int, f func(E) E) S {

	s2 := make([]E, len(s))
	copy(s2, s)

	if len(s) == 0 {
		return s2
	}

	if f != nil {
		s2[i] = f(s[i])
	}
	return s2
}

// Insert creates a new slice that duplicates `s` and then inserts values v... into the new slice at index `i`.
// Insert panics if `i` is out of range.
// In the returned slice r, the values v... are in the range [i, i+len(v)), so r[i] == v[0] and r[i+len(v)-1] == v[len(v)-1].
// This function is O(len(s)+len(v)).
func Insert[S ~[]E, E any](s S, i int, v ...E) S {

	s2 := make([]E, len(s)+len(v))
	copy(s2, s[:i])
	copy(s2[i:], v)
	copy(s2[i+len(v):], s[i:])
	return s2
}

// Remove creates a new slice that duplicates `s` and then removes the first occurrence of `v`.
// If `v` is not in `s`, the returned slice is identical to `s`.
// This function is O(len(s)).
func Remove[S ~[]E, E comparable](s S, v E) S {

	for i, e := range s {
		if e == v {
			s2 := make([]E, len(s)-1)
			copy(s2, s[:i])
			copy(s2[i:], s[i+1:])
			return s2
		}
	}

	s2 := make([]E, len(s))
	copy(s2, s)
	return s2

}

// RemoveAt creates a new slice that duplicates `s` and then removes the element at index `i`.
// RemoveAt panics if `i` is out of range.
// In the returned slice r, the values [i+1, len(s)) are shifted left by one index, so r[i] == s[i+1].
// RemoveAt is equivalent to RemoveRange(s, i, i+1).
// This function is O(len(s)).
func RemoveAt[S ~[]E, E any](s S, i int) S {

	return RemoveRange(s, i, i+1)

}

// RemoveRange creates a new slice that duplicates `s` and then removes the elements in the range [i, j).
// RemoveRange panics if `i` or `j` is out of range, or if `i > j`.
// In the returned slice r, the values [j, len(s)) are shifted left by `j-i` indices, so r[i] == s[j].
// This function is O(len(s)).
func RemoveRange[S ~[]E, E any](s S, i, j int) S {

	// bound checks. Equivalent to:
	//	if i > j {
	//		panic()
	//	}
	_ = s[i:j]

	s2 := make([]E, len(s)-(j-i))
	copy(s2, s[:i])
	copy(s2[i:], s[j:])
	return s2
}

// Push creates a new slice that duplicates `s` and then appends values v... to the end of the new slice.
// In the returned slice r, the values v... are in the range [len(s), len(s)+len(v)), so r[len(s)] == v[0] and r[len(s)+len(v)-1] == v[len(v)-1].
// Append is equivalent to Insert(s, len(s), v...).
// This function is O(len(s)+len(v)).
func Push[S ~[]E, E any](s S, v ...E) S {

	return Insert(s, len(s), v...)

}

// PushFront creates a new slice that duplicates `s` and then prepends values v... to the beginning of the new slice.
// In the returned slice r, the values v... are in the range [0, len(v)), so r[0] == v[0] and r[len(v)-1] == v[len(v)-1].
// Prepend is equivalent to Insert(s, 0, v...).
// This function is O(len(s)+len(v)).
func PushFront[S ~[]E, E any](s S, v ...E) S {

	return Insert(s, 0, v...)

}

// Pop returns a new slice that duplicates `s` from [0, len(s)-1) and the element at index `len(s)-1`.
// Pop does nothing if s is empty.
// In the returned slice r, r[i] == s[i] for all i in [0, len(s)-1).
// This function is O(len(s)).
func Pop[S ~[]E, E any](s S) (_ S, _ E) {

	if len(s) == 0 {
		return
	}

	return RemoveAt(s, len(s)-1), s[len(s)-1]

}

// PopFront returns a new slice that duplicates `s` from [1, len(s)) and the element at index 0.
// PopFront does nothing if s is empty.
// In the returned slice r, r[i] == s[i+1] for all i in [0, len(s)-1).
// This function is O(len(s)).
func PopFront[S ~[]E, E any](s S) (_ S, _ E) {

	if len(s) == 0 {
		return
	}

	return RemoveAt(s, 0), s[0]
}

// ShiftLeft creates a news slice from `s` with items got shifted by `j` positions to the left and the first `j` items removed.
// ShiftLeft keeps the length of the slice the same, so the last `j` items are set to the zero value of `E`.
// If you want to shift the first `j` items and do not wish to keep the length, use RemoveRange(s, 0, j) instead.
// ShiftLeft panics if `j` is negative.
// In the returned slice r, r[i] == s[i+j] for all i in [0, len(s)-j).
// This function is O(len(s)).
func ShiftLeft[S ~[]E, E any](s S, j int) S {

	if j > len(s) {
		j = len(s)
	}

	s2 := make([]E, len(s))
	copy(s2, s[j:])
	return s2

}

// ShiftRight creates a news slice from `s` with items got shifted by `j` positions to the right and the last `j` items removed.
// ShiftRight keeps the length of the slice the same, so the first `j` items are set to the zero value of `E`.
// If you want to shift right the last `j` items and do not wish to keep the length, use RemoveRange(s, len(s)-j, len(s)) instead.
// ShiftRight panics if `j` is negative.
// In the returned slice r, r[i] == s[i-j] for all i in [j, len(s)).
// This function is O(len(s)).
func ShiftRight[S ~[]E, E any](s S, j int) S {
	if j > len(s) {
		j = len(s)
	}

	s2 := make([]E, len(s))
	copy(s2[j:], s)
	return s2

}

// RotateLeft creates a new slice from `s` with items got rotated by `j` positions to the left.
// RotateLeft panics if `j` is negative.
// In the returned slice r, r[i] == s[(i+j)%len(s)] for all i in [0, len(s)).
// This function is O(len(s)).
func RotateLeft[S ~[]E, E any](s S, j int) S {

	if len(s) == 0 {
		return make([]E, 0)
	}

	if j > len(s) {
		j = j % len(s)
	}

	s2 := make([]E, len(s))
	copy(s2, s[j:])
	copy(s2[len(s)-j:], s[:j])
	return s2

}

// RotateRight creates a new slice from `s` with items got rotated by `j` positions to the right.
// RotateRight panics if `j` is negative.
// In the returned slice r, r[i] == s[(i-j)%len(s)] for all i in [0, len(s)).
// This function is O(len(s)).
func RotateRight[S ~[]E, E any](s S, j int) S {

	if len(s) == 0 {
		return make([]E, 0)
	}

	if j > len(s) {
		j = j % len(s)
	}

	s2 := make([]E, len(s))
	copy(s2[j:], s)
	copy(s2, s[len(s)-j:])
	return s2

}

// Reverse creates a new slice from `s` with items got reversed.
// In the returned slice r, r[i] == s[len(s)-i-1] for all i in [0, len(s)).
// This function is O(len(s)).
func Reverse[S ~[]E, E any](s S) S {

	s2 := make([]E, len(s))
	for i := range s {
		s2[i] = s[len(s)-i-1]
	}
	return s2

}

// Concat creates a new slice that is the result of concatenating `s` and `v`.
// In the returned slice r, r[i] == s[i] for all i in [0, len(s)) and r[i] == v[i-len(s)] for all i in [len(s), len(s)+len(v)).
// This function is O(len(s)+len(v)).
func Concat[S ~[]E, E any](s S, v S) S {

	s2 := make([]E, len(s)+len(v))
	copy(s2, s)
	copy(s2[len(s):], v)
	return s2

}

// ConcatAll creates a new slice that is the result of concatenating all the slices in `s`.
// In the returned slice r, r[i] == s[i] for all i in [0, len(s)) and r[i] == v[i-len(s)] for all i in [len(s), len(s)+len(v)).
// This function is O(len(s)+len(v)).
func ConcatAll[S ~[]E, E any](s ...S) S {

	if len(s) == 0 {
		return make([]E, 0)
	}

	length := 0
	for _, v := range s {
		length += len(v)
	}
	s2 := make([]E, length)
	i := 0
	for _, v := range s {
		copy(s2[i:], v)
		i += len(v)
	}

	return s2

}

// Repeat creates a new slice that is the result of repeating `s` `n` times.
// In the returned slice r, r[i] == s[i%len(s)] for all i in [0, len(s)*n).
// Repeat panics if len(s)*n overflows, and returns an empty slice if n <= 0.
// This function is O(len(s)*n).
func Repeat[S ~[]E, E any](s S, n int) S {

	if n <= 0 {
		return make([]E, 0)
	}

	//detect overflow
	length := len(s) * n
	if length/n != len(s) {
		panic("len(s)*n overflows")
	}

	s2 := make([]E, len(s)*n)
	for i := 0; i < n; i++ {
		copy(s2[i*len(s):], s)
	}

	return s2

}

// Map creates a new slice from `s` by applying `f` to each item.
// In the returned slice r, r[i] == f(i, s[i]) for all i in [0, len(s)).
// If f is nil, Map returns an empty slice.
// This function is O(len(s)).
func Map[S ~[]E, E any, T any](s S, f func(i int, e E) T) []T {

	if len(s) == 0 || f == nil {
		return make([]T, 0)
	}

	s2 := make([]T, len(s))
	for i, v := range s {
		s2[i] = f(i, v)
	}
	return s2

}

// Flat returns a shallow copy of `s` with all sub slices of `s` concatenated into it recursively up to `depth` times.
// If `depth` is 0, Flat returns a copy of `s`.
// If `depth` is negative, Flat panics.
// Flat is expensive because it can't calculate the length of the returned slice.
// So, it creates a temporary slice, and append each item (after flattening) to it causing a lot of memory re-allocations.
//func Flat[S ~[]E, E any](s S, depth int) []E {
//
//	if depth == 0 {
//		s2 := make([]E, len(s))
//		copy(s2, s)
//		return s2
//	}
//
//	if depth < 0 {
//		panic("depth must be >= 0")
//	}
//
//	if len(s) == 0 {
//		return make([]E, 0)
//	}
//
//	s2 := make([]E, 0)
//	for _, v := range s {
//		if v1, ok := any(v).([]E); ok {
//			s2 = append(s2, Flat(v1, depth-1)...)
//		} else {
//			s2 = append(s2, v1...)
//		}
//	}
//	return s2
//}

// FlatMap returns a new slice formed by applying `f` to each item in s, and then flattening the results by one level.
//func FlatMap[S ~[]E, E any, T any](s S, f func(i int, e E) []T) []T {
//
//	if len(s) == 0 || f == nil {
//		return make([]T, 0)
//	}
//
//	s2 := make([]T, 0)
//	for i, v := range s {
//		s2 = append(s2, f(i, v)...)
//	}
//	return s2
//
//}

// Filter creates a new slice from `s` by selecting only the items that satisfy the predicate `f`.
// In the returned slice r, r[i] == s[j] for all i in [0, len(r)) and f(j, s[j]) == true for all j in [0, len(s)).
// If f is nil, Filter returns an empty slice.
// This function is O(len(s)).
func Filter[S ~[]E, E any](s S, f func(i int, e E) bool) (filtered S) {

	if len(s) == 0 || f == nil {
		return
	}

	for i, v := range s {
		if f(i, v) {
			filtered = append(filtered, v)
		}
	}
	return filtered

}

// Partition return two new slices from `s` by selecting only the items that satisfy the predicate `f` and the items that don't.
// The first returned slice r1 contains all the items that satisfy the predicate `f`.
// The second returned slice r2 contains all the items that don't satisfy the predicate `f`.
// When f is called on each item, it's responsibility of the caller to make sure that the predicate is not changing the state of the slice.
func Partition[S ~[]E, E any](s S, f func(i int, e E) bool) (matched S, unmatched S) {

	if len(s) == 0 || f == nil {
		return
	}

	for i, v := range s {
		if f(i, v) {
			matched = append(matched, v)
		} else {
			unmatched = append(unmatched, v)
		}
	}
	return

}
