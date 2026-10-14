package api

import (
	"encoding/json"
	"log"
	"net/http"
)

const (
	CodeOK            = 0
	CodeBadRequest    = 40001
	CodeParseFailed   = 40002
	CodeNotFound      = 40401
	CodeModelFailed   = 50001
	CodeInternalError = 50002
)

type apiResponse struct {
	Code    int    `json:"code"`
	Data    any    `json:"data"`
	Message string `json:"message"`
}

func writeOK(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusOK, apiResponse{Code: CodeOK, Data: data, Message: "ok"})
}

func writeErr(w http.ResponseWriter, status, code int, msg string) {
	writeJSON(w, status, apiResponse{Code: code, Data: nil, Message: msg})
}

func writeJSON(w http.ResponseWriter, status int, v apiResponse) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("encode response: %v", err)
	}
}

// decodeBody 解析 JSON 请求体；失败时写出 40001 并返回 false。
func decodeBody(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeErr(w, http.StatusBadRequest, CodeBadRequest, "invalid JSON body: "+err.Error())
		return false
	}
	return true
}
