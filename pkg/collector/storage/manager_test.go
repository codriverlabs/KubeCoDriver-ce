package storage

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestNewManager(t *testing.T) {
	tests := []struct {
		name       string
		basePath   string
		dateFormat string
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "valid configuration",
			basePath:   t.TempDir(),
			dateFormat: "2006/01/02",
			wantErr:    false,
		},
		{
			name:       "empty date format",
			basePath:   t.TempDir(),
			dateFormat: "",
			wantErr:    true,
			errMsg:     "dateFormat is required",
		},
		{
			name:       "creates base directory",
			basePath:   filepath.Join(t.TempDir(), "new", "path"),
			dateFormat: "2006-01-02",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, err := NewManager(tt.basePath, tt.dateFormat)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewManager() expected error, got nil")
				} else if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("NewManager() error = %v, want %v", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("NewManager() unexpected error = %v", err)
				return
			}

			if mgr == nil {
				t.Error("NewManager() returned nil manager")
				return
			}

			if mgr.basePath != tt.basePath {
				t.Errorf("NewManager() basePath = %v, want %v", mgr.basePath, tt.basePath)
			}

			if mgr.dateFormat != tt.dateFormat {
				t.Errorf("NewManager() dateFormat = %v, want %v", mgr.dateFormat, tt.dateFormat)
			}

			// Verify directory was created
			if _, err := os.Stat(tt.basePath); os.IsNotExist(err) {
				t.Errorf("NewManager() did not create base directory")
			}
		})
	}
}

func TestSaveProfile(t *testing.T) {
	tests := []struct {
		name       string
		dateFormat string
		metadata   ProfileMetadata
		content    string
		checkPath  func(basePath string, metadata ProfileMetadata) (string, error)
	}{
		{
			name:       "hierarchical date structure",
			dateFormat: "2006/01/02",
			metadata: ProfileMetadata{
				Namespace:       "default",
				AppLabel:        "app-nginx",
				CoDriverJobName: "profile-job",
				Filename:        "output.txt",
			},
			content: "test profile data",
			checkPath: func(basePath string, metadata ProfileMetadata) (string, error) {
				// Find file in hierarchical structure: basePath/namespace/label/codriverjob/year/month/day/filename
				pattern := filepath.Join(basePath, metadata.Namespace, metadata.AppLabel, metadata.CoDriverJobName, "*", "*", "*", metadata.Filename)
				matches, err := filepath.Glob(pattern)
				if err != nil {
					return "", err
				}
				if len(matches) == 0 {
					return "", nil
				}
				return matches[0], nil
			},
		},
		{
			name:       "flat date structure",
			dateFormat: "2006-01-02",
			metadata: ProfileMetadata{
				Namespace:       "production",
				AppLabel:        "env-prod",
				CoDriverJobName: "profile-prod",
				Filename:        "perf.data",
			},
			content: "performance data",
			checkPath: func(basePath string, metadata ProfileMetadata) (string, error) {
				// Find file in flat structure: basePath/namespace/label/codriverjob/date/filename
				pattern := filepath.Join(basePath, metadata.Namespace, metadata.AppLabel, metadata.CoDriverJobName, "*", metadata.Filename)
				matches, err := filepath.Glob(pattern)
				if err != nil {
					return "", err
				}
				if len(matches) == 0 {
					return "", nil
				}
				return matches[0], nil
			},
		},
		{
			name:       "unknown app label",
			dateFormat: "2006/01/02",
			metadata: ProfileMetadata{
				Namespace:       "default",
				AppLabel:        "unknown",
				CoDriverJobName: "test-profile",
				Filename:        "test.txt",
			},
			content: "test content",
			checkPath: func(basePath string, metadata ProfileMetadata) (string, error) {
				pattern := filepath.Join(basePath, metadata.Namespace, metadata.AppLabel, metadata.CoDriverJobName, "*", "*", "*", metadata.Filename)
				matches, err := filepath.Glob(pattern)
				if err != nil {
					return "", err
				}
				if len(matches) == 0 {
					return "", nil
				}
				return matches[0], nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			basePath := t.TempDir()
			mgr, err := NewManager(basePath, tt.dateFormat)
			if err != nil {
				t.Fatalf("NewManager() error = %v", err)
			}

			reader := bytes.NewBufferString(tt.content)
			err = mgr.SaveProfile(reader, tt.metadata)
			if err != nil {
				t.Errorf("SaveProfile() error = %v", err)
				return
			}

			// Find the created file
			filePath, err := tt.checkPath(basePath, tt.metadata)
			if err != nil {
				t.Fatalf("checkPath() error = %v", err)
			}

			if filePath == "" {
				t.Errorf("SaveProfile() no file found for metadata %+v", tt.metadata)
				return
			}

			// Verify content
			savedContent, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("os.ReadFile() error = %v", err)
			}

			if string(savedContent) != tt.content {
				t.Errorf("SaveProfile() content = %v, want %v", string(savedContent), tt.content)
			}
		})
	}
}

