package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Message struct {
	Anno string `json:"anno"`
}

type destMessage struct {
	RemoteAnno string `json:"anno"`
	RemoteUser string `json:"name"`
}

type clientInfo struct {
	Connected bool
	UserId    string
}

var clients = make(map[*websocket.Conn]clientInfo)
var clientsMu sync.Mutex

func main() {

	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	// if you do this you need to strip the /static/ prefix :(

	http.HandleFunc("/api/ws", func(w http.ResponseWriter, r *http.Request) {
		usernameParam := r.URL.Query().Get("username")

		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Upgrade failed: %v", err)
			return
		}
		clientsMu.Lock()
		var exists = false
		for _, info := range clients {
			if info.UserId == usernameParam {
				exists = true
				break
			}
		}
		// check suer exists or not
		if exists {
			err := ws.WriteJSON(map[string]string{
				"error": "Username taken",
			})
			if err != nil {
				log.Printf("Write ERR: %v", err)
			}
			ws.Close()
			return
		}
		clients[ws] = clientInfo{
			Connected: true,
			UserId:    usernameParam,
		}
		clientsMu.Unlock()
		defer func() {
			clientsMu.Lock()
			delete(clients, ws)
			clientsMu.Unlock()
			ws.Close()
		}()

		for {
			var msg Message
			if err := ws.ReadJSON(&msg); err != nil {
				log.Printf("Read error: %v", err)
				break
			}
			log.Printf("\n------------------------------------------------------------------------\nGot data: %v", msg.Anno)

			clientsMu.Lock()
			var resp = destMessage{
				RemoteUser: usernameParam,
				RemoteAnno: msg.Anno,
			}
			for client, info := range clients { // client is

				if client == ws {
					resp.RemoteUser = "You"
				}

				if err := client.WriteJSON(resp); err != nil {
					log.Printf("Error: %v", err)
					client.Close()
					delete(clients, client)
				}
				log.Printf("Sending it to %v", info.UserId)
			}
			// what the fuck is the ...
			// the `...` does one thing. its ofc meaning the function accepts zero or more args...

			clientsMu.Unlock() // unlock connection.
		}
	})
	log.Println("HTTP server has started at :8011")
	http.ListenAndServe(":8011", nil)
}
