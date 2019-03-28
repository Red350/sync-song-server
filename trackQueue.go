package main

import "log"

type TrackQueue []*Track

func (q *TrackQueue) push(t *Track) {
	*q = append(*q, t)
	log.Printf("Queue after push %#v", *q)
}

func (q *TrackQueue) pop() *Track {
	t := (*q)[0]
	*q = (*q)[1:]
	return t
}

func (q *TrackQueue) isEmpty() bool {
	return len(*q) == 0
}
