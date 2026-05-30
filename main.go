package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {
	// Ensure data directory exists for SQLite
	os.MkdirAll("./data", os.ModePerm)
	initDB()

	mux := http.NewServeMux()

	// Endpoints
	mux.HandleFunc("/api/search", handleSearch)
	mux.HandleFunc("/api/playlists", handlePlaylists)
	mux.HandleFunc("/api/playlists/add", handleAddToPlaylist)
	mux.HandleFunc("/api/stream", handleStream)

	log.Println("API Server running on port 8080...")
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", mux))
}

func respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(data)
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "missing query 'q'", http.StatusBadRequest)
		return
	}
	results, err := searchYoutube(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondJSON(w, results)
}

func handlePlaylists(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == "GET" {
		// Get specific playlist if ID is provided
		if idStr := r.URL.Query().Get("id"); idStr != "" {
			id, _ := strconv.Atoi(idStr)
			songs, err := getPlaylistSongs(id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			respondJSON(w, songs)
			return
		}
		
		// Otherwise get all playlists
		playlists, err := getPlaylists()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		respondJSON(w, playlists)
	} else if r.Method == "POST" {
		name := r.URL.Query().Get("name")
		if name == "" {
			http.Error(w, "missing playlist 'name'", http.StatusBadRequest)
			return
		}
		id, err := createPlaylist(name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		respondJSON(w, map[string]interface{}{"id": id, "name": name})
	}
}

func handleAddToPlaylist(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	playlistID, _ := strconv.Atoi(r.URL.Query().Get("id"))
	
	var song YTResult
	if err := json.NewDecoder(r.Body).Decode(&song); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := addSongToPlaylist(playlistID, song); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondJSON(w, map[string]string{"status": "success"})
}

func handleStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	videoID := r.URL.Query().Get("v")
	if videoID == "" {
		http.Error(w, "missing video id 'v'", http.StatusBadRequest)
		return
	}

	directURL, err := getDirectAudioURL(videoID)
	if err != nil {
		http.Error(w, "failed to extract stream URL", http.StatusInternalServerError)
		return
	}

	// Create request
	req, err := http.NewRequest("GET", directURL, nil)
	if err != nil {
		http.Error(w, "failed to create request", http.StatusInternalServerError)
		return
	}

	// Forward Range headers for seeking capability!
	if rangeHeader := r.Header.Get("Range"); rangeHeader != "" {
		req.Header.Set("Range", rangeHeader)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "failed to stream from youtube", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Copy headers
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.Header().Set("Content-Length", resp.Header.Get("Content-Length"))
	w.Header().Set("Accept-Ranges", resp.Header.Get("Accept-Ranges"))
	if contentRange := resp.Header.Get("Content-Range"); contentRange != "" {
		w.Header().Set("Content-Range", contentRange)
	}
	
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
