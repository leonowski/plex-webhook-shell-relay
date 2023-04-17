package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os/exec"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type WebhookPayload struct {
	Event  string `json:"event"`
	Player struct {
		UUID string `json:"uuid"`
	} `json:"Player"`
}

func handlePlexWebhook(w http.ResponseWriter, r *http.Request, playerUUID string, playCmd, pauseCmd, stopCmd string) {
	err := r.ParseMultipartForm(32 << 20) // 32 MB max memory
	if err != nil {
		log.Printf("Error parsing multipart form: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	jsonData := r.FormValue("payload")
	if jsonData == "" {
		log.Println("JSON data not found in the multipart payload")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var payload WebhookPayload
	if err := json.Unmarshal([]byte(jsonData), &payload); err != nil {
		log.Printf("Error unmarshalling JSON: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("Received request from player UUID: %s", payload.Player.UUID) // Log the player UUID

	if playerUUID != "" && payload.Player.UUID != playerUUID {
		w.WriteHeader(http.StatusOK)
		return
	}

	if playerUUID != "" && payload.Player.UUID != playerUUID {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch payload.Event {
	case "media.play":
		exec.Command("sh", "-c", playCmd).Run()
	case "media.pause":
		exec.Command("sh", "-c", pauseCmd).Run()
	case "media.stop":
		exec.Command("sh", "-c", stopCmd).Run()
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func main() {
	playerUUID := flag.String("playerUUID", "", "UUID of the player")
	port := flag.String("port", "8080", "Server port")
	playCmd := flag.String("play", "", "Shell command to execute on media.play event")
	pauseCmd := flag.String("pause", "", "Shell command to execute on media.pause event")
	stopCmd := flag.String("stop", "", "Shell command to execute on media.stop event")

	flag.Parse()

	if *playCmd == "" || *pauseCmd == "" || *stopCmd == "" {
		flag.Usage()
		log.Fatal("Play, pause, and stop commands must be provided.")
	}

	router := mux.NewRouter()
	router.HandleFunc("/plex-webhook", func(w http.ResponseWriter, r *http.Request) {
		handlePlexWebhook(w, r, *playerUUID, *playCmd, *pauseCmd, *stopCmd)
	}).Methods("POST")

	handler := cors.Default().Handler(router)

	log.Printf("Listening on port %s...", *port)
	if err := http.ListenAndServe(":"+*port, handler); err != nil {
		log.Fatal(err)
	}
}
