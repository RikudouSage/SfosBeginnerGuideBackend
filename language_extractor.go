package main

import (
	"SfodBeginnerGuide/helper"
	"embed"
	"log"
	"net/http"
)

func langHandler(writer http.ResponseWriter, request *http.Request) {
	languages, err := extractLanguages(&docs)
	if err != nil {
		log.Println(err)

		helper.WriteResponse(http.StatusInternalServerError, map[string]string{
			"error": "Internal error",
		}, writer)
		return
	}

	helper.WriteOkResponse(languages, writer)
}

func extractLanguages(fs *embed.FS) ([]string, error) {
	entries, err := fs.ReadDir("docs")
	if err != nil {
		return nil, err
	}

	result := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		result = append(result, entry.Name())
	}

	return result, nil
}
