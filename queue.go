package main

type TrackQueue []Track

func (q TrackQueue) push(t Track) {
	q = append(q, t)
}

func (q TrackQueue) pop() Track {
	t := q[0]
	q = q[1:]
	return t
}
