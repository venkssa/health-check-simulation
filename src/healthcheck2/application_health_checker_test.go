package healthcheck2

import (
	"testing"
	"time"
)

func TestApplicationHealthChecker_GetOverallStatus(t *testing.T) {
	healthyService1 := NewTestHealthChecker("service1", AlwaysHelthyStatusFn)
	healthyService2 := NewTestHealthChecker("service2", AlwaysHelthyStatusFn)

	unhealthyService1 := NewTestHealthChecker("service2", AlwaysFailingStatusFn)

	healthyStatus := OverallStatus{IsHealthy: true}
	t.Run("Simple Healthy Overall Status",
		verifyGetStatus(NewSimpleApplicationHealthChecker(healthyService1, healthyService2), healthyStatus))
	t.Run("Concurrent Healthy Overall Status",
		verifyGetStatus(NewConcurrentApplicationHealthChecker(healthyService1, healthyService2), healthyStatus))

	unhealthyStatus := OverallStatus{IsHealthy: false}
	t.Run("Simple Unhealthy Overall Status",
		verifyGetStatus(NewSimpleApplicationHealthChecker(unhealthyService1, healthyService2), unhealthyStatus))
	t.Run("Concurrent Unhelathy Overall Status",
		verifyGetStatus(NewConcurrentApplicationHealthChecker(unhealthyService1, healthyService2), unhealthyStatus))
}

func verifyGetStatus(appHealthChecker ApplicationHealthChecker, expectedStatus OverallStatus) func(t *testing.T) {
	return func (t *testing.T) {
		actualStatus := appHealthChecker.GetOverallStatus()

		if actualStatus.IsHealthy != expectedStatus.IsHealthy {
			t.Errorf("Expected isHealthy to be %v but was %v",
				expectedStatus.IsHealthy, actualStatus.IsHealthy)
		}
	}
}

func TestApplicationHealthChecker_TimeGetOverallStatus(t *testing.T) {
	checkers := []HealthChecker{
		NewTestHealthChecker("service1", Delay(10*time.Millisecond, AlwaysHelthyStatusFn)),
		NewTestHealthChecker("service2", Delay(20*time.Millisecond, AlwaysHelthyStatusFn)),
	}

	t.Run("Simple", timeAndVerifyGetStatus(NewSimpleApplicationHealthChecker(checkers...), 30*time.Millisecond))
	t.Run("Concurrent",
		timeAndVerifyGetStatus(NewConcurrentApplicationHealthChecker(checkers...), 20*time.Millisecond))
}

func timeAndVerifyGetStatus(appHealthChecker ApplicationHealthChecker, expectedDuration time.Duration) func(*testing.T) {
	return func(t *testing.T) {
		startTime := time.Now()
		appHealthChecker.GetOverallStatus()
		actualDuration := time.Now().Sub(startTime)

		if actualDuration < expectedDuration {
			t.Errorf("Expected call to take %v but took %v", expectedDuration, actualDuration)
		}
	}
}
