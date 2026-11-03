package main

import (
	"log"
	"net/http"
	"sync"
	"time"

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
	Connected   bool
	UserId      string
	ClientColor string
}

var clients = make(map[*websocket.Conn]clientInfo)
var clientsMu sync.Mutex

func main() {

	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	// if you do this you need to strip the /static/ prefix :(

	http.HandleFunc("/api/ws", func(w http.ResponseWriter, r *http.Request) {
		usernameParam := r.URL.Query().Get("username")
		userColor := r.URL.Query().Get("color")

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
			clientsMu.Unlock()
			ws.WriteControl(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(
					websocket.CloseNormalClosure,
					"Username taken", // so each go thing requires `,` for some reasons
				),
				time.Now().Add(time.Second),  // WHHYYYY IS ',' REQUIRED
			)
			ws.Close()
			return
		}
		clients[ws] = clientInfo{
			Connected:   true,
			UserId:      usernameParam,
			ClientColor: userColor,
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
			SendMSG(sendMSGStruct{
				senderUsername: usernameParam,
				text:           msg.Anno,
				sender:         ws,
			})

		}
	})
	log.Println("HTTP server has started at :8011")
	http.ListenAndServe(":8011", nil)
}

type sendMSGStruct struct {
	senderUsername string
	text           string
	sender         *websocket.Conn
}

func SendMSG(data sendMSGStruct) { // so its just Javascript's data:sendMSGStruct but without the ":"
	clientsMu.Lock()
	var resp = destMessage{
		RemoteUser: data.senderUsername,
		RemoteAnno: data.text,
	}
	for client, info := range clients { // client is

		if client == data.sender {
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
