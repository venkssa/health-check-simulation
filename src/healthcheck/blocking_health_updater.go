package healthcheck

type BlockingHealthUpdater struct {
	ApplicationHealthChecker
}

func NewBlockingHealthUpdater(checker ApplicationHealthChecker) HealthUpdater {
	return BlockingHealthUpdater{checker}
}
