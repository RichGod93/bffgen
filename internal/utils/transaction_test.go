package utils

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// TestBackupDirectoryPermissions verifies that backup directories are created with restricted permissions
func TestBackupDirectoryPermissions(t *testing.T) {
	// Create a test directory
	testDir := t.TempDir()
	defer os.RemoveAll(testDir)

	// Create a transaction
	tx := NewTransaction()

	// Override the backup directory for testing
	tx.backupDir = filepath.Join(testDir, "backup", "test")

	// Create the backup directory
	if err := os.MkdirAll(tx.backupDir, BackupDirPerm); err != nil {
		t.Fatalf("Failed to create backup directory: %v", err)
	}

	// Check directory permissions
	info, err := os.Stat(tx.backupDir)
	if err != nil {
		t.Fatalf("Failed to stat backup directory: %v", err)
	}

	// Get the actual permissions
	perms := info.Mode().Perm()

	// Should be 0o700 (user read/write/execute only)
	if perms != BackupDirPerm {
		t.Errorf("Expected backup directory permissions %v, got %v", BackupDirPerm, perms)
	}

	// Verify no world-readable or group-readable bits are set
	if perms&0o077 != 0 {
		t.Errorf("Backup directory has world or group permissions: %v", perms)
	}
}

// TestBackupFilePermissions verifies that backup files are created with restricted permissions
func TestBackupFilePermissions(t *testing.T) {
	testDir := t.TempDir()
	defer os.RemoveAll(testDir)

	// Create a backup file with restricted permissions
	backupPath := filepath.Join(testDir, "backup.txt")
	testData := []byte("test backup content")

	if err := os.WriteFile(backupPath, testData, BackupFilePerm); err != nil {
		t.Fatalf("Failed to write backup file: %v", err)
	}

	// Check file permissions
	info, err := os.Stat(backupPath)
	if err != nil {
		t.Fatalf("Failed to stat backup file: %v", err)
	}

	perms := info.Mode().Perm()

	// Should be 0o600 (user read/write only)
	if perms != BackupFilePerm {
		t.Errorf("Expected backup file permissions %v, got %v", BackupFilePerm, perms)
	}

	// Verify no world-readable or group-readable bits are set
	if perms&0o077 != 0 {
		t.Errorf("Backup file has world or group permissions: %v", perms)
	}

	// Verify content
	readData, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("Failed to read backup file: %v", err)
	}

	if !bytes.Equal(readData, testData) {
		t.Errorf("Backup file content mismatch: expected %s, got %s", string(testData), string(readData))
	}
}

// TestProjectFilePermissions verifies that project files are created with standard permissions
func TestProjectFilePermissions(t *testing.T) {
	testDir := t.TempDir()
	defer os.RemoveAll(testDir)

	// Create a project file with standard permissions
	projectPath := filepath.Join(testDir, "source.go")
	testData := []byte("package main")

	if err := os.WriteFile(projectPath, testData, ProjectFilePerm); err != nil {
		t.Fatalf("Failed to write project file: %v", err)
	}

	// Check file permissions
	info, err := os.Stat(projectPath)
	if err != nil {
		t.Fatalf("Failed to stat project file: %v", err)
	}

	perms := info.Mode().Perm()

	// Should be 0o644 (user read/write, others read)
	if perms != ProjectFilePerm {
		t.Errorf("Expected project file permissions %v, got %v", ProjectFilePerm, perms)
	}
}

// TestProjectDirectoryPermissions verifies that project directories are created with standard permissions
func TestProjectDirectoryPermissions(t *testing.T) {
	testDir := t.TempDir()
	defer os.RemoveAll(testDir)

	// Create a project directory
	projectDir := filepath.Join(testDir, "src", "lib")

	if err := os.MkdirAll(projectDir, ProjectDirPerm); err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	// Check directory permissions
	info, err := os.Stat(projectDir)
	if err != nil {
		t.Fatalf("Failed to stat project directory: %v", err)
	}

	perms := info.Mode().Perm()

	// Should be 0o755 (user read/write/execute, others read/execute)
	if perms != ProjectDirPerm {
		t.Errorf("Expected project directory permissions %v, got %v", ProjectDirPerm, perms)
	}
}

// TestTransactionFileOperationPermissions verifies permissions during transaction operations
func TestTransactionFileOperationPermissions(t *testing.T) {
	testDir := t.TempDir()
	defer os.RemoveAll(testDir)

	// Create a test file
	testFile := filepath.Join(testDir, "test.txt")
	originalContent := []byte("original content")

	if err := os.WriteFile(testFile, originalContent, ProjectFilePerm); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a transaction to update the file
	tx := NewFileTransaction()
	tx.backupDir = filepath.Join(testDir, "backup")

	// Add update operation
	newContent := []byte("updated content")
	if err := tx.UpdateFile(testFile, newContent); err != nil {
		t.Fatalf("Failed to add update operation: %v", err)
	}

	// Execute transaction
	if err := tx.Execute(); err != nil {
		t.Fatalf("Failed to execute transaction: %v", err)
	}

	// Verify updated file has correct permissions
	info, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("Failed to stat updated file: %v", err)
	}

	perms := info.Mode().Perm()
	if perms != ProjectFilePerm {
		t.Errorf("Expected updated file permissions %v, got %v", ProjectFilePerm, perms)
	}

	// Verify backup file has restricted permissions
	backupPath := filepath.Join(tx.backupDir, "test.txt")
	info, err = os.Stat(backupPath)
	if err != nil {
		t.Fatalf("Failed to stat backup file: %v", err)
	}

	perms = info.Mode().Perm()
	if perms != BackupFilePerm {
		t.Errorf("Expected backup file permissions %v, got %v", BackupFilePerm, perms)
	}

	// Verify backup directory has restricted permissions
	info, err = os.Stat(tx.backupDir)
	if err != nil {
		t.Fatalf("Failed to stat backup directory: %v", err)
	}

	perms = info.Mode().Perm()
	if perms != BackupDirPerm {
		t.Errorf("Expected backup directory permissions %v, got %v", BackupDirPerm, perms)
	}
}
