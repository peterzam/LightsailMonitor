package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]bool)
var broadcaster = make(chan *Info)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	} else {
		http.ServeFile(w, r, "./index.html")
	}

}

func infoWriter(info Info) {
	go func() {
		broadcaster <- &info
	}()
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	clients[ws] = true
	broadcaster <- &info
}

func echo() {
	for {
		val := <-broadcaster
		info := fmt.Sprintf("%s|%s|%s|%s", val.ip, val.status, val.traffic, strconv.Itoa(val.time))
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, []byte(info))
			if err != nil {
				log.Printf("Websocket error: %s", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
