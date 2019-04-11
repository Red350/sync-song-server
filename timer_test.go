package main

import (
	"os"
	"testing"
	"time"
)

var testNow = time.Unix(123, 0)
var testFuture = time.Unix(124, 0)

func replaceTimeNow(t time.Time) {
	timeNow = func() time.Time {
		return t
	}
}

// Performs setup and teardown for tests.
func TestMain(m *testing.M) {
	// Replace timeNow function with a testable implementation
	// that always returns the same time.
	replaceTimeNow(testNow)

	// Run the test and exit with the result.
	retCode := m.Run()
	os.Exit(retCode)
}

func TestMillisToDuration(t *testing.T) {
	testCases := []struct {
		millis int64
		want   time.Duration
	}{
		{1, time.Duration(1000000)},
		{0, time.Duration(0)},
		{-1, time.Duration(-1000000)},
		{123456, time.Duration(123456000000)},
	}

	for _, tc := range testCases {
		got := millisToDuration(tc.millis)
		if got != tc.want {
			t.Errorf("millisToDuration %d, got: %d, want: %d", tc.millis, got, tc.want)
		}
	}
}

func TestStart_SetsStartTime(t *testing.T) {
	timer := NewMillisTimer(9999999, func() {})

	if timer.start != testNow {
		t.Errorf("Start time incorrect: got %v, want: %v", timer.start, testNow)
	}
}

func TestStop_StopsTimer(t *testing.T) {
	timer := NewMillisTimer(9999999, func() {})
	timer.Stop()

	got := timer.timer.Stop()
	if got {
		t.Errorf("Stop did not stop the timer")
	}
}

func TestTimePassed(t *testing.T) {
	timer := NewMillisTimer(9999999, func() {})
	replaceTimeNow(testFuture)

	testCases := []struct {
		offset int64
		want   int64
	}{
		{0, 1000},
		{-10, 990},
		{10, 1010},
		{-1010, -10},
	}

	for _, tc := range testCases {
		got := timer.TimePassed(tc.offset)
		if got != tc.want {
			t.Errorf("TimePassed returned incorrect time: go: %d, want: %d", got, tc.want)
		}
	}
}

func TestNowMillis(t *testing.T) {

}
