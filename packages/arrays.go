// name: arrays
// description: Array utilities
// author: Mist
// requires: sort

type Arrays struct{}

func (Arrays) Clone(arr []any) []any {
	return append([]any{}, arr...)
}

func (Arrays) SortBy(arr []any, key string) []any {
	sort.Slice(arr, func(i, j int) bool {
		m1 := arr[i].(map[string]any)
		m2 := arr[j].(map[string]any)
		return OSLtoString(m1[key]) < OSLtoString(m2[key])
	})
	return arr
}

func (Arrays) Sort(arr []any) []any {
	sort.Slice(arr, func(i, j int) bool {
		return OSLtoString(arr[i]) < OSLtoString(arr[j])
	})
	return arr
}

func (Arrays) SortByNum(arr []any, key string) []any {
	sort.Slice(arr, func(i, j int) bool {
		m1 := arr[i].(map[string]any)
		m2 := arr[j].(map[string]any)
		return OSLcastNumber(m1[key]) < OSLcastNumber(m2[key])
	})
	return arr
}

func (Arrays) Filter(arr []any, fn func(any) bool) []any {
	var filtered []any
	for _, v := range arr {
		if fn(v) {
			filtered = append(filtered, v)
		}
	}
	if len(filtered) == 0 {
		return []any{}
	}
	return filtered
}

func (Arrays) Map(arr []any, fn func(any) any) []any {
	var mapped []any
	for _, v := range arr {
		mapped = append(mapped, fn(v))
	}
	if len(mapped) == 0 {
		return []any{}
	}
	return mapped
}

func (Arrays) Reduce(arr []any, fn func(any, any) any, initial any) any {
	for _, v := range arr {
		initial = fn(initial, v)
	}
	return initial
}

func (Arrays) ReduceRight(arr []any, fn func(any, any) any, initial any) any {
	for i := len(arr) - 1; i >= 0; i-- {
		initial = fn(initial, arr[i])
	}
	return initial
}

func (Arrays) Reverse(arr []any) []any {
	for i, j := 0, len(arr)-1; i < j; i, j = i+1, j-1 {
		arr[i], arr[j] = arr[j], arr[i]
	}
	return arr
}

func (Arrays) Find(arr []any, fn func(any) bool) any {
	for _, v := range arr {
		if fn(v) {
			return v
		}
	}
	return nil
}

func (Arrays) FindIndex(arr []any, fn func(any) bool) int {
	for i, v := range arr {
		if fn(v) {
			return i
		}
	}
	return -1
}

func (Arrays) FindLastIndex(arr []any, fn func(any) bool) int {
	for i, v := range arr {
		if fn(v) {
			return i
		}
	}
	return -1
}

func (Arrays) Reverse(arr []any) []any {
	i, j := 0, len(arr)-1
	arr[i], arr[j] = arr[j], arr[i]
	return arr
}

var arrays = Arrays{}