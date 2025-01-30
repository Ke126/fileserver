package fileserver_test

import (
	"fileserver"
	"io/fs"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func TestRouting(t *testing.T) {
	// this function tests the routing of the custom fileserver
	// against the built in fileserver from the net/http package

	// important things to test are:
	// the root path
	// requests to /somefile/ should redirect to /somefile
	// requests to /somedir should redirect to /somedir/
	// dir listings should have Content-Type: text/html; charset=utf-8
	// redirects should have NO Content-Type header
	// 404 responses should have Content-Type: text/plain; charset=utf-8

	tests := []struct {
		route           string
		wantStatus      int
		wantContentType string
	}{
		{
			route:           "/",
			wantStatus:      http.StatusOK,
			wantContentType: "text/html; charset=utf-8",
		},
		{
			route:           "/a",
			wantStatus:      http.StatusOK,
			wantContentType: "text/plain; charset=utf-8",
		},
		{
			route:           "/a/",
			wantStatus:      http.StatusMovedPermanently,
			wantContentType: "",
		},
		{
			route:           "/b",
			wantStatus:      http.StatusMovedPermanently,
			wantContentType: "",
		},
		{
			route:           "/b/",
			wantStatus:      http.StatusOK,
			wantContentType: "text/html; charset=utf-8",
		},
		{
			route:           "/b/c",
			wantStatus:      http.StatusOK,
			wantContentType: "text/plain; charset=utf-8",
		},
		{
			route:           "/b/c/",
			wantStatus:      http.StatusMovedPermanently,
			wantContentType: "",
		},
		{
			route:           "/d",
			wantStatus:      http.StatusMovedPermanently,
			wantContentType: "",
		},
		{
			route:           "/d/",
			wantStatus:      http.StatusOK,
			wantContentType: "text/html; charset=utf-8",
		},
		{
			route:           "/text.txt",
			wantStatus:      http.StatusOK,
			wantContentType: "text/plain; charset=utf-8",
		},
		{
			route:           "/unknown",
			wantStatus:      http.StatusOK,
			wantContentType: "text/html; charset=utf-8",
		},
		{
			route:           "/not-vid.mp4",
			wantStatus:      http.StatusOK,
			wantContentType: "video/mp4",
		},
		{
			route:           "/?thing",
			wantStatus:      http.StatusOK,
			wantContentType: "text/html; charset=utf-8",
		},
		{
			route:           "/%3Fthing",
			wantStatus:      http.StatusOK,
			wantContentType: "text/plain; charset=utf-8",
		},
		{
			route:           "/%23thing",
			wantStatus:      http.StatusOK,
			wantContentType: "text/plain; charset=utf-8",
		},
	}

	dir := makeMockFS()
	dir2 := fileserver.FromFS(".", dir)
	customFileServer := fileserver.FileServer(dir2, slog.Default())
	stdlibFileServer := http.FileServer(http.FS(dir))

	for _, tt := range tests {
		// test custom implementation
		t.Run("custom GET "+tt.route, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.route, nil)
			res := httptest.NewRecorder()

			customFileServer.ServeHTTP(res, req)

			gotStatus := res.Result().StatusCode
			gotContentType := res.Result().Header.Get("Content-Type")

			if gotStatus != tt.wantStatus {
				t.Errorf("got %d, want %d", gotStatus, tt.wantStatus)
			}

			if gotContentType != tt.wantContentType {
				t.Errorf("got %s, want %s", gotContentType, tt.wantContentType)
			}

		})
		// test against stdlib
		t.Run("stdlib GET "+tt.route, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.route, nil)
			res := httptest.NewRecorder()

			stdlibFileServer.ServeHTTP(res, req)

			gotStatus := res.Result().StatusCode
			gotContentType := res.Result().Header.Get("Content-Type")

			if gotStatus != tt.wantStatus {
				t.Errorf("got %d, want %d", gotStatus, tt.wantStatus)
			}

			if gotContentType != tt.wantContentType {
				t.Errorf("got %s, want %s", gotContentType, tt.wantContentType)
			}
		})
	}
}

func makeMockFS() fs.FS {
	// mock fs structure:
	// / (root)
	// ├─ a
	// ├─ b/
	// │  └─ c
	// ├─ d/
	// ├─ e
	// ├─ text.txt
	// ├─ some-csv
	// ├─ not-vid.mp4
	// ├─ empty
	// ├─ ?thing
	// └─ #thing
	mock := fstest.MapFS{
		"a":   {Data: []byte("a")},   // /a
		"b/c": {Data: []byte("b/c")}, // /b/c
		"d/":  {},                    // /d/

		"text.txt":    {Data: []byte("Hello world")},               // text file with .txt extension
		"unknown":     {Data: []byte(`<!DOCTYPE html>`)},           // html without .html extension
		"not-vid.mp4": {Data: []byte("I am not actually a video")}, // text file with .mp4 extension
		"empty":       {},                                          // empty file
		"?thing":      {},                                          // filename starting with &
		"#thing":      {},                                          // filename starting with #
	}
	return mock
}
