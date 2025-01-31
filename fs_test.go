package fileserver_test

import (
	"io/fs"
	"os"
	"testing"
	"testing/fstest"

	"github.com/ke126/fileserver"
)

func TestOpen(t *testing.T) {
	t.Run("test open", func(t *testing.T) {
		dir := t.TempDir()
		fsys := fileserver.NewFS(dir)

		// TestFS automatically checks . /. ./. /
		if err := fstest.TestFS(fsys); err != nil {
			t.Fatal(err)
		}

		FillDir(dir)

		if err := fstest.TestFS(fsys, "a", "b/c", "b/d", "e"); err != nil {
			t.Fatal(err)
		}

		// check for some file that doesn't exist, should fail
		if err := fstest.TestFS(fsys, "a", "b/c", "b/d", "e", "f"); err == nil {
			t.Fatal("wanted fail")
		}

		_, err := fsys.Open(".")
		if err != nil {
			t.Errorf("open . should succeed")
		}
	})
}

func FillDir(name string) {
	// temp dir structure:
	// / (root)
	// ├─ a
	// ├─ b/
	// │  ├─ c
	// │  └─ d/
	// └─ e/

	var fileperms fs.FileMode = 0644
	var dirperms fs.FileMode = 0755

	os.WriteFile(name+"/a", []byte("this is file a"), fileperms)
	os.Mkdir(name+"/b", dirperms)
	os.WriteFile(name+"/b/c", []byte("this is file c"), fileperms)
	os.Mkdir(name+"/b/d/", dirperms)
	os.Mkdir(name+"/e/", dirperms)
}
