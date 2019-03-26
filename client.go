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
	log.Printf("Sending message %q to %s", msg, c.Username)
	return c.Conn.WriteJSON(msg)
}

// ReadIncomingMessages loops forever, reading incoming messages from this client's connection,
// and puts them in the InMsg channel.
// Should be called asynchronously.
func (c *Client) ReadIncomingMessages() {
	for {
		msg := Message{}
		if err := c.Conn.ReadJSON(&msg); err != nil {
			log.Println(err)
			return
		}
		msg.Username = c.Username
		c.InMsgs <- msg
	}
}