package healthcheck

import (
	"reflect"
	"testing"
)

func TestBlockingConcurrentTracker_Register(t *testing.T) {
	tracker := new(BlockingConcurrentTracker)
	checker := new(alwaysHealthyHealthChecker)
	tracker.Register(checker)

	names := tracker.DependentServiceNames()

	if reflect.DeepEqual(names, []string{checker.ServiceName()}) != true {
		t.Errorf("Expected %v but was %v", []string{checker.ServiceName()}, names)
	}
}

func TestBlockingConcurrentTracker_GetStatusOfDependentServices(t *testing.T) {
	tracker := new(BlockingConcurrentTracker)
	tracker.Register(new(alwaysHealthyHealthChecker))
	tracker.Register(new(alwaysFailingHealthChecker))

	overallStatus := tracker.GetStatusOfDependentServices()

	if overallStatus.IsHealthy != false {
		t.Errorf("Expected all services to be healthy but was %v", overallStatus.Msg)
	}
}
