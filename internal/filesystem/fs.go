package filesystem

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
)

type WriterFS interface {
	fs.FS
	Mkdir(name string, perm fs.FileMode) error
	Remove(name string) error
	Rename(oldpath string, newpath string) error
	WriteFile(name string, data []byte, perm fs.FileMode) error
}

// FSys should comply with the fs.FSys interface (and fstest.TestFS).
// Therefore the Open method should reject any attempts to open
// files which do not satisfy fs.ValidPath.
// For consistency reasons, the other methods Mkdir,
// Remove, Rename, and WriteFile will also enforce this,
// even though it is not necessary for the os functions.
type FSys struct {
	path string
	root fs.FS
}

var _ WriterFS = &FSys{}
var errEmptyRoot = errors.New("filesystem: filesystem with empty root")

func New(path string) *FSys {
	return &FSys{path: path, root: os.DirFS(path)}
}

func New2(path string, root fs.FS) *FSys {
	return &FSys{path: path, root: root}
}

func (fsys *FSys) Open(name string) (fs.File, error) {
	// the os.DirFS Open function already handles checking
	// for valid name and valid root name
	return fsys.root.Open(name)
}

func (fsys *FSys) Mkdir(name string, perm fs.FileMode) error {
	if fsys.path == "" {
		return errEmptyRoot
	}
	if !fs.ValidPath(name) {
		return &fs.PathError{Op: "mkdir", Path: name, Err: fs.ErrInvalid}
	}
	fullpath := filepath.Join(fsys.path, name)
	// note: the os.MkdirAll function is idempotent:
	// it only returns an error if the directory cannot be created.
	// the directory already existing is NOT an error
	return os.MkdirAll(fullpath, perm)
}

func (fsys *FSys) Remove(name string) error {
	if fsys.path == "" {
		return errEmptyRoot
	}
	if !fs.ValidPath(name) {
		return &fs.PathError{Op: "remove", Path: name, Err: fs.ErrInvalid}
	}
	fullpath := filepath.Join(fsys.path, name)
	// note: the os.RemoveAll function is idempotent:
	// it only returns an error if there is an error when removing the file/s.
	// removing nonexistent files or empty dirs is NOT an error
	return os.RemoveAll(fullpath)
}

// unlike the os.Rename function, Rename should return an error
// if the newname already exists
func (fsys *FSys) Rename(oldname string, newname string) error {
	if fsys.path == "" {
		return errEmptyRoot
	}
	if !fs.ValidPath(oldname) {
		return &fs.PathError{Op: "mkdir", Path: oldname, Err: fs.ErrInvalid}
	}
	if !fs.ValidPath(oldname) {
		return &fs.PathError{Op: "mkdir", Path: newname, Err: fs.ErrInvalid}
	}
	_, err := fsys.Open(newname)
	// here we want an error, since that signifies the file does not exist
	if err != nil {
		return fmt.Errorf("fs: cannot rename %s to %s: %w", oldname, newname, os.ErrExist)
	}
	// now we can just call os.Rename
	oldfullpath := filepath.Join(fsys.path, oldname)
	newfullpath := filepath.Join(fsys.path, newname)
	return os.Rename(oldfullpath, newfullpath)
}

// unlike the os.WriteFile function, WriteFile should return an error
// if the file already exists
func (fsys *FSys) WriteFile(name string, data []byte, perm fs.FileMode) error {
	if fsys.path == "" {
		return errEmptyRoot
	}
	if !fs.ValidPath(name) {
		return &fs.PathError{Op: "writefile", Path: name, Err: fs.ErrInvalid}
	}
	_, err := fsys.Open(name)
	// here we want an error, since that signifies the file does not exist
	if err != nil {
		return fmt.Errorf("fs: cannot create %s: %w", name, os.ErrExist)
	}

	// todo: finish this function
	return os.WriteFile(name, data, perm)
}

// Clean takes a url path, strips leading and trailing "/",
// and cleans it to something like "path/to/file" for use in an fs.FS.
// As a special case, the root path "/" is cleaned to "."
func Clean(name string) string {
	name = path.Clean(name)
	if name == "/" {
		return "."
	}
	return name[1:]
}
