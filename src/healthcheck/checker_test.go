package healthcheck

import (
	"testing"
	"time"
)

type alwaysHealthyHealthChecker struct {
	t     *testing.T
	delay time.Duration
}

func (ahhc alwaysHealthyHealthChecker) GetStatus() Status {
	time.Sleep(ahhc.delay)
	return Status{true, "AlwaysHealthyService", "All good", time.Now()}
}

func (ahhc alwaysHealthyHealthChecker) ServiceName() string {
	return "AlwaysHealthyService"
}

type alwaysFailingHealthChecker struct {
	delay time.Duration
}

func (afhc alwaysFailingHealthChecker) GetStatus() Status {
	time.Sleep(afhc.delay)
	return Status{false, "AlwaysFailingHealthChecker", "Something went wrong.", time.Now()}
}

func (afhc alwaysFailingHealthChecker) ServiceName() string {
	return "AlwaysFailingHealthChecker"
}
