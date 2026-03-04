// name: csv
// description: CSV parsing and generation utilities
// author: roturbot
// requires: encoding/csv, strings, os

type CSV struct{}

func (CSV) parse(data any) []map[string]any {
	dataStr := OSLtoString(data)

	reader := csv.NewReader(strings.NewReader(dataStr))
	records, err := reader.ReadAll()
	if err != nil {
		return []map[string]any{}
	}

	if len(records) == 0 {
		return []map[string]any{}
	}

	headers := records[0]
	result := make([]map[string]any, 0, len(records)-1)

	for i := 1; i < len(records); i++ {
		row := make(map[string]any)

		for j, header := range headers {
			if j < len(records[i]) {
				row[header] = records[i][j]
			} else {
				row[header] = ""
			}
		}

		result = append(result, row)
	}

	return result
}

func (CSV) parseRaw(data any) [][]any {
	dataStr := OSLtoString(data)

	reader := csv.NewReader(strings.NewReader(dataStr))
	records, err := reader.ReadAll()
	if err != nil {
		return [][]any{}
	}

	result := make([][]any, len(records))
	for i, record := range records {
		row := make([]any, len(record))
		for j, field := range record {
			row[j] = field
		}
		result[i] = row
	}

	return result
}

func (CSV) stringify(data map[string]any) string {
	var builder strings.Builder

	writer := csv.NewWriter(&builder)
	headers := make([]string, 0, len(data))
	values := make([]string, 0, len(data))

	for k, v := range data {
		headers = append(headers, k)
		values = append(values, OSLtoString(v))
	}

	writer.Write(headers)
	writer.Write(values)
	writer.Flush()

	return builder.String()
}

func (CSV) stringifyRows(data []map[string]any) string {
	if len(data) == 0 {
		return ""
	}

	var builder strings.Builder
	writer := csv.NewWriter(&builder)

	firstRow := data[0]
	headers := make([]string, 0, len(firstRow))

	for k := range firstRow {
		headers = append(headers, k)
	}

	writer.Write(headers)

	for _, row := range data {
		values := make([]string, len(headers))
		for i, header := range headers {
			values[i] = OSLtoString(row[header])
		}
		writer.Write(values)
	}

	writer.Flush()

	return builder.String()
}

func (CSV) stringifyArray(data [][]any) string {
	var builder strings.Builder
	writer := csv.NewWriter(&builder)

	for _, row := range data {
		values := make([]string, len(row))
		for i, field := range row {
			values[i] = OSLtoString(field)
		}
		writer.Write(values)
	}

	writer.Flush()

	return builder.String()
}

func (CSV) readFile(path any) []map[string]any {
	pathStr := OSLtoString(path)
	data := fs.ReadFile(path)

	if data == "" {
		return []map[string]any{}
	}

	return csv.parse(data)
}

func (CSV) readFileRaw(path any) [][]any {
	pathStr := OSLtoString(path)
	data := fs.ReadFile(path)

	if data == "" {
		return [][]any{}
	}

	return csv.parseRaw(data)
}

func (CSV) writeFile(path any, data map[string]any) bool {
	pathStr := OSLtoString(path)
	csvData := csv.stringify(data)

	return fs.WriteFile(path, csvData)
}

func (CSV) writeFileRows(path any, data []map[string]any) bool {
	pathStr := OSLtoString(path)
	csvData := csv.stringifyRows(data)

	return fs.WriteFile(path, csvData)
}

func (CSV) writeFileArray(path any, data [][]any) bool {
	pathStr := OSLtoString(path)
	csvData := csv.stringifyArray(data)

	return fs.WriteFile(path, csvData)
}

func (CSV) toRows(data []map[string]any) [][]any {
	if len(data) == 0 {
		return [][]any{}
	}

	result := make([][]any, len(data)+1)

	headers := make([]any, 0)
	for k := range data[0] {
		headers = append(headers, k)
	}
	result[0] = headers

	for i, row := range data {
		rowArr := make([]any, len(headers))
		for j, header := range headers {
			rowArr[j] = row[OSLtoString(header)]
		}
		result[i+1] = rowArr
	}

	return result
}

func (CSV) fromRows(rows [][]any) []map[string]any {
	if len(rows) == 0 {
		return []map[string]any{}
	}

	headers := rows[0]
	result := make([]map[string]any, len(rows)-1)

	for i := 1; i < len(rows); i++ {
		row := make(map[string]any)

		for j, header := range headers {
			if j < len(rows[i]) {
				row[OSLtoString(header)] = rows[i][j]
			} else {
				row[OSLtoString(header)] = ""
			}
		}

		result[i-1] = row
	}

	return result
}

