package main

import (
	"fileserver"
	"net/http"
)

func main() {
	thing := http.Dir(".")
	fs := fileserver.FileServer(thing)

	http.Handle("/", fs)

	http.ListenAndServe(":8080", nil)
}
