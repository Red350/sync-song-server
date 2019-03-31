package main

import (
	"fmt"
	"log"
	"math"
	"sort"
)

// HandshakeResponse represents a single outcome from a handshake request.
type HandshakeResponse struct {
	latency int
	offset  int
}

// performHandshake performs a clock handshake between client and server
// to determine the latency and clock offset for this client.
func performClockHandshake(c *Client) error {
	var responses []HandshakeResponse
	for i := 1; i <= 5; i++ {
		serverBefore := NowMillis()
		c.Send(Message{Command: Command(S_HANDSHAKE), Timestamp: serverBefore})
		msg := Message{}
		if err := c.Conn.ReadJSON(&msg); err != nil {
			return fmt.Errorf("failed to read message: %s", err)
		}

		serverAfter := NowMillis()
		if ClientCommand(msg.Command) != C_HANDSHAKE {
			return fmt.Errorf("handshake not completed, received instead: %#v", msg)
		}

		appTime := msg.Timestamp
		latency := (serverAfter - serverBefore) / 2
		// offset is the amount by which the app is ahead, negative meaning behind.
		offset := appTime - serverBefore - latency
		responses = append(responses, HandshakeResponse{latency: int(latency), offset: int(offset)})
	}

	latency, offset := determineLatencyAndOffset(responses)
	c.Latency = latency
	c.Offset = offset
	// Inform the client that the handshake is complete.
	c.Send(Message{Command: Command(S_HANDSHAKE), Timestamp: 0})
	return nil
}

func determineLatencyAndOffset(responses []HandshakeResponse) (int, int) {
	median, mad := medianAbsoluteDeviation(responses)
	upperOutlier := median + 3*mad
	lowerOutlier := median - 3*mad

	var usable []HandshakeResponse
	for _, v := range responses {
		if v.latency < upperOutlier && v.latency > lowerOutlier {
			usable = append(usable, v)
			log.Printf("%d %d Yes", v.latency, v.offset)
		} else {
			log.Printf("%d %d No", v.latency, v.offset)
		}
	}

	var lTot, oTot int
	count := len(usable)
	for _, v := range usable {
		lTot += v.latency
		oTot += v.offset
	}
	return lTot / count, oTot / count
}

// medianAbsoluteDeviation returns the median and the median absolute deviation of the response latencies.
// Assumes that the provided slice has an odd number of elements.
func medianAbsoluteDeviation(responses []HandshakeResponse) (int, int) {
	sort.Slice(responses, func(i, j int) bool { return responses[i].latency < responses[j].latency })
	median := responses[len(responses)/2].latency

	var diffs []int
	for _, v := range responses {
		diffs = append(diffs, int(math.Abs(float64(v.latency-median))))
	}
	sort.Ints(diffs)
	return median, diffs[len(diffs)/2]
}