func (CSV) getColumn(data []map[string]any, column any) []any {
	columnStr := OSLtoString(column)
	result := make([]any, len(data))

	for i, row := range data {
		result[i] = row[columnStr]
	}

	return result
}

func (CSV) filter(data []map[string]any, fn any) []map[string]any {
	result := make([]map[string]any, 0, len(data))

	for _, row := range data {
		shouldInclude := OSLcastBool(OSLcallFunc(fn, nil, []any{row}))
		if shouldInclude {
			result = append(result, row)
		}
	}

	return result
}

func (CSV) map(data []map[string]any, fn any) []map[string]any {
	result := make([]map[string]any, len(data))

	for i, row := range data {
		newRow := OSLcallFunc(fn, nil, []any{row})
		if newRowMap, ok := newRow.(map[string]any); ok {
			result[i] = newRowMap
		}
	}

	return result
}

func (CSV) reduce(data []map[string]any, initial any, fn any) any {
	result := initial

	for _, row := range data {
		result = OSLcallFunc(fn, nil, []any{result, row})
	}

	return result
}

func (CSV) groupBy(data []map[string]any, key any) map[string][]map[string]any {
	keyStr := OSLtoString(key)
	result := make(map[string][]map[string]any)

	for _, row := range data {
		value := OSLtoString(row[keyStr])
		result[value] = append(result[value], row)
	}

	return result
}

func (CSV) sortBy(data []map[string]any, key any) []map[string]any {
	keyStr := OSLtoString(key)

	result := make([]map[string]any, len(data))
	copy(result, data)

	sort.Slice(result, func(i, j int) bool {
		valI := OSLtoString(result[i][keyStr])
		valJ := OSLtoString(result[j][keyStr])
		return valI < valJ
	})

	return result
}

func (CSV) sortByNum(data []map[string]any, key any) []map[string]any {
	keyStr := OSLtoString(key)

	result := make([]map[string]any, len(data))
	copy(result, data)

	sort.Slice(result, func(i, j int) bool {
		valI := OSLcastNumber(result[i][keyStr])
		valJ := OSLcastNumber(result[j][keyStr])
		return valI < valJ
	})

	return result
}

func (CSV) aggregate(data []map[string]any, key any, value any, fn any) map[string]any {
	keyStr := OSLtoString(key)
	valueStr := OSLtoString(value)

	grouped := csv.groupBy(data, keyStr)
	result := make(map[string]any)

	for k, rows := range grouped {
		values := make([]any, len(rows))
		for i, row := range rows {
			values[i] = row[valueStr]
		}

		aggregated := OSLcallFunc(fn, nil, []any{values})
		result[k] = aggregated
	}

	return result
}

func (CSV) count(data []map[string]any, key any) map[string]int {
	keyStr := OSLtoString(key)
	grouped := csv.groupBy(data, keyStr)
	result := make(map[string]int)

	for k, rows := range grouped {
		result[k] = len(rows)
	}

	return result
}

func (CSV) unique(data []map[string]any, key any) []any {
	keyStr := OSLtoString(key)
	seen := make(map[string]bool)
	result := make([]any, 0)

	for _, row := range data {
		value := OSLtoString(row[keyStr])
		if !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}

	return result
}

func (CSV) join(data1 []map[string]any, data2 []map[string]any, key any) []map[string]any {
	keyStr := OSLtoString(key)
	result := make([]map[string]any, 0)

	for _, row1 := range data1 {
		joinKey := OSLtoString(row1[keyStr])

		for _, row2 := range data2 {
			if OSLtoString(row2[keyStr]) == joinKey {
				merged := make(map[string]any)
				for k, v := range row1 {
					merged[k] = v
				}
				for k, v := range row2 {
					merged[k+"_2"] = v
				}
				result = append(result, merged)
				break
			}
		}
	}

	return result
}

func (CSV) pivot(data []map[string]any, keyColumn any, valueColumn any, pivotColumn any) map[string]map[string]any {
	keyCol := OSLtoString(keyColumn)
	valCol := OSLtoString(valueColumn)
	pivotCol := OSLtoString(pivotColumn)

	pivotValues := csv.unique(data, pivotCol)
	result := make(map[string]map[string]any)

	for _, row := range data {
		key := OSLtoString(row[keyCol])
		pivot := OSLtoString(row[pivotCol])

		if _, ok := result[key]; !ok {
			result[key] = make(map[string]any)
		}

		result[key][pivot] = row[valCol]
	}

	return result
}

func (CSV) transpose(data [][]any) [][]any {
	if len(data) == 0 {
		return [][]any{}
	}

	rows := len(data)
	cols := len(data[0])

	result := make([][]any, cols)
	for i := range result {
		result[i] = make([]any, rows)
	}

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			result[j][i] = data[i][j]
		}
	}

	return result
}

