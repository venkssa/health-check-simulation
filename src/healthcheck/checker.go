package healthcheck

import (
	"fmt"
	"strings"
	"time"
)

type Status struct {
	IsHealthy   bool
	ServiceName string
	Msg         string
	Timestamp   time.Time
}

type HealthChecker interface {
	GetStatus() Status
	ServiceName() string
}

type Tracker interface {
	Register(HealthChecker)
	Start()
	Stop()
	GetStatusOfDependentServices() OverallStatus
	DependentServiceNames() []string
}

type HealthCheckers []HealthChecker

func (hcs HealthCheckers) DependentServiceNames() []string {
	names := make([]string, 0, len(hcs))

	for _, checker := range hcs {
		names = append(names, checker.ServiceName())
	}

	return names
}

type OverallStatus struct {
	IsHealthy                bool
	Msg                      string
	DependentServiceStatuses []Status
	Timestamp                time.Time
}

func buildOverallStatus(statuses []Status) OverallStatus {
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
		IsHealthy: isHealthy,
		Msg:       msg,
		DependentServiceStatuses: statuses,
		Timestamp:                time.Now(),
	}
}

func pendingOverallStatus(serviceNames []string) OverallStatus {
	var pendingStatuses []Status
	for _, serviceName := range serviceNames {
		pendingStatus := Status{
			IsHealthy:   false,
			ServiceName: serviceName,
			Msg:         "Health check pending.",
			Timestamp:   time.Now(),
		}
		pendingStatuses = append(pendingStatuses, pendingStatus)
	}
	return OverallStatus{
		IsHealthy: false,
		Msg:       "Health check pending.",
		DependentServiceStatuses: pendingStatuses,
		Timestamp:                time.Now(),
	}
}

func stoppedHealthCheckStatus(serviceNames []string) OverallStatus {
	var stoppedStatuses []Status
	for _, serviceName := range serviceNames {
		stoppedStatus := Status{
			IsHealthy:   false,
			ServiceName: serviceName,
			Msg:         "Health check stopped.",
			Timestamp:   time.Now(),
		}
		stoppedStatuses = append(stoppedStatuses, stoppedStatus)
	}
	return OverallStatus{
		IsHealthy: false,
		Msg:       "Health check stopped.",
		DependentServiceStatuses: stoppedStatuses,
		Timestamp:                time.Now(),
	}
}
