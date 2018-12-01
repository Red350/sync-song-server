package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// TODO: might need to mess around with the size of the channel, assuming we even want to use buffered channels at all.
var outgoingMessages = make(chan string, 10)
var clients = []Client{}
var nextClientID = 0

type Client struct {
	Conn     *websocket.Conn
	ID       int
	Incoming chan string
	Outgoing chan string
}

func NewClient(conn *websocket.Conn) Client {
	nextClientID++
	return Client{
		conn,
		nextClientID,
		make(chan string, 10),
		make(chan string, 10),
	}
}

type Lobby struct {
	ID         string `json:"id,omitempty"`
	Name       string `json:"name,omitempty"`
	NumMembers int    `json:"numMembers,omitempty"`
}

var Lobbies = []Lobby{
	{"AAAA", "Lobby 1", 3},
	{"BBBB", "Lobby 2", 4},
	{"CCCC", "Lobby 3", 1},
}

func sayHello(w http.ResponseWriter, r *http.Request) {
	fmt.Println("request received")
	message := r.URL.Path
	message = strings.TrimPrefix(message, "/")
	message = "Hello " + message

	w.Write([]byte(message))
}

func GetLobbies(w http.ResponseWriter, r *http.Request) {
	log.Printf("GetLobbies request received: %s", *r)

	json.NewEncoder(w).Encode(Lobbies)
}

func GetLobby(w http.ResponseWriter, r *http.Request) {
	log.Printf("GetLobby request received: %s", *r)

	var params = mux.Vars(r)

	for _, lobby := range Lobbies {
		if lobby.ID == params["id"] {
			json.NewEncoder(w).Encode(lobby)
		}
	}
}

func JoinLobby(w http.ResponseWriter, r *http.Request) {
	log.Printf("JoinLobby request received: %s", *r)

	var params = mux.Vars(r)

	for i, lobby := range Lobbies {
		if lobby.ID == params["id"] {
			fmt.Printf("Found lobby: %s\n", lobby)
			Lobbies[i].NumMembers++
			json.NewEncoder(w).Encode(Lobbies[i])
		}
	}
}

func CreateLobby(w http.ResponseWriter, r *http.Request) {
	log.Printf("CreateLobby request received: %s", *r)

	id := r.Form.Get("id")
	fmt.Println(id)

	w.Write([]byte("asdf"))
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("wsHandler request received: %s", *r)

	conn, err := websocket.Upgrade(w, r, w.Header(), 1024, 1024)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		return
	}
	client := NewClient(conn)
	clients = append(clients, client)
	go sendMessages(client)
	go readIncomingMessages(client)
}

func sendMessages(c Client) {
	for {
		select {
		case outMsg := <-c.Outgoing:
			if err := c.Conn.WriteMessage(websocket.TextMessage, []byte(outMsg)); err != nil {
				log.Printf("Websocket error: %s", err)
			}
		case inMsg := <-c.Incoming:
			log.Printf("Received message: %s", inMsg)
			for _, other := range clients {
				log.Printf("checking client %s", other)
				if other.ID != c.ID {
					log.Printf("Sending message: %s to %v", inMsg, other)
					if err := other.Conn.WriteMessage(websocket.TextMessage, []byte(inMsg)); err != nil {
						log.Printf("Websocket error: %s", err)
					}
				}
			}
		}
	}
}

func readIncomingMessages(c Client) {
	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		msgString := string(msg)
		c.Incoming <- msgString
	}
}

func consoleReader() {
	reader := bufio.NewReader(os.Stdin)
	for {
		input, _ := reader.ReadString('\n')
		log.Printf("Sending message %s to clients %v", input, clients)
		for _, c := range clients {
			c.Outgoing <- input
		}
	}
}

func main() {
	go consoleReader()
	router := mux.NewRouter()

	router.HandleFunc("/", sayHello).Methods("GET")
	router.HandleFunc("/lobbies", GetLobbies).Methods("GET")
	router.HandleFunc("/lobbies/{id}", GetLobby).Methods("GET")
	router.HandleFunc("/lobbies/{id}/join", JoinLobby).Methods("GET")
	router.HandleFunc("/lobbies/create", JoinLobby).Methods("POST")

	router.HandleFunc("/ws", wsHandler)

	fmt.Println("Starting server")
	log.Fatal(http.ListenAndServe(":8080", router))
	fmt.Println("here i am")
}
