package handlers

import "net/http"

func SuccessResponse(response http.ResponseWriter, serialized []byte) {
	response.WriteHeader(http.StatusOK)
	response.Write(serialized)
}

func FailureResponse(response http.ResponseWriter, erroMessage string) {
	response.WriteHeader(http.StatusNotFound)
	response.Write([]byte(erroMessage))
}
