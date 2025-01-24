package main

import (
	"net/http"
)

func main() {
	thing := http.Dir(".")
	fs := http.FileServer(thing)

	http.Handle("/", fs)

	http.ListenAndServe(":8080", nil)
}
