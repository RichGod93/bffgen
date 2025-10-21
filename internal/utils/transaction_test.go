package utils

import (
	"os"
	"testing"
)

func TestTransaction(t *testing.T) {
	// Create temp directory for testing
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldDir)

	t.Run("CreateFile", func(t *testing.T) {
		tx := NewTransaction()
		testFile := "test.txt"
		testContent := []byte("test content")

		tx.AddCreate(testFile, testContent)

		if err := tx.Execute(); err != nil {
			t.Fatalf("Failed to execute transaction: %v", err)
		}

		// Verify file was created
		content, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("Failed to read created file: %v", err)
		}

		if string(content) != string(testContent) {
			t.Errorf("Expected content '%s', got '%s'", testContent, content)
		}

		if !tx.IsCompleted() {
			t.Error("Transaction should be marked as completed")
		}
	})

	t.Run("UpdateFile", func(t *testing.T) {
		testFile := "update.txt"
		oldContent := []byte("old content")
		newContent := []byte("new content")

		// Create initial file
		os.WriteFile(testFile, oldContent, 0644)

		tx := NewTransaction()
		tx.AddUpdate(testFile, oldContent, newContent)

		if err := tx.Execute(); err != nil {
			t.Fatalf("Failed to execute transaction: %v", err)
		}

		// Verify file was updated
		content, _ := os.ReadFile(testFile)
		if string(content) != string(newContent) {
			t.Errorf("Expected content '%s', got '%s'", newContent, content)
		}

		// Verify backup was created
		backupDir := tx.GetBackupDir()
		if _, err := os.Stat(backupDir); os.IsNotExist(err) {
			t.Error("Backup directory should have been created")
		}
	})

	t.Run("DeleteFile", func(t *testing.T) {
		testFile := "delete.txt"
		content := []byte("to be deleted")

		// Create file
		os.WriteFile(testFile, content, 0644)

		tx := NewTransaction()
		tx.AddDelete(testFile, content)

		if err := tx.Execute(); err != nil {
			t.Fatalf("Failed to execute transaction: %v", err)
		}

		// Verify file was deleted
		if _, err := os.Stat(testFile); !os.IsNotExist(err) {
			t.Error("File should have been deleted")
		}
	})

	t.Run("RollbackOnError", func(t *testing.T) {
		testFile1 := "rollback1.txt"

		tx := NewTransaction()

		// Add valid create operation
		tx.AddCreate(testFile1, []byte("content1"))

		// Add invalid operation (create in non-existent directory with no parent creation)
		tx.AddCreate("/nonexistent/impossible/file.txt", []byte("will fail"))

		err := tx.Execute()
		if err == nil {
			t.Error("Transaction should have failed")
		}

		// Verify first file was rolled back
		if _, err := os.Stat(testFile1); !os.IsNotExist(err) {
			t.Error("First file should have been rolled back")
		}
	})

	t.Run("FileTransaction", func(t *testing.T) {
		ft := NewFileTransaction()

		testFile := "filetx.txt"
		content := []byte("file transaction")

		if err := ft.CreateFile(testFile, content); err != nil {
			t.Fatalf("Failed to add create: %v", err)
		}

		if err := ft.ExecuteAndCommit(); err != nil {
			t.Fatalf("Failed to execute and commit: %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(testFile); os.IsNotExist(err) {
			t.Error("File should have been created")
		}
	})

	t.Run("UpdateFileInTransaction", func(t *testing.T) {
		testFile := "updatetx.txt"
		oldContent := []byte("old")
		newContent := []byte("new")

		// Create initial file
		os.WriteFile(testFile, oldContent, 0644)

		ft := NewFileTransaction()

		if err := ft.UpdateFile(testFile, newContent); err != nil {
			t.Fatalf("Failed to add update: %v", err)
		}

		if err := ft.ExecuteAndCommit(); err != nil {
			t.Fatalf("Failed to execute: %v", err)
		}

		// Verify content
		content, _ := os.ReadFile(testFile)
		if string(content) != string(newContent) {
			t.Errorf("Expected '%s', got '%s'", newContent, content)
		}
	})

	t.Run("WrapOperation", func(t *testing.T) {
		testFile := "wrapped.txt"

		err := WrapOperation(func(tx *FileTransaction) error {
			return tx.CreateFile(testFile, []byte("wrapped content"))
		})

		if err != nil {
			t.Fatalf("WrapOperation failed: %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(testFile); os.IsNotExist(err) {
			t.Error("File should have been created")
		}
	})
}

func TestGetOperationCount(t *testing.T) {
	tx := NewTransaction()

	if tx.GetOperationCount() != 0 {
		t.Error("New transaction should have 0 operations")
	}

	tx.AddCreate("file1.txt", []byte("content"))
	tx.AddCreate("file2.txt", []byte("content"))

	if tx.GetOperationCount() != 2 {
		t.Errorf("Expected 2 operations, got %d", tx.GetOperationCount())
	}
}
