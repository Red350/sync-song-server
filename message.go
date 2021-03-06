package main

import "fmt"

type Message struct {
	// Who the message originated from (empty string implies the server).
	Username string `json:"username,omitempty"`

	// Track being referenced by the rest of this message.
	// TODO rename this to just Track.
	CurrentTrack *Track `json:"currentTrack,omitempty"`

	// Tracks in the queue.
	TrackQueue []*Track `json:"trackQueue,omitempty"`

	// Clients connected to the lobby.
	ClientNames []string `json:"clientNames,omitempty"`

	// Current lobby admin.
	Admin string `json:"admin,omitempty"`

	// Command for the user to perform e.g. play/pause.
	Command Command `json:"command,omitempty"`

	// User messages.
	UserMsg string `json:"userMsg,omitempty"`

	// Time at which a command should be executed.
	// Also used for the clock handshake.
	Timestamp int64 `json:"timestamp,omitempty"`
}

// Implement stringer interface.
func (m Message) String() string {
	str := fmt.Sprintf("Username: %q, Command: %d, Admin: %q, Clients: %s, UserMsg: %q, Timestamp: %d, TrackQueue: %p", m.Username, m.Command, m.Admin, m.ClientNames, m.UserMsg, m.TrackQueue)
	if m.CurrentTrack != nil {
		str = fmt.Sprintf("%s, Track: %#v", str, *m.CurrentTrack)
	}
	return str
}

type Track struct {
	// Spotify URI for this track.
	URI string `json:"uri,omitempty"`

	// Name of the track.
	Name string `json:"name,omitempty"`

	// Name of the artist.
	Artist string `json:"artist,omitempty"`

	// Duration of the track in millis.
	Duration int64 `json:"duration,omitempty"`

	// Song position in millis.
	Position int64 `json:"position,omitempty"`

	// User who chose this song.
	Username string `json:"username,omitempty"`
}
