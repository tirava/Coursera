package main

import "net/http"

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
	switch r.URL.Path {
	case "...":
		//srv.wrapperDoSomeJob(w, r)
		srv.wrapperDoSomeJob()
	default:
		// 404
	}
}

func (srv *OtherApi) wrapperDoSomeJob() {
	// заполнение структуры params
	// валидирование параметров
	//res, err := srv.DoSomeJob(ctx, params)
	// прочие обработки
}
