package healthcheck

import (
	"time"
)

type Status struct {
	IsHealthy   bool
	ServiceName string
	Msg         string
	Timestamp   time.Time
}

type HealthChecker interface {
	ServiceName() string
	GetStatus() Status
}
