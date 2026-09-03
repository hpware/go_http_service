package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Message struct {
	Command string `json:"command"`
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, this is a basic go app!")
	})

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	// if you do this you need to strip the /static/ prefix :(

	http.HandleFunc("/api/ws", func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Upgrade failed: %v", err)
			return
		}
		defer ws.Close()
		for {
			var msg Message
			if err := ws.ReadJSON(&msg); err != nil {
				log.Printf("Read error: %v", err)
				break
			}
			log.Printf("Got data: %v", err)
			out, err1 := exec.Command(msg.Command).CombinedOutput()
			if err := ws.WriteJSON(out); err != nil || err1 != nil {
				log.Printf("Write error: %v", err)
				break
			}
		}
	})
	log.Println("HTTP server has started at :8011")
	http.ListenAndServe(":8011", nil)
}
