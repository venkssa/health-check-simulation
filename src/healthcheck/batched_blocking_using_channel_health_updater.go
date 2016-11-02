package healthcheck

import "time"

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
				if IsStaleStatus(status, updateAfter) {
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
