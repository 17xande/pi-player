package main

import (
	"io"
	"log"
	"net/http"
)

func main() {
	addr := ":8080"
	p := Player{}
	http.HandleFunc("/", handlerHome)
	http.HandleFunc("/start", handlerStart(&p))
	http.HandleFunc("/command", handlerCommand(&p))

	log.Printf("Listening on port %s\n", addr)
	err := http.ListenAndServe(addr, nil)
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

func handlerStart(p *Player) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := p.Start("/home/pi/movies/Bee Movie.mp4")
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
			http.Error(w, "Couldn't send command. ", 500)
			log.Println("Couldn't send command: ", err)
			return
		}
		w.Write([]byte("Command sent!"))
	}
}
