package search

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"SfosBeginnerGuide/internal/helper"
)

var (
	ErrSearchDisabled          = errors.New("search capability disabled")
	ErrEmbeddingsServerMissing = errors.New("embeddings server missing")
	ErrLanguageNotFound        = errors.New("language not found")
	ErrAssetsUnavailable       = errors.New("search assets not configured")
)

type Service struct {
	Root   fs.FS
	Client *Client
}

func NewService(root fs.FS) *Service {
	return &Service{Root: root}
}

func (receiver *Service) Search(ctx context.Context, language, query string, limit int) ([]Result, error) {
	if !helper.BoolEnv("CAPABILITY_SEARCHING", false) {
		return nil, ErrSearchDisabled
	}

	baseURL := strings.TrimSpace(os.Getenv("EMBEDDINGS_SERVER"))
	if baseURL == "" {
		return nil, ErrEmbeddingsServerMissing
	}

	if strings.TrimSpace(language) == "" {
		return nil, errors.New("language is required")
	}
	if strings.TrimSpace(query) == "" {
		return nil, errors.New("query is required")
	}

	if receiver.Root == nil {
		return nil, ErrAssetsUnavailable
	}

	langFS, err := fs.Sub(receiver.Root, path.Join("docs", language))
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrLanguageNotFound, language)
	}

	client := receiver.Client
	if client == nil {
		client = &Client{
			BaseURL: baseURL,
			HTTP:    &http.Client{Timeout: 60 * time.Second},
		}
	} else if strings.TrimSpace(client.BaseURL) == "" {
		client.BaseURL = baseURL
	}

	return Search(ctx, langFS, client, query, limit)
}
