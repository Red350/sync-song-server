package main

import (
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

type Message struct {
	Username string
	Content  string
}

func (m Message) String() string {
	return fmt.Sprintf("{Username: %s, Content: %q}", m.Username, m.Content)
}

type Lobby struct {
	ID         string       `json:"id"`
	Name       string       `json:"name"`
	Genre      string       `json:"genre"`
	Public     bool         `json:"public"`
	Clients    []Client     `json:"-"`
	NumMembers int          `json:"numMembers"`
	InMsgs     chan Message `json:"-"`
}

func NewLobby(id, name, genre string, public bool) *Lobby {
	lobby := Lobby{
		ID:         id,
		Name:       name,
		Genre:      genre,
		Public:     public,
		Clients:    []Client{},
		NumMembers: 0,
		InMsgs:     make(chan Message, 10),
	}

	// TODO this should be moved to where lobbies are created
	go listenForClientMsgs(&lobby)
	return &lobby
}

func (l *Lobby) join(conn *websocket.Conn, username string) Client {
	// Each client shares the same InMsg channel, allowing the server to
	// conveniently read from all clients.
	client := NewClient(conn, username, l.InMsgs)
	l.NumMembers++

	go client.ReadIncomingMessages()

	l.Clients = append(l.Clients, client)
	return client
}

func listenForClientMsgs(l *Lobby) {
	for {
		msg := <-l.InMsgs
		log.Printf("Received message: %q from client %d", msg.Content, msg.Username)
		// TODO currently this echoes the message back to the client that sent it.
		// May want to change that in the future.
		for _, c := range l.Clients {
			if err := c.Send(msg); err != nil {
				log.Printf("Failed to send message %s to %s: %s", msg, c.Username, err)
			}
		}
	}
}
