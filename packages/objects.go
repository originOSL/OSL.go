// name: objects
// description: Object utilities
// author: Mist
// requires: os

type Objects struct{}

func (Objects) Clone(obj map[string]any) map[string]any {
	clone := make(map[string]any, len(obj))
	for k, v := range obj {
		clone[k] = v
	}
	return clone
}

func (Objects) DeepClone(obj map[string]any) map[string]any {
	clone := make(map[string]any, len(obj))
	for k, v := range obj {
		switch v := v.(type) {
		case map[string]any:
			clone[k] = Objects{}.DeepClone(v)
		case []any:
			clone[k] = deepCloneSlice(v)
		default:
			clone[k] = v
		}
	}
	return clone
}

func deepCloneSlice(arr []any) []any {
	clone := make([]any, len(arr))
	for i, v := range arr {
		switch v := v.(type) {
		case map[string]any:
			clone[i] = Objects{}.DeepClone(v)
		case []any:
			clone[i] = deepCloneSlice(v)
		default:
			clone[i] = v
		}
	}
	return clone
}

func (Objects) Merge(maps ...map[string]any) map[string]any {
	out := make(map[string]any)
	for _, m := range maps {
		for k, v := range m {
			out[k] = v
		}
	}
	return out
}

func (Objects) Keys(obj map[string]any) []string {
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	return keys
}

func (Objects) Values(obj map[string]any) []any {
	values := make([]any, 0, len(obj))
	for _, v := range obj {
		values = append(values, v)
	}
	return values
}

func (Objects) Has(obj map[string]any, key any) bool {
	_, ok := obj[OSLcastString(key)]
	return ok
}

func (Objects) Pick(obj map[string]any, keys ...string) map[string]any {
	out := make(map[string]any)
	for _, k := range keys {
		if val, ok := obj[k]; ok {
			out[k] = val
		}
	}
	return out
}

func (Objects) Omit(obj map[string]any, keys ...string) map[string]any {
	out := Objects{}.Clone(obj)
	for _, k := range keys {
		delete(out, k)
	}
	return out
}

func (Objects) Filter(obj map[string]any, fn func(key string, value any) bool) map[string]any {
	out := make(map[string]any)
	for k, v := range obj {
		if fn(k, v) {
			out[k] = v
		}
	}
	return out
}

var objects = Objects{}