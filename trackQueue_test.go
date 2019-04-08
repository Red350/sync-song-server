package main

import "testing"

func TestPush_AddsToQueue(t *testing.T) {
	q := TrackQueue{}
	q.Push(nil)
	if q[0] != nil {
		t.Errorf("Push did not add element to queue")
	}
}

func TestPop_RemovesFromQueue(t *testing.T) {
	q := TrackQueue{}
	q.Push(&Track{URI: "1"})
	track := q.Pop()
	if track.URI != "1" {
		t.Errorf("Pop did not remove element from queue")
	}
}

func TestQueueIsFIFO(t *testing.T) {
	q := TrackQueue{}
	q.Push(&Track{URI: "1"})
	q.Push(&Track{URI: "2"})
	q.Push(&Track{URI: "3"})
	first := q.Pop()
	second := q.Pop()
	third := q.Pop()
	if first.URI != "1" || second.URI != "2" || third.URI != "3" {
		t.Errorf("Queue did not return elements in FIFO fashion")
	}
}

//func TestPush(t *testing.T) {
//	testCases := []struct {
//		command ClientCommand
//		name    string
//		want    int
//	}{
//		{C_HANDSHAKE, "C_HANDSHAKE", 1},
//		{ADD_SONG, "ADD_SONG", 2},
//		{VOTE_SKIP, "VOTE_SKIP", 3},
//		{PROMOTE, "PROMOTE", 4},
//		{STATE, "STATE", 5},
//	}
//
//	for _, tc := range testCases {
//		got := int(tc.command)
//		if got != tc.want {
//			t.Errorf("%s incorrect ordinal, got: %d, want %d", tc.name, got, tc.want)
//		}
//	}
//}

func TestIsEmpty_WhenEmpty(t *testing.T) {
	q := TrackQueue{}
	if !q.IsEmpty() {
		t.Errorf("IsEmpty returned false on empty queue")
	}
}

func TestIsEmpty_WhenNotEmpty(t *testing.T) {
	q := TrackQueue{}
	q.Push(nil)
	if q.IsEmpty() {
		t.Errorf("IsEmpty returned false on empty queue")
	}
}
