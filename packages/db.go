// name: db
// description: SQLite database utilities
// author: roturbot
// requires: database/sql, github.com/mattn/go-sqlite3

type DB struct {
	conn *sql.DB
}

func sanitizeCol(col string) string {
	sanitized := ""
	for _, r := range col {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			sanitized += string(r)
		}
	}
	return sanitized
}

type DBRow struct {
	columns []string
	values  []any
}

func (DB) open(path any) *DB {
	pathStr := OSLtoString(path)

	conn, err := sql.Open("sqlite3", pathStr)
	if err != nil {
		return nil
	}

	err = conn.Ping()
	if err != nil {
		return nil
	}

	return &DB{conn: conn}
}

func (DB) openMemory() *DB {
	conn, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil
	}

	err = conn.Ping()
	if err != nil {
		return nil
	}

	return &DB{conn: conn}
}

func (d *DB) close() error {
	if d == nil || d.conn == nil {
		return nil
	}
	return d.conn.Close()
}

func (d *DB) exec(query any, args ...any) bool {
	if d == nil || d.conn == nil {
		return false
	}

	queryStr := OSLtoString(query)
	argsSlice := make([]any, len(args))
	for i, arg := range args {
		argsSlice[i] = arg
	}

	_, err := d.conn.Exec(queryStr, argsSlice...)
	return err == nil
}

func (d *DB) query(query any, args ...any) []DBRow {
	if d == nil || d.conn == nil {
		return []DBRow{}
	}

	queryStr := OSLtoString(query)
	argsSlice := make([]any, len(args))
	for i, arg := range args {
		argsSlice[i] = arg
	}

	rows, err := d.conn.Query(queryStr, argsSlice...)
	if err != nil {
		return []DBRow{}
	}
	defer rows.Close()

	var result []DBRow

	for rows.Next() {
		columns, _ := rows.Columns()
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))

		for i := range values {
			valuePtrs[i] = &values[i]
		}

		rows.Scan(valuePtrs...)

		row := DBRow{
			columns: columns,
			values:  values,
		}

		result = append(result, row)
	}

	return result
}

func (d *DB) queryOne(query any, args ...any) DBRow {
	rows := d.query(query, args...)

	if len(rows) == 0 {
		return DBRow{}
	}

	return rows[0]
}

func (d *DB) queryMap(query any, args ...any) []map[string]any {
	rows := d.query(query, args...)
	result := make([]map[string]any, len(rows))

	for i, row := range rows {
		result[i] = row.toMap()
	}

	return result
}

func (d *DB) queryMapOne(query any, args ...any) map[string]any {
	row := d.queryOne(query, args...)
	return row.toMap()
}

