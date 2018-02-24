package model

import (
	"net/http"
)

type SuccesResult struct {
	Code   int         `json:"code"`
	Result interface{} `json:"result"`
}

func NewSuccessResult(result interface{}) *SuccesResult {
	return &SuccesResult{http.StatusOK, result}
}
