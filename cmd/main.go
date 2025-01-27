package main

import (
	"fileserver"
	"fileserver/internal/filesystem"
	"net/http"
)

func main() {
	thing := filesystem.New(".")
	fs := fileserver.FileServer(thing)

	http.Handle("/", fs)

	http.ListenAndServe(":8080", nil)
}
