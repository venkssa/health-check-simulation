package healthcheck

import "time"

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
