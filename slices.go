package immutable

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

// Remove creates a new slice that duplicates `s` and then removes the element at index `i`.
// Remove panics if `i` is out of range.
// In the returned slice r, the values [i+1, len(s)) are shifted left by one index, so r[i] == s[i+1].
// Remove is equivalent to RemoveRange(s, i, i+1).
// This function is O(len(s)).
func Remove[S ~[]E, E any](s S, i int) S {

	return RemoveRange(s, i, i+1)

}

// RemoveRange creates a new slice that duplicates `s` and then removes the elements in the range [i, j).
// RemoveRange panics if `i` or `j` is out of range, or if `i > j`.
// In the returned slice r, the values [j, len(s)) are shifted left by `j-i` indices, so r[i] == s[j].
// This function is O(len(s)).
func RemoveRange[S ~[]E, E any](s S, i, j int) S {

	// bounce checks. Equivalent to:
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

	return Remove(s, len(s)-1), s[len(s)-1]

}

// PopFront returns a new slice that duplicates `s` from [1, len(s)) and the element at index 0.
// PopFront does nothing if s is empty.
// In the returned slice r, r[i] == s[i+1] for all i in [0, len(s)-1).
// This function is O(len(s)).
func PopFront[S ~[]E, E any](s S) (_ S, _ E) {

	if len(s) == 0 {
		return
	}

	return Remove(s, 0), s[0]
}

// ShiftLeft creates a news slice from `s` with items got shifted by `j` positions to the left and the first `j` items removed.
// ShiftLeft keeps the length of the slice the same, so the last `j` items are set to the zero value of `E`.
// If you want to shift the first `j` items and do not wish to keep the length, use RemoveRange(s, 0, j) instead.
// ShiftLeft panics if `j` is negative.
// In the returned slice r, r[i] == s[i+j] for all i in [0, len(s)-j).
// This function is O(j).
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
// This function is O(j).
func ShiftRight[S ~[]E, E any](s S, j int) S {
	if j > len(s) {
		j = len(s)
	}

	s2 := make([]E, len(s))
	copy(s2[j:], s)
	return s2

}
