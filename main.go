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

var Lobbies = []Lobby{}
var Lobb = NewLobby("AAAA", "Lobby 1")

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
	log.Print("wsHandler request received")

	conn, err := websocket.Upgrade(w, r, w.Header(), 1024, 1024)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		return
	}
	// TODO pass the lobby id as part of the join command
	Lobb.join(conn)
}

func consoleReader() {
	reader := bufio.NewReader(os.Stdin)
	for {
		input, _ := reader.ReadString('\n')
		for _, l := range Lobbies {
			for _, c := range l.Clients {
				c.Send(Message{ClientID: -1, Content: input})
			}
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
