// name: yaml
// description: YAML parsing and encoding utilities
// author: roturbot
// requires: gopkg.in/yaml.v3, encoding/json

type YAML struct{}

func (YAML) Parse(source any) any {
	sourceStr := OSLtoString(source)
	var result map[string]any

	err := yaml.Unmarshal([]byte(sourceStr), &result)
	if err != nil {
		sliceResult := make([]any, 0)
		err2 := yaml.Unmarshal([]byte(sourceStr), &sliceResult)
		if err2 == nil {
			return sliceResult
		}
		return map[string]any{"error": err.Error()}
	}

	return result
}

func (YAML) Stringify(data any) string {
	result, err := yaml.Marshal(data)
	if err != nil {
		return ""
	}
	return string(result)
}

func (y YAML) ToJSON(yamlData any) string {
	parsed := y.Parse(yamlData)
	jsonBytes, err := json.Marshal(parsed)
	if err != nil {
		return ""
	}
	return string(jsonBytes)
}

func (YAML) FromJSON(jsonData any) any {
	jsonStr := OSLtoString(jsonData)
	var result map[string]any

	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return nil
	}

	return result
}

func (YAML) Get(data any, key any) any {
	if dataMap, ok := data.(map[string]any); ok {
		keyStr := OSLtoString(key)
		return dataMap[keyStr]
	}
	return nil
}

func (YAML) Set(data map[string]any, key any, value any) map[string]any {
	keyStr := OSLtoString(key)
	data[keyStr] = value
	return data
}

func (y YAML) Merge(data1 map[string]any, data2 map[string]any) map[string]any {
	result := make(map[string]any)

	for k, v := range data1 {
		result[k] = v
	}

	for k, v := range data2 {
		if existing, ok := result[k]; ok {
			if existingMap, ok := existing.(map[string]any); ok {
				if newMap, ok := v.(map[string]any); ok {
					result[k] = y.Merge(existingMap, newMap)
					continue
				}
			}
		}
		result[k] = v
	}

	return result
}

func (YAML) Keys(data map[string]any) []any {
	keys := make([]any, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	return keys
}

func (YAML) Values(data map[string]any) []any {
	values := make([]any, 0, len(data))
	for _, v := range data {
		values = append(values, v)
	}
	return values
}

func (YAML) Has(data map[string]any, key any) bool {
	keyStr := OSLtoString(key)
	_, ok := data[keyStr]
	return ok
}

func (YAML) Delete(data map[string]any, key any) map[string]any {
	keyStr := OSLtoString(key)
	delete(data, keyStr)
	return data
}

var yaml = YAML{}