func (CSV) stats(data []map[string]any, key any) map[string]any {
	keyStr := OSLtoString(key)
	values := csv.getColumn(data, keyStr)

	if len(values) == 0 {
		return map[string]any{}
	}

	numValues := make([]float64, 0)
	for _, v := range values {
		numValues = append(numValues, OSLcastNumber(v))
	}

	if len(numValues) == 0 {
		return map[string]any{
			"count": len(values),
		}
	}

	return map[string]any{
		"count":   len(values),
		"min":     math.min(numValues...),
		"max":     math.max(numValues...),
		"average": math.avg(numValues),
		"sum":     math.sum(numValues),
	}
}

func (CSV) merge(data1 []map[string]any, data2 []map[string]any) []map[string]any {
	result := append(data1, data2)
	return result
}

func (CSV) chunk(data []map[string]any, size any) [][]map[string]any {
	sizeInt := int(OSLcastNumber(size))
	if sizeInt <= 0 {
		sizeInt = 10
	}

	result := make([][]map[string]any, 0, (len(data)+sizeInt-1)/sizeInt)

	for i := 0; i < len(data); i += sizeInt {
		end := i + sizeInt
		if end > len(data) {
			end = len(data)
		}
		result = append(result, data[i:end])
	}

	return result
}

func (CSV) flatten(data []map[string]any, separator any) []map[string]any {
	sep := OSLtoString(separator)
	if sep == "" {
		sep = "_"
	}

	result := make([]map[string]any, len(data))

	for i, row := range data {
		flat := make(map[string]any)

		for k, v := range row {
			if nested, ok := v.(map[string]any); ok {
				for nk, nv := range nested {
					flat[k+sep+nk] = nv
				}
			} else {
				flat[k] = v
			}
		}

		result[i] = flat
	}

	return result
}

func (CSV) sample(data []map[string]any, count any) []map[string]any {
	countInt := int(OSLcastNumber(count))

	if countInt >= len(data) {
		result := make([]map[string]any, len(data))
		copy(result, data)
		return result
	}

	result := make([]map[string]any, countInt)
	indices := make(map[int]bool)

	for len(result) < countInt {
		randIndex := OSLrandomInt(0, len(data)-1)
		if !indices[randIndex] {
			indices[randIndex] = true
			result[len(result)] = data[randIndex]
		}
	}

	return result
}

func (CSV) appendRow(data []map[string]any, row map[string]any) []map[string]any {
	result := append(data, row)
	return result
}

func (CSV) prependRow(data []map[string]any, row map[string]any) []map[string]any {
	result := append([]map[string]any{row}, data...)
	return result
}

func (CSV) insertRow(data []map[string]any, index any, row map[string]any) []map[string]any {
	indexInt := int(OSLcastNumber(index))

	if indexInt >= len(data) {
		return csv.appendRow(data, row)
	}
	if indexInt <= 0 {
		return csv.prependRow(data, row)
	}

	result := make([]map[string]any, len(data)+1)
	copy(result[:indexInt], data[:indexInt])
	result[indexInt] = row
	copy(result[indexInt+1:], data[indexInt:])

	return result
}

func (CSV) deleteRow(data []map[string]any, index any) []map[string]any {
	indexInt := int(OSLcastNumber(index))

	if indexInt < 0 || indexInt >= len(data) {
		return data
	}

	result := make([]map[string]any, len(data)-1)
	copy(result[:indexInt], data[:indexInt])
	copy(result[indexInt:], data[indexInt+1:])

	return result
}

func (CSV) addColumn(data []map[string]any, column any, defaultValue any) []map[string]any {
	columnStr := OSLtoString(column)
	defaultValueVal := defaultValue

	result := make([]map[string]any, len(data))

	for i, row := range data {
		newRow := make(map[string]any)
		for k, v := range row {
			newRow[k] = v
		}
		newRow[columnStr] = defaultValueVal
		result[i] = newRow
	}

	return result
}

func (CSV) removeColumn(data []map[string]any, column any) []map[string]any {
	columnStr := OSLtoString(column)

	result := make([]map[string]any, len(data))

	for i, row := range data {
		newRow := make(map[string]any)
		for k, v := range row {
			if k != columnStr {
				newRow[k] = v
			}
		}
		result[i] = newRow
	}

	return result
}

func (CSV) renameColumn(data []map[string]any, oldColumn any, newColumn any) []map[string]any {
	oldCol := OSLtoString(oldColumn)
	newCol := OSLtoString(newColumn)

	result := make([]map[string]any, len(data))

	for i, row := range data {
		newRow := make(map[string]any)
		for k, v := range row {
			if k == oldCol {
				newRow[newCol] = v
			} else {
				newRow[k] = v
			}
		}
		result[i] = newRow
	}

	return result
}

var csv = CSV{}
