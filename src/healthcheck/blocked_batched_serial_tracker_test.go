package healthcheck

import (
	"testing"
)

func TestBlockingBatchedSerialTracker_GetStatusOfDependentServices(t *testing.T) {
	tracker := new(BlockingBatchedSerialTracker)
	tracker.Register(&testHealthChecker{"Checker1", alwaysHelthyStatusFn})

	tracker.Start()
	defer tracker.Stop()

	overallStatus := tracker.GetStatusOfDependentServices()
	if overallStatus.IsHealthy != true {
		t.Errorf("Expected a healthy error check, but was %v", overallStatus)
	}
}

func TestBlockingBatchedSerialTracker_ConcurrentGetStatusOfDependentServices(t *testing.T) {
	tracker := new(BlockingBatchedSerialTracker)
	tracker.Register(&testHealthChecker{"Checker1", alwaysHelthyStatusFn})

	ConcurrentHealthTrackerTest(tracker)
}
