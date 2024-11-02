package handlers

import "net/http"

func SuccessResponse(response http.ResponseWriter, serialized []byte) {
	if len(serialized) == 0 {
		response.WriteHeader(http.StatusNoContent)
		return
	}
	response.Header().Add("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	response.Write(serialized)
}

func FailureResponse(response http.ResponseWriter, status int, erroMessage string) {
	response.WriteHeader(status)
	response.Write([]byte(erroMessage))
}
