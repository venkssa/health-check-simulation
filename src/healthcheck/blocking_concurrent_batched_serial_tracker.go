package healthcheck

import "time"

func blockingConcurrentBatchedSerialTracker(healthCheckers HealthCheckers, updateDuration time.Duration,
	lastStatusChan <-chan chan<- OverallStatus, stopChan <-chan struct{}) {
	go func(checkers HealthCheckers) {
		startTracker := time.After(0 * time.Millisecond)
		currentStatus := buildUnhealthyOverallStatus(checkers.DependentServiceNames(),
			"Healthcheck on services not yet complete.")

		for {
			select {
			case <-stopChan:
				if startTracker != nil {
					currentStatus = buildUnhealthyOverallStatus(checkers.DependentServiceNames(),
						"Healthcheck on services is stopped")
					startTracker = nil
					checkers = HealthCheckers([]HealthChecker{})
				}
			case <-startTracker:
				var statuses []Status
				for _, checker := range checkers {
					statuses = append(statuses, checker.GetStatus())
				}
				currentStatus = buildOverallStatus(statuses)
				startTracker = time.After(updateDuration)
			case resultChan := <-lastStatusChan:
				resultChan <- currentStatus
			}
		}

	}(append(HealthCheckers{}, healthCheckers...))
}
