package healthcheck2

import (
	"testing"
	"time"
)

func TestHealthUpdater_GetApplicationStatus(t *testing.T) {
	service1Checker := NewTestHealthChecker("Service1", AlwaysHelthyStatusFn)
	service2Checker := NewTestHealthChecker("Service2", AlwaysHelthyStatusFn)

	simpleAppHealthChecker := NewSimpleApplicationHealthChecker(service1Checker, service2Checker)
	concurrentAppHealthChecker := NewConcurrentApplicationHealthChecker(service1Checker, service2Checker)
	updateAfter := time.Duration(1 * time.Millisecond)

	expectedAppStatus := ApplicationStatus{IsHealthy: true}

	t.Run("BlockingSimpleHealthUpdater",
		testGetApplicationStatus(NewBlockingHealthUpdater(simpleAppHealthChecker), expectedAppStatus))
	t.Run("BlockingConcurrentHealthUpdater",
		testGetApplicationStatus(NewBlockingHealthUpdater(concurrentAppHealthChecker), expectedAppStatus))

	t.Run("BatchedBlockingSimpleHealthUpdater",
		testGetApplicationStatus(NewBatchedBlockingHealthUpdater(simpleAppHealthChecker, updateAfter),
			expectedAppStatus))
	t.Run("BatchedBlockingConcurrentHealthUpdater",
		testGetApplicationStatus(NewBatchedBlockingHealthUpdater(concurrentAppHealthChecker, updateAfter),
			expectedAppStatus))

	t.Run("BatchedBlockingUsingChannelsSimpleHealthUpdater", testGetApplicationStatus(
		NewBatchedBlockingUsingChannelsHealthUpdater(simpleAppHealthChecker, updateAfter),
		expectedAppStatus))
	t.Run("BatchedBlockingUsingChannelsConcurrentHealthUpdater", testGetApplicationStatus(
		NewBatchedBlockingUsingChannelsHealthUpdater(concurrentAppHealthChecker, updateAfter),
		expectedAppStatus))

	simpleAppHealthChecker, signal := simpleSignallingAppHealthChecker()
	t.Run("NonBlockingSimpleHealthUpdater", testNonBlockingGetApplicationStatus(
		NewNonBlockingHealthUpdater(simpleAppHealthChecker, updateAfter), signal, 3,
		expectedAppStatus))
	concurrentAppHealthChecker, signal = concurrentSignallingAppHealthChecker()
	t.Run("NonBlockingConcurrentHealthUpdater", testNonBlockingGetApplicationStatus(
		NewNonBlockingHealthUpdater(concurrentAppHealthChecker, updateAfter), signal, 3,
		expectedAppStatus))

	simpleAppHealthChecker, signal = simpleSignallingAppHealthChecker()
	t.Run("NonBlockingUsingChannelsSimpleHealthUpdater", testNonBlockingGetApplicationStatus(
		NewNonBlockingUsingChannelsHealthUpdater(simpleAppHealthChecker, updateAfter), signal, 3,
		expectedAppStatus))
	concurrentAppHealthChecker, signal = concurrentSignallingAppHealthChecker()
	t.Run("NonBlockingUsingChannelsConcurrentHealthUpdater", testNonBlockingGetApplicationStatus(
		NewNonBlockingUsingChannelsHealthUpdater(concurrentAppHealthChecker, updateAfter), signal, 3,
		expectedAppStatus))
}

func testGetApplicationStatus(updater HealthUpdater, expectedStatus ApplicationStatus) func(*testing.T) {
	return func(t *testing.T) {
		actualStatus := updater.GetApplicationStatus()
		if actualStatus.IsHealthy != expectedStatus.IsHealthy {
			t.Errorf("Expected ishealthy to be %v but was %v", expectedStatus, actualStatus)
		}
	}
}

func testNonBlockingGetApplicationStatus(updater HealthUpdater, signal chan struct{}, numOfSignalToWait uint32,
	expectedStatus ApplicationStatus) func(*testing.T) {
	return func(t *testing.T) {
		for idx := uint32(0); idx < numOfSignalToWait; idx++ {
			<-signal
		}
		go func() {
			<-signal
		}()
		testGetApplicationStatus(updater, expectedStatus)(t)
	}
}

func simpleSignallingAppHealthChecker() (SimpleApplicationHealthChecker, chan struct{}) {
	signal := make(chan struct{})
	simpleAppHealthChecker := NewSimpleApplicationHealthChecker(
		NewTestHealthChecker("Service1", Signal(signal, AlwaysHelthyStatusFn)),
		NewTestHealthChecker("Service2", Signal(signal, AlwaysHelthyStatusFn)))
	return simpleAppHealthChecker, signal
}

func concurrentSignallingAppHealthChecker() (ConcurrentApplicationHealthChecker, chan struct{}) {
	signal := make(chan struct{})
	concurrentAppHealthChecker := NewConcurrentApplicationHealthChecker(
		NewTestHealthChecker("Service1", Signal(signal, AlwaysHelthyStatusFn)),
		NewTestHealthChecker("Service2", Signal(signal, AlwaysHelthyStatusFn)))
	return concurrentAppHealthChecker, signal
}
