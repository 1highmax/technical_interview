package shred

import (
	"crypto/rand"
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"strings"
	"testing"
)

func createTempFile(t *testing.T, content []byte) (string, func()) {
	tmpfile, err := ioutil.TempFile("", "testfile")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	if content != nil {
		if _, err := tmpfile.Write(content); err != nil {
			t.Fatalf("Failed to write to temp file: %v", err)
		}
	}

	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	return tmpfile.Name(), func() {
		os.Remove(tmpfile.Name())
	}
}

func TestShred(t *testing.T) {
	// Test 1: Basic Functionality
	testFile, cleanup := createTempFile(t, []byte("some data"))
	defer cleanup()

	err := Shred(testFile, func(name string) (File, error) {
		f, err := os.OpenFile(name, os.O_WRONLY, 0)
		if err != nil {
			return nil, err
		}
		return &RealFile{f}, nil
	})
	if err != nil {
		t.Errorf("Shred failed: %v", err)
	}

	_, err = os.Stat(testFile)
	if !os.IsNotExist(err) {
		t.Errorf("File not deleted")
	}

	// Test 2: Non-Existent File
	err = Shred("nonexistentfile", func(name string) (File, error) {
		return &RealFile{}, errors.New("file does not exist")
	})
	if err == nil {
		t.Errorf("Expected error for non-existent file, got nil")
	}

	// Test 3: Permission Denied
	usr, err := user.Current()
	if err != nil {
		t.Fatalf("Failed to get current user: %v", err)
	}

	if usr.Uid == "0" {
		t.Skip("Skipping test as the current user is root")
	}

	testFile, cleanup = createTempFile(t, []byte("some data"))
	defer cleanup()

	if err := os.Chmod(testFile, 0444); err != nil { // Read-only
		t.Fatalf("Failed to change file permissions: %v", err)
	}

	err = Shred(testFile, func(name string) (File, error) {
		f, err := os.OpenFile(name, os.O_WRONLY, 0)
		if err != nil {
			return nil, err
		}
		return &RealFile{f}, nil
	})
	if err == nil {
		t.Errorf("Expected permission error, got nil")
	}

	// Cleanup permission change for deletion
	os.Chmod(testFile, 0644)

	// Test 4: Empty File
	testFile, cleanup = createTempFile(t, []byte(""))
	defer cleanup()

	err = Shred(testFile, func(name string) (File, error) {
		f, err := os.OpenFile(name, os.O_WRONLY, 0)
		if err != nil {
			return nil, err
		}
		return &RealFile{f}, nil
	})
	if err != nil {
		t.Errorf("Shred failed: %v", err)
	}

	_, err = os.Stat(testFile)
	if !os.IsNotExist(err) {
		t.Errorf("File not deleted")
	}

	// Test 5: Large File
	largeData := make([]byte, 10*1024*1024) // 10 MB
	_, err = rand.Read(largeData)
	if err != nil {
		t.Fatalf("Failed to generate random data: %v", err)
	}

	testFile, cleanup = createTempFile(t, largeData)
	defer cleanup()

	err = Shred(testFile, func(name string) (File, error) {
		f, err := os.OpenFile(name, os.O_WRONLY, 0)
		if err != nil {
			return nil, err
		}
		return &RealFile{f}, nil
	})
	if err != nil {
		t.Errorf("Shred failed: %v", err)
	}

	_, err = os.Stat(testFile)
	if !os.IsNotExist(err) {
		t.Errorf("File not deleted")
	}

	// Test 6: Open File Error
	err = Shred("mockfile", func(name string) (File, error) {
		return nil, errors.New("failed to open file")
	})
	if err == nil {
		t.Errorf("Expected error for open file failure, got nil")
	}

	// Test 7: File Stat Error
	err = Shred(testFile, func(name string) (File, error) {
		mockFile := &MockFile{}
		mockFile.statError = errors.New("failed to stat file")
		return mockFile, nil
	})
	if err == nil {
		t.Errorf("Expected error for file stat failure, got nil")
	}

	// Test 8: File Sync Error
	err = Shred(testFile, func(name string) (File, error) {
		mockFile := &MockFile{}
		mockFile.syncError = errors.New("failed to sync file")
		return mockFile, nil
	})
	if err == nil {
		t.Errorf("Expected error for file sync failure, got nil")
	}

	// Test 9: File Write Error
	err = Shred(testFile, func(name string) (File, error) {
		mockFile := &MockFile{}
		mockFile.writeError = errors.New("failed to write file")
		return mockFile, nil
	})
	if err == nil {
		t.Errorf("Expected error for file write failure, got nil")
	}

	// Test 10: File Close Error
	mockFile := &MockFile{closeError: errors.New("failed to close file")}
	err = Shred("mockfile", func(name string) (File, error) {
		return mockFile, nil
	})
	if err == nil {
		t.Errorf("Expected error for file close failure, got nil")
	}

	// Test 11: File Remove Error
	mockFile = &MockFile{removeError: errors.New("failed to remove file")}
	err = Shred("mockfile", func(name string) (File, error) {
		return mockFile, nil
	})
	if err == nil {
		t.Errorf("Expected error for file remove failure, got nil")
	}
}

