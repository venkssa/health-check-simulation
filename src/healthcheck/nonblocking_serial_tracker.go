package healthcheck

import (
	"sync"
	"time"
)

const defaultFrequency = 3 * time.Second

type NonBlockingSerialTracker struct {
	HealthCheckers

	frequency time.Duration
	start     sync.Once
	stop      sync.Once

	backgroundChecker *backgroundChecker
	resultUpdater     *resultUpdater
}

func (tr *NonBlockingSerialTracker) Register(checker HealthChecker) {
	tr.HealthCheckers = append(tr.HealthCheckers, checker)
}

func (tr *NonBlockingSerialTracker) Start() {
	tr.start.Do(func() {
		updateChan := make(chan OverallStatus)

		tr.backgroundChecker = &backgroundChecker{
			checkers:   append([]HealthChecker{}, tr.HealthCheckers...),
			frequency:  tr.frequency,
			updateChan: updateChan,
			stop:       make(chan struct{}),
		}

		tr.resultUpdater = &resultUpdater{
			updateChan: updateChan,
			lastKnownStatus: make(chan chan OverallStatus),
			stop:       make(chan struct{}),
		}

		go tr.backgroundChecker.Start()
		go tr.resultUpdater.Start(pendingOverallStatus(tr.DependentServiceNames()))
	})
}

func (tr *NonBlockingSerialTracker) Stop() {
	tr.stop.Do(func() {
		tr.backgroundChecker.Stop()
		tr.resultUpdater.Stop()

		tr.backgroundChecker = nil
		tr.resultUpdater = nil
	})
}

func (tr *NonBlockingSerialTracker) GetStatusOfDependentServices() OverallStatus {
	if tr.resultUpdater == nil {
		return stoppedHealthCheckStatus(tr.HealthCheckers.DependentServiceNames())
	}
	return tr.resultUpdater.GetLastUpdatedStatus()
}

type resultUpdater struct {
	updateChan      <-chan OverallStatus
	lastKnownStatus chan chan OverallStatus
	stop            chan struct{}
}

func (t *resultUpdater) Start(initialStatus OverallStatus) {
	lastKnownOverallStatus := initialStatus

	for {
		select {
		case statusChan := <-t.lastKnownStatus:
			statusChan <- lastKnownOverallStatus
		case currentOverallStatus := <-t.updateChan:
			lastKnownOverallStatus = currentOverallStatus
		case <-t.stop:
			return
		}
	}
}

func (t *resultUpdater) Stop() {
	t.stop <- struct{}{}
}

func (t *resultUpdater) GetLastUpdatedStatus() OverallStatus {
	statusChan := make(chan OverallStatus)
	t.lastKnownStatus <- statusChan
	return <- statusChan
}

type backgroundChecker struct {
	checkers   HealthCheckers
	frequency  time.Duration
	updateChan chan<- OverallStatus
	stop       chan struct{}
}

func (bc *backgroundChecker) Stop() {
	bc.stop <- struct{}{}
}

func (bc *backgroundChecker) Start() {
	if bc.frequency == 0 {
		bc.frequency = defaultFrequency
	}

	startFetch := time.After(0 * time.Millisecond)

	for {
		select {
		case <-bc.stop:
			return
		case <-startFetch:
			var statuses []Status
			for _, checker := range bc.checkers {
				statuses = append(statuses, checker.GetStatus())
			}
			startFetch =  time.After(bc.frequency)
			select {
			case bc.updateChan <- buildOverallStatus(statuses):
			}
		}
	}
}
