package main

import (
	"errors"
	"fileserver"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	logger := slog.Default()

	err := os.Mkdir("./content", 0750)
	if err != nil && errors.Is(err, os.ErrExist) {
		logger.Info("Using existing /content dir")
	} else {
		logger.Info("Created new /content dir")
	}

	fs := fileserver.NewFS("./content")
	handler := fileserver.FileServer(fs, logger)

	http.Handle("/", handler)

	logger.Info("Listening on port 8080")
	http.ListenAndServe(":8080", nil)
}
