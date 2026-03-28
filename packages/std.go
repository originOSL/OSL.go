// This is a set of funtions that are used in the compiler for OSL.go

func getGamepads() []any {
	// Stub implementation - returns empty array
	// This should be implemented with actual gamepad support
	return []any{}
}

func dist(x1, y1, x2, y2 float64) float64 {
	dx := x1 - x2
	dy := y1 - y2
	return math.Sqrt(dx*dx + dy*dy)
}

func OSLlen(s any) int {
	if s == nil {
		return 0
	}
	switch s := s.(type) {
	case string:
		return len(s)
	case []any:
		return len(s)
	case []string:
		return len(s)
	case []int:
		return len(s)
	case []float64:
		return len(s)
	case []bool:
		return len(s)
	case []byte:
		return len(s)
	case []OSLio.Reader:
		return len(s)
	}
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return 0
		}
		v = v.Elem()
	}
	if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		return v.Len()
	}
	if v.Kind() == reflect.Map {
		return v.Len()
	}
	if v.Kind() == reflect.String {
		return len(v.String())
	}
	panic("OSLlen, invalid type: " + v.Kind().String())
}

func encodeURIComponent(str string) string {
	var buf strings.Builder
	hex := "0123456789ABCDEF"

	for i := 0; i < len(str); i++ {
		c := str[i]

		if (c >= 'A' && c <= 'Z') ||
			(c >= 'a' && c <= 'z') ||
			(c >= '0' && c <= '9') ||
			c == '-' || c == '_' || c == '.' || c == '!' ||
			c == '~' || c == '*' || c == '\'' || c == '(' || c == ')' {

			buf.WriteByte(c)
		} else {
			buf.WriteByte('%')
			buf.WriteByte(hex[c>>4])
			buf.WriteByte(hex[c&15])
		}
	}

	return buf.String()
}

func decodeURIComponent(s string) string {
	result := make([]byte, 0, len(s))

	for i := 0; i < len(s); {
		if s[i] == '%' {
			if i+2 >= len(s) {
				return ""
			}

			h1 := OSLfromHex(s[i+1])
			h2 := OSLfromHex(s[i+2])
			if h1 == -1 || h2 == -1 {
				return ""
			}

			result = append(result, byte(h1<<4|h2))
			i += 3
		} else {
			result = append(result, s[i])
			i++
		}
	}

	return string(result)
}

func OSLfromHex(c byte) int {
	switch {
	case '0' <= c && c <= '9':
		return int(c - '0')
	case 'A' <= c && c <= 'F':
		return int(c - 'A' + 10)
	case 'a' <= c && c <= 'f':
		return int(c - 'a' + 10)
	default:
		return -1
	}
}

func OSLtoString(s any) string {
	switch s := s.(type) {
	case string:
		return s
	case []byte:
		return string(s)
	case []any:
		return JsonStringify(s)
	case map[string]any, map[string]string, map[string]int, map[string]float64, map[string]bool:
		return JsonStringify(s)
	case OSLio.Reader:
		data, err := OSLio.ReadAll(s)
		if err != nil {
			panic("OSLcastString: failed to read OSLio.Reader:" + err.Error())
		}
		return string(data)
	case int, int64:
		return fmt.Sprintf("%d", s)
	case float64:
		return fmt.Sprintf("%g", s)
	default:
		return fmt.Sprintf("%v", s)
	}
}

func OSLcastObject(s any) map[string]any {
	if s == nil {
		return map[string]any{}
	}
	obj, ok := s.(map[string]any)
	if ok {
		return obj
	}
	panic("OSLcastObject, invalid type: " + reflect.TypeOf(s).String())

}

func OSLcastArray(values ...any) []any {
	if len(values) == 1 {
		v := values[0]

		if arr, ok := v.([]any); ok {
			return arr
		}

		rv := reflect.ValueOf(v)

		if rv.Kind() == reflect.Ptr {
			if rv.IsNil() {
				return []any{}
			}
			rv = rv.Elem()
		}

		if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
			out := make([]any, rv.Len())
			for i := 0; i < rv.Len(); i++ {
				out[i] = rv.Index(i).Interface()
			}
			return out
		}

		return []any{v}
	}

	return values
}

