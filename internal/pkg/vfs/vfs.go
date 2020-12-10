package vfs

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
)

type VFS interface {
	ReadFile(path string) ([]byte, error)
	Glob(pattern string) ([]string, error)
}

type LinuxFS struct{}

func (_ LinuxFS) ReadFile(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

func (_ LinuxFS) Glob(pattern string) ([]string, error) {
	return filepath.Glob(pattern)
}

type ReadFileResult struct {
	Data []byte
	Err  error
}

type GlobResult struct {
	Matches []string
	Err     error
}

type FakeFS struct {
	FileContents map[string]ReadFileResult
	GlobResults  map[string]GlobResult
}

func (ff FakeFS) ReadFile(path string) ([]byte, error) {
	if res, ok := ff.FileContents[path]; ok {
		return res.Data, res.Err
	}
	return nil, fmt.Errorf("fakefs: readfile: unregistered path %q", path)
}

func (ff FakeFS) Glob(pattern string) (matches []string, err error) {
	if res, ok := ff.GlobResults[pattern]; ok {
		return res.Matches, res.Err
	}
	return nil, fmt.Errorf("fakefs: glob: unregistered path %q", pattern)
}
