package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

type Playlist struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./data/ytm.db")
	if err != nil {
		log.Fatal(err)
	}

	createTables := `
	CREATE TABLE IF NOT EXISTS playlists (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE
	);

	CREATE TABLE IF NOT EXISTS songs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		playlist_id INTEGER,
		video_id TEXT,
		title TEXT,
		uploader TEXT,
		duration REAL,
		FOREIGN KEY(playlist_id) REFERENCES playlists(id)
	);
	`
	_, err = db.Exec(createTables)
	if err != nil {
		log.Fatal(err)
	}
}

func createPlaylist(name string) (int64, error) {
	res, err := db.Exec("INSERT OR IGNORE INTO playlists (name) VALUES (?)", name)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func addSongToPlaylist(playlistID int, song YTResult) error {
	_, err := db.Exec("INSERT INTO songs (playlist_id, video_id, title, uploader, duration) VALUES (?, ?, ?, ?, ?)",
		playlistID, song.ID, song.Title, song.Uploader, song.Duration)
	return err
}

func getPlaylists() ([]Playlist, error) {
	rows, err := db.Query("SELECT id, name FROM playlists")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var p []Playlist = []Playlist{}
	for rows.Next() {
		var pl Playlist
		if err := rows.Scan(&pl.ID, &pl.Name); err == nil {
			p = append(p, pl)
		}
	}
	return p, nil
}

func getPlaylistSongs(playlistID int) ([]YTResult, error) {
	rows, err := db.Query("SELECT video_id, title, uploader, duration FROM songs WHERE playlist_id = ?", playlistID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var songs []YTResult = []YTResult{}
	for rows.Next() {
		var s YTResult
		if err := rows.Scan(&s.ID, &s.Title, &s.Uploader, &s.Duration); err == nil {
			songs = append(songs, s)
		}
	}
	return songs, nil
}
