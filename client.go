package main

import (
	"log"

	"github.com/gorilla/websocket"
)

// Client represents a single user who is connected to the server.
type Client struct {
	Conn     *websocket.Conn
	Username string
	InMsgs   chan Message
	OutMsgs  chan string
}

// NewClient is a convenience method for initialising a Client.
func NewClient(conn *websocket.Conn, username string, inMsgs chan Message) Client {
	return Client{
		Conn:     conn,
		Username: username,
		InMsgs:   inMsgs,
		OutMsgs:  make(chan string, 10),
	}
}

// Send sends a message to this client using their websocket.
func (c *Client) Send(msg Message) error {
	log.Printf("Sending to %s: %#v", c.Username, msg)
	return c.Conn.WriteJSON(msg)
}

// ReadIncomingMessages loops forever, reading incoming messages from this client's connection,
// and putting them in the InMsgs channel.
// Should be called asynchronously.
func (c *Client) ReadIncomingMessages() error {
	for {
		msg := Message{}
		if err := c.Conn.ReadJSON(&msg); err != nil {
			return err
		}
		msg.Username = c.Username
		c.InMsgs <- msg
	}
}
