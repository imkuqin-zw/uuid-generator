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