func OSLequal(a any, b any) bool {
	if a == b {
		return true
	}
	return strings.EqualFold(OSLtoString(a), OSLtoString(b))
}

func OSLnotEqual(a any, b any) bool {
	if a == b {
		return false
	}
	return !strings.EqualFold(OSLtoString(a), OSLtoString(b))
}

func OSLcastInt(i any) int {
	if i == nil {
		return 0
	}
	switch i := i.(type) {
	case string:
		f, _ := strconv.ParseFloat(string(i), 64)
		return int(f)
	case int:
		return i
	case float64:
		return int(i)
	case bool:
		if i {
			return 1
		}
		return 0
	case int8:
		return int(i)
	case int16:
		return int(i)
	case int32:
		return int(i)
	case int64:
		return int(i)
	case json.Number:
		f, _ := i.Float64()
		return int(f)
	default:
		panic("OSLcastInt, invalid type: " + reflect.TypeOf(i).String())
	}
}

func OSLlogValues(values ...any) {
	for _, v := range values {
		OSLlog(v)
	}
}

func OSLlog(v any) {
	if v == nil {
		fmt.Println("null")
	}
	switch v := v.(type) {
	case *SafeMap[string, any]:
		// Convert to regular map for JSON serialization
		keys := v.Keys()
		m := make(map[string]any, len(keys))
		for _, k := range keys {
			val, _ := v.Get(k)
			m[k] = val
		}
		fmt.Println(JsonStringify(m))
		return
	case *SafeSlice[any]:
		// Convert to regular slice for JSON serialization
		fmt.Println(JsonStringify(v.Values()))
		return
	case map[string]any:
		fmt.Println(JsonStringify(v))
		return
	case []any:
		fmt.Println(JsonStringify(v))
		return
	case string, int, float64, bool:
		fmt.Println(v)
		return
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			fmt.Println("null")
			return
		}
		rv = rv.Elem()
	}

	if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
		fmt.Println(JsonStringify(OSLcastArray(v)))
		return
	}

	if rv.Kind() == reflect.Map {
		fmt.Println(JsonStringify(OSLcastObject(v)))
		return
	}

	fmt.Println(v)
}

func OSLisFunc(v any) bool {
	if v == nil {
		return false
	}
	return reflect.TypeOf(v).Kind() == reflect.Func
}

func OSLcallFunc(fn any, self any, params []any) any {
	if fn == nil {
		return nil
	}

	if params == nil {
		params = []any{}
	}

	if self != nil {
		params = append([]any{self}, params...)
	}

	rv := reflect.ValueOf(fn)
	if rv.Kind() != reflect.Func {
		panic("OSLcallFunc: invalid type: " + reflect.TypeOf(fn).String())
	}

	ft := rv.Type()
	numIn := ft.NumIn()

	isVariadic := ft.IsVariadic()

	args := make([]reflect.Value, 0, len(params))

	for i := range params {
		var pt reflect.Type

		if isVariadic && i >= numIn-1 {
			pt = ft.In(numIn - 1).Elem()
		} else {
			pt = ft.In(i)
		}

		var av reflect.Value

		if params[i] == nil {
			switch pt.Kind() {
			case reflect.Interface, reflect.Pointer, reflect.Map,
				reflect.Slice, reflect.Func, reflect.Chan:
				av = reflect.Zero(pt)
			default:
				panic("OSLcallFunc: nil is not assignable to " + pt.String())
			}
		} else {
			av = reflect.ValueOf(params[i])

			at := av.Type()

			if at.AssignableTo(pt) {
			} else if at.ConvertibleTo(pt) {
				av = av.Convert(pt)
			} else if pt.Kind() == reflect.Interface && at.Implements(pt) {
			} else {
				panic(
					"OSLcallFunc: cannot use " + at.String() +
						" as " + pt.String(),
				)
			}
		}

		args = append(args, av)
	}

	out := rv.Call(args)

	switch len(out) {
	case 0:
		return nil
	case 1:
		return out[0].Interface()
	default:
		res := make([]any, len(out))
		for i := range out {
			res[i] = out[i].Interface()
		}
		return res
	}
}

