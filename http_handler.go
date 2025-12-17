package main

import (
	"SfodBeginnerGuide/helper"
	"SfodBeginnerGuide/structs"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
)

var mdParser = newParser(&docs, markdown)

func handler(writer http.ResponseWriter, request *http.Request) {
	defer func(request *http.Request) {
		err := helper.DrainBody(request)
		if err != nil {
			log.Println(fmt.Errorf("drain body error: %w", err))
		}
	}(request)

	if request.Method != http.MethodGet {
		helper.WriteResponse(
			http.StatusMethodNotAllowed,
			structs.NewErrorMessage("Method not allowed"),
			writer,
		)
		return
	}

	path := request.URL.Path
	file, err := mdParser.parseByPath(path)
	if errors.Is(err, os.ErrNotExist) {
		helper.WriteResponse(
			http.StatusNotFound,
			structs.NewErrorMessage("No content could be found at the requested URL"),
			writer,
		)
		return
	} else if err != nil {
		log.Println(err)
		helper.WriteResponse(
			http.StatusInternalServerError,
			structs.NewErrorMessage("Failed parsing"),
			writer,
		)
		return
	}

	helper.WriteOkResponse(file, writer)
}
