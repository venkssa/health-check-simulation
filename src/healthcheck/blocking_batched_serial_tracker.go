package healthcheck

import (
	"sync"
	"time"
)

const defaultUpdateDuration = 3 * time.Second

type BlockingBatchedSerialTracker struct {
	HealthCheckers
	UpdateDuration time.Duration

	once        sync.Once
	statusMutex sync.RWMutex
	status      OverallStatus
}

func (tr *BlockingBatchedSerialTracker) Register(checker HealthChecker) {
	tr.HealthCheckers = append(tr.HealthCheckers, checker)
}

func (tr *BlockingBatchedSerialTracker) Start() {
	tr.once.Do(func() {
		if tr.UpdateDuration == 0 {
			tr.UpdateDuration = defaultUpdateDuration
		}
		tr.status = OverallStatus{Timestamp: time.Time{}}
	})
}

func (tr *BlockingBatchedSerialTracker) Stop() {}

func (tr *BlockingBatchedSerialTracker) GetStatusOfDependentServices() OverallStatus {
	tr.statusMutex.RLock()
	currentStatus := tr.status
	tr.statusMutex.RUnlock()

	if isStaleStatus(currentStatus, tr.UpdateDuration) {
		tr.statusMutex.Lock()
		defer tr.statusMutex.Unlock()

		if isStaleStatus(currentStatus, tr.UpdateDuration) {
			var statuses []Status
			for _, checker := range tr.HealthCheckers {
				statuses = append(statuses, checker.GetStatus())
			}

			tr.status = buildOverallStatus(statuses)
		}
		currentStatus = tr.status
	}
	return currentStatus
}

func isStaleStatus(status OverallStatus, updateDuration time.Duration) bool {
	return time.Now().Sub(status.Timestamp) >= updateDuration
}
