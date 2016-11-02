package healthcheck

import (
	"sync"
	"time"
)

type BatchedBlockingHealthUpdater struct {
	checker     ApplicationHealthChecker
	status      ApplicationStatus
	updateAfter time.Duration
	lock        sync.RWMutex
}

func NewBatchedBlockingHealthUpdater(checker ApplicationHealthChecker, updateAfter time.Duration) HealthUpdater {
	return &BatchedBlockingHealthUpdater{
		status: ApplicationStatus{
			IsHealthy: false,
			Timestamp: time.Time{},
			Msg:       "Health Check not run",
		},
		checker:     checker,
		updateAfter: updateAfter}
}

func (bu *BatchedBlockingHealthUpdater) GetApplicationStatus() ApplicationStatus {
	bu.lock.RLock()
	status := bu.status
	updateAfter := bu.updateAfter
	bu.lock.RUnlock()

	if IsStaleStatus(status, updateAfter) {
		bu.lock.Lock()
		defer bu.lock.Unlock()

		if IsStaleStatus(bu.status, bu.updateAfter) {
			bu.status = bu.checker.GetApplicationStatus()
		}
		status = bu.status
	}
	return status
}
