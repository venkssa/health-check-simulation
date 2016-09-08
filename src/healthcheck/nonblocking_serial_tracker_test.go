package healthcheck

import (
	"testing"
	"time"
)

func TestNonBlockingSerialTracker_GetStatusOfDependentServices(t *testing.T) {
	tracker := &NonBlockingSerialTracker{Frequency: 10 * time.Millisecond}
	tracker.Register(&testHealthChecker{"Checker1", alwaysHelthyStatusFn})
	ConcurrentHealthTrackerTest(tracker)
}
