package healthcheck2

import "time"

type StatusFunc func(string) Status

func AlwaysHelthyStatusFn(serviceName string) Status {
	return Status{
		IsHealthy:   true,
		ServiceName: serviceName,
		Msg:         "All ok",
		Timestamp:   time.Now(),
	}
}

func AlwaysFailingStatusFn(serviceName string) Status {
	return Status{
		IsHealthy:   false,
		ServiceName: serviceName,
		Msg:         "Oops something went wrong",
		Timestamp:   time.Now(),
	}
}

func Delay(by time.Duration, fn StatusFunc) StatusFunc {
	return func(serviceName string) Status {
		time.Sleep(by)
		return fn(serviceName)
	}
}

func Signal(signal chan<- struct{}, fn StatusFunc) StatusFunc {
	return func(serviceName string) Status {
		signal <- struct{}{}
		return fn(serviceName)
	}
}

type TestHealthChecker struct {
	serviceName string
	statusFn    func(serviceName string) Status
}

func NewTestHealthChecker(serviceName string, statusFn StatusFunc) TestHealthChecker {
	return TestHealthChecker{serviceName, statusFn}
}

func (hc TestHealthChecker) GetStatus() Status {
	return hc.statusFn(hc.serviceName)
}

func (hc TestHealthChecker) ServiceName() string {
	return hc.serviceName
}
