package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

const ID_LENGTH = 4

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

// Generate random lobby IDs until one of them is unique.
func UniqueLobbyID() string {
	for {
		id := RandStringBytes(ID_LENGTH)
		if _, exists := Lobbies[id]; !exists {
			return id
		}
	}
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
	name := r.FormValue("name")
	genre := r.FormValue("genre")
	admin := r.FormValue("admin")
	mode, err := strconv.Atoi(r.FormValue("mode"))
	if err != nil {
		log.Fatal("Lobby mode is not an int: %s", err)
	}
	public, err := strconv.ParseBool(r.FormValue("public"))
	if err != nil {
		log.Fatal("Public bool formatted incorrectly: %s", err)
	}
	log.Printf("CreateLobby request received: Name: %q, Mode: %d, Genre: %q, Public: %t, Admin: %q", name, mode, genre, public, admin)

	id := UniqueLobbyID()
	l := NewLobby(id, name, LobbyMode(mode), genre, public, admin)
	Lobbies[id] = l
	// Persist the lobby in the db.
	if err := insertLobby(l); err != nil {
		panic(fmt.Sprintf("Failed to insert lobby: %s", err))
	}
	log.Printf("Lobby %q has been created with ID %q", name, id)

	w.Write([]byte(fmt.Sprintf("%s", id)))
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

	if lobby, ok := Lobbies[id]; ok {
		client := lobby.join(conn, username)
		log.Printf("%s has joined lobby %q", client.Username, lobby.ID)
	} else {
		log.Printf("Lobby with ID %q does not exist", id)
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseUnsupportedData, "Lobby does not exist"))
		conn.Close()
		return
	}
}

// This is here for convenience during developement.
func initialiseTestLobby() {
	lobby := NewLobby("SQRT", "Test Lobby", FREE_FOR_ALL, "Whatev", true, "red350")
	log.Printf("Test lobby created. Remove this before deployment.")
	if err := insertLobby(lobby); err != nil {
		panic(fmt.Sprintf("Failed to insert lobby: %s", err))
	}
}

func main() {
	log.Printf("Starting server")
	//initialiseTestLobby()
	// Load lobby state from the database.
	if err := loadFromDB(&Lobbies); err != nil {
		log.Printf("Failed to load lobbies from db: %s", err)
	}
	log.Printf("Lobby state loaded")

	//Lobbies["SQRT"].CurrentTrack = &Track{URI: "new", Name: "new", Artist: "new"}
	//Lobbies["SQRT"].TrackQueue.Push(&Track{URI: "queue", Name: "queue", Artist: "queue"})
	//Lobbies["SQRT"].TrackQueue.Push(&Track{URI: "queue2", Name: "queue", Artist: "queue"})

	//if err := persistQueue(Lobbies["SQRT"]); err != nil {
	//	log.Printf("Failed to persist queue: %s", err)
	//}
	//if err := persistCurrentTrack(Lobbies["SQRT"]); err != nil {
	//	log.Printf("Failed to persist current track: %s", err)
	//}

	//log.Printf("")
	//for k, v := range Lobbies {
	//	log.Printf("%s: %s %#v\n", k, v.Name, v.CurrentTrack)
	//	for _, track := range v.TrackQueue {
	//		log.Printf("%#v\n", track)
	//	}
	//}

	router := mux.NewRouter()

	router.HandleFunc("/lobbies", GetLobbies).Methods("GET")
	router.HandleFunc("/lobbies/{id}", GetLobby).Methods("GET")
	router.HandleFunc("/lobbies/{id}/join", JoinLobby).Queries("username", "").Methods("GET")
	router.HandleFunc("/lobbies/create", CreateLobby).Methods("POST")

	log.Printf("Server started")
	log.Fatal(http.ListenAndServe(":8080", router))
}
