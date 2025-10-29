// This is a set of funtions that are used in the compiler for OSL.go

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
	case []io.Reader:
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

func OSLcastString(s any) string {
	switch s := s.(type) {
	case string:
		return s
	case []byte:
		return string(s)
	case []any:
		return JsonStringify(s)
	case map[string]any:
		return JsonStringify(s)
	case map[string]string:
		return JsonStringify(s)
	case map[string]int:
		return JsonStringify(s)
	case map[string]float64:
		return JsonStringify(s)
	case map[string]bool:
		return JsonStringify(s)
	case io.Reader:
		data, err := io.ReadAll(s)
		if err != nil {
			panic("OSLcastString: failed to read io.Reader:" + err.Error())
			return ""
		}
		return string(data)
	default:
		return fmt.Sprintf("%v", s)
	}
}

func OSLcastObject(s any) map[string]any {
	if s == nil {
		return map[string]any{}
	}
	switch s := s.(type) {
	case map[string]any:
		return s
	default:
		panic("OSLcastObject: invalid type, " + reflect.TypeOf(s).String())
	}
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
	return strings.EqualFold(OSLcastString(a), OSLcastString(b))
}

func OSLnotEqual(a any, b any) bool {
	if a == b {
		return false
	}
	return !strings.EqualFold(OSLcastString(a), OSLcastString(b))
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

func OSLrandom(low any, high any) int {
	highInt := OSLcastInt(high)
	lowInt := OSLcastInt(low)
	return OSLcastInt(rand.Intn(int(highInt-lowInt+1))) + lowInt
}

func OSLnullishCoaless(a any, b any) any {
	if a == nil {
		return b
	}
	return a
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

// Helper function to convert bool to float64
func boolToFloat64(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}

// Helper function to convert bool to int
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func boolToStr(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

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

	v := reflect.ValueOf(a)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	key := OSLcastString(b)

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

func OSLjoin(a any, b any) string {
	a = OSLcastString(a)
	b = OSLcastString(b)
	return OSLcastString(a) + OSLcastString(b)
}

// OSLmultiply handles the * operation: multiplies numbers, repeats strings
func OSLmultiply(a any, b any) float64 {
	return OSLcastNumber(a) * OSLcastNumber(b)
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

func OSLtrim(s any, from int, to int) string {
	str := []rune(OSLcastString(s))

	start := from - 1
	end := to

	if start < 0 {
		start = 0
	}
	if end < 0 {
		end = len(str) + end + 1
	}

	if start > len(str) {
		start = len(str)
	}
	if end > len(str) {
		end = len(str)
	}

	if start > end {
		start, end = end, start
	}

	return string(str[start:end])
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

	key := OSLcastString(b)
	switch a := a.(type) {
	case map[string]any:
		_, ok := a[key]
		return ok
	case []any:
		for _, v := range a {
			if OSLcastString(v) == key {
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

	switch a := a.(type) {
	case map[string]any:
		delete(a, OSLcastString(b))
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

	key := OSLcastString(b)

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

	v := reflect.ValueOf(a)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return false
		}
		v = v.Elem()
	}

	key := OSLcastString(b)

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
		if field.IsValid() && field.CanSet() {
			val := reflect.ValueOf(value)
			if val.Type().AssignableTo(field.Type()) {
				field.Set(val)
				return true
			}
		}
		return false
	}

	return false
}

func OSLarrayJoin(a any, b any) string {
	var out string
	sep := OSLcastString(b)
	arr := OSLcastArray(a)

	for _, v := range arr {
		out += OSLcastString(v) + sep
	}

	return strings.TrimSuffix(out, sep)
}

func OSLgetKeys(a any) []any {
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
		i := 0
		for _, v := range a {
			keys[i] = OSLcastString(v)
			i++
		}
		return keys
	default:
		return []any{}
	}
}

func OSLgetValues(a any) []any {
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
	switch a := a.(type) {
	case map[string]any:
		_, ok := a[OSLcastString(b)]
		return ok
	case []any:
		for _, v := range a {
			if OSLcastString(v) == OSLcastString(b) {
				return true
			}
		}
		return false
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