func OSLsort(arr []any) []any {
	if arr == nil {
		return nil
	}

	sort.Slice(arr, func(i, j int) bool {
		return OSLtoString(arr[i]) < OSLtoString(arr[j])
	})
	return arr
}

func OSLreplace(s string, old string, new string) string {
	return strings.ReplaceAll(s, old, new)
}

func OSLreplaceFirst(s string, old string, new string) string {
	return strings.Replace(s, old, new, 1)
}

func OSLsortBy(arr []any, key any) []any {
	if arr == nil {
		return nil
	}

	if OSLisFunc(key) {
		sort.Slice(arr, func(i, j int) bool {
			ki := OSLcallFunc(key, nil, []any{arr[i]})
			kj := OSLcallFunc(key, nil, []any{arr[j]})

			return OSLless(ki, kj)
		})
		return arr
	}

	keyStr := OSLtoString(key)
	sort.Slice(arr, func(i, j int) bool {
		ai, ok1 := arr[i].(map[string]any)
		aj, ok2 := arr[j].(map[string]any)

		if !ok1 || !ok2 {
			return false
		}

		ki := ai[keyStr]
		kj := aj[keyStr]

		return OSLless(ki, kj)
	})

	return arr
}

func OSLless(a any, b any) bool {
	if a == b {
		return false
	}
	return OSLtoString(a) < OSLtoString(b)
}

func OSLgreater(a any, b any) bool {
	if a == b {
		return false
	}
	return OSLtoString(a) > OSLtoString(b)
}

func OSLcastNumber(n any) float64 {
	if n == nil {
		return 0
	}
	switch n := n.(type) {
	case string:
		f, _ := strconv.ParseFloat(string(n), 64)
		return f
	case int:
		return float64(n)
	case int64:
		return float64(n)
	case float64:
		return n
	case bool:
		if n {
			return float64(1)
		}
		return float64(0)
	case json.Number:
		f, _ := n.Float64()
		return f
	default:
		return float64(n.(float64))
	}
}

func OSLcastBool(b any) bool {
	if b == nil {
		return false
	}

	switch b := b.(type) {
	case string:
		return len(b) > 0
	case int:
		return b == 1
	case bool:
		return b
	case []any:
		return len(b) > 0
	case map[string]any:
		return len(b) > 0
	default:
		v := reflect.ValueOf(b)
		if v.Kind() == reflect.Ptr && !v.IsNil() {
			return OSLcastBool(v.Elem().Interface())
		}
		panic("OSLcastBool, invalid type: " + v.Kind().String())
	}
}

func OSLcastUsable(s any) any {
	switch s := s.(type) {
	case string, int, bool, float64, map[string]any:
		return s
	case []any:
		result := make([]any, len(s))
		for i, v := range s {
			result[i] = OSLcastUsable(v)
		}
		return result
	default:
		rv := reflect.ValueOf(s)
		if rv.Kind() == reflect.Slice {
			result := make([]any, rv.Len())
			for i := 0; i < rv.Len(); i++ {
				result[i] = OSLcastUsable(rv.Index(i).Interface())
			}
			return result
		}
		return fmt.Sprintf("%v", s)
	}
}

func OSLrandom[T int | float64](low, high T) T {
	if high <= low {
		return low
	}

	switch any(low).(type) {
	case int:
		return T(OSLrand.Intn(int(high-low)) + int(low))

	case float64:
		return (T(OSLrand.Float64()) * (high - low)) + low
	}

	panic("OSLrandom: unsupported type")
}

func OSLnullishCoaless(a any, b any) any {
	if a == nil {
		return b
	}
	return a
}

