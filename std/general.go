// This is a set of funtions that are used in the compiler for OSL.go

func OSLlen(s any) int {
	switch s := s.(type) {
	case string:
		return len(s)
	case []any:
		return len(s)
	}
	return 0
}

func OSLcastString(s any) string {
	switch s := s.(type) {
	case string:
		return s
	default:
		return fmt.Sprintf("%v", s)
	}
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
	switch i := i.(type) {
	case string:
		f, _ := strconv.ParseFloat(string(i), 64)
		return int(f)
	default:
		if i == true || i == false {
			if i == true {
				return 1
			}
			return 0
		}
		return int(i.(int))
	}
}

func OSLcastFloat(n any) float64 {
	switch n := n.(type) {
	case string:
		f, _ := strconv.ParseFloat(string(n), 64)
		return f
	default:
		if n == true || n == false {
			if n == true {
				return float64(1)
			}
			return float64(0)
		}
		return float64(n.(float64))
	}
}

func OSLcastBool(b any) bool {
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
		return b.(bool)
	}
}

// converts go types into types that osl can deal with
func OSLcastUsable(s any) any {
	switch s := s.(type) {
	case string, int, bool, float64, map[string]any, []any:
		return s
	default:
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
	enc.Encode(obj)
	return strings.TrimSpace(buf.String())
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
	a = OSLcastUsable(a)
	b = OSLcastUsable(b)
	switch a := a.(type) {
	case map[string]any:
		return a[OSLcastString(b)]
	case []any:
		idx := OSLcastInt(b) - 1
		if idx < 0 || idx >= len(a) {
			return nil
		}
		return a[idx]
	case string:
		idx := OSLcastInt(b) - 1
		if idx < 0 || idx >= len(a) {
			return nil
		}
		return string(a[idx])
	case []byte:
		idx := OSLcastInt(b) - 1
		if idx < 0 || idx >= len(a) {
			return nil
		}
		return a[idx]
	default:
		return nil
	}
}

func OSLjoin(a any, b any) string {
	a = OSLcastString(a)
	b = OSLcastString(b)
	return OSLcastString(a) + OSLcastString(b)
}

// OSLconcat handles ++ operation: concatenates all types without space
func OSLconcat(a any, b any) any {
	if ab, ok := a.(bool); ok {
		a = OSLcastInt(ab)
	}
	if bb, ok := b.(bool); ok {
		b = OSLcastInt(bb)
	}
	switch va := a.(type) {
	case string:
		switch vb := b.(type) {
		case string:
			return OSLcastString(va) + OSLcastString(vb)
		default:
			return OSLcastString(va) + OSLcastString(b.(string))
		}
	case int:
		switch vb := b.(type) {
		case string:
			return OSLcastString(va) + OSLcastString(vb)
		case int:
			return OSLcastInt(va) + OSLcastInt(vb)
		default:
			return OSLcastInt(va) + OSLcastInt(b.(int))
		}
	case []any:
		switch vb := b.(type) {
		case []any:
			return append(va, vb...)
		default:
			return append(va, b)
		}
	case map[string]any:
		switch vb := b.(type) {
		case string:
			return "[Object object]" + string(vb)
		case []any:
			// merge the array numerical keys into the object
			for i, v := range vb {
				va[fmt.Sprintf("%v", i)] = v
			}
			return va
		case map[string]any:
			// merge the object keys into the object
			for k, v := range vb {
				va[k] = v
			}
			return va
		default:
			return "[Object object]" + fmt.Sprintf("%v", b)
		}
	default:
		// Convert to string and concatenate
		return fmt.Sprintf("%v", a) + fmt.Sprintf("%v", b)
	}
}

// OSLmultiply handles the * operation: multiplies numbers, repeats strings
func OSLmultiply(a any, b any) any {
	switch va := a.(type) {
	case string:
		switch vb := b.(type) {
		case int:
			count := int(vb)
			if count <= 0 {
				return ""
			}
			return strings.Repeat(string(va), count)
		}
	}
	return float64(a.(float64)) * float64(b.(float64))
}