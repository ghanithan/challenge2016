package handlers

import "net/http"

func SuccessResponse(response http.ResponseWriter, serialized []byte) {
	response.Header().Add("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	response.Write(serialized)
}

func FailureResponse(response http.ResponseWriter, status int, erroMessage string) {
	response.WriteHeader(status)
	response.Write([]byte(erroMessage))
}
