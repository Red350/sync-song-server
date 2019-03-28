package main

import (
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

// Go equivalent to enum.
type LobbyMode int

const (
	ADMIN_CONTROLLED LobbyMode = iota + 1
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
	CurrentTrack *Track            `json:"currentTrack`
	TrackQueue   TrackQueue        `json:"trackQueue"`
	Clients      map[string]Client `json:"-"`
	SkipVotes    map[string]bool   `json"-"`
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
		SkipVotes:  make(map[string]bool),
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

	// Remove any outstanding votes for this client.
	delete(l.SkipVotes, client.Username)

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

// listenForClientMsgs listens to the lobby's InMsgs chan for any messages from clients
// and performs actions based on their content.
func (l *Lobby) listenForClientMsgs() {
	for {
		l.log("Waiting for client message")
		inMsg := <-l.InMsgs
		// TODO could do all this inside of a goroutine, otherwise a single thread is dealing with all user requests.
		// Though maybe its better not to, to avoid race conditions.
		log.Printf("Received message: %#v from client %s", inMsg, inMsg.Username)
		outMsg := Message{Username: inMsg.Username}

		// Attach a user message to the outgoing message if exists.
		if inMsg.UserMsg != "" {
			outMsg.UserMsg = inMsg.UserMsg
		}

		// Parse the command and perform any necessary actions.
		command := ClientCommand(inMsg.Command)
		switch command {
		case ADD_SONG:
			switch l.LobbyMode {
			case FREE_FOR_ALL:
				if l.CurrentTrack == nil {
					l.CurrentTrack = inMsg.CurrentTrack
					l.playCurrentTrack()
					continue
				}
				l.addToQueue(inMsg.CurrentTrack)
				outMsg.Command = Command(QUEUE)
				outMsg.TrackQueue = l.TrackQueue
			case ADMIN_CONTROLLED:
				// TODO return an error here if the user can't add a command.
				// TODO populate the out message instead of calling playToAll and continuing.
				if inMsg.Username == l.Admin {
					if l.CurrentTrack == nil {
						l.CurrentTrack = inMsg.CurrentTrack
						l.playCurrentTrack()
						continue
					}
					l.addToQueue(inMsg.CurrentTrack)
					outMsg.Command = Command(QUEUE)
					outMsg.TrackQueue = l.TrackQueue
				}
			}
		case VOTE_SKIP:
			// Vote to skip works the same in all lobby modes.
			l.log(fmt.Sprintf("Skip vote received from %s", inMsg.Username))
			l.SkipVotes[inMsg.Username] = true
			if l.countVotes() {
				// Skip to the next song
				l.log("Skip vote passed")
				if l.TrackQueue.isEmpty() {
					// TODO return an error instead of continuing once errors have been added to the message struct.
					continue
				}
				nextTrack := l.TrackQueue.pop()
				outMsg.CurrentTrack = nextTrack
				outMsg.Command = Command(SKIP)
			}
		}

		// Send the response message.
		l.sendToAll(outMsg)
	}
}

// playCurrentTrack sends a command to all lobby members to play the current track.
func (l *Lobby) playCurrentTrack() {
	log.Printf("Playing %#v", l.CurrentTrack)
	l.sendToAll(Message{
		CurrentTrack: l.CurrentTrack,
		Command:      Command(PLAY),
	})
}

// addToQueue adds the provided track to the track queue.
func (l *Lobby) addToQueue(track *Track) {
	log.Printf("Adding track to queue: %#v", track)
	l.TrackQueue.push(track)
}

// Returns true if more than half the lobby members have voted to skip, otherwise false.
func (l *Lobby) countVotes() bool {
	return len(l.SkipVotes) > (l.NumMembers / 2)
}

// sendToAll sends the provided message to all this lobby's clients.
func (l *Lobby) sendToAll(msg Message) {
	for _, c := range l.Clients {
		if err := c.Send(msg); err != nil {
			log.Printf("Failed to send message %s to %s: %s", msg, c.Username, err)
		}
	}
}

// sendState sends the current state of the lobby to a client.
func (l *Lobby) sendState(c *Client) {
	state := Message{}
	if l.CurrentTrack != nil {
		state.CurrentTrack = l.CurrentTrack
		state.Command = Command(PLAY)
	}
	log.Printf("Sending lobby state to %s", c.Username)
	c.Send(state)
}

// log logs a message with the lobby ID prefixed.
func (l *Lobby) log(msg string) {
	log.Printf("%s: %s", l.ID, msg)
}
