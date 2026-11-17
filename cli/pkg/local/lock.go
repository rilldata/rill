package local

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/gofrs/flock"
)

// lockFileName is the name of the lock file in a project's tmp directory.
const lockFileName = "rill.lock"

// lockOwner describes the process that holds a project's lock.
// It is stored as JSON in the lock file, so a process that is refused the lock can tell the user who holds it.
type lockOwner struct {
	PID int    `json:"pid"`
	URL string `json:"url,omitempty"` // Where the app is served; empty for commands that don't serve it.
}

// lockProject takes an exclusive lock on a project's tmp directory, which holds its catalog and OLAP data.
// That state only supports one Rill process at a time:
// a second process fails on the OLAP database's own file lock,
// and it takes over the catalog's controller version, which breaks the first process on its next write.
//
// If another process holds the lock, it returns an error that names it.
// The lock is an OS file lock, so the OS releases it when the process exits, even after a crash;
// the lock file itself is left in place and reused.
func lockProject(dbDirPath, url string) (*flock.Flock, error) {
	// The permissions let other users of the project open the file to check the lock.
	lock := flock.New(filepath.Join(dbDirPath, lockFileName), flock.SetPermissions(0o644))
	locked, err := lock.TryLock()
	if err != nil {
		// On file systems that can't lock files (such as some network mounts), carry on without the lock.
		// Any other error, such as a lock file that can't be opened, may hide a process that holds the lock.
		if errors.Is(err, errors.ErrUnsupported) || errors.Is(err, syscall.ENOLCK) {
			return lock, nil
		}
		return nil, fmt.Errorf("failed to lock the project: %w", err)
	}

	if !locked {
		// The owner details are best effort: the owner may not have written them yet.
		var owner lockOwner
		data, err := os.ReadFile(lock.Path())
		if err == nil {
			_ = json.Unmarshal(data, &owner)
		}
		switch {
		case owner.URL != "":
			return nil, fmt.Errorf("`rill start` is already running for this project at %s (PID %d); stop it and try again", owner.URL, owner.PID)
		case owner.PID != 0:
			return nil, fmt.Errorf("another Rill command (PID %d) is already running for this project; wait for it to finish and try again", owner.PID)
		default:
			return nil, errors.New("another Rill process is already running for this project; stop it and try again")
		}
	}

	// Record the owner for the benefit of other processes; it replaces the details of any previous owner.
	data, err := json.Marshal(lockOwner{PID: os.Getpid(), URL: url})
	if err == nil {
		_ = os.WriteFile(lock.Path(), data, 0o644)
	}

	return lock, nil
}
