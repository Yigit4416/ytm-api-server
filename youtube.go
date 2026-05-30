package main

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"strings"
)

type YTResult struct {
	ID       string  `json:"id"`
	Title    string  `json:"title"`
	Uploader string  `json:"uploader"`
	Duration float64 `json:"duration"`
}

func searchYoutube(query string) ([]YTResult, error) {
	var args []string

	if strings.HasPrefix(query, "http") {
		args = []string{"--dump-json", "--flat-playlist", query}
	} else {
		args = []string{"--dump-json", "--flat-playlist", "ytsearch15:" + query}
	}

	cmd := exec.Command("yt-dlp", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	var results []YTResult
	for _, line := range lines {
		if line == "" {
			continue
		}
		var res YTResult
		if err := json.Unmarshal([]byte(line), &res); err == nil {
			if res.ID != "" {
				results = append(results, res)
			}
		}
	}
	return results, nil
}

func getDirectAudioURL(videoID string) (string, error) {
	url := "https://www.youtube.com/watch?v=" + videoID
	cmd := exec.Command("yt-dlp", "-g", "-f", "bestaudio", url)
	out, err := cmd.Output()
	if err != nil || len(out) == 0 {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
