package fileserver

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"path"
	"slices"
	"strings"

	"fileserver/internal/filesystem"
)

var _ http.Handler = &fileHandler{}

type fileHandler struct {
	fs WriterFS
}

type WriterFS interface {
	fs.FS
	Mkdir(name string, perm fs.FileMode) error
	Remove(name string) error
	Rename(oldpath string, newpath string) error
	WriteFile(name string, data []byte, perm fs.FileMode) error
}

func FileServer(fs WriterFS) *fileHandler {
	return &fileHandler{fs}
}

func (fh *fileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "/_static/styles.css") {
		http.ServeFile(w, r, "_static/styles.css")
		return
	}

	filename := filesystem.Clean(r.URL.Path)
	fmt.Println(r.URL.Path, filename)

	f, err := fh.fs.Open(filename)
	if err != nil {
		fmt.Println(err)
		http.NotFound(w, r)
		return
	}
	defer f.Close()

	s, err := f.Stat()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	lastChar := r.URL.Path[len(r.URL.Path)-1]
	// normal file
	if !s.IsDir() {
		// redirect if path has a trailing slash
		if lastChar == '/' {
			w.Header()["Content-Type"] = nil
			http.Redirect(w, r, "../"+path.Base(r.URL.Path), http.StatusMovedPermanently)
			return
		}
		// serve the file
		rs, ok := f.(io.ReadSeeker)
		if !ok {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		http.ServeContent(w, r, s.Name(), s.ModTime(), rs)
		return
	}

	// directory
	// redirect if path does not have a trailing slash
	if lastChar != '/' {
		w.Header()["Content-Type"] = nil
		http.Redirect(w, r, path.Base(r.URL.Path)+"/", http.StatusMovedPermanently)
		return
	}
	// serve the directory listing page
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fh.listDirs(filename, w, r)
}

type dirListPage struct {
	Search      string
	Breadcrumbs []breadcrumb
	Files       []filesystem.FileView
}

type breadcrumb struct {
	Name string
	Href string
}

func makeBreadcrumbs(urlPath string) []breadcrumb {
	var breadcrumbs []breadcrumb

	// urlPath always has a leading / and no trailing /
	urlPath = path.Clean(urlPath)
	// if urlPath == "/", there is only one root breadcrumb
	if urlPath == "/" {
		breadcrumbs = append(breadcrumbs, breadcrumb{
			Name: "/",
			Href: ".",
		})
		return breadcrumbs
	}
	parts := strings.Split(urlPath, "/")
	href := "."
	for _, e := range slices.Backward(parts) {
		breadcrumbs = append(breadcrumbs, breadcrumb{
			Name: e + "/",
			Href: href,
		})
		href += "/.."
	}
	slices.Reverse(breadcrumbs)
	return breadcrumbs
}

func (fh *fileHandler) listDirs(name string, w http.ResponseWriter, r *http.Request) {
	fileViews, err := filesystem.ListFiles(fh.fs, name)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	breadcrumbs := makeBreadcrumbs(r.URL.Path)
	page := dirListPage{
		Search:      r.URL.Query().Get("q"),
		Files:       fileViews,
		Breadcrumbs: breadcrumbs,
	}

	template := template.Must(template.New("dirlist.html").ParseFiles("_static/dirlist.html"))

	w.WriteHeader(http.StatusOK)
	template.Execute(w, page)
}
