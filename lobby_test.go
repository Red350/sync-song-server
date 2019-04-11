package main

import (
	"testing"
)

func TestCountVotes(t *testing.T) {
	l := Lobby{}
	testCases := []struct {
		votes        map[string]bool
		numMembers   int
		wantResult   bool
		wantRequired int
	}{
		// Vote passed, odd members.
		{
			map[string]bool{
				"a": true,
				"b": true,
				"c": true,
			}, 5, true, 0,
		},
		// Vote passed, even members.
		{
			map[string]bool{
				"a": true,
				"b": true,
			}, 2, true, 0,
		},
		// Vote not passed, odd members.
		{
			map[string]bool{
				"a": true,
				"b": true,
			}, 5, false, 1,
		},
		// Vote not passed, even members.
		{
			map[string]bool{
				"a": true,
				"b": true,
			}, 4, false, 1,
		},
		// No votes cast.
		{
			map[string]bool{}, 5, false, 3,
		},
	}

	for _, tc := range testCases {
		l.SkipVotes = tc.votes
		l.NumMembers = tc.numMembers
		gotResult, gotRequired := l.countVotes()
		if gotResult != tc.wantResult {
			t.Errorf("CountVotes incorrect result %v, got: %t, want: %t", tc.votes, gotResult, tc.wantResult)
		}
		if gotRequired != tc.wantRequired {
			t.Errorf("CountVotes incorrect required %v, got: %d, want: %d", tc.votes, gotRequired, tc.wantRequired)
		}
	}
}
