package healthcheck

type BlockingSerialTracker struct {
	HealthCheckers
}

func (tr *BlockingSerialTracker) Register(checker HealthChecker) {
	tr.HealthCheckers = append(tr.HealthCheckers, checker)
}

// Queries every registered health checker in serial and returns the result.
func (tr *BlockingSerialTracker) GetStatusOfDependentServices() OverallStatus {
	var statuses []Status
	for _, checker := range tr.HealthCheckers {
		statuses = append(statuses, checker.GetStatus())
	}

	return buildOverallStatus(statuses)
}

func (tr *BlockingSerialTracker) Start() {}

func (tr *BlockingSerialTracker) Stop() {}
