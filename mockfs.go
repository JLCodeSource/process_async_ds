package main

import (
	"io"
	"io/fs"
	"os"
	"sort"
	"time"
)

// MockFS is a mocking filesystem for testing purposes
type MockFS []*MockFile

// MockFile is the minimal struct of a fs.FS File
type MockFile struct {
	FS      MockFS
	isDir   bool
	modTime time.Time
	mode    fs.FileMode
	name    string
	size    int64
	sys     interface{}
}

// Open is a mock implementation of File.Open
func (mfs MockFS) Open(name string) (fs.File, error) {
	for _, f := range mfs {
		if f.Name() == name {
			return f, nil
		}
	}

	if len(mfs) > 0 {
		return mfs[0].FS.Open(name)
	}

	return nil, &fs.PathError{
		Op:   "read",
		Path: name,
		Err:  os.ErrNotExist,
	}
}

// ReadDir is a mock implementation of File.ReadDir
func (mfs MockFS) ReadDir(n int) ([]fs.DirEntry, error) {
	list := make([]fs.DirEntry, 0, len(mfs))

	for _, v := range mfs {
		list = append(list, v)
	}

	sort.Slice(list, func(a, b int) bool {
		return list[a].Name() > list[b].Name()
	})

	if n < 0 {
		return list, nil
	}

	if n > len(list) {
		return list, io.EOF
	}
	return list[:n], io.EOF
}

// Name is a mock impementation of File.Name
func (m *MockFile) Name() string {
	return m.name
}

// IsDir is a mock implementation of File.IsDir
func (m *MockFile) IsDir() bool {
	return m.isDir
}

// Info is a mock implementation of File.Info
func (m *MockFile) Info() (fs.FileInfo, error) {
	return m.Stat()
}

// Stat is a mock implementation of File.Stat
func (m *MockFile) Stat() (fs.FileInfo, error) {
	return m, nil
}

// Size is a mock implemntation of File.Size
func (m *MockFile) Size() int64 {
	return m.size
}

// Mode is a mock implementation of File.Mode
func (m *MockFile) Mode() os.FileMode {
	return m.mode
}

// ModTime is a mock implementation of File.ModTime
func (m *MockFile) ModTime() time.Time {
	return m.modTime
}

// Sys is a mock implementation of File.Sys
func (m *MockFile) Sys() interface{} {
	return m.sys
}

// Type is a mock implementation of File.Type
func (m *MockFile) Type() fs.FileMode {
	return m.Mode().Type()
}

// Reap is a mock implementation of File.Read
// N.B. not implemented
func (m *MockFile) Read(p []byte) (int, error) {
	panic("not implemented")
}

// Close is a mock implementation of File.Close
func (m *MockFile) Close() error {
	return nil
}

// ReadDir is a mock implementation of File.ReadDir
func (m *MockFile) ReadDir(n int) ([]fs.DirEntry, error) {
	if !m.IsDir() {
		return nil, os.ErrNotExist
	}

	if m.FS == nil {
		return nil, nil
	}
	return m.FS.ReadDir(n)
}

// NewFile is a helper function to add MockFiles to MockFS
// N.B. Not fully implemented
func NewFile(mf MockFile) *MockFile {
	return &MockFile{
		name:    mf.name,
		modTime: mf.modTime,
		FS:      mf.FS,
		size:    mf.size,
	}
}

// NewDir is a helper function to add MockDirs to MockFS
func NewDir(name string, files ...*MockFile) *MockFile {
	return &MockFile{
		FS:    files,
		isDir: true,
		name:  name,
	}
}
