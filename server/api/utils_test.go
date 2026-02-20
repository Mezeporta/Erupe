package api

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.uber.org/zap"
)

func TestInTrustedRoot(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		trustedRoot string
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "path directly in trusted root",
			path:        "/home/user/screenshots/image.jpg",
			trustedRoot: "/home/user/screenshots",
			wantErr:     false,
		},
		{
			name:        "path with nested directories in trusted root",
			path:        "/home/user/screenshots/2024/image.jpg",
			trustedRoot: "/home/user/screenshots",
			wantErr:     false,
		},
		{
			name:        "path outside trusted root",
			path:        "/home/user/other/image.jpg",
			trustedRoot: "/home/user/screenshots",
			wantErr:     true,
			errMsg:      "path is outside of trusted root",
		},
		{
			name:        "path attempting directory traversal",
			path:        "/home/user/screenshots/../../../etc/passwd",
			trustedRoot: "/home/user/screenshots",
			wantErr:     true,
			errMsg:      "path is outside of trusted root",
		},
		{
			name:        "root directory comparison",
			path:        "/home/user/screenshots/image.jpg",
			trustedRoot: "/",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := inTrustedRoot(tt.path, tt.trustedRoot)
			if (err != nil) != tt.wantErr {
				t.Errorf("inTrustedRoot() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && err.Error() != tt.errMsg {
				t.Errorf("inTrustedRoot() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestVerifyPath(t *testing.T) {
	// Create temporary directory structure for testing
	tmpDir := t.TempDir()
	safeDir := filepath.Join(tmpDir, "safe")
	unsafeDir := filepath.Join(tmpDir, "unsafe")

	if err := os.MkdirAll(safeDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	if err := os.MkdirAll(unsafeDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create subdirectory in safe directory
	nestedDir := filepath.Join(safeDir, "subdir")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("Failed to create nested directory: %v", err)
	}

	// Create actual test files
	safeFile := filepath.Join(safeDir, "image.jpg")
	if err := os.WriteFile(safeFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	nestedFile := filepath.Join(nestedDir, "image.jpg")
	if err := os.WriteFile(nestedFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create nested test file: %v", err)
	}

	unsafeFile := filepath.Join(unsafeDir, "image.jpg")
	if err := os.WriteFile(unsafeFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create unsafe test file: %v", err)
	}

	tests := []struct {
		name        string
		path        string
		trustedRoot string
		wantErr     bool
	}{
		{
			name:        "valid path in trusted directory",
			path:        safeFile,
			trustedRoot: safeDir,
			wantErr:     false,
		},
		{
			name:        "valid nested path in trusted directory",
			path:        nestedFile,
			trustedRoot: safeDir,
			wantErr:     false,
		},
		{
			name:        "path outside trusted directory",
			path:        unsafeFile,
			trustedRoot: safeDir,
			wantErr:     true,
		},
		{
			name:        "path with .. traversal attempt",
			path:        filepath.Join(safeDir, "..", "unsafe", "image.jpg"),
			trustedRoot: safeDir,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := verifyPath(tt.path, tt.trustedRoot, zap.NewNop())
			if (err != nil) != tt.wantErr {
				t.Errorf("verifyPath() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && result == "" {
				t.Errorf("verifyPath() result should not be empty on success")
			}
			if !tt.wantErr && !strings.HasPrefix(result, tt.trustedRoot) {
				t.Errorf("verifyPath() result = %s does not start with trustedRoot = %s", result, tt.trustedRoot)
			}
		})
	}
}

func TestVerifyPathWithSymlinks(t *testing.T) {
	// Skip on systems where symlinks might not work
	tmpDir := t.TempDir()
	safeDir := filepath.Join(tmpDir, "safe")
	outsideDir := filepath.Join(tmpDir, "outside")

	if err := os.MkdirAll(safeDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	if err := os.MkdirAll(outsideDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create a file outside the safe directory
	outsideFile := filepath.Join(outsideDir, "outside.jpg")
	if err := os.WriteFile(outsideFile, []byte("outside"), 0644); err != nil {
		t.Fatalf("Failed to create outside file: %v", err)
	}

	// Try to create a symlink pointing outside (this might fail on some systems)
	symlinkPath := filepath.Join(safeDir, "link.jpg")
	if err := os.Symlink(outsideFile, symlinkPath); err != nil {
		t.Skipf("Symlinks not supported on this system: %v", err)
	}

	// Verify that symlink pointing outside is detected
	_, err := verifyPath(symlinkPath, safeDir, zap.NewNop())
	if err == nil {
		t.Errorf("verifyPath() should reject symlink pointing outside trusted root")
	}
}

func BenchmarkVerifyPath(b *testing.B) {
	tmpDir := b.TempDir()
	safeDir := filepath.Join(tmpDir, "safe")
	if err := os.MkdirAll(safeDir, 0755); err != nil {
		b.Fatalf("Failed to create test directory: %v", err)
	}

	testPath := filepath.Join(safeDir, "test.jpg")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = verifyPath(testPath, safeDir, zap.NewNop())
	}
}

func BenchmarkInTrustedRoot(b *testing.B) {
	testPath := "/home/user/screenshots/2024/01/image.jpg"
	trustedRoot := "/home/user/screenshots"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = inTrustedRoot(testPath, trustedRoot)
	}
}
