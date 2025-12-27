package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"SfosBeginnerGuide/internal/content"
	"SfosBeginnerGuide/internal/httpapi"
	"SfosBeginnerGuide/internal/markdown"
	"SfosBeginnerGuide/internal/search"
)

//go:embed docs
var docs embed.FS

func main() {
	gracefulShutdown := make(chan os.Signal, 1)
	signal.Notify(gracefulShutdown, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	port, ok := os.LookupEnv("APP_PORT")
	if !ok {
		port = "8080"
	}

	md := markdown.New()
	parser := content.NewCachedMarkdownParser(docs, md, 5*time.Minute)
	languages := content.NewFSLocalizer(docs, "docs")
	searcher := search.NewService(docs)
	handler := httpapi.NewHandler(parser, languages, searcher)

	mux := http.NewServeMux()
	mux.HandleFunc("/languages", handler.LanguagesList)
	mux.HandleFunc("/capabilities", handler.Capabilities)
	mux.HandleFunc("/search/", handler.Search)
	mux.HandleFunc("/", handler.Content)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	go func() {
		log.Println("Server is starting on port " + port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(fmt.Errorf("failed to start server: %w", err))
		}
	}()

	<-gracefulShutdown
	log.Println("Shutdown requested, shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
}
