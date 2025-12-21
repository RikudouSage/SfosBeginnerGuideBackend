package httpx

import (
	"io"
	"net/http"
)

func DrainBody(request *http.Request) error {
	_, err := io.Copy(io.Discard, request.Body)
	if err != nil {
		return err
	}

	return request.Body.Close()
}
