package healthcheck

import (
	"time"
)

type HealthUpdater interface {
	GetApplicationStatus() ApplicationStatus
}

func IsStaleStatus(status ApplicationStatus, updateAfter time.Duration) bool {
	return time.Now().Sub(status.Timestamp) >= updateAfter
}
