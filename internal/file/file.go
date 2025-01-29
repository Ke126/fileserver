package file

import (
	"io"
	"io/fs"
	"mime"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
)

// FileF is an fs.DirEntry wrapper which exposes a
// similar set of methods to describe a FileF in a formatted way.
// setting customName or customHref will override the default
// behavior of the Name and Href methods, which are to use the shortened
// name of the FileF (i.e. "my-FileF" instead of "path/to/my-FileF")
type FileF struct {
	dirEntry   fs.DirEntry
	openFunc   func() (fs.File, error)
	customName string
	customHref string
}

func (f FileF) IsDir() bool {
	return f.dirEntry.IsDir()
}

// Name returns the canonical name of a file:
// normal files must not have a trailing "/", and
// directories must have a trailing "/"
func (f FileF) Name() string {
	name := f.dirEntry.Name()
	if f.customName != "" {
		name = f.customName
	}
	if f.IsDir() {
		name += "/"
	}
	return name
}

// Href returns the relative href for the file,
// while escaping any "?" or "#" characters
func (f FileF) Href() string {
	href := f.dirEntry.Name()
	if f.customHref != "" {
		href = f.customHref
	}
	if f.IsDir() {
		href += "/"
	}
	return (&url.URL{Path: href}).String()
}

const timeFmtStr = "1/2/2006 3:04 PM"

// DateModified returns the modtime of the file
// in the format 1/2/2006 3:04 PM
func (f FileF) DateModified() string {
	i, err := f.dirEntry.Info()
	if err != nil {
		return ""
	}
	return i.ModTime().Format(timeFmtStr)
}

// MimeType returns the mimetype of the file following
// the same procedure as http.ServeContent for consistency, i.e. it
// first checks the file extension, and then sniffs the
// first 512 bytes of the file if necessary
func (f FileF) MimeType() string {
	if f.IsDir() {
		return "directory"
	}
	// first check extension if present
	ctype := mime.TypeByExtension(filepath.Ext(f.dirEntry.Name()))

	// otherwise sniff first 512 bytes
	if ctype == "" {
		file, err := f.openFunc()
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
func (f FileF) Size() string {
	if f.IsDir() {
		return ""
	}
	i, err := f.dirEntry.Info()
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
