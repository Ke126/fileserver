package filesystem

import (
	"io"
	"io/fs"
	"mime"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strconv"
)

// FileView is an fs.DirEntry wrapper which exposes a
// similar set of methods to describe a file in a formatted way
type FileView struct {
	dirEntry fs.DirEntry
	openFunc func() (fs.File, error)
}

func (fv FileView) IsDir() bool {
	return fv.dirEntry.IsDir()
}

// Name returns the canonical name of a file:
// normal files must not have a trailing "/", and
// directories must have a trailing "/"
func (fv FileView) Name() string {
	name := fv.dirEntry.Name()
	if fv.IsDir() {
		name += "/"
	}
	return name
}

// Href returns the relative href for the file,
// while escaping any "?" or "#" characters
func (fv FileView) Href() string {
	return (&url.URL{Path: fv.Name()}).String()
}

const timeFmtStr = "1/2/2006 3:04 PM"

// DateModified returns the modtime of the file
// in the format 1/2/2006 3:04 PM
func (fv FileView) DateModified() string {
	i, err := fv.dirEntry.Info()
	if err != nil {
		return ""
	}
	return i.ModTime().Format(timeFmtStr)
}

// MimeType returns the mimetype of the file following
// the same procedure as http.ServeContent for consistency, i.e. it
// first checks the file extension, and then sniffs the
// first 512 bytes of the file if necessary
func (fv FileView) MimeType() string {
	if fv.IsDir() {
		return "directory"
	}
	// first check extension if present
	ctype := mime.TypeByExtension(filepath.Ext(fv.Name()))

	// otherwise sniff first 512 bytes
	if ctype == "" {
		file, err := fv.openFunc()
		if err != nil {
			return ""
		}
		defer file.Close()

		rs, ok := file.(io.ReadSeeker)
		if !ok {
			return ""
		}

		var buf [512]byte
		n, _ := io.ReadFull(rs, buf[:])
		ctype = http.DetectContentType(buf[:n])
		_, err = rs.Seek(0, io.SeekStart)
		if err != nil {
			return ""
		}
	}
	return ctype
}

// Size returns the size of the file in a formatted string
// such as "5 B", "167 MB", or "8156 GB".
// A directory has no size
func (fv FileView) Size() string {
	if fv.IsDir() {
		return ""
	}
	i, err := fv.dirEntry.Info()
	if err != nil {
		return ""
	}
	size := int(i.Size())
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

func ListFiles(fsys fs.FS, name string) ([]FileView, error) {
	var out []FileView
	dirs, err := fs.ReadDir(fsys, name)
	if err != nil {
		return out, err
	}

	for _, dir := range dirs {
		openFunc := func() (fs.File, error) {
			fullpath := path.Join(name, dir.Name())
			return fsys.Open(fullpath)
		}
		f := FileView{dirEntry: dir, openFunc: openFunc}
		out = append(out, f)
	}

	return out, nil
}
