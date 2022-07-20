package filelock

import (
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

const (
	shortWait = 10 * time.Millisecond
	longWait  = 10 * shortWait
	filename  = "./testing"
)

func TestFileLock(t *testing.T) {
	suite.Run(t, &FileLockSuite{filepath: filename})
}

type FileLockSuite struct {
	filepath string
	suite.Suite
}

func (s *FileLockSuite) TestLockNoContention() {
	lock, err := New(s.filepath)
	s.Nil(err)
	started := make(chan struct{})
	acquired := make(chan struct{})
	go func() {
		close(started)
		err := lock.Lock()
		close(acquired)
		s.Nil(err)
	}()

	select {
	case <-started:
		// good, goroutine started.
	case <-time.After(shortWait * 2):
		s.T().Errorf("timeout waiting for goroutine to start")
	}

	select {
	case <-acquired:
		// got the lock. good.
	case <-time.After(shortWait * 2):
		s.T().Fatalf("Timed out waiting for lock acquisition.")
	}
	err = lock.Unlock()
	s.Nil(err)
}

func (s *FileLockSuite) TestLockBlocks() {
	lock, err := New(s.filepath)
	s.Nil(err)
	kill := make(chan struct{})

	// this will block until the other process has the lock.
	procDone := LockFromAnotherProc(s.T(), s.filepath, kill)

	defer func() {
		close(kill)
		// now wait for the other process to exit so the file will be unlocked.
		select {
		case <-procDone:
		case <-time.After(time.Second):
		}
	}()

	started := make(chan struct{})
	result := make(chan error)
	go func() {
		close(started)
		result <- lock.TryLock()
	}()

	select {
	case <-started:
		// good, goroutine started.
	case <-time.After(shortWait):
		s.T().Fatalf("timeout waiting for goroutine to start")
	}

	// Wait for trylock to fail.
	select {
	case err := <-result:
		// yes, I know this is redundant with the assert below, but it makes the
		// failed test message clearer.
		if err == nil {
			s.T().Fatalf("lock succeeded, but should have errored out")
		}
		// This should be the error from trylock failing.
		s.ErrorIs(err, ErrLocked)
	case <-time.After(shortWait):
		s.T().Fatalf("took too long to fail trylock")
	}

}

// LockFromAnotherProc will launch a process and block until that process has
// created the lock file.  If we time out waiting for the other process to take
// the lock, this function will fail the current test.
func LockFromAnotherProc(t *testing.T, path string, kill chan struct{}) (done chan struct{}) {
	cmd := exec.Command(os.Args[0], "-test.run", "TestLockFromOtherProcess")
	cmd.Env = append(
		// We must preserve os.Environ() on Windows,
		// or the subprocess will fail in weird and
		// wonderful ways.
		os.Environ(),
	)

	if err := cmd.Start(); err != nil {
		t.Fatalf("error starting other proc: %v", err)
	}

	done = make(chan struct{})

	go func() {
		cmd.Wait()
		close(done)
	}()

	go func() {
		select {
		case <-kill:
			// this may fail, but there's not much we can do about it
			_ = cmd.Process.Kill()
		case <-done:
		}
	}()

	for x := 0; x < 10; x++ {
		time.Sleep(shortWait)
		if _, err := os.Stat(path); err == nil {
			// file created by other process, let's continue
			break
		}
		if x == 9 {
			t.Fatalf("timed out waiting for other process to start")
		}
	}
	return done
}

func TestLockFromOtherProcess(t *testing.T) {
	lock, err := New(filename)
	if err != nil {
		t.Fatalf("error locking %q: %v", filename, err)
	}
	err = lock.Lock()
	if err != nil {
		t.Fatalf("error locking %q: %v", filename, err)
	}
	time.Sleep(longWait)
	err = lock.Unlock()
	if err != nil {
		t.Fatalf("error unlocking %q: %v", filename, err)
	}
}
