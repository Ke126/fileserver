package fileserver

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"net/url"
	"path"
	"slices"
	"strconv"
	"strings"
)

var _ http.Handler = &fileHandler{}

type fileHandler struct {
	root http.FileSystem
}

func FileServer(root http.FileSystem) *fileHandler {
	return &fileHandler{root}
}

func (fs *fileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "/_static/styles.css") {
		http.ServeFile(w, r, "_static/styles.css")
		return
	}
	// urlPath should always have a leading "/" and no trailing "/"
	urlPath := r.URL.Path
	fmt.Println(urlPath)
	if !strings.HasPrefix(urlPath, "/") {
		urlPath = "/" + urlPath
	}
	urlPath = path.Clean(urlPath)
	fmt.Println(urlPath)
	// w.WriteHeader(200)
	// return

	f, err := fs.root.Open(urlPath)
	if err != nil {
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
			http.Redirect(w, r, "../"+path.Base(urlPath), http.StatusMovedPermanently)
			return
		}
		// serve the file
		// w.WriteHeader(http.StatusOK)
		// io.CopyN(w, f, s.Size())
		http.ServeContent(w, r, s.Name(), s.ModTime(), f)
		return
	}

	// directory
	// redirect if path does not have a trailing slash
	if lastChar != '/' {
		w.Header()["Content-Type"] = nil
		http.Redirect(w, r, path.Base(urlPath)+"/", http.StatusMovedPermanently)
		return
	}
	// serve the directory listing page
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fs.dirList(f, urlPath, w, r)
}

type dirListInfo struct {
	Breadcrumbs []breadcrumb
	Files       []fileInfoView
}

type breadcrumb struct {
	Name string
	Href string
}

func makeBreadcrumbs(path string) []breadcrumb {
	// path always has a leading / and no trailing /
	var breadcrumbs []breadcrumb

	// if path == "/", there is only one root breadcrumb
	if path == "/" {
		breadcrumbs = append(breadcrumbs, breadcrumb{
			Name: "/",
			Href: ".",
		})
		return breadcrumbs
	}
	parts := strings.Split(path, "/")
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

type fileInfoView struct {
	fileInfo fs.FileInfo
}

func (f fileInfoView) Name() string {
	if f.fileInfo.IsDir() {
		return f.fileInfo.Name() + "/"
	}
	return f.fileInfo.Name()
}

func (f fileInfoView) Href() string {
	return (&url.URL{Path: f.Name()}).String()
}

func (f fileInfoView) DateModified() string {
	return f.fileInfo.ModTime().Format("1/2/2006 3:04 PM")
}

func (f fileInfoView) MimeType() string {
	if f.fileInfo.IsDir() {
		return "directory"
	}
	return ""
}

func (f fileInfoView) Size() string {
	if f.fileInfo.IsDir() {
		return ""
	}
	size := int(f.fileInfo.Size())
	if size < 1000 {
		return strconv.Itoa(size) + " B"
	}
	size /= 1000
	if size < 1000 {
		return strconv.Itoa(size) + " KB"
	}
	size /= 1000
	if size < 1000 {
		return strconv.Itoa(size) + " MB"
	}
	size /= 1000
	return strconv.Itoa(size) + " GB"
}

func (fs *fileHandler) dirList(file http.File, fullName string, w http.ResponseWriter, r *http.Request) {
	dir, err := file.Readdir(-1)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	files := make([]fileInfoView, 0)
	for _, f := range dir {
		file := fileInfoView{f}
		files = append(files, file)
	}
	breadcrumbs := makeBreadcrumbs(fullName)
	view := dirListInfo{
		Files:       files,
		Breadcrumbs: breadcrumbs,
	}

	template := template.Must(template.New("dirlist.html").ParseFiles("_static/dirlist.html"))

	w.WriteHeader(http.StatusOK)
	template.Execute(w, view)
}
