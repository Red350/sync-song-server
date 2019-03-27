package main

type Message struct {
	// Who the message originated from (empty string implies the server).
	Username string `json:"username,omitempty"`

	// Spotify URI of the current track in this lobby.
	CurrentTrack Track `json:"currentTrack,omitempty"`

	// Command for the user to perform e.g. play/pause.
	Command string `json:"command,omitempty"`

	// User messages.
	UserMsg string `json:"userMsg,omitempty"`
}

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
