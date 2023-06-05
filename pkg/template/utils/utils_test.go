package utils_test

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"kafka-polygon/pkg/template/utils"
	"net/http"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tj/assert"
)

const (
	defaultContents = "I am a fake file"
)

type fakeFile struct {
	io.ReadSeeker
	fi     *fakeFileInfo
	path   string // as opened
	entpos int
}

func (f *fakeFile) Close() error               { return nil }
func (f *fakeFile) Stat() (fs.FileInfo, error) { return f.fi, nil }
func (f *fakeFile) Readdir(count int) ([]fs.FileInfo, error) {
	if !f.fi.dir {
		return nil, fs.ErrInvalid
	}

	var fis []fs.FileInfo

	limit := f.entpos + count
	if count <= 0 || limit > len(f.fi.ents) {
		limit = len(f.fi.ents)
	}

	for ; f.entpos < limit; f.entpos++ {
		fis = append(fis, f.fi.ents[f.entpos])
	}

	if len(fis) == 0 && count > 0 {
		return fis, io.EOF
	}

	return fis, nil
}

type fakeFileInfo struct {
	dir      bool
	basename string
	modtime  time.Time
	ents     []*fakeFileInfo
	contents string
	err      error
}

func (f *fakeFileInfo) Name() string       { return f.basename }
func (f *fakeFileInfo) Sys() any           { return nil }
func (f *fakeFileInfo) ModTime() time.Time { return f.modtime }
func (f *fakeFileInfo) IsDir() bool        { return f.dir }
func (f *fakeFileInfo) Size() int64        { return int64(len(f.contents)) }
func (f *fakeFileInfo) Mode() fs.FileMode {
	if f.dir {
		return 0755 | fs.ModeDir
	}

	return 0644
}

type fakeFS map[string]*fakeFileInfo

func (fsys fakeFS) Open(name string) (http.File, error) {
	name = path.Clean(name)
	f, ok := fsys[name]

	if !ok {
		return nil, fs.ErrNotExist
	}

	if f.err != nil {
		return nil, f.err
	}

	return &fakeFile{ReadSeeker: strings.NewReader(f.contents), fi: f, path: name}, nil
}

func TestWalkFS(t *testing.T) {
	t.Parallel()

	rootDir := "./testdata"
	pFile := "file"

	tests := []struct {
		name, escaped string
	}{
		{`file`, `data1234567890`},
	}

	// We put each test file in its own directory in the fakeFS so we can look at it in isolation.
	fileServ := make(fakeFS)
	for _, test := range tests {
		fileServ[test.name] = &fakeFileInfo{
			basename: test.name,
			contents: test.escaped,
		}
	}

	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info == nil || info.IsDir() {
			return nil
		}

		assert.True(t, info.Size() > 0)

		_, err = os.Open(fmt.Sprintf("%s/%s", rootDir, path))

		return err
	}

	err := utils.Walk(&fileServ, pFile, walkFn)
	require.NoError(t, err)
}

func TestWalkFSError(t *testing.T) {
	t.Parallel()

	expErr := errors.New("this dir")
	pFile := "file"

	tests := []struct {
		name, escaped string
	}{
		{`file`, `data1234567890`},
	}

	fileServ := make(fakeFS)

	for _, test := range tests {
		testFile := &fakeFileInfo{basename: test.name}
		fileServ[test.name] = &fakeFileInfo{
			dir:     true,
			modtime: time.Unix(1000000000, 0).UTC(),
			ents:    []*fakeFileInfo{testFile},
		}
	}

	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info == nil || info.IsDir() {
			return expErr
		}

		return nil
	}

	err := utils.Walk(&fileServ, pFile, walkFn)
	require.Error(t, err)
	assert.Equal(t, fmt.Sprintf("walk %s", expErr), err.Error())
}

func TestReadFileFS(t *testing.T) {
	t.Parallel()

	pFile := "file"

	tests := []struct {
		name, escaped string
	}{
		{`file`, `data1234567890`},
	}

	fileServ := make(fakeFS)

	for _, test := range tests {
		fileServ[test.name] = &fakeFileInfo{
			basename: test.name,
			contents: test.escaped,
		}
	}

	b, err := utils.ReadFile(pFile, &fileServ)
	require.NoError(t, err)
	assert.True(t, len(b) > 0)
	assert.Equal(t, "data1234567890", string(b))
}

func TestReadFileFSError(t *testing.T) {
	t.Parallel()

	pFile := "file"

	tests := []struct {
		name, escaped string
	}{
		{`file`, `data1234567890`},
	}

	// We put each test file in its own directory in the fakeFS so we can look at it in isolation.
	fileServ := make(fakeFS)

	for _, test := range tests {
		fileServ[test.name] = &fakeFileInfo{
			err: errors.New("not validsate file"),
		}
	}

	_, err := utils.ReadFile(pFile, &fileServ)
	require.Error(t, err)
	assert.Equal(t, "readFile::Open not validsate file", err.Error())
}

func TestWalkFSIsDir(t *testing.T) {
	t.Parallel()

	dirMod := time.Unix(123, 0).UTC()
	fileMod := time.Unix(1000000000, 0).UTC()
	pDir := "/testdata"

	fileB := &fakeFileInfo{
		basename: "b",
		modtime:  fileMod,
		contents: defaultContents,
	}

	fileA := &fakeFileInfo{
		basename: "a",
		modtime:  fileMod,
		contents: defaultContents,
	}

	fFS := &fakeFS{
		"/testdata": &fakeFileInfo{
			dir:     true,
			modtime: dirMod,
			ents: []*fakeFileInfo{
				fileA,
				fileB,
			},
		},
		"/testdata/a": fileA,
		"/testdata/b": fileB,
	}

	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info == nil || info.IsDir() {
			assert.Equal(t, true, info.IsDir())
			return nil
		}

		assert.True(t, info.Size() > 0)

		return err
	}

	err := utils.Walk(fFS, pDir, walkFn)
	require.NoError(t, err)
}

func TestReadFile(t *testing.T) {
	t.Parallel()

	pFile := "./testdata/file"

	b, err := utils.ReadFile(pFile, nil)
	require.NoError(t, err)
	assert.True(t, len(b) > 0)
	assert.Equal(t, "0123456789\n", string(b))
}

func TestReadFileError(t *testing.T) {
	t.Parallel()

	pFile := "./testdata/file2"

	b, err := utils.ReadFile(pFile, nil)
	require.Error(t, err)
	assert.True(t, len(b) == 0)
	assert.Equal(t, "readFile::ReadFile open ./testdata/file2: no such file or directory", err.Error())
}
