package handler

import "net/http"

type KeepassFilter struct {
	handler func(http.ResponseWriter, *http.Request)
}

func NewKeepassFilter(f func(http.ResponseWriter, *http.Request)) KeepassFilter {
	return KeepassFilter{f}
}
