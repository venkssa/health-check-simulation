package healthcheck

import (
	"reflect"
	"testing"
)

func TestBlockingSerialTracker_Register(t *testing.T) {
	tracker := new(BlockingSerialTracker)
	checker := &testHealthChecker{"Checker1", alwaysHelthyStatusFn}
	tracker.Register(checker)

	names := tracker.DependentServiceNames()

	if reflect.DeepEqual(names, []string{checker.ServiceName()}) != true {
		t.Errorf("Expected %v but was %v", []string{checker.ServiceName()}, names)
	}
}

func TestBlockingSerialTracker_GetStatusOfDependentServices(t *testing.T) {
	tracker := new(BlockingSerialTracker)
	tracker.Register(&testHealthChecker{"Checker1", alwaysHelthyStatusFn})
	tracker.Register(&testHealthChecker{"Checker1", alwaysFailingStatusFn})

	overallStatus := tracker.GetStatusOfDependentServices()

	if overallStatus.IsHealthy != false {
		t.Errorf("Expected all services to be healthy but was %v", overallStatus.Msg)
	}
}
