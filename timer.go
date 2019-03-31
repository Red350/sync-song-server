package main

import "time"

// Inspired by https://stackoverflow.com/a/34892121/11184227
// MillisTimer allows for checking when the timer will expire.
type MillisTimer struct {
	timer *time.Timer
	start time.Time
}

func NewMillisTimer(millis int64, f func()) *MillisTimer {
	durationNanos := time.Duration(millis) * time.Millisecond
	start := time.Now()
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

func (mt *MillisTimer) TimePassed() int64 {
	return time.Now().Sub(mt.start).Nanoseconds() / int64(time.Millisecond)
}