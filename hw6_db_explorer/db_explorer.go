package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные

type DBExplorer struct {
	db     *sql.DB
	tables map[string]*Table
	//path   *regexp.Regexp
}

type Table struct {
	name    string
	columns []*Column
}

type Column struct {
	fieldName string
	typeName  string
	isNull    bool
	keyType   string
	defField  *string
	extra     string
}

type Response struct {
	Response crDBE  `json:"response,omitempty"`
	Error    string `json:"error,omitempty"`
}

type crDBE map[string]interface{}

func NewDbExplorer(db *sql.DB) (http.Handler, error) {
	dbe := DBExplorer{
		db:     db,
		tables: make(map[string]*Table),
	}
	showTables, err := dbe.db.Query("SHOW TABLES")
	if err != nil {
		log.Fatal(err)
	}
	defer showTables.Close()

	tables := make([]*Table, 0)
	table := ""
	for showTables.Next() {
		err = showTables.Scan(&table)
		if err != nil {
			log.Fatal(err)
		}
		tables = append(tables, &Table{name: table})
	}

	for _, table := range tables {
		cols, err := dbe.db.Query(fmt.Sprintf("SHOW COLUMNS FROM `%s`", table.name))
		if err != nil {
			log.Fatal(err)
		}

		for cols.Next() {
			col := &Column{}
			isNull := ""
			err = cols.Scan(&col.fieldName, &col.typeName, &isNull, &col.keyType, &col.defField, &col.extra)

			if isNull == "YES" {
				col.isNull = true
			}
			table.columns = append(table.columns, col)
		}
		dbe.tables[table.name] = table
		cols.Close()
	}

	return dbe, nil
}

func (dbe DBExplorer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	parts := strings.Split(r.URL.Path, "/")
	name, id := "", ""
	if len(parts) > 1 {
		name = parts[1]
	}
	if len(parts) > 2 {
		id = parts[2]
	}

	if name != "" {
		_, ok := dbe.tables[name]
		if !ok {
			writeJSON(w, http.StatusNotFound, "", nil, "unknown table")
			return
		}
	}

	switch r.Method {
	case "GET":
		if name == "" && id == "" {
			dbe.getTables(w, r)
		} else if name != "" && id == "" {
			dbe.getSelect(w, r, name)
		} else if name != "" && id != "" {
			dbe.getSelectID(w, r, name, id)
		}
	case "PUT":
		if name != "" && id == "" {
			dbe.insRecord(w, r, name)
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func writeJSON(w http.ResponseWriter, respStatus int, respName string, respData interface{}, error string) {
	resp := Response{}
	if error != "" {
		resp = Response{
			Error: error,
		}
	} else {
		resp = Response{
			Response: crDBE{respName: respData},
		}
	}

	respJSON, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(respStatus)
	w.Write(respJSON)
}

func (dbe *DBExplorer) getTables(w http.ResponseWriter, r *http.Request) {
	rows, err := dbe.db.Query("SHOW TABLES")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, "", nil, err.Error())
		return
	}
	defer rows.Close()

	tables := make([]string, 0)
	table := ""
	for rows.Next() {
		err = rows.Scan(&table)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, "", nil, err.Error())
			return
		}
		tables = append(tables, table)
	}
	writeJSON(w, http.StatusOK, "tables", tables, "")
}

func (dbe *DBExplorer) getSelect(w http.ResponseWriter, r *http.Request, table string) {

	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil {
		offset = 0
	}
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		limit = 5
	}

	query := fmt.Sprintf(`SELECT * FROM %s LIMIT %d OFFSET %d`, table, limit, offset)
	respAll, err := dbe.execQuery(table, query)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, "", nil, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, "records", respAll, "")
}

func (dbe *DBExplorer) getColumns(table string) []string {
	tabNames, _ := dbe.tables[table]
	colNames := make([]string, 0)
	for _, col := range tabNames.columns {
		colNames = append(colNames, col.fieldName)
	}
	return colNames
}

func (dbe *DBExplorer) execQuery(table string, query string) ([]crDBE, error) {
	rows, err := dbe.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dest := make([]interface{}, 0)

	colNames := dbe.getColumns(table)

	colTypes, _ := rows.ColumnTypes()
	for _, item := range colTypes {
		switch item.DatabaseTypeName() {
		case "INT":
			dest = append(dest, new(int))
		case "VARCHAR", "TEXT":
			dest = append(dest, new(sql.NullString))
		default:
		}
	}
	resp := make([]crDBE, 0)
	for rows.Next() {
		rows.Scan(dest...)
		row := make(crDBE, 0)
		for i, item := range dest {
			switch v := item.(type) {
			case *int:
				row[colNames[i]] = *v
			case *sql.NullString:
				if v.Valid {
					row[colNames[i]] = v.String
				} else {
					row[colNames[i]] = nil
				}
			}
		}
		resp = append(resp, row)
	}

	return resp, nil
}

func (dbe *DBExplorer) getSelectID(w http.ResponseWriter, r *http.Request, table, ids string) {
	id, _ := strconv.Atoi(ids)
	colNames := dbe.getColumns(table)

	query := fmt.Sprintf(`SELECT %s FROM %s WHERE %s = %d`, strings.Join(colNames, ", "), table, colNames[0], id)
	respId, err := dbe.execQuery(table, query)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, "", nil, err.Error())
		return
	}
	if len(respId) == 0 {
		writeJSON(w, http.StatusNotFound, "", nil, "record not found")
		return
	}
	writeJSON(w, http.StatusOK, "record", respId[0], "")
}

func (dbe *DBExplorer) insRecord(w http.ResponseWriter, r *http.Request, table string) {
	tab := dbe.tables[table]

	data := make(crDBE)
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, "", nil, err.Error())
		return
	}

	id, err := dbe.insertData(table, data)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, "", nil, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, tab.columns[0].fieldName, id, "")
}

func (dbe *DBExplorer) insertData(table string, data crDBE) (int64, error) {
	columns := dbe.tables[table].columns

	vals := make([]interface{}, 0)
	for i := 1; i < len(columns); i++ {
		val, ok := data[columns[i].fieldName]
		if !ok {
			if columns[i].isNull {
				val = nil
			} else {
				val = ""
			}
		}
		vals = append(vals, val)
	}
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, strings.Join(dbe.getColumns(table)[1:], ", "), "?"+strings.Repeat(", ?", len(vals)-1))
	result, err := dbe.db.Exec(query, vals...)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

//sql injects!!!!!!!!!!!!!!
