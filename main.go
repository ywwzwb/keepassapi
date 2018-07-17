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
	var port string
	if len(os.Args) >= 3 {
		port = os.Args[1]
		helper.Keepassdbpath = os.Args[2]
	} else {
		port = os.Getenv("KEEPASS_PORT")
		helper.Keepassdbpath = os.Getenv("KEEPASS_DBPATH")
	}
	if len(port) == 0 || len(helper.Keepassdbpath) == 0 {
		fmt.Println("Usage: keepassapi <port> <dbpath>")
		os.Exit(1)
	}
	r := mux.NewRouter()
	r.Handle("/{UUID:[A-Za-z0-9\\-_=]*}", handler.NewSimpleFilter(handler.ReadDB)).Methods("GET")
	r.Handle("/{UUID:[A-Za-z0-9\\-_=]+}", handler.NewSimpleFilter(handler.AddRecord)).Methods("POST")
	r.Handle("/{UUID:[A-Za-z0-9\\-_=]+}", handler.NewSimpleFilter(handler.UpdateRecord)).Methods("PUT")
	r.Handle("/{UUID:[A-Za-z0-9\\-_=]+}", handler.NewSimpleFilter(handler.DeleteRecord)).Methods("DELETE")
	fmt.Println("running at port:", port)
	fmt.Println("keepass db path:", helper.Keepassdbpath)
	http.ListenAndServe("0.0.0.0:"+port, r)
}
