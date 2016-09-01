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

	backgroundChecker *statusChecker
	resultUpdater     *statusUpdater
}

func (tr *NonBlockingSerialTracker) Register(checker HealthChecker) {
	tr.HealthCheckers = append(tr.HealthCheckers, checker)
}

func (tr *NonBlockingSerialTracker) Start() {
	tr.start.Do(func() {
		updateChan := make(chan OverallStatus)

		tr.backgroundChecker = &statusChecker{
			checkers:   append([]HealthChecker{}, tr.HealthCheckers...),
			frequency:  tr.frequency,
			updateChan: updateChan,
			stop:       make(chan struct{}),
		}

		tr.resultUpdater = &statusUpdater{
			updateChan: updateChan,
			lastKnownStatus: make(chan chan OverallStatus),
			stop:       make(chan struct{}),
		}

		go tr.backgroundChecker.Start()
		go tr.resultUpdater.Start(buildUnhealthyOverallStatus(tr.DependentServiceNames(),
			"Health check pending."))
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
		return buildUnhealthyOverallStatus(tr.DependentServiceNames(), "Health check stopped.")
	}
	return tr.resultUpdater.GetLastUpdatedStatus()
}

type statusUpdater struct {
	updateChan      <-chan OverallStatus
	lastKnownStatus chan chan OverallStatus
	stop            chan struct{}
}

func (su *statusUpdater) Start(initialStatus OverallStatus) {
	lastKnownOverallStatus := initialStatus

	for {
		select {
		case statusChan := <-su.lastKnownStatus:
			statusChan <- lastKnownOverallStatus
		case currentOverallStatus := <-su.updateChan:
			lastKnownOverallStatus = currentOverallStatus
		case <-su.stop:
			return
		}
	}
}

func (su *statusUpdater) Stop() {
	su.stop <- struct{}{}
}

func (su *statusUpdater) GetLastUpdatedStatus() OverallStatus {
	statusChan := make(chan OverallStatus)
	su.lastKnownStatus <- statusChan
	return <- statusChan
}

type statusChecker struct {
	checkers   HealthCheckers
	frequency  time.Duration
	updateChan chan<- OverallStatus
	stop       chan struct{}
}

func (sc *statusChecker) Stop() {
	sc.stop <- struct{}{}
}

func (sc *statusChecker) Start() {
	if sc.frequency == 0 {
		sc.frequency = defaultFrequency
	}

	startFetch := time.After(0 * time.Millisecond)

	for {
		select {
		case <-sc.stop:
			return
		case <-startFetch:
			var statuses []Status
			for _, checker := range sc.checkers {
				statuses = append(statuses, checker.GetStatus())
			}
			startFetch =  time.After(sc.frequency)
			select {
			case sc.updateChan <- buildOverallStatus(statuses):
			}
		}
	}
}
