package main

import "testing"

func TestClientCommands_CorrectOrdinals(t *testing.T) {
	testCases := []struct {
		command ClientCommand
		name    string
		want    int
	}{
		{C_HANDSHAKE, "C_HANDSHAKE", 1},
		{ADD_SONG, "ADD_SONG", 2},
		{VOTE_SKIP, "VOTE_SKIP", 3},
		{PROMOTE, "PROMOTE", 4},
		{STATE, "STATE", 5},
	}

	for _, tc := range testCases {
		got := int(tc.command)
		if got != tc.want {
			t.Errorf("%s incorrect ordinal, got: %d, want %d", tc.name, got, tc.want)
		}
	}
}

func TestServerCommands_CorrectOrdinals(t *testing.T) {
	testCases := []struct {
		command ServerCommand
		name    string
		want    int
	}{
		{S_HANDSHAKE, "S_HANDSHAKE", 1},
		{PLAY, "PLAY", 2},
		{PAUSE, "PAUSE", 3},
		{RESUME, "RESUME", 4},
		{SKIP, "SKIP", 5},
		{SEEK_TO, "SEEK_TO", 6},
		{SEEK_RELATIVE, "SEEK_RELATIVE", 7},
		{QUEUE, "QUEUE", 8},
	}

	for _, tc := range testCases {
		got := int(tc.command)
		if got != tc.want {
			t.Errorf("%s incorrect ordinal, got: %d, want %d", tc.name, got, tc.want)
		}
	}
}
