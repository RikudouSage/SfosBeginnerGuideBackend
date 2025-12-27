package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	docsDir          = "docs"
	venvPythonPath   = "embeddings/.venv/bin/python"
	uvicornAppPath   = "embeddings.server:app"
	embedMode        = "passage"
	maxWordsPerChunk = 350
	overlapWords     = 50
	requestBatchSize = 32
)

type embedRequest struct {
	Texts []string `json:"texts"`
	Mode  string   `json:"mode"`
}

type embedResponse struct {
	Vectors [][]float64 `json:"vectors"`
	Dim     int         `json:"dim"`
}

type chunkEmbedding struct {
	Index  int       `json:"index"`
	Text   string    `json:"text"`
	Vector []float32 `json:"vector"`
}

type fileEmbeddings struct {
	Source string           `json:"source"`
	Model  string           `json:"model"`
	Dim    int              `json:"dim"`
	Chunks []chunkEmbedding `json:"chunks"`
}

func main() {
	if _, err := os.Stat(docsDir); err != nil {
		failf("docs directory not found: %v", err)
	}
	if _, err := os.Stat(venvPythonPath); err != nil {
		failf("missing venv python at %s (run `make install-deps`)", venvPythonPath)
	}

	port, err := randomFreePort()
	if err != nil {
		failf("failed to get free port: %v", err)
	}

	cmd, err := startServer(port)
	if err != nil {
		failf("failed to start embeddings server: %v", err)
	}
	defer stopServer(cmd)

	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	if err = waitForHealth(baseURL, 60*time.Second); err != nil {
		failf("server failed health check: %v", err)
	}

	model := os.Getenv("EMBEDDING_MODEL")
	if model == "" {
		model = "unknown"
	}
	fmt.Printf("Embedding model: %s\n", model)

	var totalFiles int
	var totalChunks int
	err = filepath.WalkDir(docsDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".md" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}

		text := strings.TrimSpace(stripFrontMatter(string(content)))
		if text == "" {
			fmt.Printf("Embedding %s (empty after front matter)\n", path)
			return writeEmbeddings(path, model, 0, nil)
		}

		chunks := chunkText(text, maxWordsPerChunk, overlapWords)
		fmt.Printf("Embedding %s (%d chunks)\n", path, len(chunks))
		vectors, dim, err := embedChunks(baseURL, chunks)
		if err != nil {
			return fmt.Errorf("embed %s: %w", path, err)
		}

		items := make([]chunkEmbedding, 0, len(chunks))
		for i, chunk := range chunks {
			items = append(items, chunkEmbedding{
				Index:  i,
				Text:   chunk,
				Vector: vectors[i],
			})
		}

		if err := writeEmbeddings(path, model, dim, items); err != nil {
			return err
		}
		totalFiles++
		totalChunks += len(chunks)
		fmt.Printf("Wrote %s (%d chunks)\n", strings.TrimSuffix(path, filepath.Ext(path))+".json", len(chunks))
		return nil
	})
	if err != nil {
		failf("embedding run failed: %v", err)
	}
	fmt.Printf("Done. Files: %d, chunks: %d\n", totalFiles, totalChunks)
}

func randomFreePort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	addr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		return 0, errors.New("unexpected listener address")
	}
	return addr.Port, nil
}

func startServer(port int) (*exec.Cmd, error) {
	cmd := exec.Command(
		venvPythonPath,
		"-m", "uvicorn",
		uvicornAppPath,
		"--host", "127.0.0.1",
		"--port", strconv.Itoa(port),
		"--log-level", "warning",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "PYTHONUNBUFFERED=1")
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd, nil
}

func stopServer(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	_ = cmd.Process.Signal(os.Interrupt)
	done := make(chan struct{})
	go func() {
		_ = cmd.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		_ = cmd.Process.Kill()
	}
}

func waitForHealth(baseURL string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 2 * time.Second}
	for time.Now().Before(deadline) {
		resp, err := client.Get(baseURL + "/health")
		if err == nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	return errors.New("timeout waiting for /health")
}

func stripFrontMatter(input string) string {
	lines := strings.Split(input, "\n")
	if len(lines) == 0 {
		return input
	}
	if strings.TrimSpace(lines[0]) != "---" {
		return input
	}
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "---" || line == "..." {
			return strings.Join(lines[i+1:], "\n")
		}
	}
	return input
}

func chunkText(text string, maxWords int, overlap int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}
	if maxWords <= 0 {
		return []string{text}
	}
	if overlap >= maxWords {
		overlap = maxWords / 4
	}

	var chunks []string
	for start := 0; start < len(words); {
		end := start + maxWords
		if end > len(words) {
			end = len(words)
		}
		chunks = append(chunks, strings.Join(words[start:end], " "))
		if end == len(words) {
			break
		}
		start = end - overlap
		if start < 0 {
			start = 0
		}
	}
	return chunks
}

func embedChunks(baseURL string, chunks []string) ([][]float32, int, error) {
	if len(chunks) == 0 {
		return nil, 0, nil
	}

	client := &http.Client{Timeout: 60 * time.Second}
	var vectors [][]float32
	dim := 0

	for start := 0; start < len(chunks); start += requestBatchSize {
		end := start + requestBatchSize
		if end > len(chunks) {
			end = len(chunks)
		}
		reqBody := embedRequest{
			Texts: chunks[start:end],
			Mode:  embedMode,
		}
		payload, err := json.Marshal(reqBody)
		if err != nil {
			return nil, 0, err
		}

		resp, err := client.Post(baseURL+"/embed", "application/json", bytes.NewReader(payload))
		if err != nil {
			return nil, 0, err
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, 0, err
		}
		if resp.StatusCode != http.StatusOK {
			return nil, 0, fmt.Errorf("embed status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
		}

		var parsed embedResponse
		if err := json.Unmarshal(body, &parsed); err != nil {
			return nil, 0, err
		}
		if len(parsed.Vectors) != end-start {
			return nil, 0, fmt.Errorf("embed response size mismatch: got %d vectors, expected %d", len(parsed.Vectors), end-start)
		}
		if dim == 0 {
			dim = parsed.Dim
		}
		for _, vec := range parsed.Vectors {
			out := make([]float32, len(vec))
			for i, v := range vec {
				out[i] = float32(v)
			}
			vectors = append(vectors, out)
		}
	}

	return vectors, dim, nil
}

func writeEmbeddings(path, model string, dim int, chunks []chunkEmbedding) error {
	output := fileEmbeddings{
		Source: path,
		Model:  model,
		Dim:    dim,
		Chunks: chunks,
	}
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}
	outPath := strings.TrimSuffix(path, filepath.Ext(path)) + ".json"
	return os.WriteFile(outPath, data, 0o644)
}

func failf(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
