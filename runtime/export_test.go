package runtime

import (
	"testing"
	"time"
)

// SetReconcileHangTimeouts overrides the controller's hang detection timeouts for the duration of a test.
// It must be called before the test creates a runtime.
func SetReconcileHangTimeouts(t testing.TB, cancelation, shutdown time.Duration) {
	prevCancelation, prevShutdown := reconcileCancelationTimeout, invocationShutdownTimeout
	reconcileCancelationTimeout, invocationShutdownTimeout = cancelation, shutdown
	t.Cleanup(func() {
		reconcileCancelationTimeout, invocationShutdownTimeout = prevCancelation, prevShutdown
	})
}
