package fileserver

import (
	"archive/zip"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"path"
	"slices"
	"strings"

	"fileserver/internal/file"
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
			redirect := "../" + path.Base(r.URL.Path)
			if r.URL.RawQuery != "" {
				redirect += "?" + r.URL.RawQuery
			}
			http.Redirect(w, r, redirect, http.StatusMovedPermanently)
			return
		}
		// serve the file
		rs, ok := f.(io.ReadSeeker)
		if !ok {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		// if its a download, set the header
		if r.URL.Query().Has("download") {
			w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, s.Name()))
		}
		http.ServeContent(w, r, s.Name(), s.ModTime(), rs)
		return
	}

	// directory
	// redirect if path does not have a trailing slash
	if lastChar != '/' {
		w.Header()["Content-Type"] = nil
		redirect := path.Base(r.URL.Path) + "/"
		if r.URL.RawQuery != "" {
			redirect += "?" + r.URL.RawQuery
		}
		http.Redirect(w, r, redirect, http.StatusMovedPermanently)
		return
	}
	// serve the zipped directory
	if r.URL.Query().Has("download") {
		fh.serveZip(filename, s.Name(), w, r)
		return
	}
	// serve the directory listing page
	fh.listDirs(filename, w, r)
}

type dirListPage struct {
	Search      string
	Breadcrumbs []breadcrumb
	Files       []file.FileF
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

func (fh *fileHandler) serveZip(name string, shortname string, w http.ResponseWriter, r *http.Request) {
	sub, err := fs.Sub(fh.fs, name)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.zip"`, shortname))

	zipper := zip.NewWriter(w)
	err = zipper.AddFS(sub)
	if err != nil {
		http.NotFound(w, r)
	}

	zipper.Close()
}

func (fh *fileHandler) listDirs(name string, w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	var fileViews []file.FileF
	var err error
	if q == "" {
		fileViews, err = file.ListFiles(fh.fs, name)
	} else {
		fileViews, err = file.SearchFiles(fh.fs, name, q)
	}
	if err != nil {
		http.NotFound(w, r)
		return
	}

	breadcrumbs := makeBreadcrumbs(r.URL.Path)
	page := dirListPage{
		Search:      q,
		Files:       fileViews,
		Breadcrumbs: breadcrumbs,
	}

	template := template.Must(template.New("dirlist.html").ParseFiles("_static/dirlist.html"))

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	template.Execute(w, page)
}
