package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

//go:embed docs
var docs embed.FS

func main() {
	gracefulShutdown := make(chan os.Signal, 1)
	signal.Notify(gracefulShutdown, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	go func() {
		port, ok := os.LookupEnv("APP_PORT")
		if !ok {
			port = "8080"
		}

		log.Println("Server is starting on port " + port)
		err := http.ListenAndServe(":"+port, nil)
		if err != nil {
			log.Fatal(fmt.Errorf("failed to start server: %w", err))
		}
	}()

	http.HandleFunc("/languages", langHandler)
	http.HandleFunc("/", handler)

	<-gracefulShutdown
	log.Println("Shutdown requested, shutting down...")
}
