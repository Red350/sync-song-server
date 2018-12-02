package main

import (
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn    *websocket.Conn
	ID      int
	InMsgs  chan Message
	OutMsgs chan string
}

func NewClient(conn *websocket.Conn, id int, inMsgs chan Message) Client {
	return Client{
		Conn:    conn,
		ID:      id,
		InMsgs:  inMsgs,
		OutMsgs: make(chan string, 10),
	}
}

func (c *Client) Send(msg Message) error {
	return c.Conn.WriteMessage(websocket.TextMessage, []byte(msg.String()))
}

func (c *Client) ReadIncomingMessages() {
	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		c.InMsgs <- Message{c.ID, string(msg)}
	}
}
