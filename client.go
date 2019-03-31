package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// Client represents a single user who is connected to the server.
type Client struct {
	Conn     *websocket.Conn
	Username string
	Lobby    *Lobby
	Latency  int64
	// Offset is the time in millis by which the app is ahead, negative meaning it is behind.
	// Stored as int64 to avoid the need to cast when adding to message timestamp.
	Offset    int64
	sendMutex sync.Mutex
}

// NewClient is a convenience method for initialising a Client.
func NewClient(conn *websocket.Conn, username string, lobby *Lobby) Client {
	client := Client{
		Conn:     conn,
		Username: username,
		Lobby:    lobby,
	}

	client.log("Starting handshake")
	if err := performClockHandshake(&client); err != nil {
		log.Printf("Failed to perform clock handshake: %s", err)
	}
	client.log(fmt.Sprintf("Handshake complete: latency: %d, offset:%d", client.Latency, client.Offset))

	return client
}

// Send sends a message to this client using their websocket.
func (c *Client) Send(msg Message) error {
	c.sendMutex.Lock()
	defer c.sendMutex.Unlock()

	// Update the timestamp based on this client's offset.
	if ServerCommand(msg.Command) != S_HANDSHAKE && msg.Timestamp != 0 {
		c.log(fmt.Sprintf("Modifying outgoing timestamp %d by %d", msg.Timestamp, c.Offset))
		msg.Timestamp += c.Offset
	}
	c.log(fmt.Sprintf("Sending message: %#v", msg))
	return c.Conn.WriteJSON(msg)
}

// ReadIncomingMessages loops forever, reading incoming messages from this client's connection,
// and putting them in the lobby's InMsgs channel.
// Should be called asynchronously.
func (c *Client) ReadIncomingMessages() error {
	for {
		msg := Message{}
		if err := c.Conn.ReadJSON(&msg); err != nil {
			return fmt.Errorf("failed to read message: %s", err)
		}
		msg.Username = c.Username
		c.log(fmt.Sprintf("Received message: %#v", msg))
		c.Lobby.InMsgs <- msg
	}
}

func (c *Client) log(msg string) {
	c.Lobby.log(fmt.Sprintf("%s: %s", c.Username, msg))
}
