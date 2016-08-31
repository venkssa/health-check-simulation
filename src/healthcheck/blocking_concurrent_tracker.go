package healthcheck

type BlockingConcurrentTracker struct {
	HealthCheckers
}

func (tr *BlockingConcurrentTracker) Register(checker HealthChecker) {
	tr.HealthCheckers = append(tr.HealthCheckers, checker)
}

// Queries every registered health checker concurrently and returns the result.
func (tr *BlockingConcurrentTracker) GetStatusOfDependentServices() OverallStatus {
	numOfHealthCheckers := len(tr.HealthCheckers)
	statusChan := make(chan Status, numOfHealthCheckers)

	for idx := 0; idx < numOfHealthCheckers; idx++ {
		go func(checker HealthChecker, result chan<- Status) {
			result <- checker.GetStatus()
		}(tr.HealthCheckers[idx], statusChan)
	}

	statuses := make([]Status, numOfHealthCheckers)
	for idx := 0; idx < numOfHealthCheckers; idx++ {
		statuses[idx] = <-statusChan
	}

	close(statusChan)
	return buildOverallStatus(statuses)
}

func (tr *BlockingConcurrentTracker) Start() {}

func (tr *BlockingConcurrentTracker) Stop() {}
