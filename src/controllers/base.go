package controllers

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"push_service/src/core"
)

type BaseController struct{}

type JsonResponse struct {
	Error   int         `json:"error"`
	Data    interface{} `json:"data"`
	Message string      `json:"msg"`
}

// Writes the response as a standard JSON response with StatusOK
func (base *BaseController) sendOk(w http.ResponseWriter, m interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(&JsonResponse{Error: 0, Data: m, Message: ""}); err != nil {
		base.sendError(w, http.StatusInternalServerError, "Internal Server Error")
	}
}

// Writes the error response as a Standard API JSON response with a response code
func (base *BaseController) sendError(w http.ResponseWriter, errorCode int, errorMsg string) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(&JsonResponse{Error: errorCode, Data: "", Message: errorMsg})
}

//check api key
func (base *BaseController) checkApiKey(token string, params interface{}) bool {
	value := reflect.ValueOf(params)
	var data []string
	for i := 0; i < value.NumField(); i++ {
		data = append(data, fmt.Sprint(value.Field(i)))
	}

	str := strings.Join(data, core.Config.APISecret)
	md5Data := []byte(str)
	has := md5.Sum(md5Data)
	md5str := fmt.Sprintf("%x", has) //将[]byte转成16进制

	if md5str != token {
		return false
	}

	return true

}
