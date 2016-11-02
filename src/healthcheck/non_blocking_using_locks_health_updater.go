package healthcheck

import (
	"sync"
	"time"
)

type NonBlockingHealthUpdater struct {
	status ApplicationStatus
	lock   sync.RWMutex
}

func NewNonBlockingHealthUpdater(checker ApplicationHealthChecker, updateAfter time.Duration) HealthUpdater {
	updater := NonBlockingHealthUpdater{}

	go func() {
		for {
			status := checker.GetApplicationStatus()
			updater.lock.Lock()
			updater.status = status
			updater.lock.Unlock()
			time.Sleep(updateAfter)
		}
	}()

	return &updater
}

func (nu *NonBlockingHealthUpdater) GetApplicationStatus() ApplicationStatus {
	nu.lock.RLock()
	defer nu.lock.RUnlock()
	return nu.status.Copy()
}
