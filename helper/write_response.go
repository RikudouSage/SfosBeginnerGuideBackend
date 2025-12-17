package helper

import (
	"encoding/json"
	"net/http"
)

func WriteOkResponse(body any, writer http.ResponseWriter) {
	WriteResponse(http.StatusOK, body, writer)
}

func WriteResponse(status int, body any, writer http.ResponseWriter) {
	writer.Header().Add("Content-Type", "application/json")

	if body == nil {
		return
	}

	data, err := json.Marshal(body)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(`{"error": "Internal Server Error"}`))
		return
	}

	writer.WriteHeader(status)
	writer.Write(data)
}
