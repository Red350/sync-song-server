package main

import (
	"time"
)

var timeNow = func() time.Time {
	return time.Now()
}

// Inspired by https://stackoverflow.com/a/34892121/11184227
// MillisTimer allows for checking when the timer will expire.
type MillisTimer struct {
	timer *time.Timer
	start time.Time
}

func NewMillisTimer(millis int64, f func()) *MillisTimer {
	durationNanos := millisToDuration(millis)
	start := timeNow()
	return &MillisTimer{
		timer: time.AfterFunc(durationNanos, f),
		start: start,
	}
}

func (mt *MillisTimer) Stop() bool {
	if mt == nil {
		return false
	}
	return mt.timer.Stop()
}

// Returns the time passed since this timer started, offset by offsetMillis.
func (mt *MillisTimer) TimePassed(offsetMillis int64) int64 {
	return timeNow().Sub(mt.start).Nanoseconds()/int64(time.Millisecond) + offsetMillis
}

// NowMillis returns the current time in milliseconds.
func NowMillis() int64 {
	return timeNow().UnixNano() / int64(time.Millisecond)
}

// millisToDuration returns the provided time in milliseconds as a Duration.
func millisToDuration(millis int64) time.Duration {
	return time.Duration(millis) * time.Millisecond
}
