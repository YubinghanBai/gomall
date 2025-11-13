package utils

// Ptr returns a pointer to the given value
func Ptr[T any](v T) *T {
	return &v
}

// PtrValue returns the value of a pointer, or zero value if nil
func PtrValue[T any](p *T) T {
	if p != nil {
		return *p
	}
	var zero T
	return zero
}

// PtrValueOr returns the value of a pointer, or the default value if nil
func PtrValueOr[T any](p *T, defaultValue T) T {
	if p != nil {
		return *p
	}
	return defaultValue
}
