package main

type Command int

type ClientCommand Command

const (
	C_HANDSHAKE ClientCommand = iota + 1
	ADD_SONG
	VOTE_SKIP
	PROMOTE
	STATE
)

type ServerCommand Command

const (
	S_HANDSHAKE ServerCommand = iota + 1
	PLAY
	PAUSE
	RESUME
	SKIP
	SEEK_TO
	SEEK_RELATIVE
	QUEUE
)
