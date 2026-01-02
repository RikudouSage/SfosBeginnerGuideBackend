package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"SfosBeginnerGuide/internal/content"
	"SfosBeginnerGuide/internal/helper"
	"SfosBeginnerGuide/internal/httpx"
	"SfosBeginnerGuide/internal/search"
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
	Searcher  SearchService
}

type SearchService interface {
	Search(ctx context.Context, language, query string, limit int) ([]search.Result, error)
}

func NewHandler(parser content.Parser, languages content.LanguageProvider, searcher SearchService) *Handler {
	return &Handler{Parser: parser, Languages: languages, Searcher: searcher}
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
		"searching": helper.BoolEnv("CAPABILITY_SEARCHING", false),
	}

	httpx.WriteOK(capabilities, writer)
}

func (receiver *Handler) Search(writer http.ResponseWriter, request *http.Request) {
	defer httpx.DrainBody(request)

	if request.Method != "QUERY" {
		httpx.WriteJSON(
			http.StatusMethodNotAllowed,
			NewErrorResponse("Method not allowed (expected QUERY)"),
			writer,
		)
		return
	}

	type searchRequest struct {
		Query string `json:"q"`
		Top   *int   `json:"top"`
	}

	lang := strings.TrimPrefix(request.URL.Path, "/search")
	lang = strings.TrimPrefix(lang, "/")
	lang, _, _ = strings.Cut(lang, "/")
	if lang == "" {
		httpx.WriteJSON(
			http.StatusBadRequest,
			NewErrorResponse("Missing language in path (expected /search/{lang})"),
			writer,
		)
		return
	}

	var body searchRequest
	if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
		if errors.Is(err, io.EOF) {
			httpx.WriteJSON(
				http.StatusBadRequest,
				NewErrorResponse("Missing JSON body"),
				writer,
			)
			return
		}
		httpx.WriteJSON(
			http.StatusBadRequest,
			NewErrorResponse("Invalid JSON body"),
			writer,
		)
		return
	}

	query := strings.TrimSpace(body.Query)
	if query == "" {
		httpx.WriteJSON(
			http.StatusBadRequest,
			NewErrorResponse("Missing body field: q"),
			writer,
		)
		return
	}

	limit := 20
	if body.Top != nil && *body.Top > 0 {
		parsed := *body.Top
		if parsed > 100 {
			parsed = 100
		}
		limit = parsed
	}

	ctx, cancel := context.WithTimeout(request.Context(), 2*time.Minute)
	defer cancel()

	results, err := receiver.Searcher.Search(ctx, lang, query, limit)
	if err != nil {
		if errors.Is(err, search.ErrSearchDisabled) {
			httpx.WriteJSON(
				http.StatusForbidden,
				NewErrorResponse("Search capability is disabled"),
				writer,
			)
			return
		}
		if errors.Is(err, search.ErrEmbeddingsServerMissing) {
			httpx.WriteJSON(
				http.StatusInternalServerError,
				NewErrorResponse("Search capability is enabled but EMBEDDINGS_SERVER is not configured"),
				writer,
			)
			return
		}
		if errors.Is(err, search.ErrLanguageNotFound) {
			httpx.WriteJSON(
				http.StatusNotFound,
				NewErrorResponse("Unknown language"),
				writer,
			)
			return
		}
		if errors.Is(err, search.ErrAssetsUnavailable) {
			httpx.WriteJSON(
				http.StatusInternalServerError,
				NewErrorResponse("Search assets are not configured on the server"),
				writer,
			)
			return
		}
		httpx.WriteJSON(
			http.StatusInternalServerError,
			NewErrorResponse("Failed to search embeddings"),
			writer,
		)
		return
	}

	httpx.WriteOK(results, writer)
}
