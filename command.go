package main

type Command int

type ClientCommand Command

const (
	ADD_SONG ClientCommand = iota + 1
	VOTE_SKIP
)

type ServerCommand Command

const (
	PLAY ServerCommand = iota + 1
	PAUSE
	RESUME
	SKIP
	SEEK_TO
	SEEK_RELATIVE
	QUEUE
)
