package main

import (
    "encoding/json"
    "net/http"
    "github.com/gorilla/mux"
    "strings"
    "fmt"
    "log"
)

type Lobby struct {
    ID string `json:"id,omitempty"`
    Name string `json:"name,omitempty"`
    NumMembers int `json:"numMembers,omitempty"`
}

var Lobbies = []Lobby {
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

func main() {
    router := mux.NewRouter()

    router.HandleFunc("/", sayHello).Methods("GET")
    router.HandleFunc("/lobbies", GetLobbies).Methods("GET")
    router.HandleFunc("/lobbies/{id}", GetLobby).Methods("GET")
    router.HandleFunc("/lobbies/{id}/join", JoinLobby).Methods("GET")
    router.HandleFunc("/lobbies/create", JoinLobby).Methods("POST")

    fmt.Println("Starting server")
    log.Fatal(http.ListenAndServe(":8080", router))
}

