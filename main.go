package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var Lobbies = make(map[string]*Lobby)

func GetLobbies(w http.ResponseWriter, r *http.Request) {
	log.Print("GetLobbies request received")

	json.NewEncoder(w).Encode(Lobbies)
}

func GetLobby(w http.ResponseWriter, r *http.Request) {
	log.Printf("GetLobby request received")

	var params = mux.Vars(r)

	for _, lobby := range Lobbies {
		if lobby.ID == params["id"] {
			json.NewEncoder(w).Encode(lobby)
		}
	}
}

func CreateLobby(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	name := r.FormValue("name")
	log.Printf("CreateLobby request received: Name: %q, ID: %q", name, id)

	if _, ok := Lobbies[id]; ok {
		log.Printf("Lobby with ID %q already exists", id)
		w.Write([]byte(fmt.Sprintf("Lobby with ID %q already exists", id)))
		return
	}

	l := NewLobby(id, name)
	Lobbies[id] = l
	log.Printf("Lobby %q has been created with ID %q", name, id)

	w.Write([]byte(fmt.Sprintf("Lobby created with ID %q", id)))
}

func JoinLobby(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	log.Printf("JoinLobby request received: %s", id)

	conn, err := websocket.Upgrade(w, r, w.Header(), 1024, 1024)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		return
	}

	if l, ok := Lobbies[id]; ok {
		client := l.join(conn)
		log.Printf("Client %d has joined Lobby %q", client.ID, l.ID)
		client.Send(
			Message{
				ClientID: -1,
				Content:  fmt.Sprintf("{name: %s}", l.Name),
			},
		)
	} else {
		log.Printf("Lobby with ID %q does not exist", id)
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseUnsupportedData, "Lobby does not exist"))
		conn.Close()
		return
	}
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

	router.HandleFunc("/lobbies", GetLobbies).Methods("GET")
	router.HandleFunc("/lobbies/{id}", GetLobby).Methods("GET")
	router.HandleFunc("/lobbies/{id}/join", JoinLobby).Methods("GET")
	router.HandleFunc("/lobbies/create", CreateLobby).Methods("POST")

	fmt.Println("Starting server")
	log.Fatal(http.ListenAndServe(":8080", router))
}