func OSLSplit(s string, sep string) []any {
	split := strings.Split(s, sep)
	out := make([]any, len(split))
	for i, v := range split {
		out[i] = v
	}
	return out
}

func JsonStringify(obj any) string {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(obj); err != nil {
		return ""
	}
	return strings.TrimRight(buf.String(), "\n")
}

func JsonParse(str string) any {
	if strings.TrimSpace(str) == "" {
		return interface{}(nil)
	}

	var obj any
	decoder := json.NewDecoder(strings.NewReader(str))
	decoder.UseNumber()
	if err := decoder.Decode(&obj); err != nil {
		return interface{}(nil)
	}
	return obj
}

func JsonFormat(obj any) string {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(obj); err != nil {
		return ""
	}
	return strings.TrimRight(buf.String(), "\n")
}

// Math operation wrappers for OSL behavior

func input(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
}

func OSLgetItem(a any, b any) any {
	if a == nil {
		return nil
	}

	if sm, ok := a.(*SafeMap[string, any]); ok {
		val, _ := sm.Get(OSLtoString(b))
		return val
	}

	if ss, ok := a.(*SafeSlice[any]); ok {
		idx := OSLcastInt(b) - 1 // OSL 1-indexed
		val, ok := ss.Get(idx)
		if !ok {
			return nil
		}
		return val
	}

	if v, ok := a.(map[string]any); ok {
		return v[OSLtoString(b)]
	}

	v := reflect.ValueOf(a)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	key := OSLtoString(b)

	switch v.Kind() {
	case reflect.Map:
		mk := reflect.ValueOf(key)
		val := v.MapIndex(mk)
		if val.IsValid() {
			return val.Interface()
		}
	case reflect.Slice, reflect.Array:
		idx := OSLcastInt(b) - 1 // OSL 1-indexed
		if idx < 0 || idx >= v.Len() {
			return nil
		}
		return v.Index(idx).Interface()
	case reflect.Struct:
		// Try exact field name
		field := v.FieldByName(key)
		if field.IsValid() && field.CanInterface() {
			return field.Interface()
		}
		// Optionally: loop through fields and match lowercase names
		t := v.Type()
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if strings.EqualFold(f.Name, key) && v.Field(i).CanInterface() {
				return v.Field(i).Interface()
			}
		}
	case reflect.String:
		idx := OSLcastInt(b) - 1
		s := v.String()
		if idx < 0 || idx >= len(s) {
			return ""
		}
		return string(s[idx])
	default:
		panic("OSLgetItem: invalid type (" + v.Kind().String() + ")")
	}

	return nil
}

func OSLjoin[T string | []any, T2 string | []any](a T, b T2) T {
	switch aSlice := any(a).(type) {
	case []any:
		switch bVal := any(b).(type) {
		case []any:
			return any(append(aSlice, bVal...)).(T)
		}
	}

	return any(OSLtoString(a) + " " + OSLtoString(b)).(T)
}

func OSLconcat[T string | []any, T2 string | []any](a T, b T2) T {
	switch aSlice := any(a).(type) {
	case []any:
		switch bVal := any(b).(type) {
		case []any:
			return any(append(aSlice, bVal...)).(T)
		}
	}

	return any(OSLtoString(a) + OSLtoString(b)).(T)
}

func OSLadd[T float64 | int](a T, b T) T {
	return T(OSLcastNumber(a) + OSLcastNumber(b))
}

func OSLcompoundAdd(a, b any) any {
	// Handle += for both strings (add space) and numbers
	// Returns the result in the same type as 'a'
	switch a.(type) {
	case string:
		return OSLtoString(a) + " " + OSLtoString(b)
	case float64:
		return OSLcastNumber(a) + OSLcastNumber(b)
	case int:
		return int(OSLcastNumber(a) + OSLcastNumber(b))
	default:
		// For any other type, try numeric addition
		return OSLcastNumber(a) + OSLcastNumber(b)
	}
}

func OSLsub[T float64 | int](a T, b T) T {
	return T(OSLcastNumber(a) - OSLcastNumber(b))
}

