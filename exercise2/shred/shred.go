// Package shred provides functionality to securely delete files by overwriting
// their contents multiple times before removing them from the filesystem.
package shred

import (
	"crypto/rand"
	"os"
)

// File is an interface representing a file that can be written to, synced,
// closed, and removed.
type File interface {
	WriteAt(b []byte, off int64) (n int, err error)
	Sync() error
	Stat() (os.FileInfo, error)
	Close() error
	Remove() error
}

// RealFile implements the File interface for an os.File.
type RealFile struct {
	*os.File
}

// Remove deletes the file from the filesystem.
func (rf *RealFile) Remove() error {
	return os.Remove(rf.Name())
}

// openFile opens a file for writing and returns a File interface.
func openFile(name string) (File, error) {
	f, err := os.OpenFile(name, os.O_WRONLY, 0)
	if err != nil {
		return nil, err
	}
	return &RealFile{f}, nil
}

// Shred overwrites the contents of the specified file multiple times with random data
// and then deletes the file.
func Shred(path string, openFile func(name string) (File, error)) error {
	const passes = 3

	// Open the file
	file, err := openFile(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Get the file size
	fi, err := file.Stat()
	if err != nil {
		return err
	}
	size := fi.Size()

	// Overwrite the file with random data
	buf := make([]byte, size)
	for i := 0; i < passes; i++ {
		rand.Read(buf)

		_, err = file.WriteAt(buf, 0)
		if err != nil {
			return err
		}

		// Sync the file to ensure data is written to disk
		err = file.Sync()
		if err != nil {
			return err
		}
	}

	// Close the file before deleting
	err = file.Close()
	if err != nil {
		return err
	}

	// Delete the file
	err = file.Remove()
	if err != nil {
		return err
	}

	return nil
}

// ShredFile opens the specified file and shreds it.
func ShredFile(path string) error {
	return Shred(path, openFile)
}