func TestShredWrites(t *testing.T) {
	const fileSize int64 = 1024 // 1 KB file size
	mockFile := &MockFile{size: fileSize}

	err := Shred("mockfile", func(name string) (File, error) {
		return mockFile, nil
	})
	if err != nil {
		t.Fatalf("Shred failed: %v", err)
	}

	expectedWrites := 3
	if mockFile.writeCount != expectedWrites {
		t.Errorf("Expected %d writes, but got %d", expectedWrites, mockFile.writeCount)
	}
	if !mockFile.closed {
		t.Errorf("File was not closed")
	}
	if !mockFile.removed {
		t.Errorf("File was not removed")
	}
}

func TestShredFile(t *testing.T) {
	// Test 1: Basic Functionality
	testFile, cleanup := createTempFile(t, []byte("some data"))
	defer cleanup()

	err := ShredFile(testFile)
	if err != nil {
		t.Errorf("ShredFile failed: %v", err)
	}

	_, err = os.Stat(testFile)
	if !os.IsNotExist(err) {
		t.Errorf("File not deleted")
	}

	// Test 2: Non-Existent File
	err = ShredFile("nonexistentfile")
	if err == nil {
		t.Errorf("Expected error for non-existent file, got nil")
	}
}

func TestOpenFile(t *testing.T) {
	// Test 1: Open Existing File
	testFile, cleanup := createTempFile(t, []byte("some data"))
	defer cleanup()

	file, err := openFile(testFile)
	if err != nil {
		t.Fatalf("Failed to open existing file: %v", err)
	}
	defer file.Close()

	// Check if the file was opened correctly
	if _, err := file.Stat(); err != nil {
		t.Fatalf("Failed to stat opened file: %v", err)
	}

	// Test 2: Open Non-Existent File
	_, err = openFile("nonexistentfile")
	if err == nil {
		t.Errorf("Expected error when opening non-existent file, got nil")
	}

	// Test 3: Permission Denied
	usr, err := user.Current()
	if err != nil {
		t.Fatalf("Failed to get current user: %v", err)
	}

	if usr.Uid != "0" { // Skip this test if running as root
		testFile, cleanup = createTempFile(t, []byte("some data"))
		defer cleanup()

		if err := os.Chmod(testFile, 0444); err != nil { // Read-only
			t.Fatalf("Failed to change file permissions: %v", err)
		}

		_, err = openFile(testFile)
		if err == nil {
			t.Errorf("Expected permission error, got nil")
		}
	}
}

func TestMainArgumentHandling(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedOutput string
		expectedError  bool
		setup          func(t *testing.T) (string, func()) // Function to set up the environment for the test
	}{
		{
			name:           "NoArguments",
			args:           []string{},
			expectedOutput: "Usage: ",
			expectedError:  true,
		},
		{
			name:           "ValidFile",
			expectedOutput: "File successfully shredded",
			expectedError:  false,
			setup: func(t *testing.T) (string, func()) {
				return createTempFile(t, []byte("some data"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cleanup func()
			if tt.setup != nil {
				var path string
				path, cleanup = tt.setup(t)
				tt.args = append(tt.args, path)
			}

			if cleanup != nil {
				defer cleanup()
			}

			cmd := exec.Command("go", append([]string{"run", "../main.go"}, tt.args...)...)
			output, err := cmd.CombinedOutput()

			if (err != nil) != tt.expectedError {
				t.Errorf("expected error: %v, got error: %v", tt.expectedError, err)
			}

			if !strings.Contains(string(output), tt.expectedOutput) {
				t.Errorf("expected output: %s, got output: %s", tt.expectedOutput, output)
			}
		})
	}
}
