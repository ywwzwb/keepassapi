package model

import (
	"encoding/json"
	"fmt"
	"io"
)

type GeneralError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewGeneralError(code int, message string) *GeneralError {
	return &GeneralError{Code: code, Message: message}
}
func (self *GeneralError) error() string {
	return fmt.Sprintf("code: %d, msg:%s", self.Code, self.Message)
}

func (self *GeneralError) WriteIn(w io.Writer) {
	encoder := json.NewEncoder(w)
	encoder.Encode(self)
}
