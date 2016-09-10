package healthcheck2

import (
	"fmt"
	"strings"
	"time"
)

type OverallStatus struct {
	IsHealthy          bool
	Msg                string
	IndividualStatuses []Status
	Timestamp          time.Time
}

func BuildOverallStatus(statuses []Status) OverallStatus {
	var unhealthyServiceNames []string
	for _, status := range statuses {
		if !status.IsHealthy {
			unhealthyServiceNames = append(unhealthyServiceNames, status.ServiceName)
		}
	}

	isHealthy := len(unhealthyServiceNames) == 0

	msg := "All dependent services are healthy."

	if !isHealthy {
		msg = fmt.Sprintf("The list of unhealthy services are [%v]", strings.Join(unhealthyServiceNames, ","))
	}

	return OverallStatus{
		IsHealthy:          isHealthy,
		Msg:                msg,
		IndividualStatuses: statuses,
		Timestamp:          time.Now(),
	}
}

type ApplicationHealthChecker interface {
	GetOverallStatus() OverallStatus
}

type SimpleApplicationHealthChecker struct {
	checkers []HealthChecker
}

func NewSimpleApplicationHealthChecker(checkers ...HealthChecker) SimpleApplicationHealthChecker {
	return SimpleApplicationHealthChecker{checkers: checkers}
}

func (sahc SimpleApplicationHealthChecker) GetOverallStatus() OverallStatus {
	var statuses []Status
	for _, checker := range sahc.checkers {
		statuses = append(statuses, checker.GetStatus())
	}
	return BuildOverallStatus(statuses)
}

type ConcurrentApplicationHealthChecker struct {
	checkers []HealthChecker
}

func NewConcurrentApplicationHealthChecker(checkers ...HealthChecker) ConcurrentApplicationHealthChecker {
	return ConcurrentApplicationHealthChecker{checkers: checkers}
}

func (pahc ConcurrentApplicationHealthChecker) GetOverallStatus() OverallStatus {
	numOfHealthCheckers := len(pahc.checkers)
	statusChan := make(chan Status, numOfHealthCheckers)

	for idx := 0; idx < numOfHealthCheckers; idx++ {
		go func(checker HealthChecker, result chan<- Status) {
			result <- checker.GetStatus()
		}(pahc.checkers[idx], statusChan)
	}

	statuses := make([]Status, numOfHealthCheckers)
	for idx := 0; idx < numOfHealthCheckers; idx++ {
		statuses[idx] = <-statusChan
	}

	close(statusChan)
	return BuildOverallStatus(statuses)
}