func TestSaveProfile_DirectoryCreation(t *testing.T) {
	basePath := t.TempDir()
	mgr, err := NewManager(basePath, "2006/01/02")
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	metadata := ProfileMetadata{
		Namespace:       "test-ns",
		AppLabel:        "tier-backend",
		CoDriverJobName: "deep-profile",
		Filename:        "data.bin",
	}

	reader := bytes.NewBufferString("test data")
	err = mgr.SaveProfile(reader, metadata)
	if err != nil {
		t.Errorf("SaveProfile() error = %v", err)
	}

	// Verify all directories were created
	expectedDirs := []string{
		filepath.Join(basePath, "test-ns"),
		filepath.Join(basePath, "test-ns", "tier-backend"),
		filepath.Join(basePath, "test-ns", "tier-backend", "deep-profile"),
	}

	for _, dir := range expectedDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("SaveProfile() did not create directory %v", dir)
		}
	}
}

func TestSaveProfile_ReadError(t *testing.T) {
	basePath := t.TempDir()
	mgr, err := NewManager(basePath, "2006/01/02")
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Create a reader that returns an error
	errorReader := &errorReader{err: errors.New("read failed")}
	metadata := ProfileMetadata{
		Namespace:       "default",
		AppLabel:        "app-test",
		CoDriverJobName: "test-job",
		Filename:        "test.txt",
	}

	err = mgr.SaveProfile(errorReader, metadata)
	if err == nil {
		t.Error("SaveProfile() expected error, got nil")
	}
}

func TestSaveProfile_EmptyFile(t *testing.T) {
	basePath := t.TempDir()
	mgr, err := NewManager(basePath, "2006/01/02")
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	metadata := ProfileMetadata{
		Namespace:       "default",
		AppLabel:        "app-test",
		CoDriverJobName: "test-job",
		Filename:        "empty.txt",
	}

	reader := bytes.NewReader([]byte{})
	err = mgr.SaveProfile(reader, metadata)
	if err != nil {
		t.Errorf("SaveProfile() unexpected error = %v", err)
	}

	// Verify empty file was created
	pattern := filepath.Join(basePath, "default", "app-test", "test-job", "*", "*", "*", "empty.txt")
	matches, _ := filepath.Glob(pattern)
	if len(matches) == 0 {
		t.Error("SaveProfile() did not create empty file")
	}
}

func TestSaveProfile_SpecialCharacters(t *testing.T) {
	basePath := t.TempDir()
	mgr, err := NewManager(basePath, "2006/01/02")
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	tests := []struct {
		name     string
		metadata ProfileMetadata
	}{
		{
			name: "hyphen in label",
			metadata: ProfileMetadata{
				Namespace:       "default",
				AppLabel:        "app-nginx-v2",
				CoDriverJobName: "test-job",
				Filename:        "output.txt",
			},
		},
		{
			name: "underscore in filename",
			metadata: ProfileMetadata{
				Namespace:       "default",
				AppLabel:        "app-test",
				CoDriverJobName: "test-job",
				Filename:        "profile_data.txt",
			},
		},
		{
			name: "dots in filename",
			metadata: ProfileMetadata{
				Namespace:       "default",
				AppLabel:        "app-test",
				CoDriverJobName: "test-job",
				Filename:        "data.2025.10.30.txt",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewBufferString("test data")
			err := mgr.SaveProfile(reader, tt.metadata)
			if err != nil {
				t.Errorf("SaveProfile() unexpected error = %v", err)
			}
		})
	}
}

func TestNewManager_InvalidPath(t *testing.T) {
	// Try to create manager with path that can't be created
	invalidPath := "/proc/invalid/path/that/cannot/be/created"
	_, err := NewManager(invalidPath, "2006/01/02")
	if err == nil {
		t.Error("NewManager() expected error for invalid path, got nil")
	}
}

func TestSaveProfile_FileCreationError(t *testing.T) {
	basePath := t.TempDir()
	mgr, err := NewManager(basePath, "2006/01/02")
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Create a file where we want to create a directory
	conflictPath := filepath.Join(basePath, "default")
	if err := os.WriteFile(conflictPath, []byte("conflict"), 0644); err != nil {
		t.Fatalf("Failed to create conflict file: %v", err)
	}

	metadata := ProfileMetadata{
		Namespace:       "default",
		AppLabel:        "app-test",
		CoDriverJobName: "test-job",
		Filename:        "test.txt",
	}

	reader := bytes.NewBufferString("test data")
	err = mgr.SaveProfile(reader, metadata)
	if err == nil {
		t.Error("SaveProfile() expected error for file creation conflict, got nil")
	}
}

func TestSaveProfile_WriteLargeFile(t *testing.T) {
	basePath := t.TempDir()
	mgr, err := NewManager(basePath, "2006/01/02")
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	metadata := ProfileMetadata{
		Namespace:       "default",
		AppLabel:        "app-test",
		CoDriverJobName: "large-job",
		Filename:        "large.bin",
	}

	// Create a large buffer (10MB)
	largeData := make([]byte, 10*1024*1024)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	reader := bytes.NewReader(largeData)
	err = mgr.SaveProfile(reader, metadata)
	if err != nil {
		t.Errorf("SaveProfile() unexpected error for large file = %v", err)
	}

	// Verify file size
	pattern := filepath.Join(basePath, "default", "app-test", "large-job", "*", "*", "*", "large.bin")
	matches, _ := filepath.Glob(pattern)
	if len(matches) == 0 {
		t.Fatal("SaveProfile() did not create large file")
	}

	info, err := os.Stat(matches[0])
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	if info.Size() != int64(len(largeData)) {
		t.Errorf("SaveProfile() file size = %d, want %d", info.Size(), len(largeData))
	}
}

// errorReader is a helper that always returns an error
type errorReader struct {
	err error
}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, e.err
}
