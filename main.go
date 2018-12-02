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

var Lobb = NewLobby("a", "Lobby 1")
var Lobbies = make(map[string]*Lobby)

func sayHello(w http.ResponseWriter, r *http.Request) {
	fmt.Println("request received")
	message := r.URL.Path
	message = strings.TrimPrefix(message, "/")
	message = "Hello " + message

	w.Write([]byte(message))
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
	log.Print("CreateLobby request received")

	id := r.Form.Get("id")
	name := r.Form.Get("name")
	fmt.Println(id, name)
	log.Print("hello: %s", id)

	w.Write([]byte("asdf"))
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
		clientID := l.join(conn)
		log.Printf("Client %d has joined Lobby %q", clientID, l.ID)
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
	Lobbies["a"] = Lobb
	go consoleReader()
	router := mux.NewRouter()

	router.HandleFunc("/", sayHello).Methods("GET")
	router.HandleFunc("/lobbies", GetLobbies).Methods("GET")
	router.HandleFunc("/lobbies/{id}", GetLobby).Methods("GET")
	router.HandleFunc("/lobbies/{id}/join", JoinLobby).Methods("GET")
	router.HandleFunc("/lobbies/create", CreateLobby).Methods("POST")

	//router.HandleFunc("/ws", wsHandler)

	fmt.Println("Starting server")
	log.Fatal(http.ListenAndServe(":8080", router))
	fmt.Println("here i am")
}
