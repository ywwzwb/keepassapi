package handler

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"keepassapi/helper"
	"keepassapi/model"
	"math"
	"net/http"
	"strconv"
	"time"
)

// password 在header 中的键, 必须大写字母开头
const PASSWORD_HEADER_KEY = "Password"
const AES_KEY = "123456"
const DECRYPT_PASSWORD_HEADER_KEY = "DecryptPassword"

type SimpleFilter struct {
	handler func(http.ResponseWriter, *http.Request)
}

func NewSimpleFilter(f func(http.ResponseWriter, *http.Request)) SimpleFilter {
	return SimpleFilter{f}
}
func (self SimpleFilter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rawPasswordArr, ok := r.Header[PASSWORD_HEADER_KEY]
	if !ok {
		err := model.NewGeneralError(http.StatusBadRequest, "缺少password 字段")
		err.WriteIn(w)
		return
	}
	if len(rawPasswordArr) == 0 {
		err := model.NewGeneralError(http.StatusBadRequest, "password 为空")
		err.WriteIn(w)
		return
	}
	rawPasswordBase64 := rawPasswordArr[0]
	rawPasswordByte, err := base64.StdEncoding.DecodeString(rawPasswordBase64)
	if err != nil {
		err := model.NewGeneralError(http.StatusBadRequest, "password base64 解码错误")
		err.WriteIn(w)
		return
	}
	password := string(rawPasswordByte)
	unlockError := helper.SharedKeepassHelper().TryUnlock(password)
	if unlockError != nil {
		unlockError.WriteIn(w)
		return
	}
	self.handler(w, r)
}

type Filter struct {
	handler func(http.ResponseWriter, *http.Request)
}

func NewFilter(f func(http.ResponseWriter, *http.Request)) Filter {
	return Filter{f}
}
func (self Filter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !passCheckFilter(w, r) {
		return
	}
}

func passCheckFilter(w http.ResponseWriter, r *http.Request) bool {
	rawPasswordArr, ok := r.Header[PASSWORD_HEADER_KEY]
	if !ok {
		err := model.NewGeneralError(http.StatusBadRequest, "缺少password 字段")
		err.WriteIn(w)
		return false
	}
	if len(rawPasswordArr) == 0 {
		err := model.NewGeneralError(http.StatusBadRequest, "password 为空")
		err.WriteIn(w)
		return false
	}
	rawPasswordBase64 := rawPasswordArr[0]
	rawPassword, err := base64.StdEncoding.DecodeString(rawPasswordBase64)
	if err != nil {
		err := model.NewGeneralError(http.StatusBadRequest, "password base64 解码错误")
		err.WriteIn(w)
		return false
	}
	aesBlockDecrypter, err := aes.NewCipher(makeAES256Key([]byte(AES_KEY)))
	if err != nil {
		err := model.NewGeneralError(http.StatusInternalServerError, "block decripter 生成失败")
		err.WriteIn(w)
		return false
	}
	aesDecrypter := cipher.NewCFBDecrypter(aesBlockDecrypter, makeAES256IV([]byte(AES_KEY))[:aesBlockDecrypter.BlockSize()])
	decrypedPassword := make([]byte, len(rawPassword))
	aesDecrypter.XORKeyStream(decrypedPassword, rawPassword)
	if len(decrypedPassword) <= 10 {
		err := model.NewGeneralError(http.StatusBadRequest, "校验失败, 长度过短")
		err.WriteIn(w)
		return false
	}
	unixIntervalString := string(decrypedPassword[0:10])
	unixInterval, err := strconv.ParseInt(unixIntervalString, 10, 64)
	if err != nil {
		err := model.NewGeneralError(http.StatusBadRequest, "校验失败, 解析时间失败")
		err.WriteIn(w)
		return false
	}
	// 超过 30 秒就丢弃

	if math.Abs(float64(unixInterval-time.Now().Unix())) > 30 {
		// TODO: debug
		// err := model.NewGeneralError(http.StatusBadRequest, "校验失败, 超时")
		// err.WriteIn(w)
		// return false
	}
	password := string(decrypedPassword[10:])
	r.Header.Set(DECRYPT_PASSWORD_HEADER_KEY, password)
	return true
}
func makeAES256Key(key []byte) []byte {
	aes256key := bytes.Repeat(key, 32/len(key)+1)[:32]
	return aes256key
}
func makeAES256IV(key []byte) []byte {
	aes256iv := bytes.Repeat(key, 32/len(key)+1)[:32]
	return aes256iv
}
