package main

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	ID       string
	Conn     *websocket.Conn
	Nickname string
	Avatar   string
	Room     *Room
	Owner    *Client
	// Send       chan []byte
	Beat       bool
	ServerMode bool
	Status     string
	OnlineKey  string
	_OnConfig  *Client
	Mutex      sync.Mutex
}

// func (c *Client) writePump() {
// 	for msg := range c.Send {
// 		if err := c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
// 			break
// 		}
// 	}
// }

func (c *Client) startHeartbeat() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		c.Mutex.Lock()
		if c.Beat {
			c.Conn.Close()
			c.Mutex.Unlock()
			break
		}
		c.Beat = true
		c.Mutex.Unlock()
		err := c.Conn.WriteMessage(websocket.BinaryMessage, []byte("heartbeat"))
		if err != nil {
			c.Conn.Close()
		}
	}
}

func (c *Client) sendl(parts []interface{}) {
	msg, err := json.Marshal(parts)
	if err != nil {
		return
	}
	// fmt.Printf("[发送消息] clientID: %s → 内容: %s\n", c.ID, string(msg))
	c.Conn.WriteMessage(websocket.TextMessage, msg)
	// c.Send <- msg
}

func (c *Client) sendTo(id string, message string) {
	clientsLock.Lock()
	target, exists := clients[id]
	clientsLock.Unlock()
	if exists && target.Owner == c {
		target.Conn.WriteMessage(websocket.TextMessage, []byte(message))
	}
}

func (c *Client) closeClient(id string) {
	clientsLock.Lock()
	target, exists := clients[id]
	clientsLock.Unlock()
	if exists && target.Owner == c {
		target.Conn.Close()
	}
}

func getClientList() [][]interface{} {
	clientsLock.Lock()
	defer clientsLock.Unlock()

	var result [][]interface{}
	for _, c := range clients {
		if c.Room == nil {
			entry := []interface{}{c.Nickname, c.Avatar, true, c.Status, c.ID, c.OnlineKey}
			result = append(result, entry)
		}
	}
	return result
}

func updateClients() {
	clientsLock.Lock()
	defer clientsLock.Unlock()

	var result []interface{}
	for _, c := range clients {
		entry := []interface{}{c.Nickname, c.Avatar, c.Room != nil, c.Status, c.ID, c.OnlineKey}
		result = append(result, entry)
	}
	msg := modifyMessage(append([]interface{}{"updateclients"}, result))
	for _, c := range clients {
		if c.Room == nil {
			c.sendl(msg)
		}
	}
}
