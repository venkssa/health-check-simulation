package healthcheck2

import (
	"sync"
	"time"
)

type HealthUpdater interface {
	GetApplicationStatus() ApplicationStatus
}

type BlockingHealthUpdater struct {
	checker ApplicationHealthChecker
}

func NewBlockingHealthUpdater(checker ApplicationHealthChecker) HealthUpdater {
	return BlockingHealthUpdater{checker}
}

func (bu BlockingHealthUpdater) GetApplicationStatus() ApplicationStatus {
	return bu.checker.GetApplicationStatus()
}

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

	if isStaleStatus(status, updateAfter) {
		bu.lock.Lock()
		defer bu.lock.Unlock()

		if isStaleStatus(bu.status, bu.updateAfter) {
			bu.status = bu.checker.GetApplicationStatus()
		}
		status = bu.status
	}
	return status
}

func isStaleStatus(status ApplicationStatus, updateAfter time.Duration) bool {
	return time.Now().Sub(status.Timestamp) >= updateAfter
}

type BatchedBlockingUsingChannelsHealthUpdater struct {
	statusChan chan<- chan<- ApplicationStatus
}

func NewBatchedBlockingUsingChannelsHealthUpdater(checker ApplicationHealthChecker,
	updateAfter time.Duration) HealthUpdater {

	statusChan := make(chan chan<- ApplicationStatus)
	updater := &BatchedBlockingUsingChannelsHealthUpdater{statusChan: statusChan}

	go func() {
		status := ApplicationStatus{
			IsHealthy: false,
			Timestamp: time.Time{},
			Msg:       "Health Check not run",
		}
		for {
			select {
			case resultChan := <-statusChan:
				if isStaleStatus(status, updateAfter) {
					status = checker.GetApplicationStatus()
				}
				resultChan <- status.Copy()
			}
		}
	}()

	return updater
}

func (bu *BatchedBlockingUsingChannelsHealthUpdater) GetApplicationStatus() ApplicationStatus {
	resultChan := make(chan ApplicationStatus)
	bu.statusChan <- resultChan
	return <-resultChan
}

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

type NonBlockingUsingChannelsHealthUpdater struct {
	statusChan chan<- chan<- ApplicationStatus
}

func NewNonBlockingUsingChannelsHealthUpdater(checker ApplicationHealthChecker,
	updateAfter time.Duration) HealthUpdater {

	statusChan := make(chan chan<- ApplicationStatus)

	updateChan := make(chan ApplicationStatus)
	updater := &NonBlockingUsingChannelsHealthUpdater{statusChan: statusChan}

	go func() {
		updateTimer := time.After(0 * time.Second)
		for {
			select {
			case <-updateTimer:
				updateChan <- checker.GetApplicationStatus()
				updateTimer = time.After(updateAfter)
			}
		}
	}()

	go func() {
		status := ApplicationStatus{
			IsHealthy: false,
			Timestamp: time.Time{},
			Msg:       "Health Check not run",
		}
		for {
			select {
			case status = <-updateChan:
			case resultChan := <-statusChan:
				resultChan <- status.Copy()
			}
		}
	}()

	return updater
}

func (nu *NonBlockingUsingChannelsHealthUpdater) GetApplicationStatus() ApplicationStatus {
	resultChan := make(chan ApplicationStatus)
	nu.statusChan <- resultChan
	return <-resultChan
}
