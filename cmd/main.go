package main

import (
	"errors"
	"log/slog"
	"net/http"
	"os"

	"github.com/ke126/fileserver"
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
