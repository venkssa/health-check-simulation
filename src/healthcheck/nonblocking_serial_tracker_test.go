package healthcheck

import (
	"sync"
	"testing"
	"time"
)

type trackerWithWaitGroup struct {
	wg     *sync.WaitGroup
	status Status
}

func (t trackerWithWaitGroup) GetStatus() Status {
	t.wg.Done()
	return t.status
}

func (t trackerWithWaitGroup) ServiceName() string {
	return t.status.ServiceName
}

func TestNonBlockingSerialTracker_GetStatusOfDependentServices(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(4)

	healthyTracker := trackerWithWaitGroup{wg, Status{true, "HealthyService", "All good", time.Now()}}
	healthyTracker2 := trackerWithWaitGroup{wg, Status{true, "HealthyService2", "All Ok", time.Now()}}

	tracker := &NonBlockingSerialTracker{frequency: 5 * time.Millisecond}
	tracker.Register(healthyTracker)
	tracker.Register(healthyTracker2)

	tracker.Start()

	wg.Wait()
	overallStatus := tracker.GetStatusOfDependentServices()
	tracker.Stop()

	if overallStatus.IsHealthy != true {
		t.Errorf("Expected %v service to be unhealthy but was %v", new(alwaysFailingHealthChecker).ServiceName(),
			overallStatus.Msg)
	}
}
