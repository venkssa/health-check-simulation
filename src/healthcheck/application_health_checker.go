package healthcheck

import (
	"fmt"
	"strings"
	"time"
)

type ApplicationStatus struct {
	IsHealthy          bool
	Msg                string
	IndividualStatuses []Status
	Timestamp          time.Time
}

func (as ApplicationStatus) Copy() ApplicationStatus {
	as.IndividualStatuses = append([]Status{}, as.IndividualStatuses...)
	return as
}

func BuildApplicationStatus(statuses []Status) ApplicationStatus {
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

	return ApplicationStatus{
		IsHealthy:          isHealthy,
		Msg:                msg,
		IndividualStatuses: statuses,
		Timestamp:          time.Now(),
	}
}

type ApplicationHealthChecker interface {
	GetApplicationStatus() ApplicationStatus
	NumOfChecks() uint32
}

type SimpleApplicationHealthChecker struct {
	checkers []HealthChecker
}

func NewSimpleApplicationHealthChecker(checkers ...HealthChecker) SimpleApplicationHealthChecker {
	return SimpleApplicationHealthChecker{checkers: checkers}
}

func (hc SimpleApplicationHealthChecker) GetApplicationStatus() ApplicationStatus {
	var statuses []Status
	for _, checker := range hc.checkers {
		statuses = append(statuses, checker.GetStatus())
	}
	return BuildApplicationStatus(statuses)
}

func (hc SimpleApplicationHealthChecker) NumOfChecks() uint32 {
	return uint32(len(hc.checkers))
}

type ConcurrentApplicationHealthChecker struct {
	checkers []HealthChecker
}

func NewConcurrentApplicationHealthChecker(checkers ...HealthChecker) ConcurrentApplicationHealthChecker {
	return ConcurrentApplicationHealthChecker{checkers: checkers}
}

func (hc ConcurrentApplicationHealthChecker) GetApplicationStatus() ApplicationStatus {
	numOfHealthCheckers := len(hc.checkers)
	statusChan := make(chan Status, numOfHealthCheckers)

	for _, checker := range hc.checkers {
		go func(checker HealthChecker, result chan<- Status) {
			result <- checker.GetStatus()
		}(checker, statusChan)
	}

	var statuses []Status
	for range hc.checkers {
		statuses = append(statuses, <-statusChan)
	}

	close(statusChan)
	return BuildApplicationStatus(statuses)
}

func (hc ConcurrentApplicationHealthChecker) NumOfChecks() uint32 {
	return uint32(len(hc.checkers))
}
