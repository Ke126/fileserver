package main

import (
	"fileserver"
	"fileserver/internal/filesystem"
	"log/slog"
	"net/http"
)

func main() {
	fs := filesystem.New(".")
	logger := slog.Default()
	handler := fileserver.FileServer(fs, logger)

	http.Handle("/", handler)

	logger.Info("Listening on port 8080")
	http.ListenAndServe(":8080", nil)
}
