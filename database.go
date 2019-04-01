package main

import (
	"database/sql"
	"fmt"
	"log"
)

// dbConn returns a connection to the database.
func dbConn() (*sql.DB, error) {
	db, err := sql.Open("mysql", "root:@/syncsong")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %s", err)
	}
	return db, nil
}

func loadFromDB(lobbies *map[string]*Lobby) error {
	db, err := dbConn()
	if err != nil {
		return fmt.Errorf("failed to connect to db: %s", err)
	}

	lobbyRows, err := db.Query("select id, name, mode, genre, public, currentUri from lobby")
	if err != nil {
		return fmt.Errorf("failed to query lobbies: %s", err)
	}

	// Parse returned rows and add to the lobbies map.
	for lobbyRows.Next() {
		var id string
		var trackName string
		var lobbyName string
		var mode int
		var genre string
		var public bool
		var uri sql.NullString
		var artist string
		var duration int64
		var currentTrack *Track

		if err := lobbyRows.Scan(&id, &lobbyName, &mode, &genre, &public, &uri); err != nil {
			return fmt.Errorf("failed to read lobby row: %s", err)
		}
		// Query for the current track.
		if uri.Valid {
			err := db.QueryRow("select uri, name, artist, duration from track where uri=?", uri).Scan(&uri, &trackName, &artist, &duration)
			if err != nil && err != sql.ErrNoRows {
				return fmt.Errorf("failed to read current track: %s", err)
			}
			currentTrack = &Track{URI: uri.String, Name: trackName, Artist: artist, Duration: duration}
		}
		lobby := NewLobby(id, lobbyName, LobbyMode(mode), genre, public, "", currentTrack)

		// Add the queue.
		queue, err := db.Query(
			`select trackURI, name, artist from queue
            join track on(track.uri = queue.trackURI)
            where lobbyID=?
            order by rank asc`, id)
		if err != nil {
			return fmt.Errorf("failed to query queue: %s", err)
		}

		for queue.Next() {
			if err := queue.Scan(&uri, &trackName, &artist); err != nil {
				return fmt.Errorf("failed to read queue: %s", err)
			}

			lobby.TrackQueue.Push(&Track{URI: uri.String, Name: trackName, Artist: artist, Duration: duration})
		}

		(*lobbies)[id] = lobby
	}
	return nil
}

func insertLobby(lobby *Lobby) error {
	db, err := dbConn()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %s", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %s", err)
	}
	stmt, err := tx.Prepare(`
        insert into lobby(id, name, mode, genre, public, currentUri)
        values(?, ?, ?, ?, ?, null);`)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to prepare statement: %s", err)
	}
	defer stmt.Close()
	if _, err := stmt.Exec(lobby.ID, lobby.Name, lobby.LobbyMode, lobby.Genre, lobby.Public); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to execute statement: %s", err)
	}

	return tx.Commit()
}

func persistQueue(lobby *Lobby) error {
	// Connect to db.
	db, err := dbConn()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %s", err)
	}

	// Begin transaction.
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %s", err)
	}

	// Delete the old queue for this lobby.
	// Since removing the first song from the queue will result in every
	// song needing to be updated in the db, it is easier to simplye remove
	// them all and reinsert then to figure out which need to be changed.
	if err := deleteQueue(tx, lobby.ID); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete lobby queue: %s", err)
	}

	// Insert queued tracks.
	for rank, track := range lobby.TrackQueue {
		if err := insertTrack(tx, track); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert queued track: %s", err)
		}
		if err := queueTrack(tx, lobby.ID, track.URI, rank); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to queue track: %s", err)
		}
	}

	return tx.Commit()
}

func persistCurrentTrack(lobby *Lobby) error {
	// Connect to db.
	db, err := dbConn()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %s", err)
	}

	// Begin transaction.
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %s", err)
	}

	// Insert current track and update lobby's current track.
	if err := insertTrack(tx, lobby.CurrentTrack); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to insert current track: %s", err)
	}
	var uri sql.NullString
	if lobby.CurrentTrack != nil {
		uri.Valid = true
		uri.String = lobby.CurrentTrack.URI
	}
	stmt, err := tx.Prepare(`update lobby set currentUri=? where id=?`)
	if err != nil {
		return fmt.Errorf("failed to prepare update current track statement: %s", err)
	}
	defer stmt.Close()
	if _, err := stmt.Exec(uri, lobby.ID); err != nil {
		return fmt.Errorf("failed to execute update current track statement: %s", err)
	}

	return tx.Commit()
}

func deleteQueue(tx *sql.Tx, lobbyID string) error {
	stmt, err := tx.Prepare(`delete from queue where lobbyID=?`)
	if err != nil {
		return fmt.Errorf("failed prepare delete statement: %s", err)
	}
	defer stmt.Close()
	if _, err := stmt.Exec(lobbyID); err != nil {
		return fmt.Errorf("failed to execute track statement: %s", err)
	}
	return nil
}

// insertTrack inserts a track if it doesn't exist, otherwise does nothing.
func insertTrack(tx *sql.Tx, track *Track) error {
	if track == nil {
		return nil
	}
	stmt, err := tx.Prepare(`
        insert ignore into track(uri, name, artist, duration)
        values(?, ?, ?, ?);`)
	if err != nil {
		return fmt.Errorf("failed to prepare track statement: %s", err)
	}
	defer stmt.Close()
	if _, err := stmt.Exec(track.URI, track.Name, track.Artist, track.Duration); err != nil {
		return fmt.Errorf("failed to execute track statement: %s", err)
	}
	return nil
}

// queueTrack inserts a row to the Queue table with the provided values.
func queueTrack(tx *sql.Tx, lobbyID string, uri string, rank int) error {
	log.Printf("Queueing id:%s uri:%s", lobbyID, uri)
	stmt, err := tx.Prepare(`
        insert into queue(lobbyID, trackURI, rank)
        values(?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("failed to prepare queue statement: %s", err)
	}
	defer stmt.Close()
	if _, err := stmt.Exec(lobbyID, uri, rank); err != nil {
		return fmt.Errorf("failed to execute queue statement id:%s uri:%s: %s", lobbyID, uri, err)
	}
	return nil
}
