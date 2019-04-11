package main

import (
	"log"
	"os"
	"testing"
)

// Outputs logs to /dev/null during testing.
func suppressLogging() {
	f, _ := os.OpenFile("/dev/null", os.O_WRONLY|os.O_APPEND, 0666)
	log.SetOutput(f)
}

var testClient = &Client{Lobby: &Lobby{}}

func TestMedianAbsoluteDeviation(t *testing.T) {
	testCases := []struct {
		resps      []HandshakeResponse
		wantMedian int
		wantMAD    int
	}{
		// Simple input.
		{
			[]HandshakeResponse{
				{1, 0},
				{2, 0},
				{3, 0},
				{4, 0},
				{5, 0},
			}, 3, 1,
		},
		// Standard input.
		{
			[]HandshakeResponse{
				{34, 0},
				{56, 0},
				{4, 0},
				{64, 0},
				{435, 0},
			}, 56, 22,
		},
		// All zeroes.
		{
			[]HandshakeResponse{
				{0, 0},
				{0, 0},
				{0, 0},
				{0, 0},
				{0, 0},
			}, 0, 0,
		},
	}

	for _, tc := range testCases {
		gotMedian, gotMAD := medianAbsoluteDeviation(tc.resps)
		if gotMedian != tc.wantMedian {
			t.Errorf("Incorrect median for %v, got: %d, want %d", tc.resps, gotMedian, tc.wantMedian)
		}
		if gotMAD != tc.wantMAD {
			t.Errorf("Incorrect MAD for %v, got: %d, want %d", tc.resps, gotMAD, tc.wantMAD)
		}
	}
}

func TestDetermineLatencyAndOffset(t *testing.T) {
	suppressLogging()

	testCases := []struct {
		resps       []HandshakeResponse
		wantLatency int
		wantOffset  int
	}{
		// Normal input.
		{
			[]HandshakeResponse{
				{1, 1},
				{2, 3},
				{3, 3},
				{4, 14},
				{5, 99},
			}, 3, 24,
		},
		// Upper outlier.
		{
			[]HandshakeResponse{
				{12, -2},
				{34, 11},
				{30, 3},
				{15, 0},
				{12312, 12},
			}, 22, 3,
		},
		// Lower outlier.
		{
			[]HandshakeResponse{
				{150, 4},
				{160, 3},
				{233, 4},
				{170, 5},
				{1, 13},
			}, 160, 4,
		},
		// All zeroes
		{
			[]HandshakeResponse{
				{0, 0},
				{0, 0},
				{0, 0},
				{0, 0},
				{0, 0},
			}, 0, 0,
		},
	}

	for _, tc := range testCases {
		gotLatency, gotOffset := determineLatencyAndOffset(testClient, tc.resps)
		if gotLatency != tc.wantLatency {
			t.Errorf("Incorrect latency for %v, got: %d, want %d", tc.resps, gotLatency, tc.wantLatency)
		}
		if gotOffset != tc.wantOffset {
			t.Errorf("Incorrect offset for %v, got: %d, want %d", tc.resps, gotOffset, tc.wantOffset)
		}
	}
}
