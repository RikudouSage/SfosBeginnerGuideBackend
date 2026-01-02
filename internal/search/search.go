package search

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	defaultTimeout     = 60 * time.Second
	maxResultsDefault  = 20
	maxResultsLimit    = 100
	embeddingQueryMode = "query"
)

type Result struct {
	Source string  `json:"source"`
	Index  int     `json:"index"`
	Text   string  `json:"text"`
	Score  float32 `json:"score"`
}

type Client struct {
	BaseURL string
	HTTP    *http.Client
}

func (receiver *Client) EmbedQuery(ctx context.Context, query string) ([]float32, error) {
	if strings.TrimSpace(receiver.BaseURL) == "" {
		return nil, errors.New("embeddings server is empty")
	}
	reqBody := embedRequest{
		Texts: []string{query},
		Mode:  embeddingQueryMode,
	}
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	client := receiver.HTTP
	if client == nil {
		client = &http.Client{Timeout: defaultTimeout}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(receiver.BaseURL, "/")+"/embed", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embed status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var parsed embedResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	if len(parsed.Vectors) != 1 {
		return nil, fmt.Errorf("unexpected embed response size: %d", len(parsed.Vectors))
	}

	vector := make([]float32, len(parsed.Vectors[0]))
	for i, v := range parsed.Vectors[0] {
		vector[i] = float32(v)
	}
	return vector, nil
}

func Search(ctx context.Context, root fs.FS, client *Client, query string, limit int) ([]Result, error) {
	if strings.TrimSpace(query) == "" {
		return nil, errors.New("query is required")
	}
	if limit <= 0 {
		limit = maxResultsDefault
	}
	if limit > maxResultsLimit {
		limit = maxResultsLimit
	}

	if root == nil {
		return nil, errors.New("search assets are not configured")
	}
	if client == nil {
		return nil, errors.New("embeddings client is required")
	}

	queryVector, err := client.EmbedQuery(ctx, query)
	if err != nil {
		return nil, err
	}

	results, err := searchEmbeddings(root, queryVector)
	if err != nil {
		return nil, err
	}

	sort.Slice(results, func(i, j int) bool { return results[i].Score > results[j].Score })

	seen := make(map[string]struct{}, len(results))
	deduped := make([]Result, 0, len(results))
	for _, result := range results {
		if _, ok := seen[result.Source]; ok {
			continue
		}
		seen[result.Source] = struct{}{}
		deduped = append(deduped, result)
	}
	if len(deduped) > limit {
		deduped = deduped[:limit]
	}
	return deduped, nil
}

type embedRequest struct {
	Texts []string `json:"texts"`
	Mode  string   `json:"mode"`
}

type embedResponse struct {
	Vectors [][]float64 `json:"vectors"`
	Dim     int         `json:"dim"`
}

type fileEmbeddings struct {
	Source string           `json:"source"`
	Model  string           `json:"model"`
	Dim    int              `json:"dim"`
	Chunks []chunkEmbedding `json:"chunks"`
}

type chunkEmbedding struct {
	Index  int       `json:"index"`
	Text   string    `json:"text"`
	Vector []float32 `json:"vector"`
}

func searchEmbeddings(root fs.FS, query []float32) ([]Result, error) {
	if len(query) == 0 {
		return nil, errors.New("empty query vector")
	}

	queryNorm := vectorNorm(query)
	if queryNorm == 0 {
		return nil, errors.New("zero query vector norm")
	}

	var results []Result
	err := fs.WalkDir(root, ".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".json" {
			return nil
		}

		data, err := fs.ReadFile(root, path)
		if err != nil {
			return err
		}

		var file fileEmbeddings
		if err := json.Unmarshal(data, &file); err != nil {
			return nil
		}

		for _, chunk := range file.Chunks {
			score := cosineSimilarity(query, queryNorm, chunk.Vector)
			results = append(results, Result{
				Source: file.Source,
				Index:  chunk.Index,
				Text:   chunk.Text,
				Score:  score,
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return results, nil
}

func cosineSimilarity(query []float32, queryNorm float32, candidate []float32) float32 {
	if len(candidate) == 0 {
		return 0
	}
	dot := float32(0)
	for i, v := range query {
		if i >= len(candidate) {
			break
		}
		dot += v * candidate[i]
	}
	denom := queryNorm * vectorNorm(candidate)
	if denom == 0 {
		return 0
	}
	return dot / denom
}

func vectorNorm(vec []float32) float32 {
	sum := float32(0)
	for _, v := range vec {
		sum += v * v
	}
	return float32(math.Sqrt(float64(sum)))
}
