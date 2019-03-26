package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var Lobbies = make(map[string]*Lobby)

// TODO move this to somewhere better
// Courtesy of https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go/31832326#31832326
const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

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
	id := RandStringBytes(4)
	name := r.FormValue("name")
	genre := r.FormValue("genre")
	public, err := strconv.ParseBool(r.FormValue("public"))
	if err != nil {
		log.Fatal("Public bool formatted incorrectly: %s", err)
	}
	log.Printf("CreateLobby request received: Name: %q, ID: %q, Genre: %q, Public: %t", name, id, genre, public)

	if _, ok := Lobbies[id]; ok {
		log.Printf("Lobby with ID %q already exists", id)
		w.Write([]byte(fmt.Sprintf("Lobby with ID %q already exists", id)))
		return
	}

	l := NewLobby(id, name, genre, public)
	Lobbies[id] = l
	log.Printf("Lobby %q has been created with ID %q", name, id)

	w.Write([]byte(fmt.Sprintf("Lobby created with ID %q", id)))
}

func JoinLobby(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	username := r.URL.Query()["username"][0]
	log.Printf("JoinLobby request received: ID: %s, username: %s", id, username)

	conn, err := websocket.Upgrade(w, r, w.Header(), 1024, 1024)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		return
	}

	if l, ok := Lobbies[id]; ok {
		client := l.join(conn, username)
		log.Printf("%s has joined lobby %q", client.Username, l.ID)
		// TODO send the currently played song.
		client.Send(
			Message{
				Content: fmt.Sprintf("{name: %s}", l.Name),
			},
		)
	} else {
		log.Printf("Lobby with ID %q does not exist", id)
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseUnsupportedData, "Lobby does not exist"))
		conn.Close()
		return
	}
}

func consoleBroadcaster() {
	reader := bufio.NewReader(os.Stdin)
	for {
		input, _ := reader.ReadString('\n')
		for _, l := range Lobbies {
			for _, c := range l.Clients {
				log.Printf("Sending %s to %s", input, c.Username)
				c.Send(Message{Content: input})
			}
		}
	}
}

func main() {
	go consoleBroadcaster()
	router := mux.NewRouter()

	router.HandleFunc("/lobbies", GetLobbies).Methods("GET")
	router.HandleFunc("/lobbies/{id}", GetLobby).Methods("GET")
	router.HandleFunc("/lobbies/{id}/join", JoinLobby).Queries("username", "").Methods("GET")
	router.HandleFunc("/lobbies/create", CreateLobby).Methods("POST")

	fmt.Println("Starting server")
	log.Fatal(http.ListenAndServe(":8080", router))
}
