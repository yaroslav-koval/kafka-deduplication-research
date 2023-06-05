package utils

import (
	"context"
	"io"
	"kafka-polygon/pkg/cerror"
	"net/http"
	"os"
	pathpkg "path"
	"path/filepath"
	"sort"
)

// Walk walks the filesystem rooted at root, calling walkFn for each file or
// directory in the filesystem, including root. All errors that arise visiting files
// and directories are filtered by walkFn. The files are walked in lexical
// order.
func Walk(fs http.FileSystem, root string, walkFn filepath.WalkFunc) error {
	if fs == nil {
		return cerror.NewF(context.Background(),
			cerror.KindInternal, "walk::FS not set http.FileSystem (%T)", fs).
			LogError()
	}

	info, err := stat(fs, root)
	if err != nil {
		return cerror.NewF(context.Background(),
			cerror.KindInternal, "walk::stat %s", walkFn(root, nil, err)).
			LogError()
	}

	err = walk(fs, root, info, walkFn)
	if err != nil {
		return cerror.NewF(context.Background(),
			cerror.KindInternal, "walk %s", err).LogError()
	}

	return nil
}

// ReadFile returns the raw content of a file
func ReadFile(path string, fs http.FileSystem) ([]byte, error) {
	if fs != nil {
		file, err := fs.Open(path)
		if err != nil {
			return nil, cerror.NewF(context.Background(),
				cerror.KindInternal, "readFile::Open %s", err).LogError()
		}
		defer file.Close()

		b, err := io.ReadAll(file)

		if err != nil {
			return nil, cerror.NewF(context.Background(),
				cerror.KindInternal, "readFile::ReadAll %s", err).LogError()
		}

		return b, nil
	}

	bf, err := os.ReadFile(path)

	if err != nil {
		return nil, cerror.NewF(context.Background(),
			cerror.KindInternal, "readFile::ReadFile %s", err).LogError()
	}

	return bf, nil
}

// readDirNames reads the directory named by dirname and returns
// a sorted list of directory entries.
func readDirNames(fs http.FileSystem, dirname string) ([]string, error) {
	fis, err := readDir(fs, dirname)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(fis))
	for i := range fis {
		names[i] = fis[i].Name()
	}

	sort.Strings(names)

	return names, nil
}

// walk recursively descends path, calling walkFn.
//
//nolint:nestif
func walk(fs http.FileSystem, path string, info os.FileInfo, walkFn filepath.WalkFunc) error {
	err := walkFn(path, info, nil)
	if err != nil {
		if info.IsDir() && err == filepath.SkipDir {
			return nil
		}

		return err
	}

	if !info.IsDir() {
		return nil
	}

	names, err := readDirNames(fs, path)
	if err != nil {
		return walkFn(path, info, err)
	}

	for _, name := range names {
		filename := pathpkg.Join(path, name)
		fileInfo, err := stat(fs, filename)

		if err != nil {
			err = walkFn(filename, fileInfo, err)
			if err != nil && err != filepath.SkipDir {
				return err
			}
		} else {
			err = walk(fs, filename, fileInfo, walkFn)
			if err != nil {
				if !fileInfo.IsDir() || err != filepath.SkipDir {
					return err
				}
			}
		}
	}

	return nil
}

// readDir reads the contents of the directory associated with file and
// returns a slice of FileInfo values in directory order.
func readDir(fs http.FileSystem, name string) ([]os.FileInfo, error) {
	f, err := fs.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return f.Readdir(0)
}

// stat returns the FileInfo structure describing file.
func stat(fs http.FileSystem, name string) (os.FileInfo, error) {
	f, err := fs.Open(name)

	if err != nil {
		return nil, err
	}

	defer f.Close()

	return f.Stat()
}
