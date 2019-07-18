package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func SearchServer(w http.ResponseWriter, r *http.Request) {

}

func TestSearchClient(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	sc := &SearchClient{
		"111",
		ts.URL,
	}

	sr := SearchRequest{
		5,
		1,
		"555",
		"name",
		0,
	}

	result, err := sc.FindUsers(sr)
	if err != nil {
		t.Errorf("klim error: %#v", err)
	}

	fmt.Println(result)

	ts.Close()
}
