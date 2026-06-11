package utils

// Intersection returns values from left that also occur in right. It preserves
// the order and duplicates from left, matching the Python implementation.
func Intersection[T comparable](left, right []T) []T {
	result := make([]T, 0)
	for _, leftValue := range left {
		for _, rightValue := range right {
			if leftValue == rightValue {
				result = append(result, leftValue)
				break
			}
		}
	}
	return result
}
