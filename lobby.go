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

// The amount of delay in millis to be added to a command.
// TODO could calculate this based off of client latency, but half a second seems decent for now.
const COMMAND_DELAY int64 = 500

type Lobby struct {
	ID           string             `json:"id"`
	Name         string             `json:"name"`
	LobbyMode    LobbyMode          `json:"lobbyMode"`
	Genre        string             `json:"genre"`
	Public       bool               `json:"public"`
	Admin        string             `json:"admin"`
	CurrentTrack *Track             `json:"currentTrack"`
	TrackQueue   TrackQueue         `json:"trackQueue"`
	Clients      map[string]*Client `json:"-"`
	// Go does not have a way to get the keys of a map without looping over it,
	// so storing them separately is a more performant way to keep track of them.
	ClientNames []string
	SkipVotes   map[string]bool `json"-"`
	NumMembers  int             `json:"numMembers"`
	InMsgs      chan Message    `json:"-"`
	TrackTimer  *MillisTimer    `json:"-"`
}

func NewLobby(id string, name string, lobbyMode LobbyMode, genre string, public bool, admin string, track *Track) *Lobby {
	lobby := Lobby{
		ID:         id,
		Name:       name,
		LobbyMode:  lobbyMode,
		Genre:      genre,
		Public:     public,
		Admin:      admin,
		TrackQueue: TrackQueue{},
		Clients:    make(map[string]*Client),
		SkipVotes:  make(map[string]bool),
		NumMembers: 0,
		InMsgs:     make(chan Message, 10),
	}

	// TODO maybe this should be moved to where lobbies are created
	go lobby.listenForClientMsgs()

	// Check if we are resuming a lobby with an existing track.
	if track != nil {
		msg := Message{}
		lobby.playTrack(&msg, track)
		lobby.sendToAll(msg)
	}
	return &lobby
}

func (l *Lobby) join(conn *websocket.Conn, username string) Client {
	// Each client shares the same InMsg channel, allowing the server to
	// conveniently read from all clients.
	client := NewClient(conn, username, l)

	// Inform clients that a new user has joined.
	l.sendServerMessage(fmt.Sprintf("%s has joined the lobby.", username))

	// Read messages from the new client.
	go func() {
		err := client.ReadIncomingMessages()
		l.log(fmt.Sprintf("%s disconnected: %s", client.Username, err))
		l.sendServerMessage(fmt.Sprintf("%s disconnected.", client.Username))
		l.disconnect(&client)
	}()

	l.NumMembers++
	l.Clients[username] = &client
	l.ClientNames = append(l.ClientNames, username)

	// Make this user the admin if there is none.
	if l.Admin == "" {
		l.promoteToAdmin(username)
	}

	// Send the initial state of the lobby to the client.
	// Disabled for now, as the client requests state instead.
	//l.sendInitialState(&client)

	// Update all clients' state to inform them of the new client.
	l.sendStateToAll()

	return client
}

// Remove the client from the active lobby clients and update state for other clients.
func (l *Lobby) disconnect(client *Client) {
	delete(l.Clients, client.Username)
	// Find and delete the users name from ClientNames.
	for i, name := range l.ClientNames {
		if name == client.Username {
			l.ClientNames = append(l.ClientNames[:i], l.ClientNames[i+1:]...)
			break
		}
	}
	l.NumMembers--

	// Remove any outstanding votes for this client.
	delete(l.SkipVotes, client.Username)

	// Check if we need to promote someone to admin.
	if client.Username == l.Admin {
		// No clients left in the lobby, clear the admin spot.
		if len(l.Clients) == 0 {
			l.log("Lobby empty, clearing admin spot")
			l.Admin = ""
		} else {
			// Go maps are randomly ordered, so this will select a random client.
			for newAdmin := range l.Clients {
				l.promoteToAdmin(newAdmin)
				break
			}
		}
	}

	l.sendStateToAll()
}

