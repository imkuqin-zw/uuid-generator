package filelock

import (
	"errors"
	"os"
	"path/filepath"
)

var (
	// ErrTimeout indicates that the lock attempt timed out.
	ErrTimeout = errors.New("lock timeout exceeded")

	// ErrLocked indicates TryLock failed because the lock was already locked.
	ErrLocked = errors.New("file is already locked")
)

// New returns a new lock around the given file.
func New(filename string) (*Lock, error) {
	if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
		return nil, err
	}
	return &Lock{filename: filename}, nil
}
