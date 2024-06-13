package shred

import (
	"os"
	"time"
)

// MockFile is a mock implementation of the File interface used for testing.
type MockFile struct {
	size        int64
	writeCount  int
	writeError  error
	syncError   error
	statError   error
	closeError  error
	removeError error
	closed      bool
	removed     bool
}

// WriteAt mocks the WriteAt method of the File interface.
func (m *MockFile) WriteAt(b []byte, off int64) (n int, err error) {
	if m.writeError != nil {
		return 0, m.writeError
	}
	m.writeCount++
	return len(b), nil
}

// Sync mocks the Sync method of the File interface.
func (m *MockFile) Sync() error {
	return m.syncError
}

// Stat mocks the Stat method of the File interface.
func (m *MockFile) Stat() (os.FileInfo, error) {
	if m.statError != nil {
		return nil, m.statError
	}
	return &MockFileInfo{size: m.size}, nil
}

// Close mocks the Close method of the File interface.
func (m *MockFile) Close() error {
	if m.closeError != nil {
		return m.closeError
	}
	m.closed = true
	return nil
}

// Remove mocks the Remove method of the File interface.
func (m *MockFile) Remove() error {
	if m.removeError != nil {
		return m.removeError
	}
	m.removed = true
	return nil
}

// MockFileInfo is a mock implementation of the os.FileInfo interface used for testing.
type MockFileInfo struct {
	size int64
}

func (m *MockFileInfo) Name() string       { return "mockfile" }
func (m *MockFileInfo) Size() int64        { return m.size }
func (m *MockFileInfo) Mode() os.FileMode  { return 0 }
func (m *MockFileInfo) ModTime() time.Time { return time.Now() }
func (m *MockFileInfo) IsDir() bool        { return false }
func (m *MockFileInfo) Sys() interface{}   { return nil }
