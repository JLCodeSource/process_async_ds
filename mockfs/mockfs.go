package mockfs

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
	// MFModTime, MFName, MFSize are public exports to set the internals
	MFModTime  time.Time
	MFName     string
	MFSize     int64
	MFFileInfo fs.FileInfo
}

// Stat is a mock implementation of File.Stat
// N.B. It has been updated to allow Public setting of modTime
func (m *MockFile) Stat() (fs.FileInfo, error) {
	// Allow direct setting of m.modTime from outside package
	if !m.MFModTime.IsZero() {
		m.modTime = m.MFModTime
	}
	if m.MFName != "" {
		m.name = m.MFName
	}
	if m.MFSize != 0 {
		m.size = m.MFSize
	}

	return m, nil
}

// Name is a mock impementation of File.Name
// N.B. It has been updated to allow Public setting of name
func (m *MockFile) Name() string {
	if m.MFName != "" {
		m.name = m.MFName
	}
	return m.name
}

// Size is a mock implemntation of File.Size
// N.B. It has been updated to allow Public setting of Size
func (m *MockFile) Size() int64 {
	if m.MFSize != 0 {
		m.size = m.MFSize
	}
	return m.size
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

// IsDir is a mock implementation of File.IsDir
func (m *MockFile) IsDir() bool {
	return m.isDir
}

// Info is a mock implementation of File.Info
func (m *MockFile) Info() (fs.FileInfo, error) {
	return m.Stat()
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
func NewFile(m MockFile) *MockFile {
	return &MockFile{
		name:    m.name,
		modTime: m.modTime,
		FS:      m.FS,
		size:    m.size,
		// Added to modify m.modtime, m.size & m.name publicly
		MFModTime: m.MFModTime,
		MFName:    m.MFName,
		MFSize:    m.MFSize,
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
