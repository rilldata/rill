package local

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLockProject(t *testing.T) {
	// When re-executed by the test below, act as a `rill start` that holds the project until it is killed.
	if dir := os.Getenv("RILL_TEST_LOCK_DIR"); dir != "" {
		_, err := lockProject(dir, "http://localhost:9100")
		require.NoError(t, err)
		fmt.Println("locked")
		time.Sleep(time.Minute)
		return
	}

	dir := t.TempDir()
	owner := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^TestLockProject$")
	owner.Env = append(os.Environ(), "RILL_TEST_LOCK_DIR="+dir)
	stdout, err := owner.StdoutPipe()
	require.NoError(t, err)
	require.NoError(t, owner.Start())
	line, err := bufio.NewReader(stdout).ReadString('\n')
	require.NoError(t, err)
	require.Equal(t, "locked\n", line)

	// The project can't be locked while another process holds it, and the error tells where that process is.
	_, err = lockProject(dir, "")
	require.EqualError(t, err, fmt.Sprintf("`rill start` is already running for this project at http://localhost:9100 (PID %d); stop it and try again", owner.Process.Pid))

	// A crashed process doesn't leave a stale lock behind, even though its details are still in the lock file.
	require.NoError(t, owner.Process.Kill())
	_ = owner.Wait()
	lock, err := lockProject(dir, "")
	require.NoError(t, err)

	// The new owner's details replace those of the crashed one.
	_, err = lockProject(dir, "http://localhost:9009")
	require.EqualError(t, err, fmt.Sprintf("another Rill command (PID %d) is already running for this project; wait for it to finish and try again", os.Getpid()))

	// A lock file that can't be opened stops the process instead of letting it run without the lock.
	// (Root can open the file regardless of its permissions.)
	if os.Geteuid() != 0 {
		require.NoError(t, os.Chmod(lock.Path(), 0o000))
		_, err = lockProject(dir, "")
		require.ErrorIs(t, err, fs.ErrPermission)
		require.NoError(t, os.Chmod(lock.Path(), 0o644))
	}

	// A released project can be locked again.
	require.NoError(t, lock.Close())
	lock, err = lockProject(dir, "")
	require.NoError(t, err)
	require.NoError(t, lock.Close())
}
