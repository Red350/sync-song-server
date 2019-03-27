package main

type Command int

type ClientCommand Command

const (
	ADD_SONG ClientCommand = iota
	VOTE_SKIP
)

type ServerCommand Command

const (
	PLAY ServerCommand = iota
	PAUSE
	RESUME
	SKIP
	SEEK_TO
	SEEK_RELATIVE
	QUEUE
)
