package main

import "time"

func NowMillis() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
