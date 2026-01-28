package main

import (
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeFileName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "lowercase heic extension",
			input:    "photo.heic",
			expected: "photo.jpg",
		},
		{
			name:     "uppercase HEIC extension",
			input:    "photo.HEIC",
			expected: "photo.jpg",
		},
		{
			name:     "file with path lowercase",
			input:    "/path/to/image.heic",
			expected: "image.jpg",
		},
		{
			name:     "file with path uppercase",
			input:    "/path/to/image.HEIC",
			expected: "image.jpg",
		},
		{
			name:     "file without heic extension",
			input:    "photo.jpg",
			expected: "photo.jpg.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeFileName(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeFileName(%q) = %q, want %q",
					tt.input, result, tt.expected)
			}
		})
	}
}

func TestFilterHeicFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files
	heicFile := filepath.Join(tempDir, "test.heic")
	heicFileUpper := filepath.Join(tempDir, "test2.HEIC")
	nonHeicFile := filepath.Join(tempDir, "test.jpg")
	subDir := filepath.Join(tempDir, "subdir")

	if err := os.WriteFile(heicFile, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(heicFileUpper, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(nonHeicFile, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		input    []string
		expected []string
		wantErr  bool
	}{
		{
			name:     "single heic file lowercase",
			input:    []string{heicFile},
			expected: []string{heicFile},
			wantErr:  false,
		},
		{
			name:     "single heic file uppercase",
			input:    []string{heicFileUpper},
			expected: []string{heicFileUpper},
			wantErr:  false,
		},
		{
			name:     "mixed files - only heic returned",
			input:    []string{heicFile, nonHeicFile},
			expected: []string{heicFile},
			wantErr:  false,
		},
		{
			name:     "directory skipped",
			input:    []string{subDir},
			expected: []string{},
			wantErr:  false,
		},
		{
			name:     "non-existent file error",
			input:    []string{"/does/not/exist.heic"},
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "multiple heic files",
			input:    []string{heicFile, heicFileUpper},
			expected: []string{heicFile, heicFileUpper},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := filterHeicFiles(tt.input, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("filterHeicFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(result) != len(tt.expected) {
					t.Errorf("filterHeicFiles() returned %d files, want %d", len(result), len(tt.expected))
					return
				}
				for i, expected := range tt.expected {
					if result[i] != expected {
						t.Errorf("filterHeicFiles()[%d] = %q, want %q", i, result[i], expected)
					}
				}
			}
		})
	}
}

func TestEnsureOutputDir(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "create new directory",
			path:    filepath.Join(tempDir, "newdir"),
			wantErr: false,
		},
		{
			name:    "existing directory",
			path:    tempDir,
			wantErr: false,
		},
		{
			name:    "nested directory creation",
			path:    filepath.Join(tempDir, "a", "b", "c"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ensureOutputDir(tt.path, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("ensureOutputDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// Verify directory exists
				info, err := os.Stat(tt.path)
				if err != nil {
					t.Errorf("directory not created: %v", err)
				}
				if !info.IsDir() {
					t.Errorf("path exists but is not a directory")
				}
			}
		})
	}
}

func TestConvert(t *testing.T) {
	// Skip if testdata doesn't exist
	const testHeicFile = "testdata/sample.heic"
	if _, err := os.Stat(testHeicFile); os.IsNotExist(err) {
		t.Skip("testdata/sample.heic not found, skipping integration test")
	}

	tempDir := t.TempDir()
	errChan := make(chan error, 1)

	tests := []struct {
		name     string
		filePath string
		outPath  string
		quality  int
		verbose  bool
	}{
		{
			name:     "convert to temp dir default quality",
			filePath: testHeicFile,
			outPath:  tempDir,
			quality:  100,
			verbose:  false,
		},
		{
			name:     "convert with lower quality",
			filePath: testHeicFile,
			outPath:  tempDir,
			quality:  80,
			verbose:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run conversion
			convert(tt.filePath, tt.outPath, tt.quality, tt.verbose, errChan)

			// Check for errors
			select {
			case err := <-errChan:
				t.Fatalf("convert failed: %v", err)
			default:
				// Success
			}

			// Verify output file exists
			expectedOut := filepath.Join(tt.outPath, normalizeFileName(tt.filePath))
			if _, err := os.Stat(expectedOut); err != nil {
				t.Errorf("output file not created: %v", err)
				return
			}

			// Verify it's a valid JPEG
			f, err := os.Open(expectedOut)
			if err != nil {
				t.Errorf("cannot open output file: %v", err)
				return
			}
			defer func() {
				if err := f.Close(); err != nil {
					t.Errorf("failed to close file: %v", err)
				}
			}()

			_, err = jpeg.Decode(f)
			if err != nil {
				t.Errorf("output is not valid JPEG: %v", err)
			}
		})
	}
}
