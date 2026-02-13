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

var arrays = Arrays{}