package main

import (
	"database/sql"
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

type cr struct {
	Response map[string]interface{} `json:"response"`
	Tables   []string               `json:"tables"`
}

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
		log.Fatal("can't show tables", err)
	}
	tables := make([]string, 0)
	for rows.Next() {
		table := ""
		err = rows.Scan(&table)
		if err != nil {
			log.Fatal("can't scan rows", err)
		}
		tables = append(tables, table)
	}
	fmt.Println(tables)
	w.Write([]byte("qqq"))

	// надо закрывать соединение, иначе будет течь
	rows.Close()
}
