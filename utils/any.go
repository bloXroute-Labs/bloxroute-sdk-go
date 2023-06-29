package utils

func Contains[T comparable](s []T, val T) bool {
	for _, v := range s {
		if v == val {
			return true
		}
	}
	return false
}
