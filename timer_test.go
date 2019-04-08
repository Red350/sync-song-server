package main

import (
	"testing"
	"time"
)

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

func TestStop_StopsTimer(t *testing.T) {
	timer := NewMillisTimer(9999999, func() {})
	timer.Stop()

	got := timer.timer.Stop()
	if got {
		t.Errorf("Stop did not stop the timer")
	}
}
