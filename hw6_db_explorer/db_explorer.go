package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"path"
)

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные

func NewDbExplorer(db *sql.DB) (handler *pathResolver, err error) {
	pr := newPathResolver()
	//pr.Add("GET /hello", hello)
	//pr.Add("* /goodbye/*", goobye)
	pr.Add("GET /", getTables)
	return pr, nil
}

type pathResolver struct {
	handlers map[string]http.HandlerFunc
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
	//fmt.Fprint(res, "qqq")
	http.NotFound(res, req)
}

func getTables(w http.ResponseWriter, r *http.Request) {

	fmt.Fprint(w, "Z")
}
