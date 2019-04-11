package main

import (
	"testing"
)

func TestMessageString_OmitsTrackWhenEmpty(t *testing.T) {
	testCases := []struct {
		msg  Message
		want string
	}{
		{
			Message{
				Username: "blah",
				UserMsg:  "asdf",
			},
			`Username: "blah", Command: 0, Admin: "", Clients: [], UserMsg: "asdf", Timestamp: [], TrackQueue: %!p(MISSING)`,
		},
		{
			Message{
				Username:     "blah",
				UserMsg:      "asdf",
				CurrentTrack: &Track{URI: "123"},
			},
			`Username: "blah", Command: 0, Admin: "", Clients: [], UserMsg: "asdf", Timestamp: [], TrackQueue: %!p(MISSING), Track: main.Track{URI:"123", Name:"", Artist:"", Duration:0, Position:0, Username:""}`,
		},
	}

	for _, tc := range testCases {
		got := tc.msg.String()
		if got != tc.want {
			t.Errorf("String %#v, got: %s, want: %s", tc.msg, got, tc.want)
		}
	}
}
