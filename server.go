package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"

	"github.com/google/uuid"
)

var (
	upgrader    = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	clients     = make(map[string]*Client)
	clientsLock sync.Mutex
	bannedIps   = make(map[string]bool)
)

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade failed:", err)
		return
	}

	wsid := uuid.NewString()
	ip := r.RemoteAddr
	if bannedIps[ip] {
		conn.WriteMessage(websocket.TextMessage, []byte("denied:banned"))
		conn.Close()
		return
	}

	client := &Client{
		ID:   wsid,
		Conn: conn,
		Send: make(chan []byte, 256),
	}
	clientsLock.Lock()
	clients[wsid] = client
	clientsLock.Unlock()

	go client.writePump()
	go client.startHeartbeat()

	msg := modifyMessage([]interface{}{"roomlist", getRoomList(), getEventList(), getClientList(), wsid})

	client.sendl(msg)

	go func() {
		defer func() {
			if client.Room != nil && client.Room.Owner == client {
				clientsLock.Lock()
				for _, c := range clients {
					if c.Room == client.Room && c != client {
						c.sendl(modifyMessage([]string{"selfclose"}))
					}
				}
				clientsLock.Unlock()
				roomMu.Lock()
				delete(roomList, client.Room.Key)
				roomMu.Unlock()
			}
			clientsLock.Lock()
			delete(clients, wsid)
			clientsLock.Unlock()
			conn.Close()
		}()
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				break
			}
			handleMessage(client, msg)
		}
	}()
}
