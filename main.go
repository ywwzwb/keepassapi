package main

import (
	"fmt"
	"keepassapi/handler"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8000"
	}
	r := mux.NewRouter()
	r.Handle("/{path:.*}", handler.NewSimpleFilter(handler.Get)).Methods("GET")
	fmt.Println("running at port:", port)
	http.ListenAndServe("0.0.0.0:"+port, r)

}
