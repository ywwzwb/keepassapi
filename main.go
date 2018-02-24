package main

import (
	"fmt"
	"keepassapi/handler"
	"keepassapi/helper"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	port := os.Getenv("PORT")
	helper.Keepassdbpath = os.Getenv("PATH")
	r := mux.NewRouter()
	r.Handle("/{path:.*}", handler.NewSimpleFilter(handler.Get)).Methods("GET")
	fmt.Println("running at port:", port)
	fmt.Println("keepass db path:", helper.Keepassdbpath)
	http.ListenAndServe("0.0.0.0:"+port, r)
}