// listenForClientMsgs listens to the lobby's InMsgs chan for any messages from clients
// and performs actions based on their content.
func (l *Lobby) listenForClientMsgs() {
	for {
		l.log("Waiting for client message")
		inMsg := <-l.InMsgs
		// TODO could do all this inside of a goroutine, otherwise a single thread is dealing with all user requests.
		// Though maybe its better not to, to avoid race conditions.
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
				l.queueOrPlay(&outMsg, inMsg.CurrentTrack)
			case ADMIN_CONTROLLED:
				// TODO return an error here if the user isn't an admin.
				if inMsg.Username == l.Admin {
					l.queueOrPlay(&outMsg, inMsg.CurrentTrack)
				}
			}
		case VOTE_SKIP:
			// Vote to skip works the same in all lobby modes.
			l.log(fmt.Sprintf("Skip vote received from %s", inMsg.Username))
			if _, ok := l.SkipVotes[inMsg.Username]; !ok {
				// Inform all users of the vote.
				l.sendServerMessage(fmt.Sprintf("%s voted to skip.", inMsg.Username))
			}
			l.SkipVotes[inMsg.Username] = true
			if l.countVotes() {
				// Skip to the next song
				l.log("Skip vote passed")
				// Inform all users that the vote passed.
				l.sendServerMessage("Skip vote passed.")
				l.playNext(&outMsg)
			}
		case PROMOTE:
			// TODO update this to not continue, and instead send from within this function.
			if inMsg.Username == l.Admin {
				l.promoteToAdmin(inMsg.Admin)
			}
			continue
		case STATE:
			// For a state command, we only want to send the state to the client who requested it.
			l.setStateMessageWithCommand(&outMsg)
			l.Clients[inMsg.Username].Send(outMsg)
			continue
		}

		// No harm in always sending the current lobby state to ensure clients stay in sync with it.
		l.setStateMessage(&outMsg)

		// Send the response message.
		l.sendToAll(outMsg)
	}
}

// sendServerMessage asynchronously sends a server message to all users.
func (l *Lobby) sendServerMessage(msg string, a ...interface{}) {
	l.sendToAll(Message{UserMsg: fmt.Sprintf(msg, a...)})
}

// SetCurrentTrack sets the current track to the provided track, persists it to the database,
// and clears any outstanding skip votes.
func (l *Lobby) SetCurrentTrack(track *Track) {
	l.CurrentTrack = track

	// Clear any outstanding votes to skip the previous track.
	l.SkipVotes = make(map[string]bool)

	// Update the database.
	l.persistCurrentTrackState()
}

// playTrack adds the track and PLAY command to the message struct, and calls SetCurrentTrack.
// It then starts a timer to keep track of when the song will end.
func (l *Lobby) playTrack(msg *Message, track *Track) {
	l.log("Playing %#v", track)
	// Update lobby state with regards to the current track.
	l.SetCurrentTrack(track)

	// If there is no current track send a pause command.
	if track == nil {
		l.sendServerMessage("No tracks in queue, add a track to play it now.")
		msg.Command = Command(PAUSE)
		return
	}

	// Inform lobby that new track is playing.
	l.sendServerMessage("Now playing: %s - %s", track.Name, track.Artist)

	msg.CurrentTrack = track
	msg.Command = Command(PLAY)
	msg.Timestamp = NowMillis() + COMMAND_DELAY

	// Start the timer for when the track will end.
	if l.TrackTimer != nil {
		l.TrackTimer.Stop() // Stop any current timer.
	}
	l.log("Starting track timer: %s: %d", track.Name, track.Duration)
    // Set the timer for two seconds before the end of the song.
    // This will hopefully allow the command for the next song to arrive
    // before the song ends, preventing Spotify from issuing its own
    // play command.
	l.TrackTimer = NewMillisTimer(track.Duration - 2000, func() {
		l.log("Timer ended for %s, starting next song", track.Name)
		l.TrackTimer = nil
		msg := Message{}
		l.playNext(&msg)
		l.setStateMessage(&msg)
		l.sendToAll(msg)
	})
}

// playNext pops the next track from the queue, updates the database, and calls playTrack.
func (l *Lobby) playNext(msg *Message) {
	var nextTrack *Track = nil
	if !l.TrackQueue.IsEmpty() {
		nextTrack = l.TrackQueue.Pop()
		l.persistQueueState()
	}
	l.playTrack(msg, nextTrack)
}

