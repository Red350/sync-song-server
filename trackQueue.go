package main

type TrackQueue []*Track

func (q *TrackQueue) Push(t *Track) {
	*q = append(*q, t)
}

func (q *TrackQueue) Pop() *Track {
	t := (*q)[0]
	*q = (*q)[1:]
	return t
}

func (q *TrackQueue) IsEmpty() bool {
	return len(*q) == 0
}
