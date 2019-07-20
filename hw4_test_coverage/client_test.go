package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
)

type UserData struct {
	Id         int    `xml:"-" json:"-"`
	FirstName  string `xml:"-" json:"-"`
	SecondName string `xml:"last_name" json:"last_name"`
	Age        int    `xml:"age"  json:"age"`
	About      string `xml:"-"  json:"-"`
	Gender     string `xml:"-" json:"-"`
}

type Users struct {
	XMLName xml.Name   `xml:"root"`
	Users   []UserData `xml:"row"`
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("query")
	switch status {
	case "StatusInternalServerError":
		w.WriteHeader(http.StatusInternalServerError)
	case "StatusBadRequest":
		w.WriteHeader(http.StatusBadRequest)
		orderField := r.URL.Query().Get("order_field")
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
	case "badUsers":
		return
	case "ok":
		// open file
		xmlFile, err := os.Open("dataset.xml")
		if err != nil {
			fmt.Printf("error open xml file: %v", err)
			return
		}
		defer xmlFile.Close()
		xmlData, _ := ioutil.ReadAll(xmlFile)

		// parse users
		users := new(Users)
		err = xml.Unmarshal(xmlData, users)
		if err != nil {
			fmt.Printf("error unmarshal xml data: %v", err)
			return
		}

		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit > 25 {
			limit = 1
		}
		//println("lim in server:", limit)

		// make body
		body, err := json.Marshal(users.Users[:limit])
		if err != nil {
			fmt.Printf("error marshal xml data to json: %v", err)
			return
		}

		// write body
		w.Write(body)

		//fmt.Println(string(body))
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
		{
			3,
			1,
			"badUsers",
			"name",
			0,
		},
		{
			6,
			1,
			"ok",
			"name",
			0,
		},
		{
			25,
			1,
			"ok",
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
