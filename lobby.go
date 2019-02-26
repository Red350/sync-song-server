package main

import (
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

type Message struct {
	ClientID int
	Content  string
}

func (m Message) String() string {
	return fmt.Sprintf("{ClientID: %d, Content: %q}", m.ClientID, m.Content)
}

type Lobby struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Genre        string       `json:"genre"`
	Public       bool         `json:"public"`
	Clients      []Client     `json:"-"`
	NextClientID int          `json:"-"`
	NumMembers   int          `json:"numMembers"`
	InMsgs       chan Message `json:"-"`
}

func NewLobby(id, name, genre string, public bool) *Lobby {
	lobby := Lobby{
		ID:           id,
		Name:         name,
		Genre:        genre,
		Public:       public,
		Clients:      []Client{},
		NextClientID: 0,
		NumMembers:   0,
		InMsgs:       make(chan Message, 10),
	}

	// TODO this should be moved to where lobbies are created
	go listenForClientMsgs(&lobby)
	return &lobby
}

func (l *Lobby) join(conn *websocket.Conn) Client {
	client := NewClient(conn, l.NextClientID, l.InMsgs)
	l.NextClientID++
	l.NumMembers++

	go client.ReadIncomingMessages()

	l.Clients = append(l.Clients, client)
	return client
}

func listenForClientMsgs(l *Lobby) {
	for {
		msg := <-l.InMsgs
		log.Printf("Received message: %q from client %d", msg.Content, msg.ClientID)
		// TODO currently this echoes the message back to the client that sent it.
		// May want to change that in the future.
		for _, c := range l.Clients {
			log.Printf("Sending message to %d", c.ID)
			if err := c.Send(msg); err != nil {
				log.Printf("Failed to send message %s to %d: %s", msg, c.ID, err)
			}
		}
	}
}
