package file

import (
	"io/fs"
	"path/filepath"
	"strings"
)

func ListFiles(fsys fs.FS, root string) ([]FileF, error) {
	var out []FileF
	dirs, err := fs.ReadDir(fsys, root)
	if err != nil {
		return out, err
	}

	for _, dir := range dirs {
		openFunc := func() (fs.File, error) {
			fullpath := filepath.Join(root, dir.Name())
			return fsys.Open(fullpath)
		}
		f := FileF{dirEntry: dir, openFunc: openFunc}
		out = append(out, f)
	}

	return out, nil
}

func SearchFiles(fsys fs.FS, root string, search string) ([]FileF, error) {
	var out []FileF
	err := fs.WalkDir(fsys, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// skip the current dir
		if path == root {
			return nil
		}
		// if not the root "." dir, strip off the root prefix
		fullpath := path
		if root != "." {
			path = strings.TrimPrefix(path, root+"/")
		}
		// fmt.Println(path)
		// if not a match, skip
		if !strings.Contains(path, search) {
			return nil
		}
		// todo: put <mark> tags around the matched string
		// start, end, _ := strings.Cut(path, search)
		// customName := html.EscapeString(start) + "<mark>" + html.EscapeString(search) + "</mark>" + html.EscapeString(end)
		// customName := strings.Replace(path, search, "<mark>"+search+"</mark>", 1)
		openFunc := func() (fs.File, error) {
			return fsys.Open(fullpath)
		}
		f := FileF{dirEntry: d, openFunc: openFunc, customName: path, customHref: path}
		out = append(out, f)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}