func OSLmultiply[AT float64 | int | string, BT float64 | int](a AT, b BT) AT {
	if str, ok := any(a).(string); ok {
		n := OSLcastNumber(b)
		if n < 0 {
			return any("").(AT)
		}
		return any(strings.Repeat(str, int(n))).(AT)
	}

	result := OSLcastNumber(a) * OSLcastNumber(b)

	if _, ok := any(a).(int); ok {
		return any(int(result)).(AT)
	}
	return any(result).(AT)
}

func OSLdivide[T float64 | int](a T, b T) float64 {
	return float64(OSLcastNumber(a) / OSLcastNumber(b))
}

func OSLmod[T float64 | int](a T, b T) T {
	return T(math.Mod(OSLcastNumber(a), OSLcastNumber(b)))
}

func OSLmin[T float64 | int](a T, b T) T {
	if a < b {
		return a
	}
	return b
}

func OSLmax[T float64 | int](a T, b T) T {
	if a > b {
		return a
	}
	return b
}

func OSLround(n any) int {
	if n == nil {
		return 0
	}
	switch n := n.(type) {
	case int:
		return n
	case float64:
		return int(n + 0.5)
	default:
		panic("OSLround, invalid type: " + reflect.TypeOf(n).String())
	}
}

func OSLceil(n any) float64 {
	switch n := n.(type) {
	case int:
		return float64(n)
	case float64:
		return math.Ceil(n)
	default:
		panic("OSLceil, invalid type: " + reflect.TypeOf(n).String())
	}
}

func OSLfloor(n any) float64 {
	switch n := n.(type) {
	case int:
		return float64(n)
	case float64:
		return math.Floor(n)
	default:
		panic("OSLfloor, invalid type: " + reflect.TypeOf(n).String())
	}
}

func OSLtrim[S string | []any, F int | float64, T int | float64](s S, from F, to T) S {
	var items []any
	isArr := false

	if arr, ok := any(s).([]any); ok {
		items = arr
		isArr = true
	} else {
		items = make([]any, 0)
		for _, r := range []rune(OSLtoString(s)) {
			items = append(items, string(r))
		}
	}

	n := len(items)
	start := int(from) - 1
	end := int(to)

	if start < 0 {
		start = 0
	} else if start > n {
		start = n
	}
	if end < 0 {
		end = n + end + 1
	}
	if end > n {
		end = n
	} else if end < 0 {
		end = 0
	}
	if start > end {
		start, end = end, start
	}

	if isArr {
		return any(items[start:end]).(S)
	}
	result := make([]rune, len(items[start:end]))
	for i, v := range items[start:end] {
		result[i] = []rune(v.(string))[0]
	}
	return any(string(result)).(S)
}

func OSLwait(seconds float64) {
	time.Sleep(time.Duration(seconds) * time.Second)
}

func OSLsign(n any) string {
	num := OSLcastNumber(n)
	if num < 0 {
		return "-"
	} else if num > 0 {
		return "+"
	}
	return "+"
}

func OSLpow[T int | float64, F int | float64](base T, exp F) T {
	return T(math.Pow(float64(base), float64(exp)))
}

func OSLxor(a, b int) int {
	return a ^ b
}

func OSLslice(s any, start int, end int) []any {
	arr := OSLcastArray(s)
	n := len(arr)

	start = start - 1
	if start < 0 {
		start = 0
	} else if start > n {
		start = n
	}

	if end < 0 {
		end = n + end + 1
	}
	if end > n {
		end = n
	} else if end < 0 {
		end = 0
	}

	if start > end {
		start, end = end, start
	}

	return arr[start:end]
}

func OSLpadStart(s string, length int, pad string) string {
	if len(s) >= length {
		return s
	}
	return strings.Repeat(pad, length-len(s)) + s
}

func OSLpadEnd(s string, length int, pad string) string {
	if len(s) >= length {
		return s
	}
	return s + strings.Repeat(pad, length-len(s))
}

