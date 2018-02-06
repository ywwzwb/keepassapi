package main

import (
	"keepassapi/handler"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.Handle("/", handler.NewSimpleFilter(handler.ListRoot))
	// r.HandleFunc("/", handler.ListRoot)
	http.ListenAndServe("0.0.0.0:8000", r)

}
