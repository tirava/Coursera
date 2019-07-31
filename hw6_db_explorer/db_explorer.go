package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
)

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные

type DBExplorer struct {
	db     *sql.DB
	tables map[string]*Table
	path   *regexp.Regexp
}

type Table struct {
	name    string
	columns []*Column
}

type Column struct {
	fieldName string
	typeName  string
	isNull    bool
	//key        string
	//defaultVal *string
	//extra      string
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
		path:   regexp.MustCompile("/?(\\w+)?(?:/(\\d+))?"),
	}
	qTables, err := dbe.db.Query("SHOW TABLES")
	if err != nil {
		log.Fatal(err)
	}

	tables := make([]*Table, 0)
	table := ""
	for qTables.Next() {
		err = qTables.Scan(&table)
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
			var isNull string
			err = cols.Scan(&col.fieldName, &col.typeName, &isNull)//&col.key,
			//&col.defaultVal,
			//&col.extra,

			if isNull == "YES" {
				col.isNull = true
			}
			table.columns = append(table.columns, col)
		}
		dbe.tables[table.name] = table
	}
	return dbe, nil
}

func (dbe DBExplorer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pathParsed := dbe.path.FindStringSubmatch(r.URL.Path)
	tableName, tableID := pathParsed[1], pathParsed[2]

	if tableName != "" {
		_, ok := dbe.tables[tableName]
		if !ok {
			writeJSON(w, http.StatusNotFound, "", nil, "unknown table")
			return
		}
	}

	ctx := context.WithValue(context.Background(), "tableName", tableName)
	ctx = context.WithValue(ctx, "tableID", tableID)

	switch r.Method {
	case "GET":
		if tableName == "" && tableID == "" {
			dbe.getTables(w, r)
		}

	default:
		w.WriteHeader(http.StatusBadRequest)
		return
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
