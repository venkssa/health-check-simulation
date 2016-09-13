package healthcheck

import (
	"testing"
	"time"
)

func TestBlockingConcurrentBatchedSerialTracker(t *testing.T) {
	checkers := HealthCheckers{&testHealthChecker{"1", alwaysHelthyStatusFn}}

	lastKnownStatusChan := make(chan chan<- OverallStatus)
	stopChan := make(chan struct{})
	blockingConcurrentBatchedSerialTracker(checkers, 1*time.Millisecond, lastKnownStatusChan, stopChan)

	statusChan := make(chan OverallStatus)
	lastKnownStatusChan <- statusChan
	t.Log(<-statusChan)

	stopChan <- struct{}{}
	lastKnownStatusChan <- statusChan
	t.Log(<-statusChan)
	stopChan <- struct{}{}
	lastKnownStatusChan <- statusChan
	t.Log(<-statusChan)
}
