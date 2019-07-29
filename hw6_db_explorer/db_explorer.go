package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path"
)

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные

type pathResolver struct {
	handlers map[string]http.HandlerFunc
}

type crTables map[string]map[string][]string

func NewDbExplorer(db *sql.DB) (handler *pathResolver, err error) {
	fooGetTables := func(w http.ResponseWriter, r *http.Request) {
		getTables(w, r, db)
	}
	pr := newPathResolver()
	//pr.Add("GET /hello", hello)
	//pr.Add("* /goodbye/*", goobye)
	pr.Add("GET /", fooGetTables)
	return pr, nil
}

func newPathResolver() *pathResolver {
	return &pathResolver{make(map[string]http.HandlerFunc)}
}

func (p *pathResolver) Add(path string, handler http.HandlerFunc) {
	p.handlers[path] = handler
}

func (p *pathResolver) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	check := req.Method + " " + req.URL.Path
	for pattern, handlerFunc := range p.handlers {
		if ok, err := path.Match(pattern, check); ok && err == nil {
			handlerFunc(res, req)
			return
		} else if err != nil {
			fmt.Fprint(res, req)
		}
	}
	http.NotFound(res, req)
}

func getTables(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	rows, err := db.Query("SHOW TABLES;")
	if err != nil {
		//log.Fatal("can't show tables", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	tables := make([]string, 0)
	table := ""
	for rows.Next() {
		//err = rows.Scan(&table)
		err = rows.Scan(&table)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Fatal("can't scan rows", err)
		}
		tables = append(tables, table)
	}

	result := crTables{}
	result["response"] = map[string][]string{}
	result["response"]["tables"] = tables

	jsonTables, err := json.Marshal(result)
	if err != nil {
		//log.Fatal("can't marshal tables", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Write([]byte(jsonTables))

	// надо закрывать соединение, иначе будет течь
	rows.Close()
}
