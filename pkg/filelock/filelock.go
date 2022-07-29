// Copyright 2022 The imkuqin-zw Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
