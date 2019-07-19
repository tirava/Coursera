package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func SearchServer(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("query")
	orderField := r.URL.Query().Get("order_field")
	switch status {
	case "StatusInternalServerError":
		w.WriteHeader(http.StatusInternalServerError)
	case "StatusBadRequest":
		w.WriteHeader(http.StatusBadRequest)
		switch orderField {
		case "name":
			w.Write([]byte(`{"Error":"ErrorBadOrderField"}`))
		case "name1":
			w.Write([]byte(`{"Error":"ErrorBadOther"}`))
		default:
			w.Write([]byte(`111`))
		}
	case "StatusUnauthorized":
		w.WriteHeader(http.StatusUnauthorized)
	default:
		//parse user
	}

	//fmt.Println(r.URL.Query())
}

func TestSearchClient(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	sc := &SearchClient{
		"111",
		ts.URL,
	}

	srs := []SearchRequest{
		{
			-1,
			1,
			"555",
			"name",
			0,
		},
		{
			35,
			1,
			"StatusInternalServerError",
			"name",
			0,
		},
		{
			1,
			-1,
			"555",
			"name",
			0,
		},
		{
			5,
			1,
			"StatusBadRequest",
			"name",
			0,
		},
		{
			5,
			1,
			"StatusBadRequest",
			"name1",
			0,
		},
		{
			5,
			1,
			"StatusBadRequest",
			"nameX",
			0,
		},
		{
			5,
			1,
			"StatusUnauthorized",
			"name",
			0,
		},
	}

	for _, sr := range srs {

		_, err := sc.FindUsers(sr)
		if err != nil {
			//t.Errorf("klim error: %#v", err)
		}

		//fmt.Println(result)
	}

	ts.Close()
}
