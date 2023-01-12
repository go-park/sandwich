package collections

func Keys[K comparable, V any](m map[K]V) []K {
	var result []K
	for k := range m {
		result = append(result, k)
	}
	return result
}

func Values[K comparable, V any](m map[K]V) []V {
	var result []V
	for _, v := range m {
		result = append(result, v)
	}
	return result
}

func Contains[T comparable](list []T, t T) bool {
	for _, v := range list {
		if v == t {
			return true
		}
	}
	return false
}

func ContainsAny[T comparable](list []T, values ...T) bool {
	f := func(list []T, values ...T) bool {
		for _, v := range list {
			if Contains(values, v) {
				return true
			}
		}
		return false
	}
	if len(list) > len(values) {
		return f(values, list...)
	}
	return f(list, values...)
}
