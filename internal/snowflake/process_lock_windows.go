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

package snowflake

import (
	"github.com/imkuqin-zw/uuid-generator/pkg/filelock"
	"github.com/imkuqin-zw/yggdrasil/pkg/defers"
	"github.com/pkg/errors"
)

func Lock() error {
	lock, err := filelock.New("C:\\ProgramData\\github\\imkuqin_zw\\uuid_generator_snowflake_process.lock")
	if err != nil {
		return errors.WithStack(err)
	}
	if err := lock.TryLock(); err != nil {
		return errors.WithStack(err)
	}
	defers.Register(func() error {
		if err := lock.Unlock(); err != nil {
			return errors.WithStack(err)
		}
		return nil
	})
	return nil
}