func (d *DB) insert(table any, data map[string]any) int64 {
	if d == nil || d.conn == nil {
		return 0
	}

	tableStr := OSLtoString(table)

	sanitizedTable := sanitizeCol(tableStr)

	columns := make([]string, 0, len(data))
	placeholders := make([]string, 0, len(data))
	values := make([]any, 0, len(data))

	for col, val := range data {
		sanitizedCol := sanitizeCol(col)
		columns = append(columns, sanitizedCol)
		placeholders = append(placeholders, "?")
		values = append(values, val)
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", sanitizedTable, strings.Join(columns, ", "), strings.Join(placeholders, ", "))

	result, err := d.conn.Exec(query, values...)
	if err != nil {
		return 0
	}

	id, _ := result.LastInsertId()
	return id
}

func (d *DB) update(table any, data map[string]any, where any, whereArgs ...any) bool {
	if d == nil || d.conn == nil {
		return false
	}

	tableStr := OSLtoString(table)
	whereStr := OSLtoString(where)

	sanitizedTable := sanitizeCol(tableStr)

	setParts := make([]string, 0, len(data))
	values := make([]any, 0, len(data)+1+len(whereArgs))

	for col, val := range data {
		sanitizedCol := sanitizeCol(col)
		setParts = append(setParts, fmt.Sprintf("%s = ?", sanitizedCol))
		values = append(values, val)
	}

	for _, arg := range whereArgs {
		values = append(values, arg)
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s", sanitizedTable, strings.Join(setParts, ", "), whereStr)

	_, err := d.conn.Exec(query, values...)
	return err == nil
}

func (d *DB) delete(table any, where any, whereArgs ...any) bool {
	if d == nil || d.conn == nil {
		return false
	}

	tableStr := OSLtoString(table)
	whereStr := OSLtoString(where)

	sanitizedTable := sanitizeCol(tableStr)

	query := fmt.Sprintf("DELETE FROM %s WHERE %s", sanitizedTable, whereStr)
	args := make([]any, len(whereArgs))
	for i, arg := range whereArgs {
		args[i] = arg
	}

	_, err := d.conn.Exec(query, args...)
	return err == nil
}

func (d *DB) count(table any, where any, whereArgs ...any) int {
	if d == nil || d.conn == nil {
		return 0
	}

	tableStr := OSLtoString(table)
	whereStr := OSLtoString(where)

	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableStr)

	if whereStr != "" {
		query += " WHERE " + whereStr
	}

	row := d.queryOne(query, whereArgs...)
	if row.isEmpty() {
		return 0
	}

	return int(OSLcastNumber(row.get(1)))
}

func (d *DB) exists(table any, where any, whereArgs ...any) bool {
	return d.count(table, where, whereArgs...) > 0
}

func (d *DB) createTable(table any, columns map[string]string) bool {
	if d == nil || d.conn == nil {
		return false
	}

	tableStr := OSLtoString(table)
	sanitizedTable := sanitizeCol(tableStr)

	colDefs := make([]string, 0, len(columns))

	for col, colType := range columns {
		sanitizedCol := sanitizeCol(col)
		colDefs = append(colDefs, fmt.Sprintf("%s %s", sanitizedCol, colType))
	}

	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", sanitizedTable, strings.Join(colDefs, ", "))

	_, err := d.conn.Exec(query)
	return err == nil
}

func (d *DB) dropTable(table any) bool {
	if d == nil || d.conn == nil {
		return false
	}

	tableStr := OSLtoString(table)
	sanitizedTable := sanitizeCol(tableStr)

	_, err := d.conn.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", sanitizedTable))
	return err == nil
}

func (d *DB) getTables() []any {
	if d == nil || d.conn == nil {
		return []any{}
	}

	query := "SELECT name FROM sqlite_master WHERE type='table' ORDER BY name"
	rows := d.query(query)

	result := make([]any, len(rows))
	for i, row := range rows {
		result[i] = row.get("name")
	}

	return result
}

func (d *DB) getColumns(table any) []any {
	if d == nil || d.conn == nil {
		return []any{}
	}

	tableStr := OSLtoString(table)
	query := fmt.Sprintf("PRAGMA table_info(%s)", tableStr)
	rows := d.query(query)

	result := make([]any, len(rows))
	for i, row := range rows {
		result[i] = row.get("name")
	}

	return result
}

func (d *DB) begin() bool {
	if d == nil || d.conn == nil {
		return false
	}

	_, err := d.conn.Exec("BEGIN")
	return err == nil
}

func (d *DB) commit() bool {
	if d == nil || d.conn == nil {
		return false
	}

	_, err := d.conn.Exec("COMMIT")
	return err == nil
}

func (d *DB) rollback() bool {
	if d == nil || d.conn == nil {
		return false
	}

	_, err := d.conn.Exec("ROLLBACK")
	return err == nil
}

func (d *DB) transaction(fn any) error {
	if d == nil || d.conn == nil {
		return fmt.Errorf("db not connected")
	}

	if !d.begin() {
		return fmt.Errorf("failed to begin transaction")
	}

	defer func() {
		if r := recover(); r != nil {
			d.rollback()
			panic(r)
		}
	}()

	result := OSLcallFunc(fn, nil, []any{d})

	if result != nil && result.(error) != nil {
		d.rollback()
		return result.(error)
	}

	if !d.commit() {
		return fmt.Errorf("failed to commit transaction")
	}

	return nil
}

func (d *DB) lastInsertId() int64 {
	if d == nil || d.conn == nil {
		return 0
	}

	result := d.queryOne("SELECT last_insert_rowid()")
	if result.isEmpty() {
		return 0
	}

	return int64(OSLcastNumber(result.get(1)))
}

func (d *DB) rowsAffected(query any, args ...any) int {
	if d == nil || d.conn == nil {
		return 0
	}

	queryStr := OSLtoString(query)
	argsSlice := make([]any, len(args))
	for i, arg := range args {
		argsSlice[i] = arg
	}

	result, err := d.conn.Exec(queryStr, argsSlice...)
	if err != nil {
		return 0
	}

	affected, _ := result.RowsAffected()
	return int(affected)
}

func (r *DBRow) get(colIndex any) any {
	if r == nil {
		return nil
	}

	idx := OSLcastInt(colIndex) - 1
	if idx < 0 || idx >= len(r.values) {
		return nil
	}

	return r.values[idx]
}

func (r *DBRow) getByName(colName any) any {
	if r == nil {
		return nil
	}

	name := strings.ToLower(OSLtoString(colName))

	for i, col := range r.columns {
		if strings.ToLower(col) == name {
			return r.values[i]
		}
	}

	return nil
}

func (r *DBRow) toMap() map[string]any {
	if r == nil {
		return map[string]any{}
	}

	result := make(map[string]any)
	for i, col := range r.columns {
		result[col] = r.values[i]
	}

	return result
}

func (r *DBRow) toArray() []any {
	if r == nil {
		return []any{}
	}

	return r.values
}

func (r *DBRow) isEmpty() bool {
	if r == nil {
		return true
	}

	return len(r.values) == 0
}

func (r *DBRow) count() int {
	if r == nil {
		return 0
	}

	return len(r.values)
}

var db = DB{}
