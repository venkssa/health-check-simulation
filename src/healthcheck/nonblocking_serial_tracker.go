package healthcheck

import (
	"sync"
	"time"
)

const defaultFrequency = 3 * time.Second

type NonBlockingSerialTracker struct {
	HealthCheckers
	Frequency time.Duration

	start sync.Once
	stop  sync.Once

	statusChecker *statusChecker
	statusKeeper  *statusKeeper
}

func (tr *NonBlockingSerialTracker) Register(checker HealthChecker) {
	tr.HealthCheckers = append(tr.HealthCheckers, checker)
}

func (tr *NonBlockingSerialTracker) Start() {
	tr.start.Do(func() {
		updateChan := make(chan OverallStatus)

		tr.statusChecker = &statusChecker{
			checkers:   append([]HealthChecker{}, tr.HealthCheckers...),
			frequency:  tr.Frequency,
			updateChan: updateChan,
			stop:       make(chan struct{}),
		}

		tr.statusKeeper = &statusKeeper{
			updateChan:      updateChan,
			lastKnownStatus: make(chan chan OverallStatus, 5),
			stop:            make(chan struct{}),
		}

		go tr.statusChecker.Start()
		go tr.statusKeeper.Start(buildUnhealthyOverallStatus(tr.DependentServiceNames(),
			"Health check pending."))
	})
}

func (tr *NonBlockingSerialTracker) Stop() {
	tr.stop.Do(func() {
		tr.statusChecker.Stop()
		tr.statusKeeper.Stop()

		tr.statusChecker = nil
		tr.statusKeeper = nil
	})
}

func (tr *NonBlockingSerialTracker) GetStatusOfDependentServices() OverallStatus {
	if tr.statusKeeper == nil {
		return buildUnhealthyOverallStatus(tr.DependentServiceNames(), "Health check stopped.")
	}
	return tr.statusKeeper.GetLastUpdatedStatus()
}

type statusKeeper struct {
	updateChan      <-chan OverallStatus
	lastKnownStatus chan chan OverallStatus
	stop            chan struct{}
}

func (su *statusKeeper) Start(initialStatus OverallStatus) {
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

func (su *statusKeeper) Stop() {
	su.stop <- struct{}{}
}

func (su *statusKeeper) GetLastUpdatedStatus() OverallStatus {
	statusChan := make(chan OverallStatus)
	su.lastKnownStatus <- statusChan
	return <-statusChan
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
			sc.updateChan <- buildOverallStatus(statuses)
			startFetch = time.After(sc.frequency)
		}
	}
}