func OSLtypeof(s any) string {
	switch s.(type) {
	case string:
		return "string"
	case int:
		return "int"
	case float64:
		return "number"
	case bool:
		return "boolean"
	case map[string]any:
		return "object"
	case []any:
		return "array"
	default:
		return "any"
	}
}

func OSLKeyIn(b any, a any) bool {
	if a == nil {
		return false
	}

	key := OSLtoString(b)
	if sm, ok := a.(*SafeMap[string, any]); ok {
		_, exists := sm.Get(key)
		return exists
	}

	switch a := a.(type) {
	case map[string]any:
		_, ok := a[key]
		return ok
	case []any:
		for _, v := range a {
			if OSLtoString(v) == key {
				return true
			}
		}
		return false
	}

	v := reflect.ValueOf(a)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return false
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Map:
		mapKeyType := v.Type().Key()
		mk := reflect.ValueOf(key)
		if !mk.Type().AssignableTo(mapKeyType) {
			if mapKeyType.Kind() == reflect.String {
				mk = reflect.ValueOf(key)
			} else {
				return false
			}
		}
		val := v.MapIndex(mk)
		return val.IsValid()

	case reflect.Slice, reflect.Array:
		idx := OSLcastInt(b) - 1
		return idx >= 0 && idx < v.Len()

	case reflect.String:
		idx := OSLcastInt(b) - 1
		return idx >= 0 && idx < len(v.String())

	case reflect.Struct:
		if field := v.FieldByName(key); field.IsValid() {
			return true
		}
		t := v.Type()
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if strings.EqualFold(f.Name, key) {
				return true
			}
		}
		return false

	default:
		return false
	}
}

func OSLdelete(a any, b any) any {
	if a == nil {
		return nil
	}

	if sm, ok := a.(*SafeMap[string, any]); ok {
		sm.Delete(OSLtoString(b))
		return a
	}

	switch a := a.(type) {
	case map[string]any:
		delete(a, OSLtoString(b))
		return a
	case []any:
		idx := OSLcastInt(b) - 1
		if idx < 0 || idx >= len(a) {
			return a
		}
		return append(a[:idx], a[idx+1:]...)
	}

	v := reflect.ValueOf(a)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	key := OSLtoString(b)

	switch v.Kind() {
	case reflect.Map:
		mk := reflect.ValueOf(key)
		if mk.Type().AssignableTo(v.Type().Key()) {
			v.SetMapIndex(mk, reflect.Value{})
		}
		return v.Interface()

	case reflect.Slice:
		idx := OSLcastInt(b) - 1
		if idx < 0 || idx >= v.Len() {
			return v.Interface()
		}
		newSlice := reflect.AppendSlice(v.Slice(0, idx), v.Slice(idx+1, v.Len()))
		return newSlice.Interface()

	default:
		return a
	}
}

func OSLsetItem(a any, b any, value any) bool {
	if a == nil {
		return false
	}

	if sm, ok := a.(*SafeMap[string, any]); ok {
		sm.Set(OSLtoString(b), value)
		return true
	}

	if ss, ok := a.(*SafeSlice[any]); ok {
		idx := OSLcastInt(b) - 1
		return ss.Set(idx, value)
	}

	v := reflect.ValueOf(a)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return false
		}
		v = v.Elem()
	}

	key := OSLtoString(b)

	switch v.Kind() {
	case reflect.Map:
		mk := reflect.ValueOf(key)
		if !mk.IsValid() {
			return false
		}

		var mv reflect.Value
		if value == nil {
			mv = reflect.Zero(v.Type().Elem())
		} else {
			mv = reflect.ValueOf(value)
		}

		if mk.Type().AssignableTo(v.Type().Key()) && mv.Type().AssignableTo(v.Type().Elem()) {
			v.SetMapIndex(mk, mv)
			return true
		}
		return false

	case reflect.Slice:
		idx := OSLcastInt(b) - 1
		if idx < 0 || idx >= v.Len() {
			return false
		}
		elem := reflect.ValueOf(value)
		if elem.Type().AssignableTo(v.Index(idx).Type()) {
			v.Index(idx).Set(elem)
			return true
		}
		return false

	case reflect.Struct:
		field := v.FieldByName(key)
		if !field.IsValid() {
			return false
		}

		var val reflect.Value
		if value == nil {
			val = reflect.Zero(field.Type())
		} else {
			val = reflect.ValueOf(value)
		}

		return setFieldUnsafe(field, val)
	}

	return false
}

