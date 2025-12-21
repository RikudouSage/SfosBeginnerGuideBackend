package httpx

import (
	"encoding/json"
	"net/http"
)

func WriteOK(body any, writer http.ResponseWriter) {
	WriteJSON(http.StatusOK, body, writer)
}

func WriteJSON(status int, body any, writer http.ResponseWriter) {
	writer.Header().Add("Content-Type", "application/json")

	if body == nil {
		return
	}

	data, err := json.Marshal(body)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		_, _ = writer.Write([]byte(`{"error": "Internal Server Error"}`))
		return
	}

	writer.WriteHeader(status)
	_, _ = writer.Write(data)
}
