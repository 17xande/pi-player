package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

func main() {
	addr := flag.String("addr", ":8080", "The addr of the application.")
	debug := flag.Bool("debug", false, "direct commands to stdout instead of omx")
	flag.Parse()
	p := Player{debug: *debug}
	http.Handle("/api", &apiHandler{p})
	http.HandleFunc("/", handlerHome)
	http.HandleFunc("/start", handlerStart(&p))
	http.HandleFunc("/command", handlerCommand(&p))

	log.Printf("Listening on port %s\n", *addr)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// Handles requests to the index page as well as any other requests
// that don't match any other paths
func handlerHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", 404)
		log.Printf("Not found: %s", r.URL)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		log.Printf("Method not allowed: %s", r.URL)
		return
	}

	io.WriteString(w, "Welcome to the pi-player")
}

type apiHandler struct {
	player Player
}

func (a *apiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != "POST" {
		m := &resMessage{Success: false, Message: "Invalid request method: " + r.Method}
		log.Println(m.Message)
		json.NewEncoder(w).Encode(m)
		return
	}

	ct := r.Header.Get("Content-Type")
	if ct != "application/json" {
		m := &resMessage{Success: false, Message: "Invalid Content-Type: " + ct}
		log.Println(m.Message)
		json.NewEncoder(w).Encode(m)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var msg reqMessage
	err := decoder.Decode(&msg)
	defer r.Body.Close()
	if err != nil {
		m := &resMessage{Success: false, Message: "Error decoding JSON request: " + err.Error()}
		log.Println(m.Message)
		json.NewEncoder(w).Encode(m)
		return
	}

	if msg.Method == "start" {
		path, ok := msg.Arguments["path"]
		if !ok {
			m := &resMessage{Success: false, Message: "No movie path provided."}
			log.Println(m.Message)
			json.NewEncoder(w).Encode(m)
			return
		}

		var position = time.Duration(0)
		if pos, ok := msg.Arguments["position"]; ok {
			p, err := time.ParseDuration(pos)
			if err != nil {
				m := &resMessage{Success: false, Message: "Error converting video position " + err.Error()}
				log.Println(m.Message)
				json.NewEncoder(w).Encode(m)
				return
			}

			position = p
		}

		err = a.player.Start(path, position)
		if err != nil {
			m := &resMessage{Success: false, Message: "Error trying to start video " + err.Error()}
			log.Println(m.Message)
			json.NewEncoder(w).Encode(m)
			return
		}
	} else if msg.Method == "sendCommad" {
		err = a.player.SendCommand(msg.Arguments["command"])
		if err != nil {
			m := &resMessage{Success: false, Message: "Error trying to execute command: " + err.Error()}
			log.Println(m.Message)
			json.NewEncoder(w).Encode(m)
			return
		}
	} else {
		m := &resMessage{Success: false, Message: "Command not supported"}
		log.Println(m.Message)
		json.NewEncoder(w).Encode(m)
		return
	}

	m := &resMessage{
		Success: true,
		Message: fmt.Sprintf("Message Received:\ncomponent: %s\nmethod: %s\narguments: %v\n", msg.Component, msg.Method, msg.Arguments),
	}
	json.NewEncoder(w).Encode(m)
	log.Println(m.Message)
}

func handlerStart(p *Player) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		position := time.Duration(0) * time.Second
		q := r.URL.Query()
		if q["position"] != nil {
			if i, err := strconv.Atoi(q["position"][0]); err != nil {
				position = time.Duration(i) * time.Second
			}
		}

		err := p.Start("/home/pi/movies/Bee Movie.mp4", position)
		if err != nil {
			http.Error(w, "Couldn't start video. ", 500)
			log.Println("Couldn't start video: ", err)
			return
		}
		w.Write([]byte("Video Started!"))
	}
}

func handlerCommand(p *Player) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q["command"] == nil {
			w.Write([]byte("No command sent"))
			return
		}

		command := q["command"][0]
		err := p.SendCommand(command)

		if err != nil {
			http.Error(w, "Couldn't send command. ", http.StatusInternalServerError)
			log.Println("Couldn't send command: ", err)
			return
		}
		w.Write([]byte("Command sent!"))
	}
}
