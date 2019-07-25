package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type cr struct {
	Error string    `json:"error"`
	OU    OtherUser `json:"response"`
}

func (srv *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "...":
		//srv.wrapperDoSomeJob(w, r)
		srv.wrapperDoSomeJob()
	default:
		// 404
	}
}

func (srv *MyApi) wrapperDoSomeJob() {
	// заполнение структуры params
	// валидирование параметров
	//res, err := srv.DoSomeJob(ctx, params)
	// прочие обработки
}

func (srv *OtherApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch strings.Trim(r.URL.Path, " ") {
	// get url from method
	case "/user/create":
		srv.handlerCreate(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (srv *OtherApi) handlerCreate(w http.ResponseWriter, r *http.Request) {

	// get Method from method
	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// get auth bool from method
	if r.Header.Get("X-Auth") != "100500" {
		w.WriteHeader(http.StatusForbidden)
	}
	// get OtherCreateParams validators
	r.ParseForm()
	user := r.PostFormValue("username")
	account := r.PostFormValue("account_name") //paramname=account_name
	class := r.PostFormValue("class")
	level := r.PostFormValue("level")
	if user == "" { //required
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if len(user) < 3 { //min=3
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if class != "warrior" && class != "sorcerer" && class != "rouge" { //enum=warrior|sorcerer|rouge,default=warrior
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "class must be one of [warrior, sorcerer, rouge]"}`))
		return
	}
	l, _ := strconv.Atoi(level)
	if l < 1 || l > 50 { //min=1,max=50
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	params := OtherCreateParams{user, account, class, l}
	ou, _ := srv.Create(r.Context(), params)
	jOU := cr{"", *ou}
	result, _ := json.Marshal(jOU)
	w.Write(result)
}