// queueOrPlay queues the track if another track is already playing, otherwise
// plays it immediately.
func (l *Lobby) queueOrPlay(msg *Message, track *Track) {
	if l.CurrentTrack == nil {
		l.playTrack(msg, track)
	} else {
		l.addToQueue(track)
		msg.Command = Command(QUEUE)
		// This is redundant as the queue is currently always added, but that may be changed in future.
		msg.TrackQueue = l.TrackQueue
	}
}

func (l *Lobby) promoteToAdmin(newAdmin string) {
	// Check that the the user being promoted is actually a lobby member.
	if _, ok := l.Clients[newAdmin]; !ok {
		l.log(fmt.Sprintf("Failed to promote %s to admin, not a lobby member"))
		return
	}

	l.Admin = newAdmin
	promoteStr := fmt.Sprintf("%s promoted to admin", newAdmin)
	promoteMsg := Message{UserMsg: promoteStr}
	l.setStateMessage(&promoteMsg)

	l.log(promoteStr)
	l.sendToAll(promoteMsg)
}

// addToQueue adds the provided track to the track queue.
func (l *Lobby) addToQueue(track *Track) {
	l.log(fmt.Sprintf("Adding track to queue: %#v", track))
	l.TrackQueue.Push(track)
	l.persistQueueState()
}

// Returns true if more than half the lobby members have voted to skip, otherwise false.
func (l *Lobby) countVotes() bool {
	return len(l.SkipVotes) > (l.NumMembers / 2)
}

// sendToAll asynchronously sends the provided message to all this lobby's clients.
func (l *Lobby) sendToAll(msg Message) {
	go func() {
		for _, c := range l.Clients {
			if err := c.Send(msg); err != nil {
				l.log(fmt.Sprintf("Failed to send message %#v to %s: %s", msg, c.Username, err))
			}
		}
	}()
}

// setStateMessageWithCommand calls setStateMessage, but also
// adds a the relevant command to update play position.
func (l *Lobby) setStateMessageWithCommand(msg *Message) {
	l.setStateMessage(msg)
	if l.TrackTimer != nil && msg.CurrentTrack != nil {
		msg.Command = Command(SEEK_TO)
	}
}

// setStateMessage loads the lobby state into the provided message.
// Adds the timestamp of the current track offset by a second, and
// also the timestamp at which point the command should be execute.
func (l *Lobby) setStateMessage(msg *Message) {
	msg.CurrentTrack = l.CurrentTrack

	// If there is a track timer running, add the position and a timestamp
	// to the message.
	if l.TrackTimer != nil && msg.CurrentTrack != nil {
		msg.CurrentTrack.Position = l.TrackTimer.TimePassed(COMMAND_DELAY)
		msg.Timestamp = NowMillis() + COMMAND_DELAY
	}
	msg.TrackQueue = l.TrackQueue
	msg.Admin = l.Admin
	msg.ClientNames = l.ClientNames
}

// sendState sends the current state of the lobby to a client.
func (l *Lobby) sendStateToAll() {
	l.log("Sending state to all clients")
	stateMsg := Message{}
	l.setStateMessage(&stateMsg)
	l.sendToAll(stateMsg)
}

func (l *Lobby) sendInitialState(c *Client) {
	// TODO maybe introduce a state command?
	stateMsg := Message{}
	l.setStateMessage(&stateMsg)
	if l.CurrentTrack != nil {
		stateMsg.Command = Command(PLAY)
	}
	l.log(fmt.Sprintf("Sending lobby state to %s", c.Username))
	c.Send(stateMsg)
}

// persistCurrentTrackState asynchronously writes the current track to the database.
func (l *Lobby) persistCurrentTrackState() {
	go func() {
		if err := persistCurrentTrack(l); err != nil {
			l.log(fmt.Sprintf("Failed to persist current track: %s", err))
			return
		}
		l.log("Current track state written to db")
	}()
}

// persistQueueState asynchronously writes the queue to the database.
func (l *Lobby) persistQueueState() {
	go func() {
		if err := persistQueue(l); err != nil {
			l.log(fmt.Sprintf("Failed to persist queue: %s", err))
			return
		}
		l.log("Queue state written to db")
	}()
}

// log logs a message with the lobby ID prefixed.
func (l *Lobby) log(msg string, a ...interface{}) {
	log.Printf(fmt.Sprintf("%s: %s", l.ID, msg), a...)
}
