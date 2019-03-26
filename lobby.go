package main

import (
	"log"

	"github.com/gorilla/websocket"
)

type Track struct {
	// Spotify URI for this track.
	URI string `json:"uri,omitempty"`

	// Name of the track.
	Name string `json:"name,omitempty"`

	// Name of the artist.
	Artist string `json:"artist,omitempty"`

	// Song position in millis.
	Position int64 `json:"position,omitempty"`
}

type Message struct {
	// Who the message originated from (empty string implies the server).
	Username string `json:"username,omitempty"`

	// Spotify URI of the current track in this lobby.
	CurrentTrack Track `json:"currentTrack,omitempty"`

	// Command for the user to perform e.g. play/pause.
	Command string `json:"command,omitempty"`

	// User messages.
	UserMsg string `json:"userMsg,omitemtpy"`
}

type Lobby struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Genre        string       `json:"genre"`
	CurrentTrack Track        `json:"currentTrack`
	Public       bool         `json:"public"`
	Clients      []Client     `json:"-"`
	NumMembers   int          `json:"numMembers"`
	InMsgs       chan Message `json:"-"`
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
		log.Printf("Received message: %q from client %d", msg, msg.Username)
		// TODO currently this echoes the message back to the client that sent it.
		// May want to change that in the future.
		for _, c := range l.Clients {
			if err := c.Send(msg); err != nil {
				log.Printf("Failed to send message %s to %s: %s", msg, c.Username, err)
			}
		}
	}
}

// Send the current state of the lobby to a client.
func (l *Lobby) sendState(c *Client) {
	state := Message{CurrentTrack: l.CurrentTrack}
	log.Printf("Sending lobby state %q to %s", state, c.Username)
	c.Send(state)
}
