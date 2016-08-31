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
			statusReq:  make(chan struct{}),
			statusResp: make(chan OverallStatus),
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
	})
}

func (tr *NonBlockingSerialTracker) GetStatusOfDependentServices() OverallStatus {
	return tr.resultUpdater.GetLastUpdatedStatus()
}

type resultUpdater struct {
	updateChan <-chan OverallStatus
	statusReq  chan struct{}
	statusResp chan OverallStatus
	stop       chan struct{}
}

func (t *resultUpdater) Start(initialStatus OverallStatus) {
	overallStatus := initialStatus

	for {
		select {
		case <-t.statusReq:
			t.statusResp <- overallStatus
		case update := <-t.updateChan:
			overallStatus = update
		case <-t.stop:
			return
		}
	}
}

func (t *resultUpdater) Stop() {
	t.stop <- struct{}{}
}

func (t *resultUpdater) GetLastUpdatedStatus() OverallStatus {
	t.statusReq <- struct{}{}
	return <-t.statusResp
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

	for {
		select {
		case <-bc.stop:
			return
		case <-time.After(bc.frequency):
			var statuses []Status
			for _, checker := range bc.checkers {
				statuses = append(statuses, checker.GetStatus())
			}
			select {
			case bc.updateChan <- buildOverallStatus(statuses):
			}
		}
	}
}
