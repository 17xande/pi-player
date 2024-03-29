package piplayer

import (
	"context"
	"log"
	"net/http"
	"time"
)

// NewServer returns a new http.Server for the piplayer interface.
func NewServer(p *Player, addr string) *http.Server {
	mux := setupRoutes(p.conf.Mount.Dir, p)
	serv := http.Server{Addr: addr, Handler: mux}

	return &serv
}

// setupRoutes registers the routes for the server, accepting a new directory where
// the content can be found. This will be called whenever there is a change of content
// given from the settings page.
func setupRoutes(content string, p *Player) *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(p.api.statAssets))))
	// mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("pkg/piplayer/assets"))))
	mux.HandleFunc("/content/", etagWrapper(content))
	mux.HandleFunc("/login", LoginHandler(p))
	mux.HandleFunc("/logout", LogoutHandler)
	mux.HandleFunc("/control", p.HandleControl)
	mux.HandleFunc("/settings", p.conf.SettingsHandler(p))
	mux.HandleFunc("/viewer", p.HandleViewer)
	mux.HandleFunc("/ws/viewer", p.ConnViewer.HandlerWebsocket(p))
	mux.HandleFunc("/ws/control", p.ConnControl.HandlerWebsocket(p))
	mux.HandleFunc("/api", p.api.Handle(p))
	mux.HandleFunc("/", p.api.handlerHome)

	return mux
}

// etagWrapper calculates an Etag value for the requested content.
func etagWrapper(content string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fs := http.StripPrefix("/content/", http.FileServer(http.Dir(content)))

		// TODO: Everything

		fs.ServeHTTP(w, r)
	}
}

// Restart the http server.
func restart(plr *Player) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := plr.Server.Shutdown(ctx); err != nil {
		log.Printf("Error shutting down server for re-registration of routes: %v\n", err)
	}
	go func() {
		time.Sleep(5 * time.Second)
		cancel()
	}()
}

// Start the http server.
func Start(plr *Player) {
	log.Printf("Listening on port %s\n", plr.Server.Addr)
	err := plr.Server.ListenAndServe()
	if err != nil {
		log.Println("ListenAndServe: ", err)
	}

	// Restart the server.
	// TODO: This needs to become a config option
	plr.Server = NewServer(plr, ":8080")
	Start(plr)
}