func setFieldUnsafe(field reflect.Value, val reflect.Value) bool {
	if !field.CanAddr() {
		return false
	}

	if !val.Type().AssignableTo(field.Type()) {
		if val.Type().ConvertibleTo(field.Type()) {
			val = val.Convert(field.Type())
		} else {
			return false
		}
	}

	ptr := unsafe.Pointer(field.UnsafeAddr())
	reflect.NewAt(field.Type(), ptr).Elem().Set(val)
	return true
}

func OSLarrayJoin(a any, b any) string {
	var out strings.Builder
	sep := OSLtoString(b)
	arr := OSLcastArray(a)

	for _, v := range arr {
		out.WriteString(OSLtoString(v) + sep)
	}

	return strings.TrimSuffix(out.String(), sep)
}

func OSLgetKeys(a any) []any {
	if sm, ok := a.(*SafeMap[string, any]); ok {
		keys := sm.Keys()
		result := make([]any, len(keys))
		for i, k := range keys {
			result[i] = k
		}
		return result
	}

	if ss, ok := a.(*SafeSlice[any]); ok {
		length := ss.Len()
		keys := make([]any, length)
		for i := 0; i < length; i++ {
			keys[i] = i + 1 // OSL is 1-indexed
		}
		return keys
	}

	switch a := a.(type) {
	case map[string]any:
		keys := make([]any, len(a))
		i := 0
		for k := range a {
			keys[i] = k
			i++
		}
		return keys
	case []any:
		keys := make([]any, len(a))
		for i := range a {
			keys[i] = i + 1 // OSL is 1-indexed
		}
		return keys
	default:
		return []any{}
	}
}

func OSLgetValues(a any) []any {
	if sm, ok := a.(*SafeMap[string, any]); ok {
		values := sm.Values()
		result := make([]any, len(values))
		copy(result, values)
		return result
	}

	switch a := a.(type) {
	case map[string]any:
		values := make([]any, len(a))
		i := 0
		for _, v := range a {
			values[i] = v
			i++
		}
		return values
	case []any:
		values := make([]any, len(a))
		i := 0
		for _, v := range a {
			values[i] = v
			i++
		}
		return values
	default:
		return []any{}
	}
}

func OSLcontains(a any, b any) bool {
	if sm, ok := a.(*SafeMap[string, any]); ok {
		_, exists := sm.Get(OSLtoString(b))
		return exists
	}

	if ss, ok := a.(*SafeSlice[any]); ok {
		// For arrays, check if value exists
		values := ss.Values()
		for _, v := range values {
			if OSLtoString(v) == OSLtoString(b) {
				return true
			}
		}
		return false
	}

	switch a := a.(type) {
	case map[string]any:
		_, ok := a[OSLtoString(b)]
		return ok
	case []any:
		for _, v := range a {
			if OSLtoString(v) == OSLtoString(b) {
				return true
			}
		}
		return false
	case string:
		return strings.Contains(a, OSLtoString(b))
	default:
		return false
	}
}

func OSLappend(a *[]any, b any) []any {
	*a = append(*a, b)
	return *a
}

func OSLpop(a *[]any) any {
	if len(*a) == 0 {
		return nil
	}
	last := (*a)[len(*a)-1]
	*a = (*a)[:len(*a)-1]
	return last
}

func OSLshift(a *[]any) any {
	if len(*a) == 0 {
		return nil
	}
	first := (*a)[0]
	*a = append([]any{}, (*a)[1:]...)
	return first
}

