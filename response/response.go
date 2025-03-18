package response

import (
	"encoding/json"
	"net/http"
)

var emptyStruct = struct{}{}

type response struct {
	Code    int         `json:"code"`
	Message any         `json:"message"`
	Details any         `json:"details"`
	Data    interface{} `json:"data"`
}

func Unauthorized(w http.ResponseWriter, r *http.Request, message string, errors map[string]string) {
	if errors == nil {
		errors = make(map[string]string)
	}
	res := response{
		Code:    http.StatusUnauthorized,
		Message: message,
		Details: errors,
		Data:    &emptyStruct,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(res)
}

func BadRequest(w http.ResponseWriter, r *http.Request, message *string, errors map[string]string) {
	if errors == nil {
		errors = make(map[string]string)
	}
	if message == nil {
		defaultMessage := "Bad Request"
		message = &defaultMessage
	}
	res := response{
		Code:    http.StatusBadRequest,
		Message: *message,
		Details: errors,
		Data:    &emptyStruct,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(res)
}

func OK(w http.ResponseWriter, r *http.Request, data any) {
	res := response{
		Code:    http.StatusOK,
		Message: "OK",
		Details: &emptyStruct,
		Data:    data,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func InternalServerError(w http.ResponseWriter, r *http.Request) {
	res := response{
		Code:    http.StatusInternalServerError,
		Message: "Internal Server Error",
		Details: &emptyStruct,
		Data:    &emptyStruct,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(res)
}

func NotFound(w http.ResponseWriter, r *http.Request, message string) {
	if message == "" {
		message = "Not Found"
	}
	res := response{
		Code:    http.StatusNotFound,
		Message: message,
		Details: &emptyStruct,
		Data:    &emptyStruct,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(res)
}

func NoContent(w http.ResponseWriter, r *http.Request) {
	res := response{
		Code:    http.StatusNoContent,
		Message: "No Content",
		Details: &emptyStruct,
		Data:    &emptyStruct,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
	json.NewEncoder(w).Encode(res)
}
