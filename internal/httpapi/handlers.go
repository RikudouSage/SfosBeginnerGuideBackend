package httpapi

import (
	"errors"
	"log"
	"net/http"
	"os"

	"SfosBeginnerGuide/internal/content"
	. "SfosBeginnerGuide/internal/helper"
	"SfosBeginnerGuide/internal/httpx"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewErrorResponse(message string) *ErrorResponse {
	return &ErrorResponse{Error: message}
}

type Handler struct {
	Parser    content.Parser
	Languages content.LanguageProvider
}

func NewHandler(parser content.Parser, languages content.LanguageProvider) *Handler {
	return &Handler{Parser: parser, Languages: languages}
}

func (receiver *Handler) Content(writer http.ResponseWriter, request *http.Request) {
	defer httpx.DrainBody(request)

	if request.Method != http.MethodGet {
		httpx.WriteJSON(
			http.StatusMethodNotAllowed,
			NewErrorResponse("Method not allowed"),
			writer,
		)
		return
	}

	path := request.URL.Path
	file, err := receiver.Parser.ParseByPath(path)
	if errors.Is(err, os.ErrNotExist) {
		httpx.WriteJSON(
			http.StatusNotFound,
			NewErrorResponse("No content could be found at the requested URL"),
			writer,
		)
		return
	}
	if err != nil {
		log.Println(err)
		httpx.WriteJSON(
			http.StatusInternalServerError,
			NewErrorResponse("Failed parsing"),
			writer,
		)
		return
	}

	httpx.WriteOK(file, writer)
}

func (receiver *Handler) LanguagesList(writer http.ResponseWriter, _ *http.Request) {
	languages, err := receiver.Languages.List()
	if err != nil {
		log.Println(err)

		httpx.WriteJSON(http.StatusInternalServerError, NewErrorResponse("Internal error"), writer)
		return
	}

	httpx.WriteOK(languages, writer)
}

func (receiver *Handler) Capabilities(writer http.ResponseWriter, request *http.Request) {
	defer httpx.DrainBody(request)

	if request.Method != http.MethodGet {
		httpx.WriteJSON(
			http.StatusMethodNotAllowed,
			NewErrorResponse("Method not allowed"),
			writer,
		)
		return
	}

	capabilities := map[string]bool{
		"searching": BoolEnv("CAPABILITY_SEARCHING", false),
	}

	httpx.WriteOK(capabilities, writer)
}