func OSLprepend(a *[]any, b any) []any {
	*a = append([]any{b}, *a...)
	return *a
}

func OSLclone(a any) any {
	switch a := a.(type) {
	case map[string]any:
		b := make(map[string]any, len(a))
		for k, v := range a {
			b[k] = OSLclone(v)
		}
		return b
	case []any:
		b := make([]any, len(a))
		for i, v := range a {
			b[i] = OSLclone(v)
		}
		return b
	default:
		return a
	}
}

// worker handling

var OSLself any = nil

func OSLworker(props map[string]any) map[string]any {
	props["createdTime"] = time.Now()
	props["processTime"] = 0
	props["alive"] = true
	props["kill"] = func() {
		props["alive"] = false
	}
	go (func() {
		OSLself = props
		OSLcallFunc(props["oncreate"], props, nil)
		for {
			startTime := time.Now()
			OSLself = props
			OSLcallFunc(props["onframe"], props, nil)
			props["processTime"] = OSLcastNumber(props["processTime"]) + time.Since(startTime).Seconds()
			if props["alive"] != true {
				props["alive"] = false
				OSLself = props
				OSLcallFunc(props["onkill"], props, nil)
				break
			}
		}
	})()
	return props
}

type SafeMap[K comparable, V any] struct {
	mu   sync.RWMutex
	data map[K]V
}

func NewSafeMap[K comparable, V any](defaults map[K]V) *SafeMap[K, V] {
	sm := &SafeMap[K, V]{
		data: make(map[K]V, len(defaults)),
	}
	for k, v := range defaults {
		sm.data[k] = v
	}
	return sm
}

func (m *SafeMap[K, V]) Set(key K, value V) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value // regular map syntax here
}

func (m *SafeMap[K, V]) Get(key K) (V, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, ok := m.data[key] // regular map syntax here
	return value, ok
}

func (m *SafeMap[K, V]) Delete(key K) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
}

func (m *SafeMap[K, V]) Keys() []K {
	m.mu.RLock()
	defer m.mu.RUnlock()
	keys := make([]K, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys
}

func (m *SafeMap[K, V]) Values() []V {
	m.mu.RLock()
	defer m.mu.RUnlock()
	values := make([]V, 0, len(m.data))
	for _, v := range m.data {
		values = append(values, v)
	}
	return values
}

// SafeSlice is a thread-safe slice for global arrays
type SafeSlice[V any] struct {
	mu   sync.RWMutex
	data []V
}

func NewSafeSlice[V any](defaults []V) *SafeSlice[V] {
	ss := &SafeSlice[V]{
		data: make([]V, len(defaults)),
	}
	copy(ss.data, defaults)
	return ss
}

func (s *SafeSlice[V]) Append(value V) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = append(s.data, value)
}

func (s *SafeSlice[V]) Get(index int) (V, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if index < 0 || index >= len(s.data) {
		var zero V
		return zero, false
	}
	return s.data[index], true
}

func (s *SafeSlice[V]) Set(index int, value V) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if index < 0 || index >= len(s.data) {
		return false
	}
	s.data[index] = value
	return true
}

func (s *SafeSlice[V]) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.data)
}

func (s *SafeSlice[V]) Values() []V {
	s.mu.RLock()
	defer s.mu.RUnlock()
	values := make([]V, len(s.data))
	copy(values, s.data)
	return values
}

// Keyboard methods (stub implementations)
// Note: These are defined as methods on a custom string type
type OSLString string

func (s OSLString) onKeyDown() bool {
	return false
}

func (s OSLString) isKeyDown() bool {
	return false
}

func (s OSLString) toNum() float64 {
	return OSLcastNumber(string(s))
}

func atob(encoded string) string {
	data, err := OSLio.ReadAll(base64.NewDecoder(base64.StdEncoding, strings.NewReader(encoded)))
	if err != nil {
		return ""
	}
	return string(data)
}

func btoa(data string) string {
	encoded := base64.StdEncoding.EncodeToString([]byte(data))
	return encoded
}
