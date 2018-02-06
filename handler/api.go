package handler

import (
	"fmt"
	"net/http"
)

func ListRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "hello, world")
}
