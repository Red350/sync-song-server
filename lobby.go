package main

import (
	"log"

	"github.com/gorilla/websocket"
)

// Go equivalent to enum.
type LobbyMode int

const (
	ADMIN_CONTROLLED LobbyMode = iota
	FREE_FOR_ALL
	ROUND_ROBIN
)

type Lobby struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	LobbyMode    LobbyMode         `json:"lobbyMode"`
	Genre        string            `json:"genre"`
	Public       bool              `json:"public"`
	Admin        string            `json:"admin"`
	CurrentTrack Track             `json:"currentTrack`
	TrackQueue   TrackQueue        `json:"trackQueue"`
	Clients      map[string]Client `json:"-"`
	NumMembers   int               `json:"numMembers"`
	InMsgs       chan Message      `json:"-"`
}

func NewLobby(id string, name string, lobbyMode LobbyMode, genre string, public bool, admin string) *Lobby {
	lobby := Lobby{
		ID:         id,
		Name:       name,
		LobbyMode:  lobbyMode,
		Genre:      genre,
		Public:     public,
		Admin:      admin,
		TrackQueue: TrackQueue{},
		Clients:    make(map[string]Client),
		NumMembers: 0,
		InMsgs:     make(chan Message, 10),
	}

	// TODO maybe this should be moved to where lobbies are created
	go lobby.listenForClientMsgs()
	return &lobby
}

func (l *Lobby) join(conn *websocket.Conn, username string) Client {
	// Each client shares the same InMsg channel, allowing the server to
	// conveniently read from all clients.
	client := NewClient(conn, username, l.InMsgs)
	l.NumMembers++

	go func() {
		err := client.ReadIncomingMessages()
		log.Printf("User disconnected: %s", err)
		l.disconnect(&client)
	}()

	l.Clients[username] = client

	// Make this user the admin if there is none.
	if l.Admin == "" {
		l.Admin = username
	}

	// Send the state of the lobby to the client.
	l.sendState(&client)

	return client
}

// Remove the client from the active lobby clients.
func (l *Lobby) disconnect(client *Client) {
	delete(l.Clients, client.Username)
	l.NumMembers--

	// Check if we need to promote someone to admin.
	if client.Username == l.Admin {
		// Go maps are randomly ordered, so this will select a random client.
		for newAdmin := range l.Clients {
			log.Printf("Promoting %s to admin", newAdmin)
			l.Admin = newAdmin
			break
		}
	}
}

// listenForClientMsgs listens to the lobby's InMsgs chan for any messages from clients.
func (l *Lobby) listenForClientMsgs() {
	for {
		inMsg := <-l.InMsgs
		log.Printf("Received message: %q from client %d", inMsg, inMsg.Username)
		outMsg := Message{Username: inMsg.Username}

		// Attach a user message to the outgoing message if exists.
		if inMsg.UserMsg != "" {
			outMsg.UserMsg = inMsg.UserMsg
		}

		// Parse the command.
		command := ClientCommand(inMsg.Command)
		switch command {
		case ADD_SONG:
			switch l.LobbyMode {
			case FREE_FOR_ALL:
				l.addToQueue(inMsg.CurrentTrack)
			case ADMIN_CONTROLLED:
				// TODO return an error here if the user can't add a command.
				if inMsg.Username == l.Admin {
					l.addToQueue(inMsg.CurrentTrack)
				}
			}
		case VOTE_SKIP:
			// TODO this
		}

		// Send the response message.
		l.sendToAll(outMsg)
	}
}

func (l *Lobby) addToQueue(track Track) {
	log.Printf("Adding track to queue: %#v", track)
	l.TrackQueue.push(track)
}

func (l *Lobby) sendToAll(msg Message) {
	for _, c := range l.Clients {
		if err := c.Send(msg); err != nil {
			log.Printf("Failed to send message %s to %s: %s", msg, c.Username, err)
		}
	}
}

// sendState sends the current state of the lobby to a client.
func (l *Lobby) sendState(c *Client) {
	state := Message{CurrentTrack: l.CurrentTrack}
	log.Printf("Sending lobby state to %s", c.Username)
	c.Send(state)
}
