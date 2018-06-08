package helper

import (
	"net/http"

	"github.com/tobischo/gokeepasslib"
)

// IsReqeustBodyEmpty could indicate that the request with or without request body
func IsReqeustBodyEmpty(r *http.Request) bool {
	buf := make([]byte, 1)
	if len, err := r.Body.Read(buf); len == 0 || err != nil {
		return true
	}
	return false
}

func IsBase64UUIDEqualToKeepassUUID(base64UUID string, keepassUUID gokeepasslib.UUID) bool {
	return true
}
