package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sync"
)

func main() {
	addr := flag.String("addr", ":8080", "The addr of the application.")
	debug := flag.Bool("debug", false, "direct commands to stdout instead of omx")
	flag.Parse()

	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets"))))
	http.Handle("/api", &apiHandler{debug: *debug})
	http.Handle("/control", &templateHandler{filename: "control.html"})
	http.HandleFunc("/", handlerHome)
	// http.HandleFunc("/command", handlerCommand(&p))

	log.Printf("Listening on port %s\n", *addr)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

// ServeHTTP handles HTTP requests for the templates
func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// once keeps track of which of these anonymous functions have already been called,
	// and stores their result. If they are called again it just returns the stored result.
	// t.once.Do(func(){
	t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	// })
	data := map[string]string{
		"Host": r.Host,
	}

	t.templ.Execute(w, data)
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

	t := templateHandler{filename: "index.html"}
	t.ServeHTTP(w, r)
}

type apiHandler struct {
	debug   bool
	message reqMessage
}

// ServeHTTP handles all calls to the API
func (a *apiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// handle POST requests
	if r.Method != "POST" {
		m := &resMessage{Success: false, Message: "Invalid request method: " + r.Method}
		log.Println(m.Message)
		json.NewEncoder(w).Encode(m)
		return
	}

	// handle only application/json requests
	ct := r.Header.Get("Content-Type")
	if ct != "application/json" {
		m := &resMessage{Success: false, Message: "Invalid Content-Type: " + ct}
		log.Println(m.Message)
		json.NewEncoder(w).Encode(m)
		return
	}

	// decode message
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&a.message)
	defer r.Body.Close()
	if err != nil {
		m := &resMessage{Success: false, Message: "Error decoding JSON request: " + err.Error()}
		log.Println(m.Message)
		json.NewEncoder(w).Encode(m)
		return
	}

	// displach execution based on which component was called
	// in this case, the Player component
	if a.message.Component == "player" {
		p := Player{api: *a, debug: a.debug}
		p.ServeHTTP(w, r)
		// return
	}

	// return a generic success message for debugging
	m := &resMessage{
		Success: true,
		Message: fmt.Sprintf("Message Received:\ncomponent: %s\nmethod: %s\narguments: %v\n", a.message.Component, a.message.Method, a.message.Arguments),
	}
	json.NewEncoder(w).Encode(m)
	log.Println(m.Message)
}

// func handlerStart(p *Player) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		position := time.Duration(0) * time.Second
// 		q := r.URL.Query()
// 		if q["position"] != nil {
// 			if i, err := strconv.Atoi(q["position"][0]); err != nil {
// 				position = time.Duration(i) * time.Second
// 			}
// 		}

// 		err := p.Start("/home/pi/movies/Bee Movie.mp4", position)
// 		if err != nil {
// 			http.Error(w, "Couldn't start video. ", 500)
// 			log.Println("Couldn't start video: ", err)
// 			return
// 		}
// 		w.Write([]byte("Video Started!"))
// 	}
// }

// func handlerCommand(p *Player) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		q := r.URL.Query()
// 		if q["command"] == nil {
// 			w.Write([]byte("No command sent"))
// 			return
// 		}

// 		command := q["command"][0]
// 		err := p.SendCommand(command)

// 		if err != nil {
// 			http.Error(w, "Couldn't send command. ", http.StatusInternalServerError)
// 			log.Println("Couldn't send command: ", err)
// 			return
// 		}
// 		w.Write([]byte("Command sent!"))
// 	}
// }
