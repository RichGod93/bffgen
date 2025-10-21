package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Transaction represents a set of file operations that can be rolled back
type Transaction struct {
	operations []FileOperation
	backupDir  string
	completed  bool
}

// FileOperation represents a single file operation
type FileOperation struct {
	Type         OperationType
	Path         string
	OriginalData []byte
	NewData      []byte
	Executed     bool
}

// OperationType represents the type of file operation
type OperationType int

const (
	OpCreate OperationType = iota
	OpUpdate
	OpDelete
)

// NewTransaction creates a new file transaction
func NewTransaction() *Transaction {
	backupDir := filepath.Join(GetStateDir(), "backup", fmt.Sprintf("%d", time.Now().Unix()))
	return &Transaction{
		operations: make([]FileOperation, 0),
		backupDir:  backupDir,
		completed:  false,
	}
}

// AddCreate adds a file creation operation
func (t *Transaction) AddCreate(path string, data []byte) {
	t.operations = append(t.operations, FileOperation{
		Type:    OpCreate,
		Path:    path,
		NewData: data,
	})
}

// AddUpdate adds a file update operation
func (t *Transaction) AddUpdate(path string, oldData, newData []byte) {
	t.operations = append(t.operations, FileOperation{
		Type:         OpUpdate,
		Path:         path,
		OriginalData: oldData,
		NewData:      newData,
	})
}

// AddDelete adds a file deletion operation
func (t *Transaction) AddDelete(path string, data []byte) {
	t.operations = append(t.operations, FileOperation{
		Type:         OpDelete,
		Path:         path,
		OriginalData: data,
	})
}

// Execute executes all operations in the transaction
func (t *Transaction) Execute() error {
	// Create backup directory
	if err := os.MkdirAll(t.backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Execute each operation
	for i, op := range t.operations {
		if err := t.executeOperation(&op); err != nil {
			// Rollback all previously executed operations
			rollbackErr := t.rollback(i)
			if rollbackErr != nil {
				return fmt.Errorf("operation failed and rollback also failed: %v (rollback error: %v)", err, rollbackErr)
			}
			return fmt.Errorf("operation failed, rolled back %d operations: %w", i, err)
		}
		t.operations[i].Executed = true
	}

	t.completed = true
	return nil
}

// executeOperation executes a single file operation
func (t *Transaction) executeOperation(op *FileOperation) error {
	switch op.Type {
	case OpCreate:
		// Create parent directory if it doesn't exist
		dir := filepath.Dir(op.Path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", op.Path, err)
		}

		// Write new file
		if err := os.WriteFile(op.Path, op.NewData, 0644); err != nil {
			return fmt.Errorf("failed to create file %s: %w", op.Path, err)
		}

	case OpUpdate:
		// Backup original file
		backupPath := filepath.Join(t.backupDir, filepath.Base(op.Path))
		if err := os.WriteFile(backupPath, op.OriginalData, 0644); err != nil {
			return fmt.Errorf("failed to backup file %s: %w", op.Path, err)
		}

		// Write updated file
		if err := os.WriteFile(op.Path, op.NewData, 0644); err != nil {
			return fmt.Errorf("failed to update file %s: %w", op.Path, err)
		}

	case OpDelete:
		// Backup original file
		backupPath := filepath.Join(t.backupDir, filepath.Base(op.Path))
		if err := os.WriteFile(backupPath, op.OriginalData, 0644); err != nil {
			return fmt.Errorf("failed to backup file %s: %w", op.Path, err)
		}

		// Delete file
		if err := os.Remove(op.Path); err != nil {
			return fmt.Errorf("failed to delete file %s: %w", op.Path, err)
		}

	default:
		return fmt.Errorf("unknown operation type: %d", op.Type)
	}

	return nil
}

// rollback rolls back all executed operations
func (t *Transaction) rollback(lastExecuted int) error {
	var rollbackErrors []string

	// Rollback in reverse order
	for i := lastExecuted - 1; i >= 0; i-- {
		op := t.operations[i]
		if !op.Executed {
			continue
		}

		if err := t.rollbackOperation(&op); err != nil {
			rollbackErrors = append(rollbackErrors, fmt.Sprintf("%s: %v", op.Path, err))
		}
	}

	if len(rollbackErrors) > 0 {
		return fmt.Errorf("rollback errors: %v", rollbackErrors)
	}

	return nil
}

// rollbackOperation rolls back a single file operation
func (t *Transaction) rollbackOperation(op *FileOperation) error {
	switch op.Type {
	case OpCreate:
		// Remove the created file
		if err := os.Remove(op.Path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove created file: %w", err)
		}

	case OpUpdate:
		// Restore from original data
		if err := os.WriteFile(op.Path, op.OriginalData, 0644); err != nil {
			return fmt.Errorf("failed to restore file: %w", err)
		}

	case OpDelete:
		// Restore from backup
		if err := os.WriteFile(op.Path, op.OriginalData, 0644); err != nil {
			return fmt.Errorf("failed to restore deleted file: %w", err)
		}

	default:
		return fmt.Errorf("unknown operation type: %d", op.Type)
	}

	return nil
}

// Commit marks the transaction as successfully completed and cleans up backups
func (t *Transaction) Commit() error {
	if !t.completed {
		return fmt.Errorf("transaction not completed, cannot commit")
	}

	// Optionally keep backups for a while, or remove them
	// For now, we keep them for safety
	return nil
}

// Rollback manually rolls back all operations
func (t *Transaction) Rollback() error {
	return t.rollback(len(t.operations))
}

// GetBackupDir returns the backup directory path
func (t *Transaction) GetBackupDir() string {
	return t.backupDir
}

// GetOperationCount returns the number of operations in the transaction
func (t *Transaction) GetOperationCount() int {
	return len(t.operations)
}

// IsCompleted returns whether the transaction has been completed
func (t *Transaction) IsCompleted() bool {
	return t.completed
}

// FileTransaction is a helper for common file transaction patterns
type FileTransaction struct {
	*Transaction
}

// NewFileTransaction creates a new file transaction
func NewFileTransaction() *FileTransaction {
	return &FileTransaction{
		Transaction: NewTransaction(),
	}
}

// CreateFile adds a file creation to the transaction
func (ft *FileTransaction) CreateFile(path string, content []byte) error {
	// Check if file already exists
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("file already exists: %s", path)
	}

	ft.AddCreate(path, content)
	return nil
}

// UpdateFile adds a file update to the transaction
func (ft *FileTransaction) UpdateFile(path string, newContent []byte) error {
	// Read existing content
	oldContent, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, treat as create
			ft.AddCreate(path, newContent)
			return nil
		}
		return fmt.Errorf("failed to read existing file: %w", err)
	}

	ft.AddUpdate(path, oldContent, newContent)
	return nil
}

// DeleteFile adds a file deletion to the transaction
func (ft *FileTransaction) DeleteFile(path string) error {
	// Read existing content for backup
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, nothing to delete
			return nil
		}
		return fmt.Errorf("failed to read file for deletion: %w", err)
	}

	ft.AddDelete(path, content)
	return nil
}

// ExecuteAndCommit executes the transaction and commits if successful
func (ft *FileTransaction) ExecuteAndCommit() error {
	if err := ft.Execute(); err != nil {
		return err
	}

	return ft.Commit()
}

// WrapOperation wraps a function in a transaction and executes it
func WrapOperation(fn func(*FileTransaction) error) error {
	tx := NewFileTransaction()

	if err := fn(tx); err != nil {
		return err
	}

	return tx.ExecuteAndCommit()
}
