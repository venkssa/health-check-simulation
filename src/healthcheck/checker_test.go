package healthcheck

import (
	"time"
	"sync"
	"testing"
)

func alwaysHelthyStatusFn(serviceName string) Status {
	return Status{
		IsHealthy: true,
		ServiceName: serviceName,
		Msg: "All ok",
		Timestamp: time.Now(),
	}
}

func alwaysFailingStatusFn(serviceName string) Status {
	return Status{
		IsHealthy: false,
		ServiceName: serviceName,
		Msg: "Oops something went wrong",
		Timestamp: time.Now(),
	}
}

type testHealthChecker struct {
	serviceName string
	statusFn func(serviceName string) Status
}

func (hc *testHealthChecker) GetStatus() Status {
	return hc.statusFn(hc.serviceName)
}

func (hc *testHealthChecker) ServiceName() string {
	return hc.serviceName
}

func ConcurrentHealthTrackerTest(tracker Tracker) {
	tracker.Start()
	defer tracker.Stop()

	wg := &sync.WaitGroup{}
	wg.Add(2)

	for idx := 0; idx < 2; idx++ {
		go func() {
			tracker.GetStatusOfDependentServices()
			wg.Done()
		}()
	}

	wg.Wait()
}

func TestChan(t *testing.T) {
	c := make(chan Tracker, 1)

	tr := new(BlockingSerialTracker)

	tr.Register(&testHealthChecker{"firstChecker", alwaysHelthyStatusFn})

	c <- tr

	r := <- c

	t.Log(r.GetStatusOfDependentServices())
}
